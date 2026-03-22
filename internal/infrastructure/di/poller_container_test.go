package di

import (
	"testing"
	"time"

	"payrune/internal/domain/valueobjects"
)

func TestLoadConfiguredPollerChainFromLookupEmpty(t *testing.T) {
	chain, scoped, err := loadConfiguredPollerChainFromLookup(func(string) string {
		return ""
	})
	if err != nil {
		t.Fatalf("loadConfiguredPollerChainFromLookup returned error: %v", err)
	}
	if scoped {
		t.Fatal("expected no configured poller chain scope")
	}
	if chain != "" {
		t.Fatalf("expected empty chain, got %q", chain)
	}
}

func TestLoadConfiguredPollerChainFromLookup(t *testing.T) {
	chain, scoped, err := loadConfiguredPollerChainFromLookup(func(key string) string {
		if key == envPollChain {
			return " Ethereum "
		}
		return ""
	})
	if err != nil {
		t.Fatalf("loadConfiguredPollerChainFromLookup returned error: %v", err)
	}
	if !scoped {
		t.Fatal("expected configured poller chain scope")
	}
	if chain != valueobjects.ChainIDEthereum {
		t.Fatalf("unexpected poller chain: got %q", chain)
	}
}

func TestLoadBitcoinMainnetEsploraConfigFromEnvMissingURL(t *testing.T) {
	t.Setenv(envBitcoinMainnetEsploraURL, "")
	t.Setenv(envBitcoinMainnetEsploraUser, "")
	t.Setenv(envBitcoinMainnetEsploraPassword, "")
	t.Setenv(envBitcoinMainnetEsploraTimeout, "")
	t.Setenv(envBitcoinMainnetEsploraTimeoutSeconds, "")

	config := loadBitcoinMainnetEsploraConfigFromEnv()
	if config != nil {
		t.Fatalf("expected nil config when url missing, got %+v", config)
	}
}

func TestLoadBitcoinMainnetEsploraConfigFromEnv(t *testing.T) {
	t.Setenv(envBitcoinMainnetEsploraURL, " https://mempool.space/api ")
	t.Setenv(envBitcoinMainnetEsploraUser, " user ")
	t.Setenv(envBitcoinMainnetEsploraPassword, "pass")
	t.Setenv(envBitcoinMainnetEsploraTimeout, "12s")
	t.Setenv(envBitcoinMainnetEsploraTimeoutSeconds, "")

	config := loadBitcoinMainnetEsploraConfigFromEnv()
	if config == nil {
		t.Fatal("expected config, got nil")
	}
	if config.Endpoint != "https://mempool.space/api" {
		t.Fatalf("unexpected endpoint: got %q", config.Endpoint)
	}
	if config.Username != "user" {
		t.Fatalf("unexpected username: got %q", config.Username)
	}
	if config.Password != "pass" {
		t.Fatalf("unexpected password: got %q", config.Password)
	}
	if config.Timeout != 12*time.Second {
		t.Fatalf("unexpected timeout: got %s", config.Timeout)
	}
}

func TestLoadBitcoinTestnet4EsploraConfigFromEnvTimeoutSecondsOverride(t *testing.T) {
	t.Setenv(envBitcoinTestnet4EsploraURL, "https://mempool.space/testnet4/api")
	t.Setenv(envBitcoinTestnet4EsploraUser, "")
	t.Setenv(envBitcoinTestnet4EsploraPassword, "")
	t.Setenv(envBitcoinTestnet4EsploraTimeout, "4s")
	t.Setenv(envBitcoinTestnet4EsploraTimeoutSeconds, "9")

	config := loadBitcoinTestnet4EsploraConfigFromEnv()
	if config == nil {
		t.Fatal("expected config, got nil")
	}
	if config.Timeout != 9*time.Second {
		t.Fatalf("unexpected timeout from seconds override: got %s", config.Timeout)
	}
}

func TestLoadBitcoinEsploraConfigsFromEnvMainnetOnly(t *testing.T) {
	t.Setenv(envBitcoinMainnetEsploraURL, "https://mempool.space/api")
	t.Setenv(envBitcoinMainnetEsploraUser, "")
	t.Setenv(envBitcoinMainnetEsploraPassword, "")
	t.Setenv(envBitcoinMainnetEsploraTimeout, "")
	t.Setenv(envBitcoinMainnetEsploraTimeoutSeconds, "")

	t.Setenv(envBitcoinTestnet4EsploraURL, "")
	t.Setenv(envBitcoinTestnet4EsploraUser, "")
	t.Setenv(envBitcoinTestnet4EsploraPassword, "")
	t.Setenv(envBitcoinTestnet4EsploraTimeout, "")
	t.Setenv(envBitcoinTestnet4EsploraTimeoutSeconds, "")

	configs := loadBitcoinEsploraConfigsFromEnv()
	if len(configs) != 1 {
		t.Fatalf("expected one configured network, got %d", len(configs))
	}
	if _, ok := configs["mainnet"]; !ok {
		t.Fatal("expected mainnet config")
	}
	if _, ok := configs["testnet4"]; ok {
		t.Fatal("did not expect testnet4 config")
	}
}

