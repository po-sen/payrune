package bootstrap

import (
	"strings"
	"testing"
	"time"

	"payrune/internal/adapters/outbound/bitcoin"
	"payrune/internal/adapters/outbound/ethereum"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
	ethereumcreate2assets "payrune/internal/infrastructure/ethereumcreate2assets"
)

func TestOpenPostgresFromEnvMissingDatabaseURL(t *testing.T) {
	t.Setenv(envDatabaseURL, " ")

	_, err := openPostgresFromEnv()
	if err == nil {
		t.Fatal("expected missing DATABASE_URL error")
	}
	if got := err.Error(); got != "DATABASE_URL is required" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestLoadReceiptRequiredConfirmationsFromEnvDefaults(t *testing.T) {
	t.Setenv(envBitcoinMainnetRequiredConfirmations, "")
	t.Setenv(envBitcoinTestnet4RequiredConfirmations, "")
	t.Setenv(envEthereumMainnetRequiredConfirmations, "")
	t.Setenv(envEthereumSepoliaRequiredConfirmations, "")

	config, err := loadReceiptRequiredConfirmationsFromEnv()
	if err != nil {
		t.Fatalf("loadReceiptRequiredConfirmationsFromEnv returned error: %v", err)
	}

	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkIDMainnet,
	}]; got != 1 {
		t.Fatalf("unexpected mainnet confirmations: got %d", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkIDTestnet4,
	}]; got != 1 {
		t.Fatalf("unexpected testnet4 confirmations: got %d", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkIDMainnet,
	}]; got != 1 {
		t.Fatalf("unexpected ethereum mainnet confirmations: got %d", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkIDSepolia,
	}]; got != 1 {
		t.Fatalf("unexpected ethereum sepolia confirmations: got %d", got)
	}
}

func TestLoadReceiptRequiredConfirmationsFromEnvCustom(t *testing.T) {
	t.Setenv(envBitcoinMainnetRequiredConfirmations, "6")
	t.Setenv(envBitcoinTestnet4RequiredConfirmations, "2")
	t.Setenv(envEthereumMainnetRequiredConfirmations, "12")
	t.Setenv(envEthereumSepoliaRequiredConfirmations, "4")

	config, err := loadReceiptRequiredConfirmationsFromEnv()
	if err != nil {
		t.Fatalf("loadReceiptRequiredConfirmationsFromEnv returned error: %v", err)
	}

	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkIDMainnet,
	}]; got != 6 {
		t.Fatalf("unexpected mainnet confirmations: got %d", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkIDTestnet4,
	}]; got != 2 {
		t.Fatalf("unexpected testnet4 confirmations: got %d", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkIDMainnet,
	}]; got != 12 {
		t.Fatalf("unexpected ethereum mainnet confirmations: got %d", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkIDSepolia,
	}]; got != 4 {
		t.Fatalf("unexpected ethereum sepolia confirmations: got %d", got)
	}
}

func TestLoadReceiptRequiredConfirmationsFromEnvInvalid(t *testing.T) {
	t.Setenv(envBitcoinMainnetRequiredConfirmations, "abc")
	t.Setenv(envBitcoinTestnet4RequiredConfirmations, "1")
	t.Setenv(envEthereumMainnetRequiredConfirmations, "12")
	t.Setenv(envEthereumSepoliaRequiredConfirmations, "4")

	_, err := loadReceiptRequiredConfirmationsFromEnv()
	if err == nil {
		t.Fatal("expected parse error for mainnet confirmations")
	}
}

func TestLoadReceiptRequiredConfirmationsFromEnvNonPositive(t *testing.T) {
	t.Setenv(envBitcoinMainnetRequiredConfirmations, "0")
	t.Setenv(envBitcoinTestnet4RequiredConfirmations, "1")
	t.Setenv(envEthereumMainnetRequiredConfirmations, "12")
	t.Setenv(envEthereumSepoliaRequiredConfirmations, "4")

	_, err := loadReceiptRequiredConfirmationsFromEnv()
	if err == nil {
		t.Fatal("expected validation error for non-positive confirmations")
	}
}

