package di

import (
	"testing"
	"time"

	"payrune/internal/domain/valueobjects"
)

func TestLoadBitcoinRequiredConfirmationsFromEnvDefaults(t *testing.T) {
	t.Setenv(envBitcoinMainnetRequiredConfirmations, "")
	t.Setenv(envBitcoinTestnet4RequiredConfirmations, "")

	config, err := loadBitcoinRequiredConfirmationsFromEnv()
	if err != nil {
		t.Fatalf("loadBitcoinRequiredConfirmationsFromEnv returned error: %v", err)
	}

	if got := config[valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet)]; got != 1 {
		t.Fatalf("unexpected mainnet confirmations: got %d", got)
	}
	if got := config[valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4)]; got != 1 {
		t.Fatalf("unexpected testnet4 confirmations: got %d", got)
	}
}

func TestLoadBitcoinRequiredConfirmationsFromEnvCustom(t *testing.T) {
	t.Setenv(envBitcoinMainnetRequiredConfirmations, "6")
	t.Setenv(envBitcoinTestnet4RequiredConfirmations, "2")

	config, err := loadBitcoinRequiredConfirmationsFromEnv()
	if err != nil {
		t.Fatalf("loadBitcoinRequiredConfirmationsFromEnv returned error: %v", err)
	}

	if got := config[valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet)]; got != 6 {
		t.Fatalf("unexpected mainnet confirmations: got %d", got)
	}
	if got := config[valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4)]; got != 2 {
		t.Fatalf("unexpected testnet4 confirmations: got %d", got)
	}
}

func TestLoadBitcoinRequiredConfirmationsFromEnvInvalid(t *testing.T) {
	t.Setenv(envBitcoinMainnetRequiredConfirmations, "abc")
	t.Setenv(envBitcoinTestnet4RequiredConfirmations, "1")

	_, err := loadBitcoinRequiredConfirmationsFromEnv()
	if err == nil {
		t.Fatal("expected parse error for mainnet confirmations")
	}
}

func TestLoadBitcoinRequiredConfirmationsFromEnvNonPositive(t *testing.T) {
	t.Setenv(envBitcoinMainnetRequiredConfirmations, "0")
	t.Setenv(envBitcoinTestnet4RequiredConfirmations, "1")

	_, err := loadBitcoinRequiredConfirmationsFromEnv()
	if err == nil {
		t.Fatal("expected validation error for non-positive confirmations")
	}
}

func TestLoadBitcoinReceiptExpiresAfterByNetworkFromEnvDefaults(t *testing.T) {
	t.Setenv(envBitcoinMainnetReceiptExpiresAfter, "")
	t.Setenv(envBitcoinTestnet4ReceiptExpiresAfter, "")

	config, err := loadBitcoinReceiptExpiresAfterByNetworkFromEnv()
	if err != nil {
		t.Fatalf("loadBitcoinReceiptExpiresAfterByNetworkFromEnv returned error: %v", err)
	}

	if got := config[valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet)]; got != defaultBitcoinReceiptExpiresAfter {
		t.Fatalf("unexpected mainnet receipt expires after: got %s", got)
	}
	if got := config[valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4)]; got != defaultBitcoinReceiptExpiresAfter {
		t.Fatalf("unexpected testnet4 receipt expires after: got %s", got)
	}
}

func TestLoadBitcoinReceiptExpiresAfterByNetworkFromEnvCustom(t *testing.T) {
	t.Setenv(envBitcoinMainnetReceiptExpiresAfter, "240h")
	t.Setenv(envBitcoinTestnet4ReceiptExpiresAfter, "36h")

	config, err := loadBitcoinReceiptExpiresAfterByNetworkFromEnv()
	if err != nil {
		t.Fatalf("loadBitcoinReceiptExpiresAfterByNetworkFromEnv returned error: %v", err)
	}

	if got := config[valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet)]; got != 240*time.Hour {
		t.Fatalf("unexpected mainnet receipt expires after: got %s", got)
	}
	if got := config[valueobjects.NetworkID(valueobjects.BitcoinNetworkTestnet4)]; got != 36*time.Hour {
		t.Fatalf("unexpected testnet4 receipt expires after: got %s", got)
	}
}

func TestLoadBitcoinReceiptExpiresAfterByNetworkFromEnvInvalid(t *testing.T) {
	t.Setenv(envBitcoinMainnetReceiptExpiresAfter, "abc")
	t.Setenv(envBitcoinTestnet4ReceiptExpiresAfter, "36h")

	_, err := loadBitcoinReceiptExpiresAfterByNetworkFromEnv()
	if err == nil {
		t.Fatal("expected parse error for mainnet receipt expires after")
	}
}

func TestLoadBitcoinReceiptExpiresAfterByNetworkFromEnvNonPositive(t *testing.T) {
	t.Setenv(envBitcoinMainnetReceiptExpiresAfter, "0s")
	t.Setenv(envBitcoinTestnet4ReceiptExpiresAfter, "36h")

	_, err := loadBitcoinReceiptExpiresAfterByNetworkFromEnv()
	if err == nil {
		t.Fatal("expected validation error for non-positive receipt expires after")
	}
}