func TestLoadBitcoinEsploraConfigsFromEnvEmpty(t *testing.T) {
	t.Setenv(envBitcoinMainnetEsploraURL, "")
	t.Setenv(envBitcoinMainnetEsploraUser, "")
	t.Setenv(envBitcoinMainnetEsploraPassword, "")
	t.Setenv(envBitcoinMainnetEsploraTimeout, "")
	t.Setenv(envBitcoinMainnetEsploraTimeoutSeconds, "")

	t.Setenv(envBitcoinTestnet4EsploraURL, "")
	t.Setenv(envBitcoinTestnet4EsploraUser, "")
	t.Setenv(envBitcoinTestnet4EsploraPassword, "")
	t.Setenv(envBitcoinTestnet4EsploraTimeout, "")
	t.Setenv(envBitcoinTestnet4EsploraTimeoutSeconds, "")

	configs := loadBitcoinEsploraConfigsFromEnv()
	if len(configs) != 0 {
		t.Fatalf("expected no configured networks, got %d", len(configs))
	}
}

func TestLoadEthereumMainnetRPCConfigFromEnvMissingURL(t *testing.T) {
	t.Setenv(envEthereumMainnetRPCURL, "")
	t.Setenv(envEthereumMainnetRPCUser, "")
	t.Setenv(envEthereumMainnetRPCPassword, "")
	t.Setenv(envEthereumMainnetRPCTimeout, "")
	t.Setenv(envEthereumMainnetRPCTimeoutSeconds, "")

	config := loadEthereumMainnetRPCConfigFromEnv()
	if config != nil {
		t.Fatalf("expected nil config when url missing, got %+v", config)
	}
}

func TestLoadEthereumSepoliaRPCConfigFromEnv(t *testing.T) {
	t.Setenv(envEthereumSepoliaRPCURL, " https://sepolia.example ")
	t.Setenv(envEthereumSepoliaRPCUser, " user ")
	t.Setenv(envEthereumSepoliaRPCPassword, "pass")
	t.Setenv(envEthereumSepoliaRPCTimeout, "12s")
	t.Setenv(envEthereumSepoliaRPCTimeoutSeconds, "")

	config := loadEthereumSepoliaRPCConfigFromEnv()
	if config == nil {
		t.Fatal("expected config, got nil")
	}
	if config.Endpoint != "https://sepolia.example" {
		t.Fatalf("unexpected endpoint: got %q", config.Endpoint)
	}
	if config.Username != "user" {
		t.Fatalf("unexpected username: got %q", config.Username)
	}
	if config.Password != "pass" {
		t.Fatalf("unexpected password: got %q", config.Password)
	}
	if config.Timeout != 12*time.Second {
		t.Fatalf("unexpected timeout: got %s", config.Timeout)
	}
}

func TestLoadEthereumRPCConfigsFromEnvMainnetOnly(t *testing.T) {
	t.Setenv(envEthereumMainnetRPCURL, "https://mainnet.example")
	t.Setenv(envEthereumMainnetRPCUser, "")
	t.Setenv(envEthereumMainnetRPCPassword, "")
	t.Setenv(envEthereumMainnetRPCTimeout, "")
	t.Setenv(envEthereumMainnetRPCTimeoutSeconds, "")

	t.Setenv(envEthereumSepoliaRPCURL, "")
	t.Setenv(envEthereumSepoliaRPCUser, "")
	t.Setenv(envEthereumSepoliaRPCPassword, "")
	t.Setenv(envEthereumSepoliaRPCTimeout, "")
	t.Setenv(envEthereumSepoliaRPCTimeoutSeconds, "")

	configs := loadEthereumRPCConfigsFromEnv()
	if len(configs) != 1 {
		t.Fatalf("expected one configured network, got %d", len(configs))
	}
	if _, ok := configs["mainnet"]; !ok {
		t.Fatal("expected mainnet config")
	}
	if _, ok := configs["sepolia"]; ok {
		t.Fatal("did not expect sepolia config")
	}
}

func TestLoadEthereumRPCConfigsFromEnvEmpty(t *testing.T) {
	t.Setenv(envEthereumMainnetRPCURL, "")
	t.Setenv(envEthereumMainnetRPCUser, "")
	t.Setenv(envEthereumMainnetRPCPassword, "")
	t.Setenv(envEthereumMainnetRPCTimeout, "")
	t.Setenv(envEthereumMainnetRPCTimeoutSeconds, "")

	t.Setenv(envEthereumSepoliaRPCURL, "")
	t.Setenv(envEthereumSepoliaRPCUser, "")
	t.Setenv(envEthereumSepoliaRPCPassword, "")
	t.Setenv(envEthereumSepoliaRPCTimeout, "")
	t.Setenv(envEthereumSepoliaRPCTimeoutSeconds, "")

	configs := loadEthereumRPCConfigsFromEnv()
	if len(configs) != 0 {
		t.Fatalf("expected no configured networks, got %d", len(configs))
	}
}