func TestLoadReceiptExpiresAfterByScopeFromEnvDefaults(t *testing.T) {
	t.Setenv(envBitcoinMainnetReceiptExpiresAfter, "")
	t.Setenv(envBitcoinTestnet4ReceiptExpiresAfter, "")
	t.Setenv(envEthereumMainnetReceiptExpiresAfter, "")
	t.Setenv(envEthereumSepoliaReceiptExpiresAfter, "")

	config, err := loadReceiptExpiresAfterByScopeFromEnv()
	if err != nil {
		t.Fatalf("loadReceiptExpiresAfterByScopeFromEnv returned error: %v", err)
	}

	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkIDMainnet,
	}]; got != defaultBitcoinReceiptExpiresAfter {
		t.Fatalf("unexpected mainnet receipt expires after: got %s", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkIDTestnet4,
	}]; got != defaultBitcoinReceiptExpiresAfter {
		t.Fatalf("unexpected testnet4 receipt expires after: got %s", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkIDMainnet,
	}]; got != defaultBitcoinReceiptExpiresAfter {
		t.Fatalf("unexpected ethereum mainnet receipt expires after: got %s", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkIDSepolia,
	}]; got != defaultBitcoinReceiptExpiresAfter {
		t.Fatalf("unexpected ethereum sepolia receipt expires after: got %s", got)
	}
}

func TestLoadReceiptExpiresAfterByScopeFromEnvCustom(t *testing.T) {
	t.Setenv(envBitcoinMainnetReceiptExpiresAfter, "240h")
	t.Setenv(envBitcoinTestnet4ReceiptExpiresAfter, "36h")
	t.Setenv(envEthereumMainnetReceiptExpiresAfter, "72h")
	t.Setenv(envEthereumSepoliaReceiptExpiresAfter, "12h")

	config, err := loadReceiptExpiresAfterByScopeFromEnv()
	if err != nil {
		t.Fatalf("loadReceiptExpiresAfterByScopeFromEnv returned error: %v", err)
	}

	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkIDMainnet,
	}]; got != 240*time.Hour {
		t.Fatalf("unexpected mainnet receipt expires after: got %s", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkIDTestnet4,
	}]; got != 36*time.Hour {
		t.Fatalf("unexpected testnet4 receipt expires after: got %s", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkIDMainnet,
	}]; got != 72*time.Hour {
		t.Fatalf("unexpected ethereum mainnet receipt expires after: got %s", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkIDSepolia,
	}]; got != 12*time.Hour {
		t.Fatalf("unexpected ethereum sepolia receipt expires after: got %s", got)
	}
}

func TestLoadReceiptExpiresAfterByScopeFromEnvInvalid(t *testing.T) {
	t.Setenv(envBitcoinMainnetReceiptExpiresAfter, "abc")
	t.Setenv(envBitcoinTestnet4ReceiptExpiresAfter, "36h")
	t.Setenv(envEthereumMainnetReceiptExpiresAfter, "72h")
	t.Setenv(envEthereumSepoliaReceiptExpiresAfter, "12h")

	_, err := loadReceiptExpiresAfterByScopeFromEnv()
	if err == nil {
		t.Fatal("expected parse error for mainnet receipt expires after")
	}
}

func TestLoadReceiptExpiresAfterByScopeFromEnvNonPositive(t *testing.T) {
	t.Setenv(envBitcoinMainnetReceiptExpiresAfter, "0s")
	t.Setenv(envBitcoinTestnet4ReceiptExpiresAfter, "36h")
	t.Setenv(envEthereumMainnetReceiptExpiresAfter, "72h")
	t.Setenv(envEthereumSepoliaReceiptExpiresAfter, "12h")

	_, err := loadReceiptExpiresAfterByScopeFromEnv()
	if err == nil {
		t.Fatal("expected validation error for non-positive receipt expires after")
	}
}

