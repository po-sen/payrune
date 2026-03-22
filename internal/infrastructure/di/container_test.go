package di

import (
	"testing"
	"time"

	"payrune/internal/adapters/outbound/ethereum"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
	ethereumcreate2assets "payrune/internal/infrastructure/ethereumcreate2assets"
)

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
		Network: valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
	}]; got != 1 {
		t.Fatalf("unexpected mainnet confirmations: got %d", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
	}]; got != 1 {
		t.Fatalf("unexpected testnet4 confirmations: got %d", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkID("mainnet"),
	}]; got != 1 {
		t.Fatalf("unexpected ethereum mainnet confirmations: got %d", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkID("sepolia"),
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
		Network: valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
	}]; got != 6 {
		t.Fatalf("unexpected mainnet confirmations: got %d", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
	}]; got != 2 {
		t.Fatalf("unexpected testnet4 confirmations: got %d", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkID("mainnet"),
	}]; got != 12 {
		t.Fatalf("unexpected ethereum mainnet confirmations: got %d", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkID("sepolia"),
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
		Network: valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
	}]; got != defaultBitcoinReceiptExpiresAfter {
		t.Fatalf("unexpected mainnet receipt expires after: got %s", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
	}]; got != defaultBitcoinReceiptExpiresAfter {
		t.Fatalf("unexpected testnet4 receipt expires after: got %s", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkID("mainnet"),
	}]; got != defaultBitcoinReceiptExpiresAfter {
		t.Fatalf("unexpected ethereum mainnet receipt expires after: got %s", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkID("sepolia"),
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
		Network: valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
	}]; got != 240*time.Hour {
		t.Fatalf("unexpected mainnet receipt expires after: got %s", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainBitcoin,
		Network: valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4),
	}]; got != 36*time.Hour {
		t.Fatalf("unexpected testnet4 receipt expires after: got %s", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkID("mainnet"),
	}]; got != 72*time.Hour {
		t.Fatalf("unexpected ethereum mainnet receipt expires after: got %s", got)
	}
	if got := config[policies.PaymentReceiptTermsScope{
		Chain:   valueobjects.SupportedChainEthereum,
		Network: valueobjects.NetworkID("sepolia"),
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

func TestNewEthereumCreate2PolicyConfigBuildsSourceRefFromFixtureMetadata(t *testing.T) {
	network := valueobjects.NetworkID("sepolia")
	collectorAddress := "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	saltDeriver := ethereum.NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		network: "0x1111111111111111111111111111111111111111111111111111111111111111",
	})
	metadata, ok := ethereumcreate2assets.LookupDeploymentMetadata(network)
	if !ok {
		t.Fatalf("expected embedded metadata for %s", network)
	}

	initCodeHash, ok := metadata.Receiver.InitCodeHashHex(collectorAddress)
	if !ok {
		t.Fatal("expected init code hash available")
	}
	expectedSourceRef, err := ethereum.BuildCreate2AddressSourceRef(
		metadata.FactoryAddress,
		collectorAddress,
		initCodeHash,
	)
	if err != nil {
		t.Fatalf("BuildCreate2AddressSourceRef returned error: %v", err)
	}

	config := newEthereumCreate2PolicyConfig(network, collectorAddress, saltDeriver)

	if config.AddressPolicyID != "ethereum-sepolia-create2" {
		t.Fatalf("unexpected address policy id: got %q", config.AddressPolicyID)
	}
	if config.AddressReferencePrefix != "ethereum-sepolia-create2" {
		t.Fatalf("unexpected address reference prefix: got %q", config.AddressReferencePrefix)
	}
	if config.AddressSourceRef != expectedSourceRef {
		t.Fatalf("unexpected address source ref: got %q want %q", config.AddressSourceRef, expectedSourceRef)
	}
}

func TestNewEthereumCreate2PolicyConfigRequiresCollectorAddress(t *testing.T) {
	saltDeriver := ethereum.NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		valueobjects.NetworkID("mainnet"): "0x1111111111111111111111111111111111111111111111111111111111111111",
	})
	config := newEthereumCreate2PolicyConfig(valueobjects.NetworkID("mainnet"), "", saltDeriver)
	if config.AddressSourceRef != "" {
		t.Fatalf("expected disabled config when collector is missing, got %q", config.AddressSourceRef)
	}
}

func TestNewEthereumCreate2PolicyConfigRequiresSaltSecret(t *testing.T) {
	config := newEthereumCreate2PolicyConfig(valueobjects.NetworkID("mainnet"), "0x2222222222222222222222222222222222222222", ethereum.NewCreate2SaltDeriver(nil))
	if config.AddressSourceRef != "" {
		t.Fatalf("expected disabled config when derivation key is missing, got %q", config.AddressSourceRef)
	}
}
