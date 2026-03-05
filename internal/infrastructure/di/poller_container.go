package di

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"payrune/internal/adapters/outbound/bitcoin"
	blockchainadapter "payrune/internal/adapters/outbound/blockchain"
	postgresadapter "payrune/internal/adapters/outbound/persistence/postgres"
	"payrune/internal/adapters/outbound/system"
	inport "payrune/internal/application/ports/in"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/application/use_cases"
	"payrune/internal/domain/value_objects"
)

type PollerContainer struct {
	RunReceiptPollingCycleUseCase inport.RunReceiptPollingCycleUseCase
	closeFn                       func() error
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
)

type bitcoinEsploraEnvKeys struct {
	url            string
	user           string
	password       string
	timeout        string
	timeoutSeconds string
}

func NewPollerContainer() (*PollerContainer, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database connection: %w", err)
	}

	unitOfWork := postgresadapter.NewUnitOfWork(db, postgresadapter.NewTxRepositories)
	bitcoinObserver, err := bitcoin.NewBitcoinEsploraReceiptObserver(loadBitcoinEsploraConfigsFromEnv())
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	receiptObserver, err := blockchainadapter.NewChainRouterReceiptObserver(
		map[value_objects.ChainID]outport.ChainReceiptObserver{
			value_objects.ChainIDBitcoin: bitcoinObserver,
		},
	)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	clock := system.NewClock()
	runReceiptPollingCycleUseCase := use_cases.NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		receiptObserver,
		clock,
	)

	return &PollerContainer{
		RunReceiptPollingCycleUseCase: runReceiptPollingCycleUseCase,
		closeFn:                       db.Close,
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

func loadBitcoinEsploraConfigsFromEnv() map[value_objects.NetworkID]*bitcoin.BitcoinEsploraObserverConfig {
	configs := make(map[value_objects.NetworkID]*bitcoin.BitcoinEsploraObserverConfig, 2)

	if mainnetConfig := loadBitcoinMainnetEsploraConfigFromEnv(); mainnetConfig != nil {
		configs[value_objects.NetworkID(value_objects.BitcoinNetworkMainnet)] = mainnetConfig
	}
	if testnet4Config := loadBitcoinTestnet4EsploraConfigFromEnv(); testnet4Config != nil {
		configs[value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4)] = testnet4Config
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
