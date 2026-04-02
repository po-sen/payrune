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
	envDatabaseURL                          = "DATABASE_URL"
	envBitcoinMainnetRequiredConfirmations  = "BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS"
	envBitcoinTestnet4RequiredConfirmations = "BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS"
	envBitcoinMainnetReceiptExpiresAfter    = "BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER"
	envBitcoinTestnet4ReceiptExpiresAfter   = "BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER"
	envBitcoinMainnetLegacyXPub             = "BITCOIN_MAINNET_LEGACY_XPUB"
	envBitcoinMainnetSegwitXPub             = "BITCOIN_MAINNET_SEGWIT_XPUB"
	envBitcoinMainnetNativeSegwitXPub       = "BITCOIN_MAINNET_NATIVE_SEGWIT_XPUB"
	envBitcoinMainnetTaprootXPub            = "BITCOIN_MAINNET_TAPROOT_XPUB"
	envBitcoinTestnet4LegacyXPub            = "BITCOIN_TESTNET4_LEGACY_XPUB"
	envBitcoinTestnet4SegwitXPub            = "BITCOIN_TESTNET4_SEGWIT_XPUB"
	envBitcoinTestnet4NativeSegwitXPub      = "BITCOIN_TESTNET4_NATIVE_SEGWIT_XPUB"
	envBitcoinTestnet4TaprootXPub           = "BITCOIN_TESTNET4_TAPROOT_XPUB"
	envEthereumMainnetRequiredConfirmations = "ETHEREUM_MAINNET_REQUIRED_CONFIRMATIONS"
	envEthereumSepoliaRequiredConfirmations = "ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS"
	envEthereumMainnetReceiptExpiresAfter   = "ETHEREUM_MAINNET_RECEIPT_EXPIRES_AFTER"
	envEthereumSepoliaReceiptExpiresAfter   = "ETHEREUM_SEPOLIA_RECEIPT_EXPIRES_AFTER"
	envEthereumMainnetCreate2Collector      = "ETHEREUM_MAINNET_CREATE2_COLLECTOR_ADDRESS"
	envEthereumSepoliaCreate2Collector      = "ETHEREUM_SEPOLIA_CREATE2_COLLECTOR_ADDRESS"
	envEthereumMainnetCreate2DerivationKey  = "ETHEREUM_MAINNET_CREATE2_DERIVATION_KEY"
	envEthereumSepoliaCreate2DerivationKey  = "ETHEREUM_SEPOLIA_CREATE2_DERIVATION_KEY"
	defaultBitcoinRequiredConfirmations     = int32(1)
	defaultBitcoinReceiptExpiresAfter       = 7 * 24 * time.Hour
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
	chainAddressDeriver, err := blockchain.NewMultiChainAddressDeriver(
		bitcoinChainAddressDeriver,
		ethereumChainAddressDeriver,
	)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	ethereumCreate2SaltDeriver := ethereum.NewCreate2SaltDeriver(
		buildEthereumCreate2DerivationKeys(
			os.Getenv(envEthereumMainnetCreate2DerivationKey),
			os.Getenv(envEthereumSepoliaCreate2DerivationKey),
		),
	)
	addressIssuancePolicies := buildAddressIssuancePoliciesFromEnv(ethereumCreate2SaltDeriver)
	if err := validateConfiguredAddressIssuancePolicies(addressIssuancePolicies, bitcoinDeriver); err != nil {
		_ = db.Close()
		return nil, err
	}
	addressPolicyReader := policyadapter.NewAddressPolicyReader(addressIssuancePolicies)
	listAddressPoliciesUseCase := usecases.NewListAddressPoliciesUseCase(addressPolicyReader)
	generateAddressUseCase := usecases.NewGenerateAddressUseCase(chainAddressDeriver, addressPolicyReader)
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
			GenerateAddress:        httpcontroller.NewGenerateAddressController(generateAddressUseCase),
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
		fallback,
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
) []policies.AddressIssuancePolicy {
	return buildAddressIssuancePolicies(os.Getenv, ethereumCreate2SaltDeriver)
}

