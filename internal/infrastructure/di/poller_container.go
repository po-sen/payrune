package di

import (
	"os"
	"strconv"
	"strings"
	"time"

	scheduleradapter "payrune/internal/adapters/inbound/scheduler"
	"payrune/internal/adapters/outbound/bitcoin"
	blockchainadapter "payrune/internal/adapters/outbound/blockchain"
	ethereumadapter "payrune/internal/adapters/outbound/ethereum"
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

	envEthereumMainnetBlockscoutURL            = "ETHEREUM_MAINNET_BLOCKSCOUT_URL"
	envEthereumMainnetBlockscoutTimeout        = "ETHEREUM_MAINNET_BLOCKSCOUT_TIMEOUT"
	envEthereumMainnetBlockscoutTimeoutSeconds = "ETHEREUM_MAINNET_BLOCKSCOUT_TIMEOUT_SECONDS"

	envEthereumSepoliaBlockscoutURL            = "ETHEREUM_SEPOLIA_BLOCKSCOUT_URL"
	envEthereumSepoliaBlockscoutTimeout        = "ETHEREUM_SEPOLIA_BLOCKSCOUT_TIMEOUT"
	envEthereumSepoliaBlockscoutTimeoutSeconds = "ETHEREUM_SEPOLIA_BLOCKSCOUT_TIMEOUT_SECONDS"
)

type bitcoinEsploraEnvKeys struct {
	url            string
	user           string
	password       string
	timeout        string
	timeoutSeconds string
}

type blockscoutEnvKeys struct {
	url            string
	timeout        string
	timeoutSeconds string
}

func NewPollerContainer() (*PollerContainer, error) {
	db, err := postgresdriver.OpenFromEnv()
	if err != nil {
		return nil, err
	}

	unitOfWork := postgresadapter.NewUnitOfWork(db)
	bitcoinObserver, err := bitcoin.NewBitcoinEsploraReceiptObserver(loadBitcoinEsploraConfigsFromEnv())
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	observers := map[valueobjects.ChainID]outport.ChainReceiptObserver{
		valueobjects.ChainIDBitcoin: bitcoinObserver,
	}
	if ethereumConfigs := loadEthereumBlockscoutConfigsFromEnv(); len(ethereumConfigs) > 0 {
		ethereumObserver, err := ethereumadapter.NewBlockscoutReceiptObserver(ethereumConfigs)
		if err != nil {
			_ = db.Close()
			return nil, err
		}
		observers[valueobjects.ChainIDEthereum] = ethereumObserver
	}
	receiptObserver, err := blockchainadapter.NewMultiChainReceiptObserver(
		observers,
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

func loadEthereumMainnetBlockscoutConfigFromEnv() *ethereumadapter.BlockscoutObserverConfig {
	return loadBlockscoutConfig(blockscoutEnvKeys{
		url:            envEthereumMainnetBlockscoutURL,
		timeout:        envEthereumMainnetBlockscoutTimeout,
		timeoutSeconds: envEthereumMainnetBlockscoutTimeoutSeconds,
	})
}

func loadEthereumSepoliaBlockscoutConfigFromEnv() *ethereumadapter.BlockscoutObserverConfig {
	return loadBlockscoutConfig(blockscoutEnvKeys{
		url:            envEthereumSepoliaBlockscoutURL,
		timeout:        envEthereumSepoliaBlockscoutTimeout,
		timeoutSeconds: envEthereumSepoliaBlockscoutTimeoutSeconds,
	})
}

func loadEthereumBlockscoutConfigsFromEnv() map[valueobjects.NetworkID]*ethereumadapter.BlockscoutObserverConfig {
	configs := make(map[valueobjects.NetworkID]*ethereumadapter.BlockscoutObserverConfig, 2)

	if mainnetConfig := loadEthereumMainnetBlockscoutConfigFromEnv(); mainnetConfig != nil {
		configs[valueobjects.NetworkID("mainnet")] = mainnetConfig
	}
	if sepoliaConfig := loadEthereumSepoliaBlockscoutConfigFromEnv(); sepoliaConfig != nil {
		configs[valueobjects.NetworkID("sepolia")] = sepoliaConfig
	}

	return configs
}

func loadBitcoinEsploraConfig(keys bitcoinEsploraEnvKeys) *bitcoin.BitcoinEsploraObserverConfig {
	endpoint := strings.TrimSpace(os.Getenv(keys.url))
	if endpoint == "" {
		return nil
	}

	timeout := 10 * time.Second
	if rawTimeout := strings.TrimSpace(os.Getenv(keys.timeout)); rawTimeout != "" {
		if parsed, err := time.ParseDuration(rawTimeout); err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	if timeoutSecondsRaw := strings.TrimSpace(os.Getenv(keys.timeoutSeconds)); timeoutSecondsRaw != "" {
		if parsedSeconds, err := strconv.Atoi(timeoutSecondsRaw); err == nil && parsedSeconds > 0 {
			timeout = time.Duration(parsedSeconds) * time.Second
		}
	}

	return &bitcoin.BitcoinEsploraObserverConfig{
		Endpoint: endpoint,
		Username: strings.TrimSpace(os.Getenv(keys.user)),
		Password: os.Getenv(keys.password),
		Timeout:  timeout,
	}
}

func loadBlockscoutConfig(keys blockscoutEnvKeys) *ethereumadapter.BlockscoutObserverConfig {
	endpoint := strings.TrimSpace(os.Getenv(keys.url))
	if endpoint == "" {
		return nil
	}

	timeout := 10 * time.Second
	if rawTimeout := strings.TrimSpace(os.Getenv(keys.timeout)); rawTimeout != "" {
		if parsed, err := time.ParseDuration(rawTimeout); err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	if timeoutSecondsRaw := strings.TrimSpace(os.Getenv(keys.timeoutSeconds)); timeoutSecondsRaw != "" {
		if parsedSeconds, err := strconv.Atoi(timeoutSecondsRaw); err == nil && parsedSeconds > 0 {
			timeout = time.Duration(parsedSeconds) * time.Second
		}
	}

	return &ethereumadapter.BlockscoutObserverConfig{
		BaseURL: endpoint,
		Timeout: timeout,
	}
}
