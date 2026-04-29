package ethereum

import (
	"context"
	"strings"
	"testing"

	outport "payrune/internal/application/ports/outbound"
	ethereumcreate2assets "payrune/internal/infrastructure/ethereumcreate2assets"
)

func TestEthereumAddressIssuanceReadinessCheckerAllowsBitcoinPolicy(t *testing.T) {
	checker, err := NewAddressIssuanceReadinessChecker(nil)
	if err != nil {
		t.Fatalf("NewAddressIssuanceReadinessChecker returned error: %v", err)
	}

	err = checker.CheckIssuanceReadiness(context.Background(), outport.AddressIssuancePolicyRecord{
		AddressPolicyID:   "bitcoin-mainnet-legacy",
		Chain:             outport.SupportedChainBitcoin,
		Network:           outport.NetworkIDMainnet,
		Scheme:            outport.AddressSchemeLegacy,
		Decimals:          8,
		AddressSpaceRef:   "xpub",
		IssuanceRefPrefix: "m/44'/0'/0'",
	})
	if err != nil {
		t.Fatalf("expected bitcoin policy to bypass ethereum readiness, got %v", err)
	}
}

func TestEthereumAddressIssuanceReadinessCheckerRejectsMissingRPCConfig(t *testing.T) {
	checker, err := NewAddressIssuanceReadinessChecker(nil)
	if err != nil {
		t.Fatalf("NewAddressIssuanceReadinessChecker returned error: %v", err)
	}

	err = checker.CheckIssuanceReadiness(
		context.Background(),
		newEthereumReadinessPolicy("ethereum-sepolia-create2", outport.NetworkIDSepolia, "", 18),
	)
	if err == nil {
		t.Fatalf("expected unavailable error, got %v", err)
	}
	if !strings.Contains(err.Error(), "rpc client is not configured") {
		t.Fatalf("expected rpc configuration error, got %q", err)
	}
}

func TestEthereumAddressIssuanceReadinessCheckerChecksNativeFactory(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	defer server.Close()

	metadata, artifact := mustLoadFactoryFixture(t, outport.NetworkIDSepolia)
	state.codesByAddress[strings.ToLower(metadata.FactoryAddress)] = artifact.RuntimeCodeHex

	checker := newEthereumReadinessCheckerForTest(t, server.URL)
	err := checker.CheckIssuanceReadiness(
		context.Background(),
		newEthereumReadinessPolicy("ethereum-sepolia-create2", outport.NetworkIDSepolia, "", 18),
	)
	if err != nil {
		t.Fatalf("expected native ethereum policy to be ready, got %v", err)
	}
	if got := strings.Join(state.requestedCodes, ","); got != strings.ToLower(metadata.FactoryAddress) {
		t.Fatalf("unexpected code checks: got %q", got)
	}
	if len(state.requestedTokenBalances) != 0 || len(state.requestedTokenDecimals) != 0 {
		t.Fatalf("expected no token calls for native policy, got balances=%v decimals=%v", state.requestedTokenBalances, state.requestedTokenDecimals)
	}
}

func TestEthereumAddressIssuanceReadinessCheckerRejectsMissingFactoryCode(t *testing.T) {
	_, server := newTestEthereumRPCServer(t)
	defer server.Close()

	checker := newEthereumReadinessCheckerForTest(t, server.URL)
	err := checker.CheckIssuanceReadiness(
		context.Background(),
		newEthereumReadinessPolicy("ethereum-sepolia-create2", outport.NetworkIDSepolia, "", 18),
	)
	if err == nil {
		t.Fatalf("expected unavailable error, got %v", err)
	}
	if !strings.Contains(err.Error(), "factory contract is not deployed") {
		t.Fatalf("expected missing factory message, got %q", err)
	}
}

func TestEthereumAddressIssuanceReadinessCheckerRejectsFactoryRuntimeHashMismatch(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	defer server.Close()

	metadata, _ := mustLoadFactoryFixture(t, outport.NetworkIDSepolia)
	state.codesByAddress[strings.ToLower(metadata.FactoryAddress)] = "0x60006000"

	checker := newEthereumReadinessCheckerForTest(t, server.URL)
	err := checker.CheckIssuanceReadiness(
		context.Background(),
		newEthereumReadinessPolicy("ethereum-sepolia-create2", outport.NetworkIDSepolia, "", 18),
	)
	if err == nil {
		t.Fatalf("expected unavailable error, got %v", err)
	}
	if !strings.Contains(err.Error(), "factory runtime hash mismatch") {
		t.Fatalf("expected runtime hash mismatch message, got %q", err)
	}
}