func buildAddressIssuancePolicies(
	envValue func(string) string,
	ethereumCreate2SaltDeriver *ethereum.Create2SaltDeriver,
) []policies.AddressIssuancePolicy {
	return []policies.AddressIssuancePolicy{
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinMainnetLegacy,
			valueobjects.NetworkIDMainnet,
			string(valueobjects.AddressSchemeLegacy),
			envValue(envBitcoinMainnetLegacyXPub),
			"m/44'/0'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinMainnetSegwit,
			valueobjects.NetworkIDMainnet,
			string(valueobjects.AddressSchemeSegwit),
			envValue(envBitcoinMainnetSegwitXPub),
			"m/49'/0'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinMainnetNativeSegwit,
			valueobjects.NetworkIDMainnet,
			string(valueobjects.AddressSchemeNativeSegwit),
			envValue(envBitcoinMainnetNativeSegwitXPub),
			"m/84'/0'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinMainnetTaproot,
			valueobjects.NetworkIDMainnet,
			string(valueobjects.AddressSchemeTaproot),
			envValue(envBitcoinMainnetTaprootXPub),
			"m/86'/0'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinTestnet4Legacy,
			valueobjects.NetworkIDTestnet4,
			string(valueobjects.AddressSchemeLegacy),
			envValue(envBitcoinTestnet4LegacyXPub),
			"m/44'/1'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinTestnet4Segwit,
			valueobjects.NetworkIDTestnet4,
			string(valueobjects.AddressSchemeSegwit),
			envValue(envBitcoinTestnet4SegwitXPub),
			"m/49'/1'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinTestnet4NativeSegwit,
			valueobjects.NetworkIDTestnet4,
			string(valueobjects.AddressSchemeNativeSegwit),
			envValue(envBitcoinTestnet4NativeSegwitXPub),
			"m/84'/1'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinTestnet4Taproot,
			valueobjects.NetworkIDTestnet4,
			string(valueobjects.AddressSchemeTaproot),
			envValue(envBitcoinTestnet4TaprootXPub),
			"m/86'/1'/0'",
		),
		newEthereumCreate2AddressIssuancePolicy(
			valueobjects.NetworkIDMainnet,
			envValue(envEthereumMainnetCreate2Collector),
			ethereumCreate2SaltDeriver,
		),
		newEthereumCreate2AddressIssuancePolicy(
			valueobjects.NetworkIDSepolia,
			envValue(envEthereumSepoliaCreate2Collector),
			ethereumCreate2SaltDeriver,
		),
	}
}

func validateConfiguredAddressIssuancePolicies(
	policies []policies.AddressIssuancePolicy,
	bitcoinDeriver *bitcoin.HDXPubAddressDeriver,
) error {
	for _, policy := range policies {
		normalized := policy.Normalize()
		if !normalized.IsEnabled() || normalized.Chain != valueobjects.SupportedChainBitcoin {
			continue
		}
		if bitcoinDeriver == nil {
			return fmt.Errorf(
				"bitcoin policy %q cannot be validated: bitcoin address deriver is not configured",
				normalized.AddressPolicyID,
			)
		}
		if err := bitcoinDeriver.ValidateXPub(normalized.IssuanceConfig.AddressSpaceRef); err != nil {
			envKey := bitcoinXPubEnvKey(normalized.AddressPolicyID)
			if envKey == "" {
				return fmt.Errorf(
					"bitcoin policy %q has invalid xpub: %w",
					normalized.AddressPolicyID,
					err,
				)
			}
			return fmt.Errorf(
				"bitcoin policy %q has invalid xpub in %s: %w",
				normalized.AddressPolicyID,
				envKey,
				err,
			)
		}
	}

	return nil
}

func newBitcoinAddressIssuancePolicy(
	addressPolicyID valueobjects.AddressPolicyID,
	network valueobjects.NetworkID,
	scheme string,
	addressSpaceRef string,
	issuanceRefPrefix string,
) policies.AddressIssuancePolicy {
	return policies.AddressIssuancePolicy{
		AddressPolicyID: addressPolicyID,
		Chain:           valueobjects.SupportedChainBitcoin,
		Network:         network,
		Scheme:          valueobjects.AddressScheme(scheme),
		MinorUnit:       "satoshi",
		Decimals:        8,
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSpaceRef:   addressSpaceRef,
			IssuanceRefPrefix: issuanceRefPrefix,
		},
	}.Normalize()
}

func newEthereumCreate2AddressIssuancePolicy(
	network valueobjects.NetworkID,
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
		MinorUnit:       "wei",
		Decimals:        18,
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

func newPaymentReceiptTermsScope(
	chain valueobjects.SupportedChain,
	network valueobjects.NetworkID,
) policies.PaymentReceiptTermsScope {
	return policies.PaymentReceiptTermsScope{
		Chain:   chain,
		Network: network,
	}
}
