package main

import "testing"

func TestLoadReceiptWebhookDispatchConfigFromEnvSuccess(t *testing.T) {
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_INTERVAL", "15s")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE", "50")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_CLAIM_TTL", "30s")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_MAX_ATTEMPTS", "10")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_RETRY_DELAY", "1m")

	config, err := loadReceiptWebhookDispatchConfigFromEnv()
	if err != nil {
		t.Fatalf("loadReceiptWebhookDispatchConfigFromEnv returned error: %v", err)
	}
	if config.BatchSize != 50 || config.MaxAttempts != 10 {
		t.Fatalf("unexpected config: %+v", config)
	}
}

func TestLoadReceiptWebhookDispatchConfigFromEnvRequiresValues(t *testing.T) {
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_INTERVAL", "")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE", "50")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_CLAIM_TTL", "30s")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_MAX_ATTEMPTS", "10")
	t.Setenv("RECEIPT_WEBHOOK_DISPATCH_RETRY_DELAY", "1m")

	_, err := loadReceiptWebhookDispatchConfigFromEnv()
	if err == nil {
		t.Fatal("expected missing env error")
	}
}
