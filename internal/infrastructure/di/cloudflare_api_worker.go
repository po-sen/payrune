package di

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	httpadapter "payrune/internal/adapters/inbound/http"
	httpcontroller "payrune/internal/adapters/inbound/http/controllers"
	"payrune/internal/adapters/outbound/bitcoin"
	"payrune/internal/adapters/outbound/blockchain"
	"payrune/internal/adapters/outbound/ethereum"
	cloudflarepostgres "payrune/internal/adapters/outbound/persistence/cloudflarepostgres"
	policyadapter "payrune/internal/adapters/outbound/policy"
	"payrune/internal/adapters/outbound/system"
	"payrune/internal/application/usecases"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
	cloudflarepostgresdriver "payrune/internal/infrastructure/drivers/cloudflarepostgres"
)

const (
	envBitcoinMainnetLegacyXPub               = "BITCOIN_MAINNET_LEGACY_XPUB"
	envBitcoinMainnetSegwitXPub               = "BITCOIN_MAINNET_SEGWIT_XPUB"
	envBitcoinMainnetNativeSegwitXPub         = "BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB"
	envBitcoinMainnetTaprootXPub              = "BITCOIN_MAINNET_TAPROOT_XPUB"
	envBitcoinTestnet4LegacyXPub              = "BITCOIN_TESTNET4_LEGACY_XPUB"
	envBitcoinTestnet4SegwitXPub              = "BITCOIN_TESTNET4_SEGWIT_XPUB"
	envBitcoinTestnet4NativeSegwitXPub        = "BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB"
	envBitcoinTestnet4TaprootXPub             = "BITCOIN_TESTNET4_TAPROOT_XPUB"
	cfEnvBitcoinMainnetRequiredConfirmations  = "BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS"
	cfEnvBitcoinTestnet4RequiredConfirmations = "BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS"
	cfEnvBitcoinMainnetReceiptExpiresAfter    = "BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER"
	cfEnvBitcoinTestnet4ReceiptExpiresAfter   = "BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER"
	cfEnvEthereumMainnetRequiredConfirmations = "ETHEREUM_MAINNET_REQUIRED_CONFIRMATIONS"
	cfEnvEthereumSepoliaRequiredConfirmations = "ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS"
	cfEnvEthereumMainnetReceiptExpiresAfter   = "ETHEREUM_MAINNET_RECEIPT_EXPIRES_AFTER"
	cfEnvEthereumSepoliaReceiptExpiresAfter   = "ETHEREUM_SEPOLIA_RECEIPT_EXPIRES_AFTER"
	cfEnvEthereumMainnetCreate2Collector      = "ETHEREUM_MAINNET_CREATE2_COLLECTOR_ADDRESS"
	cfEnvEthereumSepoliaCreate2Collector      = "ETHEREUM_SEPOLIA_CREATE2_COLLECTOR_ADDRESS"
	cfEnvEthereumMainnetCreate2DerivationKey  = "ETHEREUM_MAINNET_CREATE2_DERIVATION_KEY"
	cfEnvEthereumSepoliaCreate2DerivationKey  = "ETHEREUM_SEPOLIA_CREATE2_DERIVATION_KEY"
	cfDefaultBitcoinRequiredConfirmations     = int32(2)
	cfDefaultBitcoinReceiptExpiresAfter       = 24 * time.Hour
)

