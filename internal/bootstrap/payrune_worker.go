package bootstrap

import (
	"context"
	"fmt"
	"strings"
)

const (
	workerOperationAPI               = "api"
	workerOperationPoller            = "poller"
	workerOperationWebhookDispatcher = "webhook_dispatcher"
)

func DispatchCloudflareWorkerOperationJSON(
	ctx context.Context,
	operation string,
	payload string,
) (string, error) {
	switch strings.TrimSpace(operation) {
	case workerOperationAPI:
		return HandleCloudflareAPIRequestJSON(ctx, payload)
	case workerOperationPoller:
		return HandleCloudflarePollerRequestJSON(ctx, payload)
	case workerOperationWebhookDispatcher:
		return HandleCloudflareReceiptWebhookDispatcherRequestJSON(ctx, payload)
	default:
		return "", fmt.Errorf("unsupported payrune worker operation: %s", operation)
	}
}
