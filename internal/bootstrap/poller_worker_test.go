package bootstrap

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestHandleCloudflarePollerRequestJSONInvalidJSON(t *testing.T) {
	_, err := HandleCloudflarePollerRequestJSON(context.Background(), "{")
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestHandleCloudflarePollerRequestJSONValidationError(t *testing.T) {
	payload, err := json.Marshal(pollerWorkerRequestEnvelope{
		Env: map[string]string{
			envPollNetwork: "mainnet",
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	_, err = HandleCloudflarePollerRequestJSON(context.Background(), string(payload))
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "POLL_CHAIN is required when POLL_NETWORK is set") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildCloudflarePollerRequestDefaults(t *testing.T) {
	request, err := buildCloudflarePollerRequest(map[string]string{})
	if err != nil {
		t.Fatalf("buildCloudflarePollerRequest returned error: %v", err)
	}

	if request.BatchSize != cloudflarePollerDefaultBatchSize {
		t.Fatalf("unexpected batch size: got %d", request.BatchSize)
	}
	if request.RescheduleInterval != cloudflarePollerDefaultRescheduleInterval {
		t.Fatalf("unexpected reschedule interval: got %s", request.RescheduleInterval)
	}
	if request.ClaimTTL != cloudflarePollerDefaultClaimTTL {
		t.Fatalf("unexpected claim ttl: got %s", request.ClaimTTL)
	}
	if request.Chain != "" {
		t.Fatalf("expected empty chain, got %q", request.Chain)
	}
	if request.Network != "" {
		t.Fatalf("expected empty network, got %q", request.Network)
	}
}

func TestBuildCloudflarePollerRequestCustomValues(t *testing.T) {
	request, err := buildCloudflarePollerRequest(map[string]string{
		envPollBatchSize:          "7",
		envPollRescheduleInterval: "3m",
		envPollClaimTTL:           "45s",
		envPollChain:              "bitcoin",
		envPollNetwork:            "mainnet",
	})
	if err != nil {
		t.Fatalf("buildCloudflarePollerRequest returned error: %v", err)
	}

	if request.BatchSize != 7 {
		t.Fatalf("unexpected batch size: got %d", request.BatchSize)
	}
	if request.RescheduleInterval != 3*time.Minute {
		t.Fatalf("unexpected reschedule interval: got %s", request.RescheduleInterval)
	}
	if request.ClaimTTL != 45*time.Second {
		t.Fatalf("unexpected claim ttl: got %s", request.ClaimTTL)
	}
	if request.Chain != "bitcoin" {
		t.Fatalf("unexpected chain: got %q", request.Chain)
	}
	if request.Network != "mainnet" {
		t.Fatalf("unexpected network: got %q", request.Network)
	}
}

func TestBuildCloudflarePollerRequestRequiresChainWhenNetworkSet(t *testing.T) {
	_, err := buildCloudflarePollerRequest(map[string]string{
		envPollNetwork: "mainnet",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
