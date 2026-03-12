package di

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	inboundadapter "payrune/internal/adapters/inbound/cloudflareworker"
	"payrune/internal/adapters/outbound/bitcoin"
	"payrune/internal/adapters/outbound/blockchain"
	cloudflarepostgres "payrune/internal/adapters/outbound/persistence/cloudflarepostgres"
	policyadapter "payrune/internal/adapters/outbound/policy"
	"payrune/internal/adapters/outbound/system"
	"payrune/internal/application/usecases"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
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
	cfDefaultBitcoinRequiredConfirmations     = int32(2)
	cfDefaultBitcoinReceiptExpiresAfter       = 24 * time.Hour
)

func BuildCloudflareAPIHTTPHandler(env map[string]string, bridgeID string) (http.Handler, error) {
	clock := system.NewClock()
	healthUseCase := usecases.NewCheckHealthUseCase(clock)

	requiredConfirmationsByNetwork, err := loadCloudflareBitcoinRequiredConfirmations(env)
	if err != nil {
		return nil, err
	}
	receiptExpiresAfterByNetwork, err := loadCloudflareBitcoinReceiptExpiresAfter(env)
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
	)
	if err != nil {
		return nil, err
	}

	addressPolicyReader := policyadapter.NewAddressPolicyReader([]policyadapter.AddressPolicyConfig{
		{
			AddressPolicyID:      "bitcoin-mainnet-legacy",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:               string(valueobjects.BitcoinAddressSchemeLegacy),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envMapValue(env, envBitcoinMainnetLegacyXPub),
			DerivationPathPrefix: "m/44'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-segwit",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:               string(valueobjects.BitcoinAddressSchemeSegwit),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envMapValue(env, envBitcoinMainnetSegwitXPub),
			DerivationPathPrefix: "m/49'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-native-segwit",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:               string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envMapValue(env, envBitcoinMainnetNativeSegwitXPub),
			DerivationPathPrefix: "m/84'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-taproot",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			Scheme:               string(valueobjects.BitcoinAddressSchemeTaproot),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envMapValue(env, envBitcoinMainnetTaprootXPub),
			DerivationPathPrefix: "m/86'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-legacy",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:               string(valueobjects.BitcoinAddressSchemeLegacy),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envMapValue(env, envBitcoinTestnet4LegacyXPub),
			DerivationPathPrefix: "m/44'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-segwit",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:               string(valueobjects.BitcoinAddressSchemeSegwit),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envMapValue(env, envBitcoinTestnet4SegwitXPub),
			DerivationPathPrefix: "m/49'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-native-segwit",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:               string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envMapValue(env, envBitcoinTestnet4NativeSegwitXPub),
			DerivationPathPrefix: "m/84'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-taproot",
			Chain:                valueobjects.SupportedChainBitcoin,
			Network:              valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			Scheme:               string(valueobjects.BitcoinAddressSchemeTaproot),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envMapValue(env, envBitcoinTestnet4TaprootXPub),
			DerivationPathPrefix: "m/86'/1'/0'",
		},
	})

	listAddressPoliciesUseCase := usecases.NewListAddressPoliciesUseCase(addressPolicyReader)
	generateAddressUseCase := usecases.NewGenerateAddressUseCase(chainAddressDeriver, addressPolicyReader)
	bridge := cloudflarepostgres.NewJSBridge()
	dbExecutor := cloudflarepostgres.NewExecutor(bridgeID, bridge)
	unitOfWork := cloudflarepostgres.NewUnitOfWork(bridgeID, bridge)
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
	getPaymentAddressStatusUseCase := usecases.NewGetPaymentAddressStatusUseCase(
		cloudflarepostgres.NewPaymentAddressStatusFinder(dbExecutor),
		addressPolicyReader,
	)

	return inboundadapter.NewAPIHandler(inboundadapter.APIDependencies{
		CheckHealthUseCase:             healthUseCase,
		ListAddressPoliciesUseCase:     listAddressPoliciesUseCase,
		GenerateAddressUseCase:         generateAddressUseCase,
		AllocatePaymentAddressUseCase:  allocatePaymentAddressUseCase,
		GetPaymentAddressStatusUseCase: getPaymentAddressStatusUseCase,
	}), nil
}

func loadCloudflareBitcoinRequiredConfirmations(env map[string]string) (map[valueobjects.NetworkID]int32, error) {
	mainnetConfirmations, err := parsePositiveInt32MapWithDefault(env, cfEnvBitcoinMainnetRequiredConfirmations, cfDefaultBitcoinRequiredConfirmations)
	if err != nil {
		return nil, err
	}
	testnet4Confirmations, err := parsePositiveInt32MapWithDefault(env, cfEnvBitcoinTestnet4RequiredConfirmations, cfDefaultBitcoinRequiredConfirmations)
	if err != nil {
		return nil, err
	}

	return map[valueobjects.NetworkID]int32{
		valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet):  mainnetConfirmations,
		valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4): testnet4Confirmations,
	}, nil
}

func loadCloudflareBitcoinReceiptExpiresAfter(env map[string]string) (map[valueobjects.NetworkID]time.Duration, error) {
	mainnetExpiresAfter, err := parseDurationMapWithDefault(env, cfEnvBitcoinMainnetReceiptExpiresAfter, cfDefaultBitcoinReceiptExpiresAfter)
	if err != nil {
		return nil, err
	}
	testnet4ExpiresAfter, err := parseDurationMapWithDefault(env, cfEnvBitcoinTestnet4ReceiptExpiresAfter, cfDefaultBitcoinReceiptExpiresAfter)
	if err != nil {
		return nil, err
	}

	return map[valueobjects.NetworkID]time.Duration{
		valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet):  mainnetExpiresAfter,
		valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4): testnet4ExpiresAfter,
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
