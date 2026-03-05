package main

import "testing"

func TestLoadPollerConfigFromEnvSuccess(t *testing.T) {
	t.Setenv("POLL_CHAIN", " BitCoin ")
	t.Setenv("POLL_NETWORK", " TestNet4 ")
	t.Setenv("POLL_INTERVAL", "5s")
	t.Setenv("POLL_CLAIM_TTL", "8s")
	t.Setenv("POLL_BATCH_SIZE", "12")
	t.Setenv("POLL_REQUIRED_CONFIRMATIONS", "2")

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
	if config.DefaultRequiredConfirmations != 2 {
		t.Fatalf("unexpected confirmations: got %d", config.DefaultRequiredConfirmations)
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
