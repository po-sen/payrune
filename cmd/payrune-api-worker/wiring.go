//go:build js && wasm

package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	httpcontroller "payrune/internal/adapters/inbound/http/controllers"
	"payrune/internal/adapters/outbound/bitcoin"
	"payrune/internal/adapters/outbound/blockchain"
	cloudflarepostgres "payrune/internal/adapters/outbound/persistence/cloudflarepostgres"
	policyadapter "payrune/internal/adapters/outbound/policy"
	"payrune/internal/adapters/outbound/system"
	"payrune/internal/application/use_cases"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/value_objects"
)

const (
	envBitcoinMainnetLegacyXPub             = "BITCOIN_MAINNET_LEGACY_XPUB"
	envBitcoinMainnetSegwitXPub             = "BITCOIN_MAINNET_SEGWIT_XPUB"
	envBitcoinMainnetNativeSegwitXPub       = "BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB"
	envBitcoinMainnetTaprootXPub            = "BITCOIN_MAINNET_TAPROOT_XPUB"
	envBitcoinTestnet4LegacyXPub            = "BITCOIN_TESTNET4_LEGACY_XPUB"
	envBitcoinTestnet4SegwitXPub            = "BITCOIN_TESTNET4_SEGWIT_XPUB"
	envBitcoinTestnet4NativeSegwitXPub      = "BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB"
	envBitcoinTestnet4TaprootXPub           = "BITCOIN_TESTNET4_TAPROOT_XPUB"
	envBitcoinMainnetRequiredConfirmations  = "BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS"
	envBitcoinTestnet4RequiredConfirmations = "BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS"
	envBitcoinMainnetReceiptExpiresAfter    = "BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER"
	envBitcoinTestnet4ReceiptExpiresAfter   = "BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER"
	defaultBitcoinRequiredConfirmations     = int32(2)
	defaultBitcoinReceiptExpiresAfter       = 24 * time.Hour
)

func buildHTTPHandler(env map[string]string, bridgeID string) (http.Handler, error) {
	clock := system.NewClock()
	healthUseCase := use_cases.NewCheckHealthUseCase(clock)
	healthController := httpcontroller.NewHealthController(healthUseCase)

	requiredConfirmationsByNetwork, err := loadBitcoinRequiredConfirmations(env)
	if err != nil {
		return nil, err
	}
	receiptExpiresAfterByNetwork, err := loadBitcoinReceiptExpiresAfter(env)
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
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			Scheme:               string(value_objects.BitcoinAddressSchemeLegacy),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envValue(env, envBitcoinMainnetLegacyXPub),
			DerivationPathPrefix: "m/44'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-segwit",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			Scheme:               string(value_objects.BitcoinAddressSchemeSegwit),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envValue(env, envBitcoinMainnetSegwitXPub),
			DerivationPathPrefix: "m/49'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-native-segwit",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			Scheme:               string(value_objects.BitcoinAddressSchemeNativeSegwit),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envValue(env, envBitcoinMainnetNativeSegwitXPub),
			DerivationPathPrefix: "m/84'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-mainnet-taproot",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkMainnet),
			Scheme:               string(value_objects.BitcoinAddressSchemeTaproot),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envValue(env, envBitcoinMainnetTaprootXPub),
			DerivationPathPrefix: "m/86'/0'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-legacy",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
			Scheme:               string(value_objects.BitcoinAddressSchemeLegacy),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envValue(env, envBitcoinTestnet4LegacyXPub),
			DerivationPathPrefix: "m/44'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-segwit",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
			Scheme:               string(value_objects.BitcoinAddressSchemeSegwit),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envValue(env, envBitcoinTestnet4SegwitXPub),
			DerivationPathPrefix: "m/49'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-native-segwit",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
			Scheme:               string(value_objects.BitcoinAddressSchemeNativeSegwit),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envValue(env, envBitcoinTestnet4NativeSegwitXPub),
			DerivationPathPrefix: "m/84'/1'/0'",
		},
		{
			AddressPolicyID:      "bitcoin-testnet4-taproot",
			Chain:                value_objects.SupportedChainBitcoin,
			Network:              value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
			Scheme:               string(value_objects.BitcoinAddressSchemeTaproot),
			MinorUnit:            "satoshi",
			Decimals:             8,
			AccountPublicKey:     envValue(env, envBitcoinTestnet4TaprootXPub),
			DerivationPathPrefix: "m/86'/1'/0'",
		},
	})

	listAddressPoliciesUseCase := use_cases.NewListAddressPoliciesUseCase(addressPolicyReader)
	generateAddressUseCase := use_cases.NewGenerateAddressUseCase(chainAddressDeriver, addressPolicyReader)
	bridge := cloudflarepostgres.NewJSBridge()
	dbExecutor := cloudflarepostgres.NewExecutor(bridgeID, bridge)
	unitOfWork := cloudflarepostgres.NewUnitOfWork(bridgeID, bridge)
	allocationIssuancePolicy := policies.NewPaymentAddressAllocationIssuancePolicy(
		requiredConfirmationsByNetwork,
		receiptExpiresAfterByNetwork,
	)
	allocatePaymentAddressUseCase := use_cases.NewAllocatePaymentAddressUseCase(
		unitOfWork,
		chainAddressDeriver,
		addressPolicyReader,
		allocationIssuancePolicy,
		clock,
	)
	getPaymentAddressStatusUseCase := use_cases.NewGetPaymentAddressStatusUseCase(
		cloudflarepostgres.NewPaymentAddressStatusFinder(dbExecutor),
		addressPolicyReader,
	)

	chainAddressController := httpcontroller.NewChainAddressController(
		listAddressPoliciesUseCase,
		generateAddressUseCase,
		allocatePaymentAddressUseCase,
		getPaymentAddressStatusUseCase,
	)

	mux := http.NewServeMux()
	healthController.RegisterRoutes(mux)
	chainAddressController.RegisterRoutes(mux)
	return mux, nil
}