func BuildCloudflareAPIHTTPHandler(env map[string]string, bridgeID string) (http.Handler, error) {
	clock := system.NewClock()
	healthUseCase := usecases.NewCheckHealthUseCase(clock)

	requiredConfirmationsByScope, err := loadCloudflareReceiptRequiredConfirmations(env)
	if err != nil {
		return nil, err
	}
	receiptExpiresAfterByScope, err := loadCloudflareReceiptExpiresAfter(env)
	if err != nil {
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
		return nil, err
	}
	ethereumCreate2SaltDeriver := ethereum.NewCreate2SaltDeriver(
		buildEthereumCreate2DerivationKeys(
			envMapValue(env, cfEnvEthereumMainnetCreate2DerivationKey),
			envMapValue(env, cfEnvEthereumSepoliaCreate2DerivationKey),
		),
	)

	addressPolicyReader := policyadapter.NewAddressPolicyReader([]policyadapter.AddressPolicyConfig{
		{
			AddressPolicyID:        "bitcoin-mainnet-legacy",
			Chain:                  valueobjects.SupportedChainBitcoin,
			Network:                valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:                 string(valueobjects.BitcoinAddressSchemeLegacy),
			MinorUnit:              "satoshi",
			Decimals:               8,
			AddressSourceRef:       envMapValue(env, envBitcoinMainnetLegacyXPub),
			AddressReferencePrefix: "m/44'/0'/0'",
		},
		{
			AddressPolicyID:        "bitcoin-mainnet-segwit",
			Chain:                  valueobjects.SupportedChainBitcoin,
			Network:                valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:                 string(valueobjects.BitcoinAddressSchemeSegwit),
			MinorUnit:              "satoshi",
			Decimals:               8,
			AddressSourceRef:       envMapValue(env, envBitcoinMainnetSegwitXPub),
			AddressReferencePrefix: "m/49'/0'/0'",
		},
		{
			AddressPolicyID:        "bitcoin-mainnet-native-segwit",
			Chain:                  valueobjects.SupportedChainBitcoin,
			Network:                valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:                 string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			MinorUnit:              "satoshi",
			Decimals:               8,
			AddressSourceRef:       envMapValue(env, envBitcoinMainnetNativeSegwitXPub),
			AddressReferencePrefix: "m/84'/0'/0'",
		},
		{
			AddressPolicyID:        "bitcoin-mainnet-taproot",
			Chain:                  valueobjects.SupportedChainBitcoin,
			Network:                valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:                 string(valueobjects.BitcoinAddressSchemeTaproot),
			MinorUnit:              "satoshi",
			Decimals:               8,
			AddressSourceRef:       envMapValue(env, envBitcoinMainnetTaprootXPub),
			AddressReferencePrefix: "m/86'/0'/0'",
		},
		{
			AddressPolicyID:        "bitcoin-testnet4-legacy",
			Chain:                  valueobjects.SupportedChainBitcoin,
			Network:                valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:                 string(valueobjects.BitcoinAddressSchemeLegacy),
			MinorUnit:              "satoshi",
			Decimals:               8,
			AddressSourceRef:       envMapValue(env, envBitcoinTestnet4LegacyXPub),
			AddressReferencePrefix: "m/44'/1'/0'",
		},
		{
			AddressPolicyID:        "bitcoin-testnet4-segwit",
			Chain:                  valueobjects.SupportedChainBitcoin,
			Network:                valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:                 string(valueobjects.BitcoinAddressSchemeSegwit),
			MinorUnit:              "satoshi",
			Decimals:               8,
			AddressSourceRef:       envMapValue(env, envBitcoinTestnet4SegwitXPub),
			AddressReferencePrefix: "m/49'/1'/0'",
		},
		{
			AddressPolicyID:        "bitcoin-testnet4-native-segwit",
			Chain:                  valueobjects.SupportedChainBitcoin,
			Network:                valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:                 string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			MinorUnit:              "satoshi",
			Decimals:               8,
			AddressSourceRef:       envMapValue(env, envBitcoinTestnet4NativeSegwitXPub),
			AddressReferencePrefix: "m/84'/1'/0'",
		},
		{
			AddressPolicyID:        "bitcoin-testnet4-taproot",
			Chain:                  valueobjects.SupportedChainBitcoin,
			Network:                valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:                 string(valueobjects.BitcoinAddressSchemeTaproot),
			MinorUnit:              "satoshi",
			Decimals:               8,
			AddressSourceRef:       envMapValue(env, envBitcoinTestnet4TaprootXPub),
			AddressReferencePrefix: "m/86'/1'/0'",
		},
		newEthereumCreate2PolicyConfig(
			valueobjects.NetworkID("mainnet"),
			envMapValue(env, cfEnvEthereumMainnetCreate2Collector),
			ethereumCreate2SaltDeriver,
		),
		newEthereumCreate2PolicyConfig(
			valueobjects.NetworkID("sepolia"),
			envMapValue(env, cfEnvEthereumSepoliaCreate2Collector),
			ethereumCreate2SaltDeriver,
		),
	})

	listAddressPoliciesUseCase := usecases.NewListAddressPoliciesUseCase(addressPolicyReader)
	generateAddressUseCase := usecases.NewGenerateAddressUseCase(chainAddressDeriver, addressPolicyReader)
	bridge := cloudflarepostgresdriver.NewJSBridge()
	dbExecutor := cloudflarepostgres.NewExecutor(bridgeID, bridge)
	unitOfWork := cloudflarepostgres.NewUnitOfWork(bridgeID, bridge)
	allocationIssuancePolicy := policies.NewPaymentAddressAllocationIssuancePolicy(
		requiredConfirmationsByScope,
		receiptExpiresAfterByScope,
	)
	allocatePaymentAddressUseCase := usecases.NewAllocatePaymentAddressUseCase(
		unitOfWork,
		chainAddressDeriver,
		ethereumCreate2SaltDeriver,
		addressPolicyReader,
		allocationIssuancePolicy,
		clock,
	)
	getPaymentAddressStatusUseCase := usecases.NewGetPaymentAddressStatusUseCase(
		cloudflarepostgres.NewPaymentAddressStatusFinder(dbExecutor),
		addressPolicyReader,
	)

	return httpadapter.NewPublicRouter(httpadapter.RouterControllers{
		Health: httpcontroller.NewHealthController(healthUseCase),
		ChainAddress: httpcontroller.NewChainAddressController(
			listAddressPoliciesUseCase,
			generateAddressUseCase,
			allocatePaymentAddressUseCase,
			getPaymentAddressStatusUseCase,
		),
	}), nil
}

