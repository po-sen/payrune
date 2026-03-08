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

	httpcontroller "payrune/internal/adapters/inbound/http/controllers"
	"payrune/internal/adapters/outbound/bitcoin"
	"payrune/internal/adapters/outbound/blockchain"
	postgresadapter "payrune/internal/adapters/outbound/persistence/postgres"
	policyadapter "payrune/internal/adapters/outbound/policy"
	"payrune/internal/adapters/outbound/system"
	"payrune/internal/application/use_cases"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/value_objects"
)

const (
	envBitcoinMainnetRequiredConfirmations  = "BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS"
	envBitcoinTestnet4RequiredConfirmations = "BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS"
	envBitcoinMainnetReceiptExpiresAfter    = "BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER"
	envBitcoinTestnet4ReceiptExpiresAfter   = "BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER"
	defaultBitcoinRequiredConfirmations     = int32(1)
	defaultBitcoinReceiptExpiresAfter       = 7 * 24 * time.Hour
)

type Container struct {
	HealthController       *httpcontroller.HealthController
	ChainAddressController *httpcontroller.ChainAddressController
	closeFn                func() error
}

func NewContainer() (*Container, error) {
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

	clock := system.NewClock()
	healthUseCase := use_cases.NewCheckHealthUseCase(clock)
	healthController := httpcontroller.NewHealthController(healthUseCase)
	requiredConfirmationsByNetwork, err := loadBitcoinRequiredConfirmationsFromEnv()
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	receiptExpiresAfterByNetwork, err := loadBitcoinReceiptExpiresAfterByNetworkFromEnv()
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	bitcoinDeriver := bitcoin.NewHDXPubAddressDeriver(
		bitcoin.NewLegacyAddressEncoder(),
		bitcoin.NewSegwitAddressEncoder(),
		bitcoin.NewNativeSegwitAddressEncoder(),
		bitcoin.NewTaprootAddressEncoder(),
	)
	chainAddressDeriver, err := blockchain.NewMultiChainAddressDeriver(
		bitcoin.NewChainAddressDeriver(bitcoinDeriver),
	)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	addressPolicyReader := policyadapter.NewAddressPolicyReader([]policyadapter.AddressPolicyConfig{
		{
			AddressPolicyID:      "bitcoin-mainnet-legacy",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			Scheme:               string(value_objects.BitcoinAddressSchemeLegacy),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_MAINNET_LEGACY_XPUB"),
			DerivationPathPrefix: "m/44'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-segwit",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			Scheme:               string(value_objects.BitcoinAddressSchemeSegwit),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_MAINNET_SEGWIT_XPUB"),
			DerivationPathPrefix: "m/49'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-native-segwit",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			Scheme:               string(value_objects.BitcoinAddressSchemeNativeSegwit),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB"),
			DerivationPathPrefix: "m/84'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-taproot",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			Scheme:               string(value_objects.BitcoinAddressSchemeTaproot),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_MAINNET_TAPROOT_XPUB"),
			DerivationPathPrefix: "m/86'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-legacy",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
			Scheme:               string(value_objects.BitcoinAddressSchemeLegacy),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_TESTNET4_LEGACY_XPUB"),
			DerivationPathPrefix: "m/44'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-segwit",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
			Scheme:               string(value_objects.BitcoinAddressSchemeSegwit),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_TESTNET4_SEGWIT_XPUB"),
			DerivationPathPrefix: "m/49'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-native-segwit",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
			Scheme:               string(value_objects.BitcoinAddressSchemeNativeSegwit),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB"),
			DerivationPathPrefix: "m/84'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-taproot",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
			Scheme:               string(value_objects.BitcoinAddressSchemeTaproot),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_TESTNET4_TAPROOT_XPUB"),
			DerivationPathPrefix: "m/86'/1'/0'",
		},
	})
	listAddressPoliciesUseCase := use_cases.NewListAddressPoliciesUseCase(addressPolicyReader)
	generateAddressUseCase := use_cases.NewGenerateAddressUseCase(chainAddressDeriver, addressPolicyReader)
	allocationStore := postgresadapter.NewPaymentAddressAllocationStore(db)
	paymentAddressStatusFinder := postgresadapter.NewPaymentAddressStatusFinder(db)
	idempotencyStore := postgresadapter.NewPaymentAddressIdempotencyStore(db)
	unitOfWork := postgresadapter.NewUnitOfWork(db)
	allocationIssuancePolicy := policies.NewPaymentAddressAllocationIssuancePolicy(
		requiredConfirmationsByNetwork,
		receiptExpiresAfterByNetwork,
	)
	allocatePaymentAddressUseCase := use_cases.NewAllocatePaymentAddressUseCase(
		unitOfWork,
		allocationStore,
		idempotencyStore,
		chainAddressDeriver,
		addressPolicyReader,
		allocationIssuancePolicy,
		clock,
	)
	getPaymentAddressStatusUseCase := use_cases.NewGetPaymentAddressStatusUseCase(
		paymentAddressStatusFinder,
		addressPolicyReader,
	)
	chainAddressController := httpcontroller.NewChainAddressController(
		listAddressPoliciesUseCase,
		generateAddressUseCase,
		allocatePaymentAddressUseCase,
		getPaymentAddressStatusUseCase,
	)

	return &Container{
		HealthController:       healthController,
		ChainAddressController: chainAddressController,
		closeFn:                db.Close,
	}, nil
}

func (c *Container) Close() error {
	if c.closeFn == nil {
		return nil
	}
	return c.closeFn()
}

func loadBitcoinRequiredConfirmationsFromEnv() (map[value_objects.NetworkID]int32, error) {
	mainnetConfirmations, err := parsePositiveInt32EnvWithDefault(
		envBitcoinMainnetRequiredConfirmations,
		defaultBitcoinRequiredConfirmations,
	)
	if err != nil {
		return nil, err
	}
	testnet4Confirmations, err := parsePositiveInt32EnvWithDefault(
		envBitcoinTestnet4RequiredConfirmations,
		defaultBitcoinRequiredConfirmations,
	)
	if err != nil {
		return nil, err
	}

	return map[value_objects.NetworkID]int32{
		value_objects.NetworkID(value_objects.BitcoinNetworkMainnet):  mainnetConfirmations,
		value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4): testnet4Confirmations,
	}, nil
}

func loadBitcoinReceiptExpiresAfterByNetworkFromEnv() (map[value_objects.NetworkID]time.Duration, error) {
	mainnetExpiresAfter, err := parsePositiveDurationEnvWithDefault(
		envBitcoinMainnetReceiptExpiresAfter,
		defaultBitcoinReceiptExpiresAfter,
	)
	if err != nil {
		return nil, err
	}
	testnet4ExpiresAfter, err := parsePositiveDurationEnvWithDefault(
		envBitcoinTestnet4ReceiptExpiresAfter,
		defaultBitcoinReceiptExpiresAfter,
	)
	if err != nil {
		return nil, err
	}

	return map[value_objects.NetworkID]time.Duration{
		value_objects.NetworkID(value_objects.BitcoinNetworkMainnet):  mainnetExpiresAfter,
		value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4): testnet4ExpiresAfter,
	}, nil
}

func parsePositiveInt32EnvWithDefault(key string, fallback int32) (int32, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return int32(parsed), nil
}

func parsePositiveDurationEnvWithDefault(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return parsed, nil
}
