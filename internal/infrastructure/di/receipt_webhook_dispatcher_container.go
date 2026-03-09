package di

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	postgresadapter "payrune/internal/adapters/outbound/persistence/postgres"
	"payrune/internal/adapters/outbound/system"
	webhookadapter "payrune/internal/adapters/outbound/webhook"
	inport "payrune/internal/application/ports/in"
	"payrune/internal/application/use_cases"
)

type ReceiptWebhookDispatcherContainer struct {
	RunReceiptWebhookDispatchCycleUseCase inport.RunReceiptWebhookDispatchCycleUseCase
	closeFn                               func() error
}

const (
	envPaymentReceiptWebhookURL                = "PAYMENT_RECEIPT_WEBHOOK_URL"
	envPaymentReceiptWebhookSecret             = "PAYMENT_RECEIPT_WEBHOOK_SECRET"
	envPaymentReceiptWebhookTimeout            = "PAYMENT_RECEIPT_WEBHOOK_TIMEOUT"
	envPaymentReceiptWebhookInsecureSkipVerify = "PAYMENT_RECEIPT_WEBHOOK_INSECURE_SKIP_VERIFY"
)

func NewReceiptWebhookDispatcherContainer() (*ReceiptWebhookDispatcherContainer, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database connection: %w", err)
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
	useCase := use_cases.NewRunReceiptWebhookDispatchCycleUseCase(unitOfWork, notifier, clock)

	return &ReceiptWebhookDispatcherContainer{
		RunReceiptWebhookDispatchCycleUseCase: useCase,
		closeFn:                               db.Close,
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
