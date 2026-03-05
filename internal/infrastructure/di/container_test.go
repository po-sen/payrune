package di

import (
	"testing"

	"payrune/internal/domain/value_objects"
)

func TestLoadBitcoinRequiredConfirmationsFromEnvDefaults(t *testing.T) {
	t.Setenv(envBitcoinMainnetRequiredConfirmations, "")
	t.Setenv(envBitcoinTestnet4RequiredConfirmations, "")

	config, err := loadBitcoinRequiredConfirmationsFromEnv()
	if err != nil {
		t.Fatalf("loadBitcoinRequiredConfirmationsFromEnv returned error: %v", err)
	}

	if got := config[value_objects.BitcoinNetworkMainnet]; got != 1 {
		t.Fatalf("unexpected mainnet confirmations: got %d", got)
	}
	if got := config[value_objects.BitcoinNetworkTestnet4]; got != 1 {
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

	if got := config[value_objects.BitcoinNetworkMainnet]; got != 6 {
		t.Fatalf("unexpected mainnet confirmations: got %d", got)
	}
	if got := config[value_objects.BitcoinNetworkTestnet4]; got != 2 {
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
