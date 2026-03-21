package di

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/sha3"

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
	postgresdriver "payrune/internal/infrastructure/drivers/postgres"
)

const (
	envBitcoinMainnetRequiredConfirmations  = "BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS"
	envBitcoinTestnet4RequiredConfirmations = "BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS"
	envBitcoinMainnetReceiptExpiresAfter    = "BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER"
	envBitcoinTestnet4ReceiptExpiresAfter   = "BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER"
	envEthereumMainnetRequiredConfirmations = "ETHEREUM_MAINNET_REQUIRED_CONFIRMATIONS"
	envEthereumSepoliaRequiredConfirmations = "ETHEREUM_SEPOLIA_REQUIRED_CONFIRMATIONS"
	envEthereumMainnetReceiptExpiresAfter   = "ETHEREUM_MAINNET_RECEIPT_EXPIRES_AFTER"
	envEthereumSepoliaReceiptExpiresAfter   = "ETHEREUM_SEPOLIA_RECEIPT_EXPIRES_AFTER"
	envEthereumMainnetCreate2Collector      = "ETHEREUM_MAINNET_CREATE2_COLLECTOR_ADDRESS"
	envEthereumSepoliaCreate2Collector      = "ETHEREUM_SEPOLIA_CREATE2_COLLECTOR_ADDRESS"
	defaultBitcoinRequiredConfirmations     = int32(1)
	defaultBitcoinReceiptExpiresAfter       = 7 * 24 * time.Hour
	ethereumCreate2FixtureReceiverInitCode  = "0x60006000556001600055"
)

type Container struct {
	APIHandler http.Handler
	closeFn    func() error
}

type ethereumCreate2DeploymentMetadata struct {
	FactoryAddress string
	Receiver       ethereumCreate2ReceiverArtifact
}

type ethereumCreate2ReceiverArtifact struct {
	InitCodeHex string
}

var ethereumCreate2DeploymentMetadataByNetwork = map[valueobjects.NetworkID]ethereumCreate2DeploymentMetadata{
	// T-003 will replace these deterministic fixtures with checked-in deployment
	// metadata and receiver artifacts from the actual CREATE2 contracts.
	valueobjects.NetworkID("mainnet"): {
		FactoryAddress: "0x1111111111111111111111111111111111111111",
		Receiver: ethereumCreate2ReceiverArtifact{
			InitCodeHex: ethereumCreate2FixtureReceiverInitCode,
		},
	},
	valueobjects.NetworkID("sepolia"): {
		FactoryAddress: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Receiver: ethereumCreate2ReceiverArtifact{
			InitCodeHex: ethereumCreate2FixtureReceiverInitCode,
		},
	},
}