func loadCloudflareReceiptRequiredConfirmations(
	env map[string]string,
) (map[policies.PaymentReceiptTermsScope]int32, error) {
	mainnetConfirmations, err := parsePositiveInt32MapWithDefault(env, cfEnvBitcoinMainnetRequiredConfirmations, cfDefaultBitcoinRequiredConfirmations)
	if err != nil {
		return nil, err
	}
	testnet4Confirmations, err := parsePositiveInt32MapWithDefault(env, cfEnvBitcoinTestnet4RequiredConfirmations, cfDefaultBitcoinRequiredConfirmations)
	if err != nil {
		return nil, err
	}
	ethereumMainnetConfirmations, err := parsePositiveInt32MapWithDefault(
		env,
		cfEnvEthereumMainnetRequiredConfirmations,
		cfDefaultBitcoinRequiredConfirmations,
	)
	if err != nil {
		return nil, err
	}
	ethereumSepoliaConfirmations, err := parsePositiveInt32MapWithDefault(
		env,
		cfEnvEthereumSepoliaRequiredConfirmations,
		cfDefaultBitcoinRequiredConfirmations,
	)
	if err != nil {
		return nil, err
	}

	return map[policies.PaymentReceiptTermsScope]int32{
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
		): mainnetConfirmations,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
		): testnet4Confirmations,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainEthereum,
			valueobjects.NetworkID("mainnet"),
		): ethereumMainnetConfirmations,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainEthereum,
			valueobjects.NetworkID("sepolia"),
		): ethereumSepoliaConfirmations,
	}, nil
}

func loadCloudflareReceiptExpiresAfter(
	env map[string]string,
) (map[policies.PaymentReceiptTermsScope]time.Duration, error) {
	mainnetExpiresAfter, err := parseDurationMapWithDefault(env, cfEnvBitcoinMainnetReceiptExpiresAfter, cfDefaultBitcoinReceiptExpiresAfter)
	if err != nil {
		return nil, err
	}
	testnet4ExpiresAfter, err := parseDurationMapWithDefault(env, cfEnvBitcoinTestnet4ReceiptExpiresAfter, cfDefaultBitcoinReceiptExpiresAfter)
	if err != nil {
		return nil, err
	}
	ethereumMainnetExpiresAfter, err := parseDurationMapWithDefault(
		env,
		cfEnvEthereumMainnetReceiptExpiresAfter,
		cfDefaultBitcoinReceiptExpiresAfter,
	)
	if err != nil {
		return nil, err
	}
	ethereumSepoliaExpiresAfter, err := parseDurationMapWithDefault(
		env,
		cfEnvEthereumSepoliaReceiptExpiresAfter,
		cfDefaultBitcoinReceiptExpiresAfter,
	)
	if err != nil {
		return nil, err
	}

	return map[policies.PaymentReceiptTermsScope]time.Duration{
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
		): mainnetExpiresAfter,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
		): testnet4ExpiresAfter,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainEthereum,
			valueobjects.NetworkID("mainnet"),
		): ethereumMainnetExpiresAfter,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainEthereum,
			valueobjects.NetworkID("sepolia"),
		): ethereumSepoliaExpiresAfter,
	}, nil
}

func parsePositiveInt32MapWithDefault(env map[string]string, key string, fallback int32) (int32, error) {
	rawValue := envMapValue(env, key)
	if rawValue == "" {
		return fallback, nil
	}

	parsedValue, err := strconv.ParseInt(rawValue, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be a positive integer: %w", key, err)
	}
	if parsedValue <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}

	return int32(parsedValue), nil
}

func envMapValue(env map[string]string, key string) string {
	return strings.TrimSpace(env[key])
}

func parseDurationMapWithDefault(env map[string]string, key string, fallback time.Duration) (time.Duration, error) {
	rawValue := strings.TrimSpace(env[key])
	if rawValue == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(rawValue)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return duration, nil
}
