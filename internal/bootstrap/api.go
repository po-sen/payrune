package bootstrap

import (
	"context"
	"database/sql"
	"errors"
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
	"payrune/internal/application/usecases"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
	ethereumcreate2assets "payrune/internal/infrastructure/ethereumcreate2assets"
	postgresinfra "payrune/internal/infrastructure/postgres"
)

const (
	envDatabaseURL                              = "DATABASE_URL"
	envBitcoinMainnetRequiredConfirmations      = "BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS"
	envBitcoinTestnet4RequiredConfirmations     = "BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS"
	envBitcoinMainnetReceiptExpiresAfter        = "BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER"
	envBitcoinTestnet4ReceiptExpiresAfter       = "BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER"
	envBitcoinMainnetLegacyXPub                 = "BITCOIN_MAINNET_LEGACY_XPUB"
	envBitcoinMainnetSegwitXPub                 = "BITCOIN_MAINNET_SEGWIT_XPUB"
	envBitcoinMainnetNativeSegwitXPub           = "BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB"
	envBitcoinMainnetTaprootXPub                = "BITCOIN_MAINNET_TAPROOT_XPUB"
	envBitcoinTestnet4LegacyXPub                = "BITCOIN_TESTNET4_LEGACY_XPUB"
	envBitcoinTestnet4SegwitXPub                = "BITCOIN_TESTNET4_SEGWIT_XPUB"
	envBitcoinTestnet4NativeSegwitXPub          = "BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB"
	envBitcoinTestnet4TaprootXPub               = "BITCOIN_TESTNET4_TAPROOT_XPUB"
	envBitcoinMainnetLegacyEnabled              = "BITCOIN_MAINNET_LEGACY_ENABLED"
	envBitcoinMainnetSegwitEnabled              = "BITCOIN_MAINNET_SEGWIT_ENABLED"
	envBitcoinMainnetNativeSegwitEnabled        = "BITCOIN_MAINNET_NATIVE_SEGWIT_ENABLED"
	envBitcoinMainnetTaprootEnabled             = "BITCOIN_MAINNET_TAPROOT_ENABLED"
	envBitcoinTestnet4LegacyEnabled             = "BITCOIN_TESTNET4_LEGACY_ENABLED"
	envBitcoinTestnet4SegwitEnabled             = "BITCOIN_TESTNET4_SEGWIT_ENABLED"
	envBitcoinTestnet4NativeSegwitEnabled       = "BITCOIN_TESTNET4_NATIVE_SEGWIT_ENABLED"
	envBitcoinTestnet4TaprootEnabled            = "BITCOIN_TESTNET4_TAPROOT_ENABLED"
	envEthereumMainnetRequiredConfirmations     = "ETHEREUM_MAINNET_REQUIRED_CONFIRMATIONS"
	envEthereumSepoliaRequiredConfirmations     = "ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS"
	envEthereumMainnetReceiptExpiresAfter       = "ETHEREUM_MAINNET_RECEIPT_EXPIRES_AFTER"
	envEthereumSepoliaReceiptExpiresAfter       = "ETHEREUM_SEPOLIA_RECEIPT_EXPIRES_AFTER"
	envEthereumMainnetCreate2Collector          = "ETHEREUM_MAINNET_CREATE2_COLLECTOR_ADDRESS"
	envEthereumSepoliaCreate2Collector          = "ETHEREUM_SEPOLIA_CREATE2_COLLECTOR_ADDRESS"
	envEthereumMainnetCreate2DerivationKey      = "ETHEREUM_MAINNET_CREATE2_DERIVATION_KEY"
	envEthereumSepoliaCreate2DerivationKey      = "ETHEREUM_SEPOLIA_CREATE2_DERIVATION_KEY"
	envEthereumMainnetUSDTAssetReference        = "ETHEREUM_MAINNET_USDT_ASSET_REFERENCE"
	envEthereumSepoliaUSDTAssetReference        = "ETHEREUM_SEPOLIA_USDT_ASSET_REFERENCE"
	envEthereumMainnetCreate2Enabled            = "ETHEREUM_MAINNET_CREATE2_ENABLED"
	envEthereumMainnetUSDTCreate2Enabled        = "ETHEREUM_MAINNET_USDT_CREATE2_ENABLED"
	envEthereumSepoliaCreate2Enabled            = "ETHEREUM_SEPOLIA_CREATE2_ENABLED"
	envEthereumSepoliaUSDTCreate2Enabled        = "ETHEREUM_SEPOLIA_USDT_CREATE2_ENABLED"
	defaultBitcoinRequiredConfirmations         = int32(1)
	defaultEthereumSepoliaRequiredConfirmations = int32(12)
	defaultBitcoinReceiptExpiresAfter           = 7 * 24 * time.Hour
)