func TestNewEthereumCreate2AddressIssuancePolicyBuildsSourceRefFromFixtureMetadata(t *testing.T) {
	network := valueobjects.NetworkIDSepolia
	collectorAddress := "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	saltDeriver := ethereum.NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		network: "0x1111111111111111111111111111111111111111111111111111111111111111",
	})
	metadata, ok := ethereumcreate2assets.LookupDeploymentMetadata(string(network))
	if !ok {
		t.Fatalf("expected embedded metadata for %s", network)
	}

	initCodeHash, ok := metadata.Receiver.InitCodeHashHex(collectorAddress)
	if !ok {
		t.Fatal("expected init code hash available")
	}
	expectedSourceRef, err := ethereum.BuildCreate2AddressSpaceRef(
		metadata.FactoryAddress,
		collectorAddress,
		initCodeHash,
	)
	if err != nil {
		t.Fatalf("BuildCreate2AddressSpaceRef returned error: %v", err)
	}

	policy := newEthereumCreate2AddressIssuancePolicy(network, collectorAddress, saltDeriver)

	if policy.AddressPolicyID != valueobjects.AddressPolicyIDEthereumSepoliaCreate2 {
		t.Fatalf("unexpected address policy id: got %q", policy.AddressPolicyID)
	}
	if policy.IssuanceConfig.IssuanceRefPrefix != "" {
		t.Fatalf("unexpected address reference prefix: got %q", policy.IssuanceConfig.IssuanceRefPrefix)
	}
	if policy.IssuanceConfig.AddressSpaceRef != expectedSourceRef {
		t.Fatalf("unexpected address source ref: got %q want %q", policy.IssuanceConfig.AddressSpaceRef, expectedSourceRef)
	}
}

func TestNewEthereumCreate2AddressIssuancePolicyRequiresCollectorAddress(t *testing.T) {
	saltDeriver := ethereum.NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		valueobjects.NetworkIDMainnet: "0x1111111111111111111111111111111111111111111111111111111111111111",
	})
	policy := newEthereumCreate2AddressIssuancePolicy(valueobjects.NetworkIDMainnet, "", saltDeriver)
	if policy.IssuanceConfig.AddressSpaceRef != "" {
		t.Fatalf("expected disabled policy when collector is missing, got %q", policy.IssuanceConfig.AddressSpaceRef)
	}
}

func TestNewEthereumCreate2AddressIssuancePolicyRequiresSaltSecret(t *testing.T) {
	policy := newEthereumCreate2AddressIssuancePolicy(
		valueobjects.NetworkIDMainnet,
		"0x2222222222222222222222222222222222222222",
		ethereum.NewCreate2SaltDeriver(nil),
	)
	if policy.IssuanceConfig.AddressSpaceRef != "" {
		t.Fatalf("expected disabled policy when derivation key is missing, got %q", policy.IssuanceConfig.AddressSpaceRef)
	}
}

func TestBuildAddressIssuancePoliciesUsesProvidedEnvLookup(t *testing.T) {
	saltDeriver := ethereum.NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		valueobjects.NetworkIDMainnet: "0x1111111111111111111111111111111111111111111111111111111111111111",
	})
	env := map[string]string{
		envBitcoinMainnetLegacyXPub:        " xpub-mainnet-legacy ",
		envEthereumMainnetCreate2Collector: "0x2222222222222222222222222222222222222222",
	}

	policies := buildAddressIssuancePolicies(func(key string) string {
		return env[key]
	}, saltDeriver)

	bitcoinPolicy := findAddressIssuancePolicyByID(policies, valueobjects.AddressPolicyIDBitcoinMainnetLegacy)
	if bitcoinPolicy.IssuanceConfig.AddressSpaceRef != "xpub-mainnet-legacy" {
		t.Fatalf("unexpected bitcoin address source ref: got %q", bitcoinPolicy.IssuanceConfig.AddressSpaceRef)
	}
	if bitcoinPolicy.IssuanceConfig.IssuanceRefPrefix != "m/44'/0'/0'" {
		t.Fatalf(
			"unexpected bitcoin address reference prefix: got %q",
			bitcoinPolicy.IssuanceConfig.IssuanceRefPrefix,
		)
	}

	ethereumPolicy := findAddressIssuancePolicyByID(policies, valueobjects.AddressPolicyIDEthereumMainnetCreate2)
	if ethereumPolicy.Chain != valueobjects.SupportedChainEthereum {
		t.Fatalf("unexpected ethereum policy chain: got %q", ethereumPolicy.Chain)
	}
	if ethereumPolicy.IssuanceConfig.IssuanceRefPrefix != "" {
		t.Fatalf(
			"unexpected ethereum address reference prefix: got %q",
			ethereumPolicy.IssuanceConfig.IssuanceRefPrefix,
		)
	}
	if ethereumPolicy.IssuanceConfig.AddressSpaceRef == "" {
		t.Fatal("expected ethereum create2 source ref to be populated")
	}
}

