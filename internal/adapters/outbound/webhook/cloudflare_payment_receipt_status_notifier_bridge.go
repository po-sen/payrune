package webhook

import (
	"context"
	"time"
)

type CloudflarePaymentReceiptStatusWebhookPostInput struct {
	Binding string
	Path    string
	Timeout time.Duration
	Headers map[string]string
	Body    []byte
}

type CloudflarePaymentReceiptStatusWebhookBridge interface {
	PostJSON(ctx context.Context, input CloudflarePaymentReceiptStatusWebhookPostInput) error
}