func NewContainer() (*Container, error) {
	db, err := postgresdriver.OpenFromEnv()
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
	chainAddressDeriver, err := blockchain.NewMultiChainAddressDeriver(
		bitcoin.NewChainAddressDeriver(bitcoinDeriver),
		ethereum.NewChainAddressDeriver(),
	)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	addressPolicyReader := policyadapter.NewAddressPolicyReader(buildAddressPolicyConfigsFromEnv())
	listAddressPoliciesUseCase := usecases.NewListAddressPoliciesUseCase(addressPolicyReader)
	generateAddressUseCase := usecases.NewGenerateAddressUseCase(chainAddressDeriver, addressPolicyReader)
	unitOfWork := postgresadapter.NewUnitOfWork(db)
	allocationIssuancePolicy := policies.NewPaymentAddressAllocationIssuancePolicy(
		requiredConfirmationsByScope,
		receiptExpiresAfterByScope,
	)
	allocatePaymentAddressUseCase := usecases.NewAllocatePaymentAddressUseCase(
		unitOfWork,
		chainAddressDeriver,
		addressPolicyReader,
		allocationIssuancePolicy,
		clock,
	)
	getPaymentAddressStatusUseCase := usecases.NewGetPaymentAddressStatusUseCase(
		postgresadapter.NewPaymentAddressStatusFinder(db),
		addressPolicyReader,
	)
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

func loadReceiptRequiredConfirmationsFromEnv() (map[policies.PaymentReceiptTermsScope]int32, error) {
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
	ethereumMainnetConfirmations, err := parsePositiveInt32EnvWithDefault(
		envEthereumMainnetRequiredConfirmations,
		defaultBitcoinRequiredConfirmations,
	)
	if err != nil {
		return nil, err
	}
	ethereumSepoliaConfirmations, err := parsePositiveInt32EnvWithDefault(
		envEthereumSepoliaRequiredConfirmations,
		defaultBitcoinRequiredConfirmations,
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

func loadReceiptExpiresAfterByScopeFromEnv() (map[policies.PaymentReceiptTermsScope]time.Duration, error) {
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
	ethereumMainnetExpiresAfter, err := parsePositiveDurationEnvWithDefault(
		envEthereumMainnetReceiptExpiresAfter,
		defaultBitcoinReceiptExpiresAfter,
	)
	if err != nil {
		return nil, err
	}
	ethereumSepoliaExpiresAfter, err := parsePositiveDurationEnvWithDefault(
		envEthereumSepoliaReceiptExpiresAfter,
		defaultBitcoinReceiptExpiresAfter,
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

func buildAddressPolicyConfigsFromEnv() []policyadapter.AddressPolicyConfig {
	return []policyadapter.AddressPolicyConfig{
		newBitcoinAddressPolicyConfig(
			"bitcoin-mainnet-legacy",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeLegacy),
			os.Getenv(envBitcoinMainnetLegacyXPub),
			"m/44'/0'/0'",
		),
		newBitcoinAddressPolicyConfig(
			"bitcoin-mainnet-segwit",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeSegwit),
			os.Getenv(envBitcoinMainnetSegwitXPub),
			"m/49'/0'/0'",
		),
		newBitcoinAddressPolicyConfig(
			"bitcoin-mainnet-native-segwit",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			os.Getenv(envBitcoinMainnetNativeSegwitXPub),
			"m/84'/0'/0'",
		),
		newBitcoinAddressPolicyConfig(
			"bitcoin-mainnet-taproot",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
			string(valueobjects.BitcoinAddressSchemeTaproot),
			os.Getenv(envBitcoinMainnetTaprootXPub),
			"m/86'/0'/0'",
		),
		newBitcoinAddressPolicyConfig(
			"bitcoin-testnet4-legacy",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			string(valueobjects.BitcoinAddressSchemeLegacy),
			os.Getenv(envBitcoinTestnet4LegacyXPub),
			"m/44'/1'/0'",
		),
		newBitcoinAddressPolicyConfig(
			"bitcoin-testnet4-segwit",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			string(valueobjects.BitcoinAddressSchemeSegwit),
			os.Getenv(envBitcoinTestnet4SegwitXPub),
			"m/49'/1'/0'",
		),
		newBitcoinAddressPolicyConfig(
			"bitcoin-testnet4-native-segwit",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			os.Getenv(envBitcoinTestnet4NativeSegwitXPub),
			"m/84'/1'/0'",
		),
		newBitcoinAddressPolicyConfig(
			"bitcoin-testnet4-taproot",
			valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
			string(valueobjects.BitcoinAddressSchemeTaproot),
			os.Getenv(envBitcoinTestnet4TaprootXPub),
			"m/86'/1'/0'",
		),
		newEthereumCreate2PolicyConfig(
			valueobjects.NetworkID("mainnet"),
			os.Getenv(envEthereumMainnetCreate2Collector),
		),
		newEthereumCreate2PolicyConfig(
			valueobjects.NetworkID("sepolia"),
			os.Getenv(envEthereumSepoliaCreate2Collector),
		),
	}
}

func newBitcoinAddressPolicyConfig(
	addressPolicyID string,
	network valueobjects.NetworkID,
	scheme string,
	addressSourceRef string,
	addressReferencePrefix string,
) policyadapter.AddressPolicyConfig {
	return policyadapter.AddressPolicyConfig{
		AddressPolicyID:        addressPolicyID,
		Chain:                  valueobjects.SupportedChainBitcoin,
		Network:                network,
		Scheme:                 scheme,
		MinorUnit:              "satoshi",
		Decimals:               8,
		AddressSourceRef:       addressSourceRef,
		AddressReferencePrefix: addressReferencePrefix,
	}
}

func newEthereumCreate2PolicyConfig(
	network valueobjects.NetworkID,
	collectorAddress string,
) policyadapter.AddressPolicyConfig {
	return policyadapter.AddressPolicyConfig{
		AddressPolicyID:        fmt.Sprintf("ethereum-%s-create2", network),
		Chain:                  valueobjects.SupportedChainEthereum,
		Network:                network,
		Scheme:                 "create2",
		MinorUnit:              "wei",
		Decimals:               18,
		AddressSourceRef:       buildEthereumCreate2AddressSourceRef(network, collectorAddress),
		AddressReferencePrefix: fmt.Sprintf("ethereum-%s-create2", network),
	}
}

func buildEthereumCreate2AddressSourceRef(
	network valueobjects.NetworkID,
	collectorAddress string,
) string {
	metadata, ok := ethereumCreate2DeploymentMetadataByNetwork[network]
	if !ok {
		return ""
	}
	return buildEthereumCreate2AddressSourceRefFromMetadata(metadata, collectorAddress)
}

func buildEthereumCreate2AddressSourceRefFromMetadata(
	metadata ethereumCreate2DeploymentMetadata,
	collectorAddress string,
) string {
	collectorAddress = strings.TrimSpace(collectorAddress)
	if collectorAddress == "" {
		return ""
	}

	initCodeHash, ok := metadata.Receiver.InitCodeHashHex()
	if !ok {
		return ""
	}

	// Factory metadata and collector define the issuance address space here.
	// The operator signer that will later submit deploy/sweep transactions is
	// a separate runtime concern and must not affect prediction inputs.
	sourceRef, err := ethereum.BuildCreate2AddressSourceRef(
		strings.TrimSpace(metadata.FactoryAddress),
		collectorAddress,
		initCodeHash,
	)
	if err != nil {
		return ""
	}
	return sourceRef
}

func (a ethereumCreate2ReceiverArtifact) InitCodeHashHex() (string, bool) {
	initCodeHex := strings.TrimSpace(a.InitCodeHex)
	if initCodeHex == "" {
		return "", false
	}
	if !strings.HasPrefix(initCodeHex, "0x") && !strings.HasPrefix(initCodeHex, "0X") {
		return "", false
	}

	initCode, err := hex.DecodeString(initCodeHex[2:])
	if err != nil || len(initCode) == 0 {
		return "", false
	}

	hasher := sha3.NewLegacyKeccak256()
	_, _ = hasher.Write(initCode)
	return "0x" + hex.EncodeToString(hasher.Sum(nil)), true
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