func TestValidateConfiguredAddressIssuancePoliciesRejectsInvalidBitcoinXPub(t *testing.T) {
	policies := []policies.AddressIssuancePolicy{
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinTestnet4NativeSegwit,
			valueobjects.NetworkIDTestnet4,
			string(valueobjects.AddressSchemeNativeSegwit),
			"tpubDCiB1iLoNxaaj4MTSk2DoTuwUpEfgm4E3vAcTnvG64rR1smhcEsoTeqNCB4af1XHGspgNfWBA3ccpXiwX5JtxwZMTFct6DQWzrKundqdwEa",
			"m/84'/1'/0'",
		),
	}

	err := validateConfiguredAddressIssuancePolicies(policies, newBootstrapBitcoinDeriver())
	if err == nil {
		t.Fatal("expected invalid xpub validation error")
	}
	if !strings.Contains(err.Error(), string(valueobjects.AddressPolicyIDBitcoinTestnet4NativeSegwit)) {
		t.Fatalf("expected policy id in error, got %q", err)
	}
	if !strings.Contains(err.Error(), envBitcoinTestnet4NativeSegwitXPub) {
		t.Fatalf("expected env key in error, got %q", err)
	}
	if !strings.Contains(err.Error(), "bad extended key checksum") {
		t.Fatalf("expected checksum parse error, got %q", err)
	}
}

func TestValidateConfiguredAddressIssuancePoliciesAcceptsValidBitcoinPolicies(t *testing.T) {
	policies := []policies.AddressIssuancePolicy{
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinTestnet4Legacy,
			valueobjects.NetworkIDTestnet4,
			string(valueobjects.AddressSchemeLegacy),
			"tpubDDoLYVq7AUqYP63QvYZxnxk1pCJnWDWdzu9w3BYTP9dJAX47xknZiEKUheaAahn6zBNT5ndCzY2x6MQ8iVj7QpFwuhm5bDF6Ggt3q1Rn2Qs",
			"m/44'/1'/0'",
		),
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinTestnet4NativeSegwit,
			valueobjects.NetworkIDTestnet4,
			string(valueobjects.AddressSchemeNativeSegwit),
			"tpubDCiB1iLoNxaaj4MTSk2DoTuwUpEfgm4E3vAcTnvG64rR1smhcEsoTeqNCB4af1XHGspgNfWBA3ccpXiwX5JtxwZMTFct6DQWzrKundqdwEq",
			"m/84'/1'/0'",
		),
	}

	if err := validateConfiguredAddressIssuancePolicies(policies, newBootstrapBitcoinDeriver()); err != nil {
		t.Fatalf("expected valid bitcoin policies to pass validation: %v", err)
	}
}

