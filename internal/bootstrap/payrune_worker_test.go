package bootstrap

import (
	"context"
	"testing"
)

func TestDispatchCloudflareWorkerOperationJSONRejectsUnsupportedOperation(t *testing.T) {
	_, err := DispatchCloudflareWorkerOperationJSON(context.Background(), "unknown", "{}")
	if err == nil {
		t.Fatal("expected unsupported operation error")
	}
}
