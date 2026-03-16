package di

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	httpadapter "payrune/internal/adapters/inbound/http"
	httpcontroller "payrune/internal/adapters/inbound/http/controllers"
	"payrune/internal/adapters/outbound/bitcoin"
	"payrune/internal/adapters/outbound/blockchain"
	"payrune/internal/adapters/outbound/ethereum"
	postgresadapter "payrune/internal/adapters/outbound/persistence/postgres"
	policyadapter "payrune/internal/adapters/outbound/policy"
	"payrune/internal/adapters/outbound/system"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/application/usecases"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
	postgresdriver "payrune/internal/infrastructure/drivers/postgres"
)

const (
	envBitcoinMainnetRequiredConfirmations  = "BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS"
	envBitcoinTestnet4RequiredConfirmations = "BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS"
	envBitcoinMainnetReceiptExpiresAfter    = "BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER"
	envBitcoinTestnet4ReceiptExpiresAfter   = "BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER"
	envEthereumSepoliaUSDTTokenAddress      = "ETHEREUM_SEPOLIA_USDT_TOKEN_ADDRESS"
	defaultBitcoinRequiredConfirmations     = int32(1)
	defaultBitcoinReceiptExpiresAfter       = 7 * 24 * time.Hour
	ethereumCreate2ForwarderScheme          = "create2_forwarder"
	ethereumMainnetUSDTTokenAddress         = "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	evmFactoryAddressFingerprintAlgorithm   = "evm-factory-address-v1"
)

type Container struct {
	APIHandler http.Handler
	closeFn    func() error
}

