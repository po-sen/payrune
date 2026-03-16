package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEVMFactoryDeployConfigFromManifest(t *testing.T) {
	manifestPath := writeManifest(t, `{
  "chainId": "11155111",
  "contractAddress": "0x1111111111111111111111111111111111111111",
  "collector": "0x2222222222222222222222222222222222222222",
  "vaultCreationCodeHash": "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "deploymentTransactionHash": "0xabc",
  "deployedAt": "2026-03-16T07:00:00Z"
}`)

	config, err := parseEVMFactoryDeployConfig([]string{
		"-deployment-manifest=" + manifestPath,
	})
	if err != nil {
		t.Fatalf("parseEVMFactoryDeployConfig returned error: %v", err)
	}
	if config.Network != "sepolia" {
		t.Fatalf("unexpected network: got %q", config.Network)
	}
}

func TestParseEVMFactoryDeployConfigLoadsDefaultsFromEnv(t *testing.T) {
	t.Setenv("ETHEREUM_SEPOLIA_RPC_URL", "https://sepolia.example")
	t.Setenv("ETHEREUM_SEPOLIA_DEPLOY_PRIVATE_KEY", "0xabc")
	t.Setenv("ETHEREUM_SEPOLIA_COLLECTOR_ADDRESS", "0x2222222222222222222222222222222222222222")
	t.Setenv("ETHEREUM_SEPOLIA_FACTORY_DEPLOYMENT_MANIFEST", "deployments/ethereum/build/deployments/test.json")
	t.Setenv("ETHEREUM_SEPOLIA_DEPLOY_CONFIRMATIONS", "3")

	config, err := parseEVMFactoryDeployConfig([]string{
		"-network=sepolia",
	})
	if err != nil {
		t.Fatalf("parseEVMFactoryDeployConfig returned error: %v", err)
	}
	if config.RPCURL != "https://sepolia.example" {
		t.Fatalf("unexpected rpc url: got %q", config.RPCURL)
	}
	if config.Confirmations != 3 {
		t.Fatalf("unexpected confirmations: got %d", config.Confirmations)
	}
	if config.OutputManifestPath != "deployments/ethereum/build/deployments/test.json" {
		t.Fatalf("unexpected output manifest path: got %q", config.OutputManifestPath)
	}
}

func TestParseEVMFactoryDeployConfigFlagOverride(t *testing.T) {
	config, err := parseEVMFactoryDeployConfig([]string{
		"-network=mainnet",
		"-rpc-url=https://mainnet.example",
		"-deploy-private-key=0xabc",
		"-collector-address=0x2222222222222222222222222222222222222222",
		"-confirmations=5",
		"-output-manifest=/tmp/factory.json",
	})
	if err != nil {
		t.Fatalf("parseEVMFactoryDeployConfig returned error: %v", err)
	}
	if config.Network != "mainnet" {
		t.Fatalf("unexpected network: got %q", config.Network)
	}
	if config.Confirmations != 5 {
		t.Fatalf("unexpected confirmations: got %d", config.Confirmations)
	}
}

func TestParseEVMFactoryDeployConfigRejectsInvalidInput(t *testing.T) {
	manifestPath := writeManifest(t, `{
  "chainId": "11155111",
  "contractAddress": "0x1111111111111111111111111111111111111111",
  "collector": "0x2222222222222222222222222222222222222222",
  "vaultCreationCodeHash": "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "deployedAt": "2026-03-16T07:00:00Z"
}`)

	tests := []struct {
		name string
		args []string
	}{
		{name: "missing network", args: []string{"-rpc-url=https://mainnet.example", "-deploy-private-key=0xabc", "-collector-address=0x2222222222222222222222222222222222222222"}},
		{name: "invalid network", args: []string{"-network=goerli", "-rpc-url=https://goerli.example", "-deploy-private-key=0xabc", "-collector-address=0x2222222222222222222222222222222222222222"}},
		{name: "missing rpc url", args: []string{"-network=mainnet", "-deploy-private-key=0xabc", "-collector-address=0x2222222222222222222222222222222222222222"}},
		{name: "invalid collector address", args: []string{"-network=mainnet", "-rpc-url=https://mainnet.example", "-deploy-private-key=0xabc", "-collector-address=bad"}},
		{name: "manifest network mismatch", args: []string{"-network=mainnet", "-deployment-manifest=" + manifestPath}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseEVMFactoryDeployConfig(tc.args); err == nil {
				t.Fatalf("expected error for args %v", tc.args)
			}
		})
	}
}

func writeManifest(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	return path
}
