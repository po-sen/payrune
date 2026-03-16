package di

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

func TestLoadEVMSweeperSecretConfigsFromEnvSuccess(t *testing.T) {
	t.Setenv(envEthereumMainnetRPCURL, "https://mainnet.example")
	t.Setenv(envEthereumMainnetSweeperPrivateKey, "0xabc123")
	t.Setenv(envEthereumSepoliaRPCURL, "https://sepolia.example")
	t.Setenv(envEthereumSepoliaSweeperPrivateKey, "")

	configs, err := LoadEVMSweeperSecretConfigsFromEnv()
	if err != nil {
		t.Fatalf("LoadEVMSweeperSecretConfigsFromEnv returned error: %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("unexpected config count: got %d", len(configs))
	}
	if configs[valueobjects.NetworkID("mainnet")].RPCURL != "https://mainnet.example" {
		t.Fatalf("unexpected mainnet rpc url: %q", configs[valueobjects.NetworkID("mainnet")].RPCURL)
	}
}

func TestBuildEVMSweeperNetworkConfigsSuccess(t *testing.T) {
	configs, err := BuildEVMSweeperNetworkConfigs(
		map[valueobjects.NetworkID]EVMSweeperSecretConfig{
			valueobjects.NetworkID("mainnet"): {
				Network:           valueobjects.NetworkID("mainnet"),
				RPCURL:            "https://mainnet.example",
				SweeperPrivateKey: "0xabc123",
			},
			valueobjects.NetworkID("sepolia"): {
				Network: valueobjects.NetworkID("sepolia"),
				RPCURL:  "https://sepolia.example",
			},
		},
		[]outport.EVMFactoryRecord{
			{
				ID:               10,
				Network:          valueobjects.NetworkID("mainnet"),
				FactoryAddress:   "0x1111111111111111111111111111111111111111",
				CollectorAddress: "0x2222222222222222222222222222222222222222",
				Status:           outport.EVMFactoryStatusActive,
			},
			{
				ID:               11,
				Network:          valueobjects.NetworkID("sepolia"),
				FactoryAddress:   "0x3333333333333333333333333333333333333333",
				CollectorAddress: "0x4444444444444444444444444444444444444444",
				Status:           outport.EVMFactoryStatusActive,
			},
		},
	)
	if err != nil {
		t.Fatalf("BuildEVMSweeperNetworkConfigs returned error: %v", err)
	}

	networks := ConfiguredEVMSweeperNetworks(configs)
	expectedNetworks := []string{"mainnet", "sepolia"}
	if !slices.Equal(networks, expectedNetworks) {
		t.Fatalf("unexpected network list: got %v want %v", networks, expectedNetworks)
	}
	if configs[valueobjects.NetworkID("sepolia")].FactoryID != 11 {
		t.Fatalf("unexpected factory id: got %d", configs[valueobjects.NetworkID("sepolia")].FactoryID)
	}
}

func TestBuildEVMSweeperNetworkConfigsSkipsFactoryWithoutSecretRuntime(t *testing.T) {
	configs, err := BuildEVMSweeperNetworkConfigs(
		map[valueobjects.NetworkID]EVMSweeperSecretConfig{},
		[]outport.EVMFactoryRecord{
			{
				Network:          valueobjects.NetworkID("sepolia"),
				FactoryAddress:   "0x1111111111111111111111111111111111111111",
				CollectorAddress: "0x2222222222222222222222222222222222222222",
				Status:           outport.EVMFactoryStatusActive,
			},
		},
	)
	if err != nil {
		t.Fatalf("BuildEVMSweeperNetworkConfigs returned error: %v", err)
	}
	if len(configs) != 0 {
		t.Fatalf("expected no configured runtimes, got %d", len(configs))
	}
}

func TestBuildEVMSweeperNetworkConfigsAllowsNetworkWithoutActiveFactory(t *testing.T) {
	configs, err := BuildEVMSweeperNetworkConfigs(
		map[valueobjects.NetworkID]EVMSweeperSecretConfig{
			valueobjects.NetworkID("sepolia"): {
				Network: valueobjects.NetworkID("sepolia"),
				RPCURL:  "https://sepolia.example",
			},
		},
		nil,
	)
	if err != nil {
		t.Fatalf("BuildEVMSweeperNetworkConfigs returned error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("unexpected config count: got %d", len(configs))
	}
	if configs[valueobjects.NetworkID("sepolia")].FactoryAddress != "" {
		t.Fatalf("expected empty factory address, got %q", configs[valueobjects.NetworkID("sepolia")].FactoryAddress)
	}
}

func TestLoadEthereumFactoryDeploymentManifestSuccess(t *testing.T) {
	manifestPath := writeEthereumFactoryManifest(t, `{
  "contractName": "DepositVaultFactory",
  "chainId": "11155111",
  "contractAddress": "0x1111111111111111111111111111111111111111",
  "collector": "0x2222222222222222222222222222222222222222",
  "vaultCreationCodeHash": "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "deploymentTransactionHash": "0xabc",
  "deployedAt": "2026-03-16T06:00:00Z"
}`)

	manifest, err := LoadEthereumFactoryDeploymentManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadEthereumFactoryDeploymentManifest returned error: %v", err)
	}
	if manifest.ChainID != "11155111" {
		t.Fatalf("unexpected chain id: %q", manifest.ChainID)
	}
	if manifest.DeploymentTransactionHash != "0xabc" {
		t.Fatalf("unexpected tx hash: %q", manifest.DeploymentTransactionHash)
	}
	if manifest.VaultCreationCodeHash != "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("unexpected vault creation code hash: %q", manifest.VaultCreationCodeHash)
	}
	if !manifest.DeployedAt.Equal(time.Date(2026, 3, 16, 6, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected deployed at: %v", manifest.DeployedAt)
	}
}

func TestParseEthereumNetworkFromChainID(t *testing.T) {
	network, ok := ParseEthereumNetworkFromChainID("11155111")
	if !ok || network != valueobjects.NetworkID("sepolia") {
		t.Fatalf("unexpected parse result: %q %t", network, ok)
	}
}

func TestLoadEthereumFactoryDeployDefaultsFromEnv(t *testing.T) {
	t.Setenv(envEthereumSepoliaRPCURL, "https://sepolia.example")
	t.Setenv(envEthereumSepoliaDeployPrivateKey, "0xabc123")
	t.Setenv(envEthereumSepoliaDeployConfirmations, "3")
	t.Setenv(envEthereumSepoliaCollectorAddress, "0x2222222222222222222222222222222222222222")
	t.Setenv(envEthereumSepoliaFactoryManifest, "/tmp/factory.json")

	defaults, ok, err := LoadEthereumFactoryDeployDefaultsFromEnv(valueobjects.NetworkID("sepolia"))
	if err != nil {
		t.Fatalf("LoadEthereumFactoryDeployDefaultsFromEnv returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected defaults")
	}
	if defaults.RPCURL != "https://sepolia.example" {
		t.Fatalf("unexpected rpc url: %q", defaults.RPCURL)
	}
	if defaults.DeployPrivateKey != "0xabc123" {
		t.Fatalf("unexpected deploy private key: %q", defaults.DeployPrivateKey)
	}
	if defaults.Confirmations != 3 {
		t.Fatalf("unexpected confirmations: %d", defaults.Confirmations)
	}
	if defaults.DeploymentManifestPath != "/tmp/factory.json" {
		t.Fatalf("unexpected manifest path: %q", defaults.DeploymentManifestPath)
	}
	if defaults.CollectorAddress != "0x2222222222222222222222222222222222222222" {
		t.Fatalf("unexpected collector address: %q", defaults.CollectorAddress)
	}
}

func TestLoadEthereumFactoryDeployDefaultsFromEnvRejectsInvalidConfirmations(t *testing.T) {
	t.Setenv(envEthereumMainnetDeployConfirmations, "bad")

	_, _, err := LoadEthereumFactoryDeployDefaultsFromEnv(valueobjects.NetworkID("mainnet"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func writeEthereumFactoryManifest(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "factory-manifest.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	return path
}
