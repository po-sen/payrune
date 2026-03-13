package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	outport "payrune/internal/application/ports/outbound"
)

type cloudflarePaymentReceiptStatusWebhookNotifier struct {
	binding string
	path    string
	secret  []byte
	timeout time.Duration
	bridge  CloudflarePaymentReceiptStatusWebhookBridge
}

func NewCloudflarePaymentReceiptStatusWebhookNotifier(
	config PaymentReceiptWebhookNotifierConfig,
	bridge CloudflarePaymentReceiptStatusWebhookBridge,
) (outport.PaymentReceiptStatusNotifier, error) {
	if bridge == nil {
		return nil, errors.New("cloudflare payment receipt webhook bridge is not configured")
	}
	if config.InsecureSkipVerify {
		return nil, errors.New("payment receipt webhook insecure skip verify is not supported in cloudflare worker runtime")
	}

	binding := strings.TrimSpace(config.CloudflareBinding)
	path := strings.TrimSpace(config.CloudflarePath)
	if binding == "" {
		return nil, errors.New("payment receipt webhook cloudflare binding is required")
	}
	if path == "" {
		path = "/receipt-status"
	}
	if !strings.HasPrefix(path, "/") {
		return nil, errors.New("payment receipt webhook cloudflare path must start with /")
	}

	secret := strings.TrimSpace(config.Secret)
	if secret == "" {
		return nil, errors.New("payment receipt webhook secret is required")
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultWebhookTimeout
	}

	return &cloudflarePaymentReceiptStatusWebhookNotifier{
		binding: binding,
		path:    path,
		secret:  []byte(secret),
		timeout: timeout,
		bridge:  bridge,
	}, nil
}

func (n *cloudflarePaymentReceiptStatusWebhookNotifier) NotifyStatusChanged(
	ctx context.Context,
	input outport.NotifyPaymentReceiptStatusChangedInput,
) error {
	body, err := json.Marshal(paymentReceiptStatusChangedPayload{
		EventType:             outport.PaymentReceiptStatusChangedEventType,
		EventVersion:          outport.PaymentReceiptStatusChangedEventVersion,
		NotificationID:        input.NotificationID,
		PaymentAddressID:      input.PaymentAddressID,
		CustomerReference:     input.CustomerReference,
		PreviousStatus:        input.PreviousStatus,
		CurrentStatus:         input.CurrentStatus,
		ObservedTotalMinor:    input.ObservedTotalMinor,
		ConfirmedTotalMinor:   input.ConfirmedTotalMinor,
		UnconfirmedTotalMinor: input.UnconfirmedTotalMinor,
		StatusChangedAt:       input.StatusChangedAt.UTC(),
	})
	if err != nil {
		return err
	}

	return n.bridge.PostJSON(ctx, CloudflarePaymentReceiptStatusWebhookPostInput{
		Binding: n.binding,
		Path:    n.path,
		Timeout: n.timeout,
		Headers: map[string]string{
			"Content-Type":              "application/json",
			"X-Payrune-Event":           outport.PaymentReceiptStatusChangedEventType,
			"X-Payrune-Event-Version":   fmt.Sprintf("%d", outport.PaymentReceiptStatusChangedEventVersion),
			"X-Payrune-Notification-ID": fmt.Sprintf("%d", input.NotificationID),
			"X-Payrune-Signature-256":   "sha256=" + computeWebhookSignature(n.secret, body),
		},
		Body: body,
	})
}
