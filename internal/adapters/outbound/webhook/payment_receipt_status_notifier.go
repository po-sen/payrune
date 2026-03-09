package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	outport "payrune/internal/application/ports/out"
)

const defaultWebhookTimeout = 10 * time.Second

type PaymentReceiptWebhookNotifierConfig struct {
	URL                string
	Secret             string
	Timeout            time.Duration
	InsecureSkipVerify bool
	Client             *http.Client
}

type paymentReceiptStatusWebhookNotifier struct {
	endpoint string
	secret   []byte
	client   *http.Client
}

func NewPaymentReceiptStatusWebhookNotifier(
	config PaymentReceiptWebhookNotifierConfig,
) (outport.PaymentReceiptStatusNotifier, error) {
	endpoint := strings.TrimSpace(config.URL)
	if endpoint == "" {
		return nil, errors.New("payment receipt webhook url is required")
	}
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("payment receipt webhook url is invalid: %w", err)
	}
	if parsedURL.Scheme != "https" {
		return nil, errors.New("payment receipt webhook url must use https")
	}
	if parsedURL.Host == "" {
		return nil, errors.New("payment receipt webhook url host is required")
	}
	secret := strings.TrimSpace(config.Secret)
	if secret == "" {
		return nil, errors.New("payment receipt webhook secret is required")
	}

	client := config.Client
	if client == nil {
		timeout := config.Timeout
		if timeout <= 0 {
			timeout = defaultWebhookTimeout
		}
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: config.InsecureSkipVerify,
		}
		client = &http.Client{
			Timeout:   timeout,
			Transport: transport,
		}
	}

	return &paymentReceiptStatusWebhookNotifier{
		endpoint: parsedURL.String(),
		secret:   []byte(secret),
		client:   client,
	}, nil
}

func (n *paymentReceiptStatusWebhookNotifier) NotifyStatusChanged(
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

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, n.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Payrune-Event", outport.PaymentReceiptStatusChangedEventType)
	request.Header.Set("X-Payrune-Event-Version", fmt.Sprintf("%d", outport.PaymentReceiptStatusChangedEventVersion))
	request.Header.Set("X-Payrune-Notification-ID", fmt.Sprintf("%d", input.NotificationID))
	request.Header.Set("X-Payrune-Signature-256", "sha256="+computeWebhookSignature(n.secret, body))

	response, err := n.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", response.StatusCode)
	}
	return nil
}

type paymentReceiptStatusChangedPayload struct {
	EventType             string    `json:"eventType"`
	EventVersion          int       `json:"eventVersion"`
	NotificationID        int64     `json:"notificationId"`
	PaymentAddressID      int64     `json:"paymentAddressId"`
	CustomerReference     string    `json:"customerReference,omitempty"`
	PreviousStatus        string    `json:"previousStatus"`
	CurrentStatus         string    `json:"currentStatus"`
	ObservedTotalMinor    int64     `json:"observedTotalMinor"`
	ConfirmedTotalMinor   int64     `json:"confirmedTotalMinor"`
	UnconfirmedTotalMinor int64     `json:"unconfirmedTotalMinor"`
	StatusChangedAt       time.Time `json:"statusChangedAt"`
}

func computeWebhookSignature(secret []byte, body []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
