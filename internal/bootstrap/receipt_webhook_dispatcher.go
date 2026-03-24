package bootstrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	scheduleradapter "payrune/internal/adapters/inbound/scheduler"
	postgresadapter "payrune/internal/adapters/outbound/persistence/postgres"
	"payrune/internal/adapters/outbound/system"
	webhookadapter "payrune/internal/adapters/outbound/webhook"
	"payrune/internal/application/usecases"
)

const (
	defaultReceiptWebhookDispatcherInterval    = 15 * time.Second
	defaultReceiptWebhookDispatcherBatchSize   = 50
	defaultReceiptWebhookDispatcherClaimTTL    = 30 * time.Second
	defaultReceiptWebhookDispatcherMaxAttempts = int32(10)
	defaultReceiptWebhookDispatcherRetryDelay  = time.Minute

	envPaymentReceiptWebhookURL                = "PAYMENT_RECEIPT_WEBHOOK_URL"
	envPaymentReceiptWebhookSecret             = "PAYMENT_RECEIPT_WEBHOOK_SECRET"
	envPaymentReceiptWebhookTimeout            = "PAYMENT_RECEIPT_WEBHOOK_TIMEOUT"
	envPaymentReceiptWebhookInsecureSkipVerify = "PAYMENT_RECEIPT_WEBHOOK_INSECURE_SKIP_VERIFY"
)

type ReceiptWebhookDispatcherConfig struct {
	Interval    time.Duration
	BatchSize   int
	ClaimTTL    time.Duration
	MaxAttempts int32
	RetryDelay  time.Duration
}

type receiptWebhookDispatcherContainer struct {
	WebhookDispatcherHandler *scheduleradapter.WebhookDispatcherHandler
	closeFn                  func() error
}

type configError struct {
	message string
}

func LoadReceiptWebhookDispatcherConfigFromEnv() (ReceiptWebhookDispatcherConfig, error) {
	return loadReceiptWebhookDispatcherConfigFromLookup(os.Getenv)
}

func RunReceiptWebhookDispatcher(ctx context.Context, config ReceiptWebhookDispatcherConfig) error {
	if config.Interval <= 0 {
		config.Interval = defaultReceiptWebhookDispatcherInterval
	}
	if config.BatchSize <= 0 {
		config.BatchSize = defaultReceiptWebhookDispatcherBatchSize
	}
	if config.ClaimTTL <= 0 {
		config.ClaimTTL = defaultReceiptWebhookDispatcherClaimTTL
	}
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = defaultReceiptWebhookDispatcherMaxAttempts
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = defaultReceiptWebhookDispatcherRetryDelay
	}

	container, err := newReceiptWebhookDispatcherContainer()
	if err != nil {
		return err
	}
	defer func() {
		_ = container.Close()
	}()

	runCycle := func() {
		output, err := container.WebhookDispatcherHandler.Handle(ctx, scheduleradapter.WebhookDispatcherRequest{
			BatchSize:   config.BatchSize,
			DispatchTTL: config.ClaimTTL,
			RetryDelay:  config.RetryDelay,
			MaxAttempts: config.MaxAttempts,
		})
		if err != nil {
			log.Printf("receipt webhook dispatch cycle failed: err=%v", err)
			return
		}

		log.Printf(
			"receipt webhook dispatch cycle complete claimed=%d sent=%d retried=%d failed=%d",
			output.ClaimedCount,
			output.SentCount,
			output.RetriedCount,
			output.FailedCount,
		)
	}

	runCycle()

	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			runCycle()
		}
	}
}

func loadReceiptWebhookDispatcherConfigFromLookup(
	lookup func(string) string,
) (ReceiptWebhookDispatcherConfig, error) {
	interval, err := parseRequiredPositiveDurationEnv(lookup, "RECEIPT_WEBHOOK_DISPATCH_INTERVAL")
	if err != nil {
		return ReceiptWebhookDispatcherConfig{}, err
	}
	batchSize, err := parseRequiredPositiveIntEnv(lookup, "RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE")
	if err != nil {
		return ReceiptWebhookDispatcherConfig{}, err
	}
	claimTTL, err := parseRequiredPositiveDurationEnv(lookup, "RECEIPT_WEBHOOK_DISPATCH_CLAIM_TTL")
	if err != nil {
		return ReceiptWebhookDispatcherConfig{}, err
	}
	maxAttempts, err := parseRequiredPositiveInt32Env(lookup, "RECEIPT_WEBHOOK_DISPATCH_MAX_ATTEMPTS")
	if err != nil {
		return ReceiptWebhookDispatcherConfig{}, err
	}
	retryDelay, err := parseRequiredPositiveDurationEnv(lookup, "RECEIPT_WEBHOOK_DISPATCH_RETRY_DELAY")
	if err != nil {
		return ReceiptWebhookDispatcherConfig{}, err
	}

	return ReceiptWebhookDispatcherConfig{
		Interval:    interval,
		BatchSize:   batchSize,
		ClaimTTL:    claimTTL,
		MaxAttempts: maxAttempts,
		RetryDelay:  retryDelay,
	}, nil
}

func newReceiptWebhookDispatcherContainer() (*receiptWebhookDispatcherContainer, error) {
	db, err := openPostgresFromEnv()
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

	return &receiptWebhookDispatcherContainer{
		WebhookDispatcherHandler: scheduleradapter.NewWebhookDispatcherHandler(
			scheduleradapter.WebhookDispatcherDependencies{
				RunReceiptWebhookDispatchCycleUseCase: useCase,
			},
		),
		closeFn: db.Close,
	}, nil
}

func (c *receiptWebhookDispatcherContainer) Close() error {
	if c.closeFn == nil {
		return nil
	}
	return c.closeFn()
}

func loadPaymentReceiptWebhookNotifierConfigFromEnv() (webhookadapter.PaymentReceiptWebhookNotifierConfig, error) {
	timeout, err := parseReceiptWebhookDispatcherPositiveDurationEnvWithDefault(
		envPaymentReceiptWebhookTimeout,
		0,
	)
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

func parseRequiredPositiveDurationEnv(lookup func(string) string, key string) (time.Duration, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		return 0, configRequiredError(key)
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, configInvalidError(key, "must be a valid duration", err)
	}
	if value <= 0 {
		return 0, configSimpleError(key, "must be greater than zero")
	}
	return value, nil
}

func parseRequiredPositiveIntEnv(lookup func(string) string, key string) (int, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		return 0, configRequiredError(key)
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, configInvalidError(key, "must be an integer", err)
	}
	if value <= 0 {
		return 0, configSimpleError(key, "must be greater than zero")
	}
	return value, nil
}

func parseRequiredPositiveInt32Env(lookup func(string) string, key string) (int32, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		return 0, configRequiredError(key)
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, configInvalidError(key, "must be a 32-bit integer", err)
	}
	if value <= 0 {
		return 0, configSimpleError(key, "must be greater than zero")
	}
	return int32(value), nil
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

func parseReceiptWebhookDispatcherPositiveDurationEnvWithDefault(
	key string,
	fallback time.Duration,
) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return parsed, nil
}

func configRequiredError(key string) error {
	return configSimpleError(key, "is required")
}

func configSimpleError(key string, message string) error {
	return &configError{message: key + " " + message}
}

func configInvalidError(key string, message string, err error) error {
	return &configError{message: key + " " + message + ": " + err.Error()}
}

func (e *configError) Error() string {
	return e.message
}
