package di

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	httpcontroller "payrune/internal/adapters/inbound/http/controllers"
	"payrune/internal/adapters/outbound/bitcoin"
	postgresadapter "payrune/internal/adapters/outbound/persistence/postgres"
	policyadapter "payrune/internal/adapters/outbound/policy"
	"payrune/internal/adapters/outbound/system"
	"payrune/internal/application/use_cases"
	"payrune/internal/domain/value_objects"
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

	bitcoinDeriver := bitcoin.NewHDXPubAddressDeriver(
		bitcoin.NewLegacyAddressEncoder(),
		bitcoin.NewSegwitAddressEncoder(),
		bitcoin.NewNativeSegwitAddressEncoder(),
		bitcoin.NewTaprootAddressEncoder(),
	)
	addressPolicyReader := policyadapter.NewAddressPolicyReader([]policyadapter.AddressPolicyConfig{
		{
			AddressPolicyID:      "bitcoin-mainnet-legacy",
			Chain:                value_objects.ChainBitcoin,
			Network:              value_objects.BitcoinNetworkMainnet,
			Scheme:               value_objects.BitcoinAddressSchemeLegacy,
			MinorUnit:            "satoshi",
			Decimals:             8,
			XPub:                 os.Getenv("BITCOIN_MAINNET_LEGACY_XPUB"),
			DerivationPathPrefix: "m/44'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-segwit",
			Chain:                value_objects.ChainBitcoin,
			Network:              value_objects.BitcoinNetworkMainnet,
			Scheme:               value_objects.BitcoinAddressSchemeSegwit,
			MinorUnit:            "satoshi",
			Decimals:             8,
			XPub:                 os.Getenv("BITCOIN_MAINNET_SEGWIT_XPUB"),
			DerivationPathPrefix: "m/49'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-native-segwit",
			Chain:                value_objects.ChainBitcoin,
			Network:              value_objects.BitcoinNetworkMainnet,
			Scheme:               value_objects.BitcoinAddressSchemeNativeSegwit,
			MinorUnit:            "satoshi",
			Decimals:             8,
			XPub:                 os.Getenv("BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB"),
			DerivationPathPrefix: "m/84'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-taproot",
			Chain:                value_objects.ChainBitcoin,
			Network:              value_objects.BitcoinNetworkMainnet,
			Scheme:               value_objects.BitcoinAddressSchemeTaproot,
			MinorUnit:            "satoshi",
			Decimals:             8,
			XPub:                 os.Getenv("BITCOIN_MAINNET_TAPROOT_XPUB"),
			DerivationPathPrefix: "m/86'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-legacy",
			Chain:                value_objects.ChainBitcoin,
			Network:              value_objects.BitcoinNetworkTestnet4,
			Scheme:               value_objects.BitcoinAddressSchemeLegacy,
			MinorUnit:            "satoshi",
			Decimals:             8,
			XPub:                 os.Getenv("BITCOIN_TESTNET4_LEGACY_XPUB"),
			DerivationPathPrefix: "m/44'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-segwit",
			Chain:                value_objects.ChainBitcoin,
			Network:              value_objects.BitcoinNetworkTestnet4,
			Scheme:               value_objects.BitcoinAddressSchemeSegwit,
			MinorUnit:            "satoshi",
			Decimals:             8,
			XPub:                 os.Getenv("BITCOIN_TESTNET4_SEGWIT_XPUB"),
			DerivationPathPrefix: "m/49'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-native-segwit",
			Chain:                value_objects.ChainBitcoin,
			Network:              value_objects.BitcoinNetworkTestnet4,
			Scheme:               value_objects.BitcoinAddressSchemeNativeSegwit,
			MinorUnit:            "satoshi",
			Decimals:             8,
			XPub:                 os.Getenv("BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB"),
			DerivationPathPrefix: "m/84'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-taproot",
			Chain:                value_objects.ChainBitcoin,
			Network:              value_objects.BitcoinNetworkTestnet4,
			Scheme:               value_objects.BitcoinAddressSchemeTaproot,
			MinorUnit:            "satoshi",
			Decimals:             8,
			XPub:                 os.Getenv("BITCOIN_TESTNET4_TAPROOT_XPUB"),
			DerivationPathPrefix: "m/86'/1'/0'",
		},
	})
	listAddressPoliciesUseCase := use_cases.NewListAddressPoliciesUseCase(addressPolicyReader)
	generateAddressUseCase := use_cases.NewGenerateAddressUseCase(bitcoinDeriver, addressPolicyReader)
	unitOfWork := postgresadapter.NewUnitOfWork(db, postgresadapter.NewTxRepositories)
	allocatePaymentAddressUseCase := use_cases.NewAllocatePaymentAddressUseCase(
		unitOfWork,
		bitcoinDeriver,
		addressPolicyReader,
	)
	chainAddressController := httpcontroller.NewChainAddressController(
		listAddressPoliciesUseCase,
		generateAddressUseCase,
		allocatePaymentAddressUseCase,
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