func TestEthereumAddressIssuanceReadinessCheckerChecksTokenContract(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	defer server.Close()

	metadata, artifact := mustLoadFactoryFixture(t, outport.NetworkIDSepolia)
	tokenAddress := "0xd077a400968890eacc75cdc901f0356c943e4fdb"

	state.codesByAddress[strings.ToLower(metadata.FactoryAddress)] = artifact.RuntimeCodeHex
	state.codesByAddress[tokenAddress] = "0x60006000"
	state.tokenBalancesByKey[ethereumTokenBalanceKey(tokenAddress, zeroEthereumAddress, "latest")] = "0x0"
	state.tokenDecimalsByAddress[tokenAddress] = "0x6"

	checker := newEthereumReadinessCheckerForTest(t, server.URL)
	err := checker.CheckIssuanceReadiness(
		context.Background(),
		newEthereumReadinessPolicy("ethereum-sepolia-usdt-create2", outport.NetworkIDSepolia, strings.ToUpper(tokenAddress), 6),
	)
	if err != nil {
		t.Fatalf("expected token policy to be ready, got %v", err)
	}
	if len(state.requestedTokenBalances) != 1 {
		t.Fatalf("expected one balanceOf call, got %v", state.requestedTokenBalances)
	}
	if len(state.requestedTokenDecimals) != 1 || state.requestedTokenDecimals[0] != tokenAddress {
		t.Fatalf("unexpected token decimals calls: got %v", state.requestedTokenDecimals)
	}
}

func TestEthereumAddressIssuanceReadinessCheckerRejectsTokenDecimalsMismatch(t *testing.T) {
	state, server := newTestEthereumRPCServer(t)
	defer server.Close()

	metadata, artifact := mustLoadFactoryFixture(t, outport.NetworkIDSepolia)
	tokenAddress := "0xd077a400968890eacc75cdc901f0356c943e4fdb"

	state.codesByAddress[strings.ToLower(metadata.FactoryAddress)] = artifact.RuntimeCodeHex
	state.codesByAddress[tokenAddress] = "0x60006000"
	state.tokenBalancesByKey[ethereumTokenBalanceKey(tokenAddress, zeroEthereumAddress, "latest")] = "0x0"
	state.tokenDecimalsByAddress[tokenAddress] = "0x12"

	checker := newEthereumReadinessCheckerForTest(t, server.URL)
	err := checker.CheckIssuanceReadiness(
		context.Background(),
		newEthereumReadinessPolicy("ethereum-sepolia-usdt-create2", outport.NetworkIDSepolia, tokenAddress, 6),
	)
	if err == nil {
		t.Fatalf("expected unavailable error, got %v", err)
	}
	if !strings.Contains(err.Error(), "token decimals mismatch") {
		t.Fatalf("expected decimals mismatch error, got %q", err)
	}
}

func newEthereumReadinessCheckerForTest(
	t *testing.T,
	endpoint string,
) *EthereumAddressIssuanceReadinessChecker {
	t.Helper()

	checker, err := NewAddressIssuanceReadinessChecker(map[string]*EthereumRPCObserverConfig{
		outport.NetworkIDSepolia: {Endpoint: endpoint},
	})
	if err != nil {
		t.Fatalf("NewAddressIssuanceReadinessChecker returned error: %v", err)
	}
	return checker
}

func mustLoadFactoryFixture(
	t *testing.T,
	network string,
) (ethereumcreate2assets.DeploymentMetadata, ethereumcreate2assets.ReceiverArtifact) {
	t.Helper()

	metadata, ok := ethereumcreate2assets.LookupDeploymentMetadata(string(network))
	if !ok {
		t.Fatalf("expected metadata for %s", network)
	}
	artifact, ok := ethereumcreate2assets.LookupReceiverArtifact(ethereumcreate2assets.FactoryArtifactName)
	if !ok {
		t.Fatal("expected factory artifact")
	}
	return metadata, artifact
}

func newEthereumReadinessPolicy(
	addressPolicyID string,
	network string,
	assetReference string,
	decimals uint8,
) outport.AddressIssuancePolicyRecord {
	return outport.AddressIssuancePolicyRecord{
		AddressPolicyID: addressPolicyID,
		Chain:           outport.SupportedChainEthereum,
		Network:         network,
		Scheme:          outport.AddressSchemeCreate2,
		AssetReference:  assetReference,
		Decimals:        decimals,
		Enabled:         true,
		AddressSpaceRef: "create2.v1:factory=0x1111111111111111111111111111111111111111;collector=0x2222222222222222222222222222222222222222;init_code_hash=0x3333333333333333333333333333333333333333333333333333333333333333",
	}
}
