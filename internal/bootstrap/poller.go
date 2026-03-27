package bootstrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	scheduleradapter "payrune/internal/adapters/inbound/scheduler"
	"payrune/internal/adapters/outbound/bitcoin"
	blockchainadapter "payrune/internal/adapters/outbound/blockchain"
	"payrune/internal/adapters/outbound/ethereum"
	postgresadapter "payrune/internal/adapters/outbound/persistence/postgres"
	"payrune/internal/adapters/outbound/system"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/application/usecases"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

const (
	defaultPollerTickInterval = 15 * time.Second
	defaultPollerClaimTTL     = 30 * time.Second
	defaultPollerBatchSize    = 50

	envPollTickInterval       = "POLL_TICK_INTERVAL"
	envPollRescheduleInterval = "POLL_RESCHEDULE_INTERVAL"
	envPollBatchSize          = "POLL_BATCH_SIZE"
	envPollClaimTTL           = "POLL_CLAIM_TTL"
	envPollChain              = "POLL_CHAIN"
	envPollNetwork            = "POLL_NETWORK"

	envBitcoinMainnetEsploraURL            = "BITCOIN_MAINNET_ESPLORA_URL"
	envBitcoinMainnetEsploraUser           = "BITCOIN_MAINNET_ESPLORA_USER"
	envBitcoinMainnetEsploraPassword       = "BITCOIN_MAINNET_ESPLORA_PASSWORD"
	envBitcoinMainnetEsploraTimeout        = "BITCOIN_MAINNET_ESPLORA_TIMEOUT"
	envBitcoinMainnetEsploraTimeoutSeconds = "BITCOIN_MAINNET_ESPLORA_TIMEOUT_SECONDS"

	envBitcoinTestnet4EsploraURL            = "BITCOIN_TESTNET4_ESPLORA_URL"
	envBitcoinTestnet4EsploraUser           = "BITCOIN_TESTNET4_ESPLORA_USER"
	envBitcoinTestnet4EsploraPassword       = "BITCOIN_TESTNET4_ESPLORA_PASSWORD"
	envBitcoinTestnet4EsploraTimeout        = "BITCOIN_TESTNET4_ESPLORA_TIMEOUT"
	envBitcoinTestnet4EsploraTimeoutSeconds = "BITCOIN_TESTNET4_ESPLORA_TIMEOUT_SECONDS"

	envEthereumMainnetRPCURL            = "ETHEREUM_MAINNET_RPC_URL"
	envEthereumMainnetRPCUser           = "ETHEREUM_MAINNET_RPC_USER"
	envEthereumMainnetRPCPassword       = "ETHEREUM_MAINNET_RPC_PASSWORD"
	envEthereumMainnetRPCTimeout        = "ETHEREUM_MAINNET_RPC_TIMEOUT"
	envEthereumMainnetRPCTimeoutSeconds = "ETHEREUM_MAINNET_RPC_TIMEOUT_SECONDS"

	envEthereumSepoliaRPCURL            = "ETHEREUM_SEPOLIA_RPC_URL"
	envEthereumSepoliaRPCUser           = "ETHEREUM_SEPOLIA_RPC_USER"
	envEthereumSepoliaRPCPassword       = "ETHEREUM_SEPOLIA_RPC_PASSWORD"
	envEthereumSepoliaRPCTimeout        = "ETHEREUM_SEPOLIA_RPC_TIMEOUT"
	envEthereumSepoliaRPCTimeoutSeconds = "ETHEREUM_SEPOLIA_RPC_TIMEOUT_SECONDS"
)

type PollerConfig struct {
	TickInterval       time.Duration
	RescheduleInterval time.Duration
	BatchSize          int
	ClaimTTL           time.Duration
	Chain              valueobjects.ChainID
	Network            valueobjects.NetworkID
}

type pollerContainer struct {
	PollerHandler *scheduleradapter.PollerHandler
	closeFn       func() error
}

type pollerDispatchConfig struct {
	RescheduleInterval time.Duration
	BatchSize          int
	ClaimTTL           time.Duration
	Chain              valueobjects.ChainID
	Network            valueobjects.NetworkID
}

type pollerDispatchDefaults struct {
	RescheduleInterval time.Duration
	BatchSize          int
	ClaimTTL           time.Duration
}

type bitcoinEsploraEnvKeys struct {
	url            string
	user           string
	password       string
	timeout        string
	timeoutSeconds string
}

type ethereumRPCEndpointEnvKeys struct {
	url            string
	user           string
	password       string
	timeout        string
	timeoutSeconds string
}

func LoadPollerConfigFromEnv() (PollerConfig, error) {
	return loadPollerConfigFromLookup(os.Getenv)
}