type apiContainer struct {
	APIHandler http.Handler
	closeFn    func() error
}

func RunAPI(ctx context.Context, addr string) error {
	container, err := newAPIContainer()
	if err != nil {
		return err
	}
	defer func() {
		_ = container.Close()
	}()

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           container.APIHandler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	err = httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func newAPIContainer() (*apiContainer, error) {
	db, err := openPostgresFromEnv()
	if err != nil {
		return nil, err
	}

	clock := system.NewClock()
	healthUseCase := usecases.NewCheckHealthUseCase(clock)
	healthController := httpcontroller.NewHealthController(healthUseCase)
	requiredConfirmationsByScope, err := loadReceiptRequiredConfirmationsFromEnv()
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	receiptExpiresAfterByScope, err := loadReceiptExpiresAfterByScopeFromEnv()
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
	bitcoinChainAddressDeriver := bitcoin.NewChainAddressDeriver(bitcoinDeriver)
	ethereumChainAddressDeriver := ethereum.NewChainAddressDeriver()
	ethereumCreate2SaltDeriver := ethereum.NewCreate2SaltDeriver(
		buildEthereumCreate2DerivationKeys(
			os.Getenv(envEthereumMainnetCreate2DerivationKey),
			os.Getenv(envEthereumSepoliaCreate2DerivationKey),
		),
	)
	addressIssuancePolicies, err := buildAddressIssuancePoliciesFromEnv(ethereumCreate2SaltDeriver)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := validateConfiguredAddressIssuancePolicies(addressIssuancePolicies, bitcoinDeriver); err != nil {
		_ = db.Close()
		return nil, err
	}
	readinessChecker, err := ethereum.NewAddressIssuanceReadinessChecker(loadEthereumRPCConfigsFromEnv())
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := validateEnabledEthereumIssuanceReadiness(context.Background(), addressIssuancePolicies, readinessChecker); err != nil {
		_ = db.Close()
		return nil, err
	}
	addressPolicyReader := policyadapter.NewAddressPolicyReader(addressIssuancePolicies)
	listAddressPoliciesUseCase := usecases.NewListAddressPoliciesUseCase(addressPolicyReader)
	unitOfWork := postgresadapter.NewUnitOfWork(db)
	allocationIssuancePolicy := policies.NewPaymentAddressAllocationIssuancePolicy(
		requiredConfirmationsByScope,
		receiptExpiresAfterByScope,
	)
	issuedAddressDeriver, err := blockchain.NewMultiChainIssuedPaymentAddressDeriver(
		bitcoin.NewIssuedPaymentAddressDeriver(bitcoinChainAddressDeriver),
		ethereum.NewIssuedPaymentAddressDeriver(ethereumChainAddressDeriver, ethereumCreate2SaltDeriver),
	)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	allocatePaymentAddressUseCase := usecases.NewAllocatePaymentAddressUseCase(
		unitOfWork,
		issuedAddressDeriver,
		addressPolicyReader,
		allocationIssuancePolicy,
		clock,
	)
	getPaymentAddressStatusUseCase := usecases.NewGetPaymentAddressStatusUseCase(
		postgresadapter.NewPaymentAddressStatusFinder(db),
		addressPolicyReader,
	)
	return &apiContainer{
		APIHandler: httpadapter.NewPublicRouter(httpadapter.RouterControllers{
			Health:                 healthController,
			ListAddressPolicies:    httpcontroller.NewListAddressPoliciesController(listAddressPoliciesUseCase),
			AllocatePaymentAddress: httpcontroller.NewAllocatePaymentAddressController(allocatePaymentAddressUseCase),
			GetPaymentAddressStatus: httpcontroller.NewGetPaymentAddressStatusController(
				getPaymentAddressStatusUseCase,
			),
		}),
		closeFn: db.Close,
	}, nil
}

func (c *apiContainer) Close() error {
	if c.closeFn == nil {
		return nil
	}
	return c.closeFn()
}

func openPostgresFromEnv() (*sql.DB, error) {
	databaseURL := strings.TrimSpace(os.Getenv(envDatabaseURL))
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	return postgresinfra.Open(databaseURL)
}

func loadReceiptRequiredConfirmationsFromEnv() (map[policies.PaymentReceiptTermsScope]int32, error) {
	return loadReceiptRequiredConfirmationsFromLookup(os.Getenv, defaultBitcoinRequiredConfirmations)
}

func loadReceiptExpiresAfterByScopeFromEnv() (map[policies.PaymentReceiptTermsScope]time.Duration, error) {
	return loadReceiptExpiresAfterByScopeFromLookup(os.Getenv, defaultBitcoinReceiptExpiresAfter)
}

func loadReceiptRequiredConfirmationsFromLookup(
	lookup func(string) string,
	fallback int32,
) (map[policies.PaymentReceiptTermsScope]int32, error) {
	mainnetConfirmations, err := parsePositiveInt32LookupWithDefault(
		lookup,
		envBitcoinMainnetRequiredConfirmations,
		fallback,
	)
	if err != nil {
		return nil, err
	}
	testnet4Confirmations, err := parsePositiveInt32LookupWithDefault(
		lookup,
		envBitcoinTestnet4RequiredConfirmations,
		fallback,
	)
	if err != nil {
		return nil, err
	}
	ethereumMainnetConfirmations, err := parsePositiveInt32LookupWithDefault(
		lookup,
		envEthereumMainnetRequiredConfirmations,
		fallback,
	)
	if err != nil {
		return nil, err
	}
	ethereumSepoliaConfirmations, err := parsePositiveInt32LookupWithDefault(
		lookup,
		envEthereumSepoliaRequiredConfirmations,
		defaultEthereumSepoliaRequiredConfirmations,
	)
	if err != nil {
		return nil, err
	}

	return map[policies.PaymentReceiptTermsScope]int32{
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkIDMainnet,
		): mainnetConfirmations,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkIDTestnet4,
		): testnet4Confirmations,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainEthereum,
			valueobjects.NetworkIDMainnet,
		): ethereumMainnetConfirmations,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainEthereum,
			valueobjects.NetworkIDSepolia,
		): ethereumSepoliaConfirmations,
	}, nil
}

