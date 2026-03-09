package di

import (
	"testing"
	"time"

	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
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

func TestLoadReceiptTrackingLifecyclePolicyFromEnvDefaults(t *testing.T) {
	t.Setenv(envPaymentReceiptPaidUnconfirmedExpiryExtension, "")

	policy, err := loadReceiptTrackingLifecyclePolicyFromEnv()
	if err != nil {
		t.Fatalf("loadReceiptTrackingLifecyclePolicyFromEnv returned error: %v", err)
	}
	updated, err := policy.ApplyObservation(
		newPollerLifecycleTestTracking(t),
		value_objects.PaymentReceiptObservation{
			ObservedTotalMinor:    1000,
			ConfirmedTotalMinor:   0,
			UnconfirmedTotalMinor: 1000,
			ConflictTotalMinor:    0,
			LatestBlockHeight:     10,
		},
		time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("ApplyObservation returned error: %v", err)
	}
	expectedExpiresAt := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	if updated.ExpiresAt == nil || !updated.ExpiresAt.Equal(expectedExpiresAt) {
		t.Fatalf("unexpected default paid unconfirmed extension: got %v", updated.ExpiresAt)
	}
}

func TestLoadReceiptTrackingLifecyclePolicyFromEnvCustom(t *testing.T) {
	t.Setenv(envPaymentReceiptPaidUnconfirmedExpiryExtension, "240h")

	policy, err := loadReceiptTrackingLifecyclePolicyFromEnv()
	if err != nil {
		t.Fatalf("loadReceiptTrackingLifecyclePolicyFromEnv returned error: %v", err)
	}
	updated, err := policy.ApplyObservation(
		newPollerLifecycleTestTracking(t),
		value_objects.PaymentReceiptObservation{
			ObservedTotalMinor:    1000,
			ConfirmedTotalMinor:   0,
			UnconfirmedTotalMinor: 1000,
			ConflictTotalMinor:    0,
			LatestBlockHeight:     10,
		},
		time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("ApplyObservation returned error: %v", err)
	}
	expectedExpiresAt := time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC)
	if updated.ExpiresAt == nil || !updated.ExpiresAt.Equal(expectedExpiresAt) {
		t.Fatalf("unexpected paid unconfirmed extension: got %v", updated.ExpiresAt)
	}
}

func TestLoadReceiptTrackingLifecyclePolicyFromEnvInvalid(t *testing.T) {
	t.Setenv(envPaymentReceiptPaidUnconfirmedExpiryExtension, "bad")

	_, err := loadReceiptTrackingLifecyclePolicyFromEnv()
	if err == nil {
		t.Fatal("expected parse error for paid unconfirmed extension")
	}
}

func TestLoadReceiptTrackingLifecyclePolicyFromEnvNonPositive(t *testing.T) {
	t.Setenv(envPaymentReceiptPaidUnconfirmedExpiryExtension, "0s")

	_, err := loadReceiptTrackingLifecyclePolicyFromEnv()
	if err == nil {
		t.Fatal("expected validation error for non-positive paid unconfirmed extension")
	}
}

func newPollerLifecycleTestTracking(t *testing.T) entities.PaymentReceiptTracking {
	t.Helper()

	tracking, err := entities.NewPaymentReceiptTracking(
		1,
		"bitcoin-testnet4-native-segwit",
		value_objects.ChainIDBitcoin,
		value_objects.NetworkID(value_objects.BitcoinNetworkTestnet4),
		"tb1qpollerpolicy",
		time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC),
		1000,
		1,
	)
	if err != nil {
		t.Fatalf("NewPaymentReceiptTracking returned error: %v", err)
	}
	expiresAt := time.Date(2026, 3, 7, 11, 0, 0, 0, time.UTC)
	tracking.ExpiresAt = &expiresAt
	return tracking
}
