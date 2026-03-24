package bootstrap

import (
	"testing"
	"time"
)

func TestLoadReceiptWebhookDispatcherConfigFromEnvSuccess(t *testing.T) {
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_INTERVAL", "15s")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE", "50")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_CLAIM_TTL", "30s")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_MAX_ATTEMPTS", "10")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_RETRY_DELAY", "1m")

	config, err := LoadReceiptWebhookDispatcherConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadReceiptWebhookDispatcherConfigFromEnv returned error: %v", err)
	}
	if config.BatchSize != 50 || config.MaxAttempts != 10 {
		t.Fatalf("unexpected config: %+v", config)
	}
}

func TestLoadReceiptWebhookDispatcherConfigFromEnvRequiresValues(t *testing.T) {
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_INTERVAL", "")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE", "50")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_CLAIM_TTL", "30s")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_MAX_ATTEMPTS", "10")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_RETRY_DELAY", "1m")

	_, err := LoadReceiptWebhookDispatcherConfigFromEnv()
	if err == nil {
		t.Fatal("expected missing env error")
	}
}

func TestLoadPaymentReceiptWebhookNotifierConfigFromEnv(t *testing.T) {
	t.Setenv(envPaymentReceiptWebhookURL, "https://wallet.example.com/webhook")
	t.Setenv(envPaymentReceiptWebhookSecret, "secret")
	t.Setenv(envPaymentReceiptWebhookTimeout, "12s")
	t.Setenv(envPaymentReceiptWebhookInsecureSkipVerify, "true")

	config, err := loadPaymentReceiptWebhookNotifierConfigFromEnv()
	if err != nil {
		t.Fatalf("loadPaymentReceiptWebhookNotifierConfigFromEnv returned error: %v", err)
	}
	if config.URL != "https://wallet.example.com/webhook" {
		t.Fatalf("unexpected url: got %q", config.URL)
	}
	if config.Secret != "secret" {
		t.Fatalf("unexpected secret: got %q", config.Secret)
	}
	if config.Timeout != 12*time.Second {
		t.Fatalf("unexpected timeout: got %s", config.Timeout)
	}
	if !config.InsecureSkipVerify {
		t.Fatal("expected insecure skip verify to be true")
	}
}

func TestLoadPaymentReceiptWebhookNotifierConfigFromEnvInvalidTimeout(t *testing.T) {
	t.Setenv(envPaymentReceiptWebhookURL, "https://wallet.example.com/webhook")
	t.Setenv(envPaymentReceiptWebhookSecret, "secret")
	t.Setenv(envPaymentReceiptWebhookTimeout, "bad")
	t.Setenv(envPaymentReceiptWebhookInsecureSkipVerify, "")

	_, err := loadPaymentReceiptWebhookNotifierConfigFromEnv()
	if err == nil {
		t.Fatal("expected invalid timeout error")
	}
}

func TestLoadPaymentReceiptWebhookNotifierConfigFromEnvInvalidBool(t *testing.T) {
	t.Setenv(envPaymentReceiptWebhookURL, "https://wallet.example.com/webhook")
	t.Setenv(envPaymentReceiptWebhookSecret, "secret")
	t.Setenv(envPaymentReceiptWebhookTimeout, "12s")
	t.Setenv(envPaymentReceiptWebhookInsecureSkipVerify, "not-bool")

	_, err := loadPaymentReceiptWebhookNotifierConfigFromEnv()
	if err == nil {
		t.Fatal("expected invalid bool error")
	}
}