func loadReceiptExpiresAfterByScopeFromLookup(
	lookup func(string) string,
	fallback time.Duration,
) (map[policies.PaymentReceiptTermsScope]time.Duration, error) {
	mainnetExpiresAfter, err := parsePositiveDurationLookupWithDefault(
		lookup,
		envBitcoinMainnetReceiptExpiresAfter,
		fallback,
	)
	if err != nil {
		return nil, err
	}
	testnet4ExpiresAfter, err := parsePositiveDurationLookupWithDefault(
		lookup,
		envBitcoinTestnet4ReceiptExpiresAfter,
		fallback,
	)
	if err != nil {
		return nil, err
	}
	ethereumMainnetExpiresAfter, err := parsePositiveDurationLookupWithDefault(
		lookup,
		envEthereumMainnetReceiptExpiresAfter,
		fallback,
	)
	if err != nil {
		return nil, err
	}
	ethereumSepoliaExpiresAfter, err := parsePositiveDurationLookupWithDefault(
		lookup,
		envEthereumSepoliaReceiptExpiresAfter,
		fallback,
	)
	if err != nil {
		return nil, err
	}

	return map[policies.PaymentReceiptTermsScope]time.Duration{
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkIDMainnet,
		): mainnetExpiresAfter,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainBitcoin,
			valueobjects.NetworkIDTestnet4,
		): testnet4ExpiresAfter,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainEthereum,
			valueobjects.NetworkIDMainnet,
		): ethereumMainnetExpiresAfter,
		newPaymentReceiptTermsScope(
			valueobjects.SupportedChainEthereum,
			valueobjects.NetworkIDSepolia,
		): ethereumSepoliaExpiresAfter,
	}, nil
}