func TestValidateConfiguredAddressIssuancePoliciesSkipsDisabledAndNonBitcoinPolicies(t *testing.T) {
	policies := []policies.AddressIssuancePolicy{
		newBitcoinAddressIssuancePolicy(
			valueobjects.AddressPolicyIDBitcoinTestnet4NativeSegwit,
			valueobjects.NetworkIDTestnet4,
			string(valueobjects.AddressSchemeNativeSegwit),
			"",
			"m/84'/1'/0'",
		),
		{
			AddressPolicyID: valueobjects.AddressPolicyIDEthereumSepoliaCreate2,
			Chain:           valueobjects.SupportedChainEthereum,
			Network:         valueobjects.NetworkIDSepolia,
			Scheme:          "create2",
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef: "create2.v1:factory=0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa;collector=0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb;init_code_hash=0x1111111111111111111111111111111111111111111111111111111111111111",
			},
		},
	}

	if err := validateConfiguredAddressIssuancePolicies(policies, newBootstrapBitcoinDeriver()); err != nil {
		t.Fatalf("expected skipped policies to pass validation: %v", err)
	}
}

func TestValidateConfiguredAddressIssuancePoliciesRejectsInvalidEthereumAssetReference(t *testing.T) {
	policies := []policies.AddressIssuancePolicy{
		{
			AddressPolicyID: valueobjects.AddressPolicyIDEthereumSepoliaUSDTCreate2,
			Chain:           valueobjects.SupportedChainEthereum,
			Network:         valueobjects.NetworkIDSepolia,
			Scheme:          valueobjects.AddressSchemeCreate2,
			AssetReference:  "0xnot-a-token",
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef: "create2.v1:factory=0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa;collector=0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb;init_code_hash=0x1111111111111111111111111111111111111111111111111111111111111111",
			},
		},
	}

	err := validateConfiguredAddressIssuancePolicies(policies, newBootstrapBitcoinDeriver())
	if err == nil {
		t.Fatal("expected invalid ethereum asset reference validation error")
	}
	if !strings.Contains(err.Error(), string(valueobjects.AddressPolicyIDEthereumSepoliaUSDTCreate2)) {
		t.Fatalf("expected policy id in error, got %q", err)
	}
	if !strings.Contains(err.Error(), envEthereumSepoliaUSDTAssetReference) {
		t.Fatalf("expected env key in error, got %q", err)
	}
}

func TestValidateConfiguredAddressIssuancePoliciesRejectsMissingEthereumAssetReferenceForUSDTPolicy(t *testing.T) {
	policies := []policies.AddressIssuancePolicy{
		{
			AddressPolicyID: valueobjects.AddressPolicyIDEthereumSepoliaUSDTCreate2,
			Chain:           valueobjects.SupportedChainEthereum,
			Network:         valueobjects.NetworkIDSepolia,
			Scheme:          valueobjects.AddressSchemeCreate2,
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef: "create2.v1:factory=0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa;collector=0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb;init_code_hash=0x1111111111111111111111111111111111111111111111111111111111111111",
			},
		},
	}

	err := validateConfiguredAddressIssuancePolicies(policies, newBootstrapBitcoinDeriver())
	if err == nil {
		t.Fatal("expected missing ethereum asset reference validation error")
	}
	if !strings.Contains(err.Error(), string(valueobjects.AddressPolicyIDEthereumSepoliaUSDTCreate2)) {
		t.Fatalf("expected policy id in error, got %q", err)
	}
	if !strings.Contains(err.Error(), envEthereumSepoliaUSDTAssetReference) {
		t.Fatalf("expected env key in error, got %q", err)
	}
}

func findAddressIssuancePolicyByID(
	issuancePolicies []policies.AddressIssuancePolicy,
	addressPolicyID valueobjects.AddressPolicyID,
) policies.AddressIssuancePolicy {
	for _, policy := range issuancePolicies {
		if policy.AddressPolicyID == addressPolicyID.Normalize() {
			return policy
		}
	}
	return policies.AddressIssuancePolicy{}
}

func newBootstrapBitcoinDeriver() *bitcoin.HDXPubAddressDeriver {
	return bitcoin.NewHDXPubAddressDeriver(
		bitcoin.NewLegacyAddressEncoder(),
		bitcoin.NewSegwitAddressEncoder(),
		bitcoin.NewNativeSegwitAddressEncoder(),
		bitcoin.NewTaprootAddressEncoder(),
	)
}
