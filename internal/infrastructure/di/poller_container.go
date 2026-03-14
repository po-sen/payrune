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

	scheduleradapter "payrune/internal/adapters/inbound/scheduler"
	"payrune/internal/adapters/outbound/bitcoin"
	blockchainadapter "payrune/internal/adapters/outbound/blockchain"
	postgresadapter "payrune/internal/adapters/outbound/persistence/postgres"
	"payrune/internal/adapters/outbound/system"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/application/usecases"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
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

	unitOfWork := postgresadapter.NewUnitOfWork(db)
	bitcoinObserver, err := bitcoin.NewBitcoinEsploraReceiptObserver(loadBitcoinEsploraConfigsFromEnv())
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	receiptObserver, err := blockchainadapter.NewMultiChainReceiptObserver(
		map[valueobjects.ChainID]outport.ChainReceiptObserver{
			valueobjects.ChainIDBitcoin: bitcoinObserver,
		},
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