func NewContainer() (*Container, error) {
	db, err := postgresdriver.OpenFromEnv()
	if err != nil {
		return nil, err
	}

	clock := system.NewClock()
	healthUseCase := usecases.NewCheckHealthUseCase(clock)
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
	activeEVMFactories, err := loadActiveEVMFactoriesForContainer(db)
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
		ethereum.NewChainAddressDeriver(),
	)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	addressPolicyReader := policyadapter.NewAddressPolicyReader([]policyadapter.AddressPolicyConfig{
		{
			AddressPolicyID:      "bitcoin-mainnet-legacy",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:               string(valueobjects.BitcoinAddressSchemeLegacy),
			AssetCode:            "btc",
			AssetType:            "native",
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_MAINNET_LEGACY_XPUB"),
			DerivationPathPrefix: "m/44'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-segwit",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:               string(valueobjects.BitcoinAddressSchemeSegwit),
			AssetCode:            "btc",
			AssetType:            "native",
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_MAINNET_SEGWIT_XPUB"),
			DerivationPathPrefix: "m/49'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-native-segwit",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:               string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			AssetCode:            "btc",
			AssetType:            "native",
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB"),
			DerivationPathPrefix: "m/84'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-taproot",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:               string(valueobjects.BitcoinAddressSchemeTaproot),
			AssetCode:            "btc",
			AssetType:            "native",
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_MAINNET_TAPROOT_XPUB"),
			DerivationPathPrefix: "m/86'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-legacy",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:               string(valueobjects.BitcoinAddressSchemeLegacy),
			AssetCode:            "btc",
			AssetType:            "native",
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_TESTNET4_LEGACY_XPUB"),
			DerivationPathPrefix: "m/44'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-segwit",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:               string(valueobjects.BitcoinAddressSchemeSegwit),
			AssetCode:            "btc",
			AssetType:            "native",
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_TESTNET4_SEGWIT_XPUB"),
			DerivationPathPrefix: "m/49'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-native-segwit",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:               string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			AssetCode:            "btc",
			AssetType:            "native",
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB"),
			DerivationPathPrefix: "m/84'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-taproot",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:               string(valueobjects.BitcoinAddressSchemeTaproot),
			AssetCode:            "btc",
			AssetType:            "native",
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     os.Getenv("BITCOIN_TESTNET4_TAPROOT_XPUB"),
			DerivationPathPrefix: "m/86'/1'/0'",
		},
		{
			AddressPolicyID:          "ethereum-mainnet-eth",
			Chain:                    valueobjects.SupportedChainEthereum,
			Network:                  valueobjects.NetworkID("mainnet"),
			Scheme:                   ethereumCreate2ForwarderScheme,
			AssetCode:                "eth",
			AssetType:                "native",
			MinorUnit:                "wei",
			Decimals:                 18,
			AccountPublicKey:         activeEVMFactories[valueobjects.NetworkID("mainnet")].FactoryAddress,
			PublicKeyFingerprintAlgo: evmFactoryAddressFingerprintAlgorithm,
			PublicKeyFingerprint:     evmFactoryFingerprint(activeEVMFactories[valueobjects.NetworkID("mainnet")].FactoryAddress),
		},
		{
			AddressPolicyID:          "ethereum-mainnet-usdt",
			Chain:                    valueobjects.SupportedChainEthereum,
			Network:                  valueobjects.NetworkID("mainnet"),
			Scheme:                   ethereumCreate2ForwarderScheme,
			AssetCode:                "usdt",
			AssetType:                "erc20",
			TokenAddress:             ethereumMainnetUSDTTokenAddress,
			MinorUnit:                "microUsdt",
			Decimals:                 6,
			AccountPublicKey:         activeEVMFactories[valueobjects.NetworkID("mainnet")].FactoryAddress,
			PublicKeyFingerprintAlgo: evmFactoryAddressFingerprintAlgorithm,
			PublicKeyFingerprint:     evmFactoryFingerprint(activeEVMFactories[valueobjects.NetworkID("mainnet")].FactoryAddress),
		},
		{
			AddressPolicyID:          "ethereum-sepolia-eth",
			Chain:                    valueobjects.SupportedChainEthereum,
			Network:                  valueobjects.NetworkID("sepolia"),
			Scheme:                   ethereumCreate2ForwarderScheme,
			AssetCode:                "eth",
			AssetType:                "native",
			MinorUnit:                "wei",
			Decimals:                 18,
			AccountPublicKey:         activeEVMFactories[valueobjects.NetworkID("sepolia")].FactoryAddress,
			PublicKeyFingerprintAlgo: evmFactoryAddressFingerprintAlgorithm,
			PublicKeyFingerprint:     evmFactoryFingerprint(activeEVMFactories[valueobjects.NetworkID("sepolia")].FactoryAddress),
		},
		{
			AddressPolicyID: "ethereum-sepolia-usdt",
			Chain:           valueobjects.SupportedChainEthereum,
			Network:         valueobjects.NetworkID("sepolia"),
			Scheme:          ethereumCreate2ForwarderScheme,
			AssetCode:       "usdt",
			AssetType:       "erc20",
			TokenAddress:    strings.TrimSpace(os.Getenv(envEthereumSepoliaUSDTTokenAddress)),
			MinorUnit:       "microUsdt",
			Decimals:        6,
			AccountPublicKey: enabledEthereumSepoliaUSDTFactoryAddress(
				strings.TrimSpace(os.Getenv(envEthereumSepoliaUSDTTokenAddress)),
				activeEVMFactories[valueobjects.NetworkID("sepolia")].FactoryAddress,
			),
			PublicKeyFingerprintAlgo: evmFactoryAddressFingerprintAlgorithm,
			PublicKeyFingerprint: evmFactoryFingerprint(enabledEthereumSepoliaUSDTFactoryAddress(
				strings.TrimSpace(os.Getenv(envEthereumSepoliaUSDTTokenAddress)),
				activeEVMFactories[valueobjects.NetworkID("sepolia")].FactoryAddress,
			)),
		},
	})
	listAddressPoliciesUseCase := usecases.NewListAddressPoliciesUseCase(addressPolicyReader)
	generateAddressUseCase := usecases.NewGenerateAddressUseCase(chainAddressDeriver, addressPolicyReader)
	unitOfWork := postgresadapter.NewUnitOfWork(db)
	allocationIssuancePolicy := policies.NewPaymentAddressAllocationIssuancePolicy(
		requiredConfirmationsByNetwork,
		receiptExpiresAfterByNetwork,
	)
	allocatePaymentAddressUseCase := usecases.NewAllocatePaymentAddressUseCase(
		unitOfWork,
		chainAddressDeriver,
		addressPolicyReader,
		allocationIssuancePolicy,
		clock,
	)
	getPaymentAddressStatusUseCase := usecases.NewGetPaymentAddressStatusUseCase(postgresadapter.NewPaymentAddressStatusFinder(db))
	chainAddressController := httpcontroller.NewChainAddressController(
		listAddressPoliciesUseCase,
		generateAddressUseCase,
		allocatePaymentAddressUseCase,
		getPaymentAddressStatusUseCase,
	)

	return &Container{
		APIHandler: httpadapter.NewPublicRouter(httpadapter.RouterControllers{
			Health:       healthController,
			ChainAddress: chainAddressController,
		}),
		closeFn: db.Close,
	}, nil
}

func (c *Container) Close() error {
	if c.closeFn == nil {
		return nil
	}
	return c.closeFn()
}

func loadBitcoinRequiredConfirmationsFromEnv() (map[valueobjects.NetworkID]int32, error) {
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

	return map[valueobjects.NetworkID]int32{
		valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet):  mainnetConfirmations,
		valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4): testnet4Confirmations,
	}, nil
}

func loadBitcoinReceiptExpiresAfterByNetworkFromEnv() (map[valueobjects.NetworkID]time.Duration, error) {
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

	return map[valueobjects.NetworkID]time.Duration{
		valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet):  mainnetExpiresAfter,
		valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4): testnet4ExpiresAfter,
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

func loadActiveEVMFactoriesForContainer(db *sql.DB) (map[valueobjects.NetworkID]outport.EVMFactoryRecord, error) {
	records, err := postgresadapter.NewEVMFactoryStore(db).ListActive(context.Background())
	if err != nil {
		return nil, err
	}

	result := make(map[valueobjects.NetworkID]outport.EVMFactoryRecord, len(records))
	for _, record := range records {
		result[record.Network] = record
	}
	return result, nil
}

func evmFactoryFingerprint(factoryAddress string) string {
	return strings.ToLower(strings.TrimSpace(factoryAddress))
}

func enabledEthereumSepoliaUSDTFactoryAddress(tokenAddress string, factoryAddress string) string {
	if strings.TrimSpace(tokenAddress) == "" {
		return ""
	}
	return strings.TrimSpace(factoryAddress)
}