func parsePositiveInt32LookupWithDefault(
	lookup func(string) string,
	key string,
	fallback int32,
) (int32, error) {
	raw := strings.TrimSpace(lookup(key))
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

func parsePositiveDurationLookupWithDefault(
	lookup func(string) string,
	key string,
	fallback time.Duration,
) (time.Duration, error) {
	raw := strings.TrimSpace(lookup(key))
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

func buildAddressIssuancePoliciesFromEnv(
	ethereumCreate2SaltDeriver *ethereum.Create2SaltDeriver,
) ([]policies.AddressIssuancePolicy, error) {
	return buildAddressIssuancePolicies(os.Getenv, ethereumCreate2SaltDeriver)
}

func buildAddressIssuancePolicies(
	envValue func(string) string,
	ethereumCreate2SaltDeriver *ethereum.Create2SaltDeriver,
) ([]policies.AddressIssuancePolicy, error) {
	bitcoinMainnetLegacyEnabled, err := parseEnabledPolicyEnv(envValue, envBitcoinMainnetLegacyEnabled)
	if err != nil {
		return nil, err
	}
	bitcoinMainnetSegwitEnabled, err := parseEnabledPolicyEnv(envValue, envBitcoinMainnetSegwitEnabled)
	if err != nil {
		return nil, err
	}
	bitcoinMainnetNativeSegwitEnabled, err := parseEnabledPolicyEnv(envValue, envBitcoinMainnetNativeSegwitEnabled)
	if err != nil {
		return nil, err
	}
	bitcoinMainnetTaprootEnabled, err := parseEnabledPolicyEnv(envValue, envBitcoinMainnetTaprootEnabled)
	if err != nil {
		return nil, err
	}
	bitcoinTestnet4LegacyEnabled, err := parseEnabledPolicyEnv(envValue, envBitcoinTestnet4LegacyEnabled)
	if err != nil {
		return nil, err
	}
	bitcoinTestnet4SegwitEnabled, err := parseEnabledPolicyEnv(envValue, envBitcoinTestnet4SegwitEnabled)
	if err != nil {
		return nil, err
	}
	bitcoinTestnet4NativeSegwitEnabled, err := parseEnabledPolicyEnv(envValue, envBitcoinTestnet4NativeSegwitEnabled)
	if err != nil {
		return nil, err
	}
	bitcoinTestnet4TaprootEnabled, err := parseEnabledPolicyEnv(envValue, envBitcoinTestnet4TaprootEnabled)
	if err != nil {
		return nil, err
	}
	ethereumMainnetCreate2Enabled, err := parseEnabledPolicyEnv(envValue, envEthereumMainnetCreate2Enabled)
	if err != nil {
		return nil, err
	}
	ethereumMainnetUSDTCreate2Enabled, err := parseEnabledPolicyEnv(envValue, envEthereumMainnetUSDTCreate2Enabled)
	if err != nil {
		return nil, err
	}
	ethereumSepoliaCreate2Enabled, err := parseEnabledPolicyEnv(envValue, envEthereumSepoliaCreate2Enabled)
	if err != nil {
		return nil, err
	}
	ethereumSepoliaUSDTCreate2Enabled, err := parseEnabledPolicyEnv(envValue, envEthereumSepoliaUSDTCreate2Enabled)
	if err != nil {
		return nil, err
	}

	return []policies.AddressIssuancePolicy{
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinMainnetLegacy,
			valueobjects.NetworkIDMainnet,
			string(valueobjects.AddressSchemeLegacy),
			bitcoinMainnetLegacyEnabled,
			envValue(envBitcoinMainnetLegacyXPub),
			"m/44'/0'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinMainnetSegwit,
			valueobjects.NetworkIDMainnet,
			string(valueobjects.AddressSchemeSegwit),
			bitcoinMainnetSegwitEnabled,
			envValue(envBitcoinMainnetSegwitXPub),
			"m/49'/0'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinMainnetNativeSegwit,
			valueobjects.NetworkIDMainnet,
			string(valueobjects.AddressSchemeNativeSegwit),
			bitcoinMainnetNativeSegwitEnabled,
			envValue(envBitcoinMainnetNativeSegwitXPub),
			"m/84'/0'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinMainnetTaproot,
			valueobjects.NetworkIDMainnet,
			string(valueobjects.AddressSchemeTaproot),
			bitcoinMainnetTaprootEnabled,
			envValue(envBitcoinMainnetTaprootXPub),
			"m/86'/0'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinTestnet4Legacy,
			valueobjects.NetworkIDTestnet4,
			string(valueobjects.AddressSchemeLegacy),
			bitcoinTestnet4LegacyEnabled,
			envValue(envBitcoinTestnet4LegacyXPub),
			"m/44'/1'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinTestnet4Segwit,
			valueobjects.NetworkIDTestnet4,
			string(valueobjects.AddressSchemeSegwit),
			bitcoinTestnet4SegwitEnabled,
			envValue(envBitcoinTestnet4SegwitXPub),
			"m/49'/1'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinTestnet4NativeSegwit,
			valueobjects.NetworkIDTestnet4,
			string(valueobjects.AddressSchemeNativeSegwit),
			bitcoinTestnet4NativeSegwitEnabled,
			envValue(envBitcoinTestnet4NativeSegwitXPub),
			"m/84'/1'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinTestnet4Taproot,
			valueobjects.NetworkIDTestnet4,
			string(valueobjects.AddressSchemeTaproot),
			bitcoinTestnet4TaprootEnabled,
			envValue(envBitcoinTestnet4TaprootXPub),
			"m/86'/1'/0'",
		),
		newEthereumCreate2AddressIssuancePolicy(
			valueobjects.NetworkIDMainnet,
			ethereumMainnetCreate2Enabled,
			envValue(envEthereumMainnetCreate2Collector),
			ethereumCreate2SaltDeriver,
		),
		newEthereumUSDTCreate2AddressIssuancePolicy(
			valueobjects.NetworkIDMainnet,
			ethereumMainnetUSDTCreate2Enabled,
			envValue(envEthereumMainnetCreate2Collector),
			envValue(envEthereumMainnetUSDTAssetReference),
			ethereumCreate2SaltDeriver,
		),
		newEthereumCreate2AddressIssuancePolicy(
			valueobjects.NetworkIDSepolia,
			ethereumSepoliaCreate2Enabled,
			envValue(envEthereumSepoliaCreate2Collector),
			ethereumCreate2SaltDeriver,
		),
		newEthereumUSDTCreate2AddressIssuancePolicy(
			valueobjects.NetworkIDSepolia,
			ethereumSepoliaUSDTCreate2Enabled,
			envValue(envEthereumSepoliaCreate2Collector),
			envValue(envEthereumSepoliaUSDTAssetReference),
			ethereumCreate2SaltDeriver,
		),
	}, nil
}

func validateConfiguredAddressIssuancePolicies(
	policies []policies.AddressIssuancePolicy,
	bitcoinDeriver *bitcoin.HDXPubAddressDeriver,
) error {
	for _, policy := range policies {
		normalized := policy.Normalize()
		if !normalized.Enabled {
			continue
		}
		switch normalized.Chain {
		case valueobjects.SupportedChainBitcoin:
			if err := validateConfiguredEnabledBitcoinAddressIssuancePolicy(normalized, bitcoinDeriver); err != nil {
				return err
			}
		case valueobjects.SupportedChainEthereum:
			if err := validateConfiguredEnabledEthereumAddressIssuancePolicy(normalized); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateEnabledEthereumIssuanceReadiness(
	ctx context.Context,
	policies []policies.AddressIssuancePolicy,
	checker *ethereum.EthereumAddressIssuanceReadinessChecker,
) error {
	if checker == nil {
		return errors.New("ethereum issuance readiness checker is not configured")
	}

	for _, policy := range policies {
		normalized := policy.Normalize()
		if normalized.Chain != valueobjects.SupportedChainEthereum || !normalized.Enabled {
			continue
		}
		if err := checker.CheckIssuanceReadiness(ctx, normalized); err != nil {
			return err
		}
	}

	return nil
}

func validateConfiguredEnabledBitcoinAddressIssuancePolicy(
	policy policies.AddressIssuancePolicy,
	bitcoinDeriver *bitcoin.HDXPubAddressDeriver,
) error {
	if bitcoinDeriver == nil {
		return fmt.Errorf(
			"bitcoin policy %q cannot be validated: bitcoin address deriver is not configured",
			policy.AddressPolicyID,
		)
	}
	if strings.TrimSpace(policy.IssuanceConfig.AddressSpaceRef) == "" {
		envKey := bitcoinXPubEnvKey(policy.AddressPolicyID)
		if envKey == "" {
			return fmt.Errorf(
				"bitcoin policy %q is enabled but xpub is missing",
				policy.AddressPolicyID,
			)
		}
		return fmt.Errorf(
			"bitcoin policy %q is enabled but %s is missing",
			policy.AddressPolicyID,
			envKey,
		)
	}
	if err := bitcoinDeriver.ValidateXPub(policy.IssuanceConfig.AddressSpaceRef); err != nil {
		envKey := bitcoinXPubEnvKey(policy.AddressPolicyID)
		if envKey == "" {
			return fmt.Errorf(
				"bitcoin policy %q has invalid xpub: %w",
				policy.AddressPolicyID,
				err,
			)
		}
		return fmt.Errorf(
			"bitcoin policy %q has invalid xpub in %s: %w",
			policy.AddressPolicyID,
			envKey,
			err,
		)
	}
	return nil
}

func validateConfiguredEnabledEthereumAddressIssuancePolicy(
	policy policies.AddressIssuancePolicy,
) error {
	if strings.TrimSpace(policy.IssuanceConfig.AddressSpaceRef) == "" {
		collectorEnvKey := ethereumCreate2CollectorEnvKey(policy.AddressPolicyID)
		derivationEnvKey := ethereumCreate2DerivationKeyEnvKey(policy.AddressPolicyID)
		if collectorEnvKey == "" || derivationEnvKey == "" {
			return fmt.Errorf(
				"ethereum policy %q is enabled but create2 static configuration is incomplete",
				policy.AddressPolicyID,
			)
		}
		return fmt.Errorf(
			"ethereum policy %q is enabled but create2 static configuration is incomplete: %s and %s are required",
			policy.AddressPolicyID,
			collectorEnvKey,
			derivationEnvKey,
		)
	}
	assetReference := strings.TrimSpace(policy.AssetReference)
	envKey := ethereumAssetReferenceEnvKey(policy.AddressPolicyID)
	if ethereumPolicyRequiresAssetReference(policy.AddressPolicyID) && assetReference == "" {
		return fmt.Errorf(
			"ethereum policy %q is enabled but %s is missing",
			policy.AddressPolicyID,
			envKey,
		)
	}
	if assetReference == "" {
		return nil
	}
	if _, err := ethereum.NormalizeEthereumAddress(assetReference, "asset reference"); err != nil {
		if envKey == "" {
			return fmt.Errorf(
				"ethereum policy %q has invalid asset reference: %w",
				policy.AddressPolicyID,
				err,
			)
		}
		return fmt.Errorf(
			"ethereum policy %q has invalid asset reference in %s: %w",
			policy.AddressPolicyID,
			envKey,
			err,
		)
	}
	return nil
}

func newBitcoinAddressIssuancePolicy(
	addressPolicyID valueobjects.AddressPolicyID,
	network valueobjects.NetworkID,
	scheme string,
	enabled bool,
	addressSpaceRef string,
	issuanceRefPrefix string,
) policies.AddressIssuancePolicy {
	return policies.AddressIssuancePolicy{
		AddressPolicyID: addressPolicyID,
		Chain:           valueobjects.SupportedChainBitcoin,
		Network:         network,
		Scheme:          valueobjects.AddressScheme(scheme),
		Decimals:        8,
		Enabled:         enabled,
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSpaceRef:   addressSpaceRef,
			IssuanceRefPrefix: issuanceRefPrefix,
		},
	}.Normalize()
}

func newEthereumCreate2AddressIssuancePolicy(
	network valueobjects.NetworkID,
	enabled bool,
	collectorAddress string,
	ethereumCreate2SaltDeriver *ethereum.Create2SaltDeriver,
) policies.AddressIssuancePolicy {
	addressSpaceRef := ""
	if ethereumCreate2SaltDeriver != nil && ethereumCreate2SaltDeriver.HasNetwork(network) {
		addressSpaceRef = ethereumcreate2assets.BuildAddressSpaceRef(string(network), collectorAddress)
	}

	return policies.AddressIssuancePolicy{
		AddressPolicyID: valueobjects.EthereumCreate2AddressPolicyID(network),
		Chain:           valueobjects.SupportedChainEthereum,
		Network:         network,
		Scheme:          valueobjects.AddressSchemeCreate2,
		Decimals:        18,
		Enabled:         enabled,
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSpaceRef: addressSpaceRef,
		},
	}.Normalize()
}

func newEthereumUSDTCreate2AddressIssuancePolicy(
	network valueobjects.NetworkID,
	enabled bool,
	collectorAddress string,
	assetReference string,
	ethereumCreate2SaltDeriver *ethereum.Create2SaltDeriver,
) policies.AddressIssuancePolicy {
	addressSpaceRef := ""
	if ethereumCreate2SaltDeriver != nil && ethereumCreate2SaltDeriver.HasNetwork(network) {
		addressSpaceRef = ethereumcreate2assets.BuildAddressSpaceRef(string(network), collectorAddress)
	}

	return policies.AddressIssuancePolicy{
		AddressPolicyID: valueobjects.EthereumUSDTCreate2AddressPolicyID(network),
		Chain:           valueobjects.SupportedChainEthereum,
		Network:         network,
		Scheme:          valueobjects.AddressSchemeCreate2,
		AssetReference:  strings.TrimSpace(assetReference),
		Decimals:        6,
		Enabled:         enabled,
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSpaceRef: addressSpaceRef,
		},
	}.Normalize()
}

func buildEthereumCreate2DerivationKeys(
	mainnetKey string,
	sepoliaKey string,
) map[valueobjects.NetworkID]string {
	return map[valueobjects.NetworkID]string{
		valueobjects.NetworkIDMainnet: strings.TrimSpace(mainnetKey),
		valueobjects.NetworkIDSepolia: strings.TrimSpace(sepoliaKey),
	}
}

func bitcoinXPubEnvKey(addressPolicyID valueobjects.AddressPolicyID) string {
	switch addressPolicyID.Normalize() {
	case valueobjects.AddressPolicyIDBitcoinMainnetLegacy:
		return envBitcoinMainnetLegacyXPub
	case valueobjects.AddressPolicyIDBitcoinMainnetSegwit:
		return envBitcoinMainnetSegwitXPub
	case valueobjects.AddressPolicyIDBitcoinMainnetNativeSegwit:
		return envBitcoinMainnetNativeSegwitXPub
	case valueobjects.AddressPolicyIDBitcoinMainnetTaproot:
		return envBitcoinMainnetTaprootXPub
	case valueobjects.AddressPolicyIDBitcoinTestnet4Legacy:
		return envBitcoinTestnet4LegacyXPub
	case valueobjects.AddressPolicyIDBitcoinTestnet4Segwit:
		return envBitcoinTestnet4SegwitXPub
	case valueobjects.AddressPolicyIDBitcoinTestnet4NativeSegwit:
		return envBitcoinTestnet4NativeSegwitXPub
	case valueobjects.AddressPolicyIDBitcoinTestnet4Taproot:
		return envBitcoinTestnet4TaprootXPub
	default:
		return ""
	}
}

func parseEnabledPolicyEnv(lookup func(string) string, key string) (bool, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		return false, nil
	}
	enabled, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return enabled, nil
}

