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
	"payrune/internal/domain/entities"
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
	addressPolicyReader := policyadapter.NewAddressPolicyReader(
		buildAddressIssuancePoliciesFromEnv(ethereumCreate2SaltDeriver),
	)
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
) []entities.AddressIssuancePolicy {
	return buildAddressIssuancePolicies(os.Getenv, ethereumCreate2SaltDeriver)
}

func buildAddressIssuancePolicies(
	envValue func(string) string,
	ethereumCreate2SaltDeriver *ethereum.Create2SaltDeriver,
) []entities.AddressIssuancePolicy {
	return []entities.AddressIssuancePolicy{
		newBitcoinAddressIssuancePolicy(
			"bitcoin-mainnet-legacy",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeLegacy),
			envValue(envBitcoinMainnetLegacyXPub),
			"m/44'/0'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			"bitcoin-mainnet-segwit",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeSegwit),
			envValue(envBitcoinMainnetSegwitXPub),
			"m/49'/0'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			"bitcoin-mainnet-native-segwit",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			envValue(envBitcoinMainnetNativeSegwitXPub),
			"m/84'/0'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			"bitcoin-mainnet-taproot",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeTaproot),
			envValue(envBitcoinMainnetTaprootXPub),
			"m/86'/0'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			"bitcoin-testnet4-legacy",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			string(valueobjects.BitcoinAddressSchemeLegacy),
			envValue(envBitcoinTestnet4LegacyXPub),
			"m/44'/1'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			"bitcoin-testnet4-segwit",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			string(valueobjects.BitcoinAddressSchemeSegwit),
			envValue(envBitcoinTestnet4SegwitXPub),
			"m/49'/1'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			"bitcoin-testnet4-native-segwit",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			envValue(envBitcoinTestnet4NativeSegwitXPub),
			"m/84'/1'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			"bitcoin-testnet4-taproot",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			string(valueobjects.BitcoinAddressSchemeTaproot),
			envValue(envBitcoinTestnet4TaprootXPub),
			"m/86'/1'/0'",
		),
		newEthereumCreate2AddressIssuancePolicy(
			valueobjects.NetworkID("mainnet"),
			envValue(envEthereumMainnetCreate2Collector),
			ethereumCreate2SaltDeriver,
		),
		newEthereumCreate2AddressIssuancePolicy(
			valueobjects.NetworkID("sepolia"),
			envValue(envEthereumSepoliaCreate2Collector),
			ethereumCreate2SaltDeriver,
		),
	}
}

func newBitcoinAddressIssuancePolicy(
	addressPolicyID string,
	network valueobjects.NetworkID,
	scheme string,
	addressSourceRef string,
	addressReferencePrefix string,
) entities.AddressIssuancePolicy {
	return entities.AddressIssuancePolicy{
		AddressPolicy: entities.AddressPolicy{
			AddressPolicyID: addressPolicyID,
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         network,
			Scheme:          scheme,
			MinorUnit:       "satoshi",
			Decimals:        8,
		},
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSourceRef:       addressSourceRef,
			AddressReferencePrefix: addressReferencePrefix,
		},
	}.Normalize()
}

func newEthereumCreate2AddressIssuancePolicy(
	network valueobjects.NetworkID,
	collectorAddress string,
	ethereumCreate2SaltDeriver *ethereum.Create2SaltDeriver,
) entities.AddressIssuancePolicy {
	addressSourceRef := ""
	if ethereumCreate2SaltDeriver != nil && ethereumCreate2SaltDeriver.HasNetwork(network) {
		addressSourceRef = ethereumcreate2assets.BuildAddressSourceRef(string(network), collectorAddress)
	}

	return entities.AddressIssuancePolicy{
		AddressPolicy: entities.AddressPolicy{
			AddressPolicyID: fmt.Sprintf("ethereum-%s-create2", network),
			Chain:           valueobjects.SupportedChainEthereum,
			Network:         network,
			Scheme:          "create2",
			MinorUnit:       "wei",
			Decimals:        18,
		},
		IssuanceConfig: valueobjects.AddressIssuanceConfig{
			AddressSourceRef:       addressSourceRef,
			AddressReferencePrefix: fmt.Sprintf("ethereum-%s-create2", network),
		},
	}.Normalize()
}

func buildEthereumCreate2DerivationKeys(
	mainnetKey string,
	sepoliaKey string,
) map[valueobjects.NetworkID]string {
	return map[valueobjects.NetworkID]string{
		valueobjects.NetworkID("mainnet"): strings.TrimSpace(mainnetKey),
		valueobjects.NetworkID("sepolia"): strings.TrimSpace(sepoliaKey),
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
