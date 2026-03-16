package main

import (
	"testing"
	"time"
)

func TestParseEVMSweeperConfigSuccess(t *testing.T) {
	config, err := parseEVMSweeperConfig([]string{
		"-network=sepolia",
		"-asset-code=USDT",
		"-payment-address-ids=101, 202",
		"-before-issued-at=2026-03-16T10:20:30Z",
		"-batch-size=25",
		"-dry-run=false",
	})
	if err != nil {
		t.Fatalf("parseEVMSweeperConfig returned error: %v", err)
	}

	if config.Network != "sepolia" {
		t.Fatalf("unexpected network: got %q", config.Network)
	}
	if config.AssetCode != "usdt" {
		t.Fatalf("unexpected asset code: got %q", config.AssetCode)
	}
	if len(config.PaymentAddressIDs) != 2 || config.PaymentAddressIDs[0] != 101 || config.PaymentAddressIDs[1] != 202 {
		t.Fatalf("unexpected payment address ids: %+v", config.PaymentAddressIDs)
	}
	expectedTime := time.Date(2026, 3, 16, 10, 20, 30, 0, time.UTC)
	if !config.BeforeIssuedAt.Equal(expectedTime) {
		t.Fatalf("unexpected beforeIssuedAt: got %v", config.BeforeIssuedAt)
	}
	if config.BatchSize != 25 {
		t.Fatalf("unexpected batch size: got %d", config.BatchSize)
	}
	if config.DryRun {
		t.Fatalf("expected dry run disabled")
	}
}

func TestParseEVMSweeperConfigDefaults(t *testing.T) {
	config, err := parseEVMSweeperConfig(nil)
	if err != nil {
		t.Fatalf("parseEVMSweeperConfig returned error: %v", err)
	}
	if !config.DryRun {
		t.Fatalf("expected dry run enabled by default")
	}
	if config.Network != "" || config.AssetCode != "" || len(config.PaymentAddressIDs) != 0 || !config.BeforeIssuedAt.IsZero() {
		t.Fatalf("unexpected non-zero defaults: %+v", config)
	}
}

func TestParseEVMSweeperConfigRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "invalid network", args: []string{"-network=eth/mainnet"}},
		{name: "invalid asset code", args: []string{"-asset-code=usdt/mainnet"}},
		{name: "invalid payment address ids", args: []string{"-payment-address-ids=101,abc"}},
		{name: "invalid before issued at", args: []string{"-before-issued-at=not-a-time"}},
		{name: "negative batch size", args: []string{"-batch-size=-1"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseEVMSweeperConfig(tc.args); err == nil {
				t.Fatalf("expected error for args %v", tc.args)
			}
		})
	}
}