func ethereumPolicyRequiresAssetReference(addressPolicyID valueobjects.AddressPolicyID) bool {
	switch addressPolicyID.Normalize() {
	case valueobjects.AddressPolicyIDEthereumMainnetUSDTCreate2,
		valueobjects.AddressPolicyIDEthereumSepoliaUSDTCreate2:
		return true
	default:
		return false
	}
}

func ethereumCreate2CollectorEnvKey(addressPolicyID valueobjects.AddressPolicyID) string {
	switch addressPolicyID.Normalize() {
	case valueobjects.AddressPolicyIDEthereumMainnetCreate2,
		valueobjects.AddressPolicyIDEthereumMainnetUSDTCreate2:
		return envEthereumMainnetCreate2Collector
	case valueobjects.AddressPolicyIDEthereumSepoliaCreate2,
		valueobjects.AddressPolicyIDEthereumSepoliaUSDTCreate2:
		return envEthereumSepoliaCreate2Collector
	default:
		return ""
	}
}

func ethereumCreate2DerivationKeyEnvKey(addressPolicyID valueobjects.AddressPolicyID) string {
	switch addressPolicyID.Normalize() {
	case valueobjects.AddressPolicyIDEthereumMainnetCreate2,
		valueobjects.AddressPolicyIDEthereumMainnetUSDTCreate2:
		return envEthereumMainnetCreate2DerivationKey
	case valueobjects.AddressPolicyIDEthereumSepoliaCreate2,
		valueobjects.AddressPolicyIDEthereumSepoliaUSDTCreate2:
		return envEthereumSepoliaCreate2DerivationKey
	default:
		return ""
	}
}

func ethereumAssetReferenceEnvKey(addressPolicyID valueobjects.AddressPolicyID) string {
	switch addressPolicyID.Normalize() {
	case valueobjects.AddressPolicyIDEthereumMainnetUSDTCreate2:
		return envEthereumMainnetUSDTAssetReference
	case valueobjects.AddressPolicyIDEthereumSepoliaUSDTCreate2:
		return envEthereumSepoliaUSDTAssetReference
	default:
		return ""
	}
}

func newPaymentReceiptTermsScope(
	chain valueobjects.SupportedChain,
	network valueobjects.NetworkID,
) policies.PaymentReceiptTermsScope {
	return policies.PaymentReceiptTermsScope{
		Chain:   chain,
		Network: network,
	}
}