func RunPoller(ctx context.Context, config PollerConfig) error {
	if config.TickInterval <= 0 {
		config.TickInterval = defaultPollerTickInterval
	}
	if config.BatchSize <= 0 {
		config.BatchSize = defaultPollerBatchSize
	}
	if config.ClaimTTL <= 0 {
		config.ClaimTTL = defaultPollerClaimTTL
	}
	container, err := newPollerContainer()
	if err != nil {
		return err
	}
	defer func() {
		_ = container.Close()
	}()

	runCycle := func() {
		output, err := container.PollerHandler.Handle(ctx, scheduleradapter.PollerRequest{
			BatchSize:          config.BatchSize,
			RescheduleInterval: config.RescheduleInterval,
			ClaimTTL:           config.ClaimTTL,
			Chain:              config.Chain,
			Network:            config.Network,
		})
		if err != nil {
			log.Printf("poll cycle failed: err=%v", err)
			return
		}

		log.Printf(
			"poll cycle complete claimed=%d updated=%d terminal_failed=%d processing_errors=%d",
			output.ClaimedCount,
			output.UpdatedCount,
			output.TerminalFailedCount,
			output.ProcessingErrorCount,
		)
	}

	runCycle()

	ticker := time.NewTicker(config.TickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			runCycle()
		}
	}
}

func loadPollerConfigFromLookup(lookup func(string) string) (PollerConfig, error) {
	tickInterval, err := parsePollerPositiveDurationLookupWithDefault(lookup, envPollTickInterval, 0)
	if err != nil {
		return PollerConfig{}, err
	}
	dispatchConfig, err := loadPollerDispatchConfigFromLookup(lookup, pollerDispatchDefaults{})
	if err != nil {
		return PollerConfig{}, err
	}

	return PollerConfig{
		TickInterval:       tickInterval,
		RescheduleInterval: dispatchConfig.RescheduleInterval,
		BatchSize:          dispatchConfig.BatchSize,
		ClaimTTL:           dispatchConfig.ClaimTTL,
		Chain:              dispatchConfig.Chain,
		Network:            dispatchConfig.Network,
	}, nil
}

func newPollerContainer() (*pollerContainer, error) {
	db, err := openPostgresFromEnv()
	if err != nil {
		return nil, err
	}

	scopeChain, scoped, err := loadConfiguredPollerChainFromEnv()
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	unitOfWork := postgresadapter.NewUnitOfWork(db)
	chainObservers := make(map[valueobjects.ChainID]outport.ChainReceiptObserver, 2)
	if !scoped || scopeChain == valueobjects.ChainIDBitcoin {
		bitcoinConfigs := loadBitcoinEsploraConfigsFromEnv()
		if len(bitcoinConfigs) > 0 {
			bitcoinObserver, err := bitcoin.NewBitcoinEsploraReceiptObserver(bitcoinConfigs)
			if err != nil {
				_ = db.Close()
				return nil, err
			}
			chainObservers[valueobjects.ChainIDBitcoin] = bitcoinObserver
		}
	}
	if !scoped || scopeChain == valueobjects.ChainIDEthereum {
		ethereumConfigs := loadEthereumRPCConfigsFromEnv()
		if len(ethereumConfigs) > 0 {
			ethereumObserver, err := ethereum.NewEthereumRPCReceiptObserver(ethereumConfigs)
			if err != nil {
				_ = db.Close()
				return nil, err
			}
			chainObservers[valueobjects.ChainIDEthereum] = ethereumObserver
		}
	}
	receiptObserver, err := blockchainadapter.NewMultiChainReceiptObserver(chainObservers)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	clock := system.NewClock()
	runReceiptPollingCycleUseCase := usecases.NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		receiptObserver,
		clock,
		policies.NewPaymentReceiptTrackingLifecyclePolicy(),
	)

	return &pollerContainer{
		PollerHandler: scheduleradapter.NewPollerHandler(scheduleradapter.PollerDependencies{
			RunReceiptPollingCycleUseCase: runReceiptPollingCycleUseCase,
		}),
		closeFn: db.Close,
	}, nil
}

func (c *pollerContainer) Close() error {
	if c.closeFn == nil {
		return nil
	}
	return c.closeFn()
}

