package di

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	scheduleradapter "payrune/internal/adapters/inbound/scheduler"
	postgresadapter "payrune/internal/adapters/outbound/persistence/postgres"
	"payrune/internal/adapters/outbound/system"
	webhookadapter "payrune/internal/adapters/outbound/webhook"
	"payrune/internal/application/usecases"
	postgresdriver "payrune/internal/infrastructure/drivers/postgres"
)

type ReceiptWebhookDispatcherContainer struct {
	WebhookDispatcherHandler *scheduleradapter.WebhookDispatcherHandler
	closeFn                  func() error
}

const (
	envPaymentReceiptWebhookURL                = "PAYMENT_RECEIPT_WEBHOOK_URL"
	envPaymentReceiptWebhookSecret             = "PAYMENT_RECEIPT_WEBHOOK_SECRET"
	envPaymentReceiptWebhookTimeout            = "PAYMENT_RECEIPT_WEBHOOK_TIMEOUT"
	envPaymentReceiptWebhookInsecureSkipVerify = "PAYMENT_RECEIPT_WEBHOOK_INSECURE_SKIP_VERIFY"
)

func NewReceiptWebhookDispatcherContainer() (*ReceiptWebhookDispatcherContainer, error) {
	db, err := postgresdriver.OpenFromEnv()
	if err != nil {
		return nil, err
	}

	notifierConfig, err := loadPaymentReceiptWebhookNotifierConfigFromEnv()
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	notifier, err := webhookadapter.NewPaymentReceiptStatusWebhookNotifier(notifierConfig)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	unitOfWork := postgresadapter.NewUnitOfWork(db)
	clock := system.NewClock()
	useCase := usecases.NewRunReceiptWebhookDispatchCycleUseCase(unitOfWork, notifier, clock)

	return &ReceiptWebhookDispatcherContainer{
		WebhookDispatcherHandler: scheduleradapter.NewWebhookDispatcherHandler(
			scheduleradapter.WebhookDispatcherDependencies{
				RunReceiptWebhookDispatchCycleUseCase: useCase,
			},
		),
		closeFn: db.Close,
	}, nil
}

func (c *ReceiptWebhookDispatcherContainer) Close() error {
	if c.closeFn == nil {
		return nil
	}
	return c.closeFn()
}

func loadPaymentReceiptWebhookNotifierConfigFromEnv() (webhookadapter.PaymentReceiptWebhookNotifierConfig, error) {
	timeout, err := parsePositiveDurationEnvWithDefault(envPaymentReceiptWebhookTimeout, 0)
	if err != nil {
		return webhookadapter.PaymentReceiptWebhookNotifierConfig{}, err
	}
	insecureSkipVerify, err := parseBoolEnv(envPaymentReceiptWebhookInsecureSkipVerify)
	if err != nil {
		return webhookadapter.PaymentReceiptWebhookNotifierConfig{}, err
	}

	return webhookadapter.PaymentReceiptWebhookNotifierConfig{
		URL:                strings.TrimSpace(os.Getenv(envPaymentReceiptWebhookURL)),
		Secret:             os.Getenv(envPaymentReceiptWebhookSecret),
		Timeout:            timeout,
		InsecureSkipVerify: insecureSkipVerify,
	}, nil
}

func parseBoolEnv(key string) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return false, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return value, nil
}
