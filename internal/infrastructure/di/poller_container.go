package di

import (
	"fmt"
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
	postgresdriver "payrune/internal/infrastructure/drivers/postgres"
)

type PollerContainer struct {
	PollerHandler *scheduleradapter.PollerHandler
	closeFn       func() error
}

const (
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

func NewPollerContainer() (*PollerContainer, error) {
	db, err := postgresdriver.OpenFromEnv()
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
	receiptObserver, err := blockchainadapter.NewMultiChainReceiptObserver(
		chainObservers,
	)
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

	return &PollerContainer{
		PollerHandler: scheduleradapter.NewPollerHandler(scheduleradapter.PollerDependencies{
			RunReceiptPollingCycleUseCase: runReceiptPollingCycleUseCase,
		}),
		closeFn: db.Close,
	}, nil
}

func (c *PollerContainer) Close() error {
	if c.closeFn == nil {
		return nil
	}
	return c.closeFn()
}

func loadConfiguredPollerChainFromEnv() (valueobjects.ChainID, bool, error) {
	return loadConfiguredPollerChainFromLookup(os.Getenv)
}

func loadConfiguredPollerChainFromLookup(
	lookup func(string) string,
) (valueobjects.ChainID, bool, error) {
	rawChain := strings.TrimSpace(lookup(envPollChain))
	if rawChain == "" {
		return "", false, nil
	}

	chain, ok := valueobjects.ParseChainID(rawChain)
	if !ok {
		return "", false, fmt.Errorf("%s is invalid", envPollChain)
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
