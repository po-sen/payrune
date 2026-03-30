package bootstrap

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

func TestLoadPollerConfigFromEnvSuccess(t *testing.T) {
	t.Setenv("POLL_CHAIN", " BitCoin ")
	t.Setenv("POLL_NETWORK", " TestNet4 ")
	t.Setenv("POLL_TICK_INTERVAL", "5s")
	t.Setenv("POLL_RESCHEDULE_INTERVAL", "2m")
	t.Setenv("POLL_CLAIM_TTL", "8s")
	t.Setenv("POLL_BATCH_SIZE", "12")

	config, err := LoadPollerConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadPollerConfigFromEnv returned error: %v", err)
	}

	if config.Chain != "bitcoin" {
		t.Fatalf("unexpected normalized chain: got %q", config.Chain)
	}
	if config.Network != "testnet4" {
		t.Fatalf("unexpected normalized network: got %q", config.Network)
	}
	if config.BatchSize != 12 {
		t.Fatalf("unexpected batch size: got %d", config.BatchSize)
	}
	if config.TickInterval != 5*time.Second {
		t.Fatalf("unexpected tick interval: got %s", config.TickInterval)
	}
	if config.RescheduleInterval != 2*time.Minute {
		t.Fatalf("unexpected reschedule interval: got %s", config.RescheduleInterval)
	}
}

func TestLoadPollerConfigFromEnvIgnoresLegacyInterval(t *testing.T) {
	t.Setenv("POLL_CHAIN", "bitcoin")
	t.Setenv("POLL_NETWORK", "mainnet")
	t.Setenv("POLL_INTERVAL", "45s")

	config, err := LoadPollerConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadPollerConfigFromEnv returned error: %v", err)
	}

	if config.TickInterval != 0 {
		t.Fatalf("expected zero tick interval when only legacy env is set: got %s", config.TickInterval)
	}
	if config.RescheduleInterval != 0 {
		t.Fatalf("expected zero reschedule interval when only legacy env is set: got %s", config.RescheduleInterval)
	}
}

func TestLoadPollerConfigFromEnvDoesNotMixLegacyInterval(t *testing.T) {
	t.Setenv("POLL_CHAIN", "bitcoin")
	t.Setenv("POLL_NETWORK", "mainnet")
	t.Setenv("POLL_INTERVAL", "45s")
	t.Setenv("POLL_TICK_INTERVAL", "10s")

	config, err := LoadPollerConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadPollerConfigFromEnv returned error: %v", err)
	}

	if config.TickInterval != 10*time.Second {
		t.Fatalf("unexpected explicit tick interval: got %s", config.TickInterval)
	}
	if config.RescheduleInterval != 0 {
		t.Fatalf("expected zero reschedule interval without explicit env: got %s", config.RescheduleInterval)
	}
}

func TestLoadPollerConfigFromEnvRequiresChainWhenNetworkSet(t *testing.T) {
	t.Setenv("POLL_CHAIN", "")
	t.Setenv("POLL_NETWORK", "mainnet")

	_, err := LoadPollerConfigFromEnv()
	if err == nil {
		t.Fatal("expected validation error when network set without chain")
	}
}

func TestParsePollerChainLookupAllowsCustomChain(t *testing.T) {
	chain, err := parsePollerChainLookup(func(key string) string {
		if key == "POLL_CHAIN" {
			return "Eth"
		}
		return ""
	}, "POLL_CHAIN")
	if err != nil {
		t.Fatalf("parsePollerChainLookup returned error: %v", err)
	}
	if chain != "eth" {
		t.Fatalf("unexpected normalized chain: got %q", chain)
	}
}

func TestParsePollerNetworkLookupValidation(t *testing.T) {
	_, err := parsePollerNetworkLookup(func(key string) string {
		if key == "POLL_NETWORK" {
			return "main/net"
		}
		return ""
	}, "POLL_NETWORK")
	if err == nil {
		t.Fatal("expected invalid network error")
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

func TestFormatPollCycleStartLog(t *testing.T) {
	got := formatPollCycleStartLog(PollerConfig{
		Chain:     valueobjects.ChainIDEthereum,
		Network:   valueobjects.NetworkID("sepolia"),
		BatchSize: 2,
	})

	want := "poll cycle start chain=ethereum network=sepolia batch=2"
	if got != want {
		t.Fatalf("unexpected poll cycle start log: got %q want %q", got, want)
	}
}
