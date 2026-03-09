package di

import (
	"testing"
	"time"
)

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