func loadPollerDispatchConfigFromLookup(
	lookup func(string) string,
	defaults pollerDispatchDefaults,
) (pollerDispatchConfig, error) {
	rescheduleInterval, err := parsePollerPositiveDurationLookupWithDefault(
		lookup,
		envPollRescheduleInterval,
		defaults.RescheduleInterval,
	)
	if err != nil {
		return pollerDispatchConfig{}, err
	}
	claimTTL, err := parsePollerPositiveDurationLookupWithDefault(
		lookup,
		envPollClaimTTL,
		defaults.ClaimTTL,
	)
	if err != nil {
		return pollerDispatchConfig{}, err
	}
	batchSize, err := parsePollerPositiveIntLookupWithDefault(
		lookup,
		envPollBatchSize,
		defaults.BatchSize,
	)
	if err != nil {
		return pollerDispatchConfig{}, err
	}
	chain, err := parsePollerChainLookup(lookup, envPollChain)
	if err != nil {
		return pollerDispatchConfig{}, err
	}
	network, err := parsePollerNetworkLookup(lookup, envPollNetwork)
	if err != nil {
		return pollerDispatchConfig{}, err
	}
	if network != "" && chain == "" {
		return pollerDispatchConfig{}, fmt.Errorf("%s is required when %s is set", envPollChain, envPollNetwork)
	}

	return pollerDispatchConfig{
		RescheduleInterval: rescheduleInterval,
		BatchSize:          batchSize,
		ClaimTTL:           claimTTL,
		Chain:              chain,
		Network:            network,
	}, nil
}

func parsePollerPositiveDurationLookupWithDefault(
	lookup func(string) string,
	key string,
	fallback time.Duration,
) (time.Duration, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return value, nil
}

func parsePollerPositiveIntLookupWithDefault(
	lookup func(string) string,
	key string,
	fallback int,
) (int, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return value, nil
}

func parsePollerChainLookup(lookup func(string) string, key string) (valueobjects.ChainID, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		return "", nil
	}

	chain, ok := valueobjects.ParseChainID(raw)
	if !ok {
		return "", fmt.Errorf("%s is invalid", key)
	}
	return chain, nil
}

func parsePollerNetworkLookup(lookup func(string) string, key string) (valueobjects.NetworkID, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		return "", nil
	}

	network, ok := valueobjects.ParseNetworkID(raw)
	if !ok {
		return "", fmt.Errorf("%s is invalid", key)
	}
	return network, nil
}

func loadConfiguredPollerChainFromEnv() (valueobjects.ChainID, bool, error) {
	return loadConfiguredPollerChainFromLookup(os.Getenv)
}

func loadConfiguredPollerChainFromLookup(
	lookup func(string) string,
) (valueobjects.ChainID, bool, error) {
	chain, err := parsePollerChainLookup(lookup, envPollChain)
	if err != nil {
		return "", false, err
	}
	if chain == "" {
		return "", false, nil
	}
	return chain, true, nil
}

func loadBitcoinMainnetEsploraConfigFromEnv() *bitcoin.BitcoinEsploraObserverConfig {
	return loadBitcoinEsploraConfig(bitcoinEsploraEnvKeys{
		url:            envBitcoinMainnetEsploraURL,
		user:           envBitcoinMainnetEsploraUser,
		password:       envBitcoinMainnetEsploraPassword,
		timeout:        envBitcoinMainnetEsploraTimeout,
		timeoutSeconds: envBitcoinMainnetEsploraTimeoutSeconds,
	})
}

func loadBitcoinTestnet4EsploraConfigFromEnv() *bitcoin.BitcoinEsploraObserverConfig {
	return loadBitcoinEsploraConfig(bitcoinEsploraEnvKeys{
		url:            envBitcoinTestnet4EsploraURL,
		user:           envBitcoinTestnet4EsploraUser,
		password:       envBitcoinTestnet4EsploraPassword,
		timeout:        envBitcoinTestnet4EsploraTimeout,
		timeoutSeconds: envBitcoinTestnet4EsploraTimeoutSeconds,
	})
}

func loadBitcoinEsploraConfigsFromEnv() map[valueobjects.NetworkID]*bitcoin.BitcoinEsploraObserverConfig {
	configs := make(map[valueobjects.NetworkID]*bitcoin.BitcoinEsploraObserverConfig, 2)

	if mainnetConfig := loadBitcoinMainnetEsploraConfigFromEnv(); mainnetConfig != nil {
		configs[valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet)] = mainnetConfig
	}
	if testnet4Config := loadBitcoinTestnet4EsploraConfigFromEnv(); testnet4Config != nil {
		configs[valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4)] = testnet4Config
	}

	return configs
}

func loadBitcoinEsploraConfig(keys bitcoinEsploraEnvKeys) *bitcoin.BitcoinEsploraObserverConfig {
	return loadBitcoinEsploraConfigFromLookup(os.Getenv, keys)
}

