package bootstrap

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestHandleCloudflareReceiptWebhookDispatcherRequestJSONInvalidJSON(t *testing.T) {
	_, err := HandleCloudflareReceiptWebhookDispatcherRequestJSON(context.Background(), "{")
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestHandleCloudflareReceiptWebhookDispatcherRequestJSONValidationError(t *testing.T) {
	payload, err := json.Marshal(receiptWebhookDispatcherWorkerRequestEnvelope{
		Env: map[string]string{
			envReceiptWebhookDispatchBatchSize: "bad",
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	_, err = HandleCloudflareReceiptWebhookDispatcherRequestJSON(context.Background(), string(payload))
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), envReceiptWebhookDispatchBatchSize) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildCloudflareReceiptWebhookDispatcherRequestDefaults(t *testing.T) {
	request, err := buildCloudflareReceiptWebhookDispatcherRequest(map[string]string{})
	if err != nil {
		t.Fatalf("buildCloudflareReceiptWebhookDispatcherRequest returned error: %v", err)
	}

	if request.BatchSize != cfDefaultReceiptWebhookDispatchBatchSize {
		t.Fatalf("unexpected batch size: got %d", request.BatchSize)
	}
	if request.DispatchTTL != cfDefaultReceiptWebhookDispatchClaimTTL {
		t.Fatalf("unexpected dispatch ttl: got %s", request.DispatchTTL)
	}
	if request.MaxAttempts != cfDefaultReceiptWebhookDispatchMaxAttempts {
		t.Fatalf("unexpected max attempts: got %d", request.MaxAttempts)
	}
	if request.RetryDelay != cfDefaultReceiptWebhookDispatchRetryDelay {
		t.Fatalf("unexpected retry delay: got %s", request.RetryDelay)
	}
}

func TestLoadCloudflareReceiptWebhookNotifierConfigDefaults(t *testing.T) {
	config, err := loadCloudflareReceiptWebhookNotifierConfig(map[string]string{})
	if err != nil {
		t.Fatalf("loadCloudflareReceiptWebhookNotifierConfig returned error: %v", err)
	}

	if config.CloudflareBinding != cfReceiptWebhookMockBinding {
		t.Fatalf("unexpected binding: got %q", config.CloudflareBinding)
	}
	if config.CloudflarePath != cfReceiptWebhookMockPath {
		t.Fatalf("unexpected path: got %q", config.CloudflarePath)
	}
	if config.Timeout != 10*time.Second {
		t.Fatalf("unexpected timeout: got %s", config.Timeout)
	}
	if config.InsecureSkipVerify {
		t.Fatal("expected insecure skip verify to default to false")
	}
}

func TestLoadCloudflareReceiptWebhookNotifierConfigInvalidBool(t *testing.T) {
	_, err := loadCloudflareReceiptWebhookNotifierConfig(map[string]string{
		envPaymentReceiptWebhookInsecureSkipVerify: "not-bool",
	})
	if err == nil {
		t.Fatal("expected invalid bool error")
	}
}
