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
	policyadapter "payrune/internal/adapters/outbound/policy"
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

	envReceiptWebhookDispatchInterval    = "RECEIPT_WEBHOOK_DISPATCH_INTERVAL"
	envReceiptWebhookDispatchBatchSize   = "RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE"
	envReceiptWebhookDispatchClaimTTL    = "RECEIPT_WEBHOOK_DISPATCH_CLAIM_TTL"
	envReceiptWebhookDispatchMaxAttempts = "RECEIPT_WEBHOOK_DISPATCH_MAX_ATTEMPTS"
	envReceiptWebhookDispatchRetryDelay  = "RECEIPT_WEBHOOK_DISPATCH_RETRY_DELAY"

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

type receiptWebhookDispatchSettings struct {
	BatchSize   int
	DispatchTTL time.Duration
	MaxAttempts int32
	RetryDelay  time.Duration
}

type receiptWebhookDispatchDefaults struct {
	BatchSize   int
	DispatchTTL time.Duration
	MaxAttempts int32
	RetryDelay  time.Duration
}

type receiptWebhookNotifierSettings struct {
	Secret             string
	Timeout            time.Duration
	InsecureSkipVerify bool
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
	interval, err := parseReceiptWebhookDispatcherPositiveDurationLookup(
		lookup,
		envReceiptWebhookDispatchInterval,
		0,
		true,
	)
	if err != nil {
		return ReceiptWebhookDispatcherConfig{}, err
	}
	dispatchSettings, err := loadReceiptWebhookDispatchSettingsFromLookup(
		lookup,
		receiptWebhookDispatchDefaults{},
		true,
	)
	if err != nil {
		return ReceiptWebhookDispatcherConfig{}, err
	}

	return ReceiptWebhookDispatcherConfig{
		Interval:    interval,
		BatchSize:   dispatchSettings.BatchSize,
		ClaimTTL:    dispatchSettings.DispatchTTL,
		MaxAttempts: dispatchSettings.MaxAttempts,
		RetryDelay:  dispatchSettings.RetryDelay,
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
	addressPolicyReader := policyadapter.NewAddressPolicyReader(
		buildAddressIssuancePolicies(os.Getenv, nil),
	)

	unitOfWork := postgresadapter.NewUnitOfWork(db)
	clock := system.NewClock()
	useCase := usecases.NewRunReceiptWebhookDispatchCycleUseCase(unitOfWork, addressPolicyReader, notifier, clock)

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
	settings, err := loadReceiptWebhookNotifierSettingsFromLookup(os.Getenv, 0)
	if err != nil {
		return webhookadapter.PaymentReceiptWebhookNotifierConfig{}, err
	}

	return webhookadapter.PaymentReceiptWebhookNotifierConfig{
		URL:                strings.TrimSpace(os.Getenv(envPaymentReceiptWebhookURL)),
		Secret:             settings.Secret,
		Timeout:            settings.Timeout,
		InsecureSkipVerify: settings.InsecureSkipVerify,
	}, nil
}

func loadReceiptWebhookDispatchSettingsFromLookup(
	lookup func(string) string,
	defaults receiptWebhookDispatchDefaults,
	required bool,
) (receiptWebhookDispatchSettings, error) {
	batchSize, err := parseReceiptWebhookDispatcherPositiveIntLookup(
		lookup,
		envReceiptWebhookDispatchBatchSize,
		defaults.BatchSize,
		required,
	)
	if err != nil {
		return receiptWebhookDispatchSettings{}, err
	}
	dispatchTTL, err := parseReceiptWebhookDispatcherPositiveDurationLookup(
		lookup,
		envReceiptWebhookDispatchClaimTTL,
		defaults.DispatchTTL,
		required,
	)
	if err != nil {
		return receiptWebhookDispatchSettings{}, err
	}
	maxAttempts, err := parseReceiptWebhookDispatcherPositiveInt32Lookup(
		lookup,
		envReceiptWebhookDispatchMaxAttempts,
		defaults.MaxAttempts,
		required,
	)
	if err != nil {
		return receiptWebhookDispatchSettings{}, err
	}
	retryDelay, err := parseReceiptWebhookDispatcherPositiveDurationLookup(
		lookup,
		envReceiptWebhookDispatchRetryDelay,
		defaults.RetryDelay,
		required,
	)
	if err != nil {
		return receiptWebhookDispatchSettings{}, err
	}

	return receiptWebhookDispatchSettings{
		BatchSize:   batchSize,
		DispatchTTL: dispatchTTL,
		MaxAttempts: maxAttempts,
		RetryDelay:  retryDelay,
	}, nil
}

func loadReceiptWebhookNotifierSettingsFromLookup(
	lookup func(string) string,
	timeoutFallback time.Duration,
) (receiptWebhookNotifierSettings, error) {
	timeout, err := parseReceiptWebhookDispatcherPositiveDurationLookup(
		lookup,
		envPaymentReceiptWebhookTimeout,
		timeoutFallback,
		false,
	)
	if err != nil {
		return receiptWebhookNotifierSettings{}, err
	}
	insecureSkipVerify, err := parseReceiptWebhookDispatcherBoolLookup(
		lookup,
		envPaymentReceiptWebhookInsecureSkipVerify,
	)
	if err != nil {
		return receiptWebhookNotifierSettings{}, err
	}

	return receiptWebhookNotifierSettings{
		Secret:             lookup(envPaymentReceiptWebhookSecret),
		Timeout:            timeout,
		InsecureSkipVerify: insecureSkipVerify,
	}, nil
}

func parseReceiptWebhookDispatcherPositiveDurationLookup(
	lookup func(string) string,
	key string,
	fallback time.Duration,
	required bool,
) (time.Duration, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		if required {
			return 0, configRequiredError(key)
		}
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		if required {
			return 0, configInvalidError(key, "must be a valid duration", err)
		}
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	if value <= 0 {
		if required {
			return 0, configSimpleError(key, "must be greater than zero")
		}
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return value, nil
}

func parseReceiptWebhookDispatcherPositiveIntLookup(
	lookup func(string) string,
	key string,
	fallback int,
	required bool,
) (int, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		if required {
			return 0, configRequiredError(key)
		}
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		if required {
			return 0, configInvalidError(key, "must be an integer", err)
		}
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	if value <= 0 {
		if required {
			return 0, configSimpleError(key, "must be greater than zero")
		}
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return value, nil
}

func parseReceiptWebhookDispatcherPositiveInt32Lookup(
	lookup func(string) string,
	key string,
	fallback int32,
	required bool,
) (int32, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		if required {
			return 0, configRequiredError(key)
		}
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		if required {
			return 0, configInvalidError(key, "must be a 32-bit integer", err)
		}
		return 0, fmt.Errorf("%s must be a positive integer: %w", key, err)
	}
	if value <= 0 {
		if required {
			return 0, configSimpleError(key, "must be greater than zero")
		}
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return int32(value), nil
}

func parseReceiptWebhookDispatcherBoolLookup(lookup func(string) string, key string) (bool, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		return false, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return value, nil
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
