package main

import (
	"testing"
	"time"
)

func TestLoadPollerConfigFromEnvSuccess(t *testing.T) {
	t.Setenv("POLL_CHAIN", " BitCoin ")
	t.Setenv("POLL_NETWORK", " TestNet4 ")
	t.Setenv("POLL_TICK_INTERVAL", "5s")
	t.Setenv("RECEIPT_POLL_INTERVAL", "2m")
	t.Setenv("POLL_CLAIM_TTL", "8s")
	t.Setenv("POLL_BATCH_SIZE", "12")

	config, err := loadPollerConfigFromEnv()
	if err != nil {
		t.Fatalf("loadPollerConfigFromEnv returned error: %v", err)
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
	if config.ReceiptPollInterval != 2*time.Minute {
		t.Fatalf("unexpected receipt poll interval: got %s", config.ReceiptPollInterval)
	}
}

func TestLoadPollerConfigFromEnvIgnoresLegacyInterval(t *testing.T) {
	t.Setenv("POLL_CHAIN", "bitcoin")
	t.Setenv("POLL_NETWORK", "mainnet")
	t.Setenv("POLL_INTERVAL", "45s")

	config, err := loadPollerConfigFromEnv()
	if err != nil {
		t.Fatalf("loadPollerConfigFromEnv returned error: %v", err)
	}

	if config.TickInterval != 0 {
		t.Fatalf("expected zero tick interval when only legacy env is set: got %s", config.TickInterval)
	}
	if config.ReceiptPollInterval != 0 {
		t.Fatalf("expected zero receipt poll interval when only legacy env is set: got %s", config.ReceiptPollInterval)
	}
}

func TestLoadPollerConfigFromEnvDoesNotMixLegacyInterval(t *testing.T) {
	t.Setenv("POLL_CHAIN", "bitcoin")
	t.Setenv("POLL_NETWORK", "mainnet")
	t.Setenv("POLL_INTERVAL", "45s")
	t.Setenv("POLL_TICK_INTERVAL", "10s")

	config, err := loadPollerConfigFromEnv()
	if err != nil {
		t.Fatalf("loadPollerConfigFromEnv returned error: %v", err)
	}

	if config.TickInterval != 10*time.Second {
		t.Fatalf("unexpected explicit tick interval: got %s", config.TickInterval)
	}
	if config.ReceiptPollInterval != 0 {
		t.Fatalf("expected zero receipt poll interval without explicit env: got %s", config.ReceiptPollInterval)
	}
}

func TestLoadPollerConfigFromEnvRequiresChainWhenNetworkSet(t *testing.T) {
	t.Setenv("POLL_CHAIN", "")
	t.Setenv("POLL_NETWORK", "mainnet")

	_, err := loadPollerConfigFromEnv()
	if err == nil {
		t.Fatal("expected validation error when network set without chain")
	}
}

func TestParseChainEnvAllowsCustomChain(t *testing.T) {
	t.Setenv("POLL_CHAIN", "Eth")

	chain, err := parseChainEnv("POLL_CHAIN")
	if err != nil {
		t.Fatalf("parseChainEnv returned error: %v", err)
	}
	if chain != "eth" {
		t.Fatalf("unexpected normalized chain: got %q", chain)
	}
}

func TestParseNetworkEnvValidation(t *testing.T) {
	t.Setenv("POLL_NETWORK", "main/net")

	_, err := parseNetworkEnv("POLL_NETWORK")
	if err == nil {
		t.Fatal("expected invalid network error")
	}
}