func envValue(env map[string]string, key string) string {
	return strings.TrimSpace(env[key])
}

func loadBitcoinRequiredConfirmations(env map[string]string) (map[value_objects.NetworkID]int32, error) {
	mainnetConfirmations, err := parsePositiveInt32EnvWithDefault(env, envBitcoinMainnetRequiredConfirmations, defaultBitcoinRequiredConfirmations)
	if err != nil {
		return nil, err
	}
	testnet4Confirmations, err := parsePositiveInt32EnvWithDefault(env, envBitcoinTestnet4RequiredConfirmations, defaultBitcoinRequiredConfirmations)
	if err != nil {
		return nil, err
	}

	return map[value_objects.NetworkID]int32{
		value_objects.NetworkID(value_objects.BitcoinNetworkMainnet):  mainnetConfirmations,
		value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4): testnet4Confirmations,
	}, nil
}

func loadBitcoinReceiptExpiresAfter(env map[string]string) (map[value_objects.NetworkID]time.Duration, error) {
	mainnetExpiresAfter, err := parseDurationEnvWithDefault(env, envBitcoinMainnetReceiptExpiresAfter, defaultBitcoinReceiptExpiresAfter)
	if err != nil {
		return nil, err
	}
	testnet4ExpiresAfter, err := parseDurationEnvWithDefault(env, envBitcoinTestnet4ReceiptExpiresAfter, defaultBitcoinReceiptExpiresAfter)
	if err != nil {
		return nil, err
	}

	return map[value_objects.NetworkID]time.Duration{
		value_objects.NetworkID(value_objects.BitcoinNetworkMainnet):  mainnetExpiresAfter,
		value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4): testnet4ExpiresAfter,
	}, nil
}

func parsePositiveInt32EnvWithDefault(env map[string]string, key string, fallback int32) (int32, error) {
	rawValue := strings.TrimSpace(env[key])
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

func parseDurationEnvWithDefault(env map[string]string, key string, fallback time.Duration) (time.Duration, error) {
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