func loadBitcoinEsploraConfigFromLookup(
	lookup func(string) string,
	keys bitcoinEsploraEnvKeys,
) *bitcoin.BitcoinEsploraObserverConfig {
	endpoint := strings.TrimSpace(lookup(keys.url))
	if endpoint == "" {
		return nil
	}

	timeout := 10 * time.Second
	if rawTimeout := strings.TrimSpace(lookup(keys.timeout)); rawTimeout != "" {
		if parsed, err := time.ParseDuration(rawTimeout); err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	if timeoutSecondsRaw := strings.TrimSpace(lookup(keys.timeoutSeconds)); timeoutSecondsRaw != "" {
		if parsedSeconds, err := strconv.Atoi(timeoutSecondsRaw); err == nil && parsedSeconds > 0 {
			timeout = time.Duration(parsedSeconds) * time.Second
		}
	}

	return &bitcoin.BitcoinEsploraObserverConfig{
		Endpoint: endpoint,
		Username: strings.TrimSpace(lookup(keys.user)),
		Password: lookup(keys.password),
		Timeout:  timeout,
	}
}

func loadEthereumMainnetRPCConfigFromEnv() *ethereum.EthereumRPCObserverConfig {
	return loadEthereumRPCConfigFromLookup(os.Getenv, ethereumRPCEndpointEnvKeys{
		url:            envEthereumMainnetRPCURL,
		user:           envEthereumMainnetRPCUser,
		password:       envEthereumMainnetRPCPassword,
		timeout:        envEthereumMainnetRPCTimeout,
		timeoutSeconds: envEthereumMainnetRPCTimeoutSeconds,
	})
}

func loadEthereumSepoliaRPCConfigFromEnv() *ethereum.EthereumRPCObserverConfig {
	return loadEthereumRPCConfigFromLookup(os.Getenv, ethereumRPCEndpointEnvKeys{
		url:            envEthereumSepoliaRPCURL,
		user:           envEthereumSepoliaRPCUser,
		password:       envEthereumSepoliaRPCPassword,
		timeout:        envEthereumSepoliaRPCTimeout,
		timeoutSeconds: envEthereumSepoliaRPCTimeoutSeconds,
	})
}

func loadEthereumRPCConfigsFromEnv() map[valueobjects.NetworkID]*ethereum.EthereumRPCObserverConfig {
	return loadEthereumRPCConfigsFromLookup(os.Getenv)
}

func loadEthereumRPCConfigsFromLookup(
	lookup func(string) string,
) map[valueobjects.NetworkID]*ethereum.EthereumRPCObserverConfig {
	configs := make(map[valueobjects.NetworkID]*ethereum.EthereumRPCObserverConfig, 2)

	if mainnetConfig := loadEthereumRPCConfigFromLookup(lookup, ethereumRPCEndpointEnvKeys{
		url:            envEthereumMainnetRPCURL,
		user:           envEthereumMainnetRPCUser,
		password:       envEthereumMainnetRPCPassword,
		timeout:        envEthereumMainnetRPCTimeout,
		timeoutSeconds: envEthereumMainnetRPCTimeoutSeconds,
	}); mainnetConfig != nil {
		configs[valueobjects.NetworkID("mainnet")] = mainnetConfig
	}
	if sepoliaConfig := loadEthereumRPCConfigFromLookup(lookup, ethereumRPCEndpointEnvKeys{
		url:            envEthereumSepoliaRPCURL,
		user:           envEthereumSepoliaRPCUser,
		password:       envEthereumSepoliaRPCPassword,
		timeout:        envEthereumSepoliaRPCTimeout,
		timeoutSeconds: envEthereumSepoliaRPCTimeoutSeconds,
	}); sepoliaConfig != nil {
		configs[valueobjects.NetworkID("sepolia")] = sepoliaConfig
	}

	return configs
}

func loadEthereumRPCConfigFromLookup(
	lookup func(string) string,
	keys ethereumRPCEndpointEnvKeys,
) *ethereum.EthereumRPCObserverConfig {
	endpoint := strings.TrimSpace(lookup(keys.url))
	if endpoint == "" {
		return nil
	}

	timeout := 10 * time.Second
	if rawTimeout := strings.TrimSpace(lookup(keys.timeout)); rawTimeout != "" {
		if parsed, err := time.ParseDuration(rawTimeout); err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	if timeoutSecondsRaw := strings.TrimSpace(lookup(keys.timeoutSeconds)); timeoutSecondsRaw != "" {
		if parsedSeconds, err := strconv.Atoi(timeoutSecondsRaw); err == nil && parsedSeconds > 0 {
			timeout = time.Duration(parsedSeconds) * time.Second
		}
	}

	return &ethereum.EthereumRPCObserverConfig{
		Endpoint: endpoint,
		Username: strings.TrimSpace(lookup(keys.user)),
		Password: lookup(keys.password),
		Timeout:  timeout,
	}
}
