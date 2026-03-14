package di

import (
	"fmt"
	"strconv"
	"time"

	scheduleradapter "payrune/internal/adapters/inbound/scheduler"
	cloudflarepostgres "payrune/internal/adapters/outbound/persistence/cloudflarepostgres"
	"payrune/internal/adapters/outbound/system"
	webhookadapter "payrune/internal/adapters/outbound/webhook"
	"payrune/internal/application/usecases"
)

const (
	cfEnvReceiptWebhookDispatchBatchSize   = "RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE"
	cfEnvReceiptWebhookDispatchClaimTTL    = "RECEIPT_WEBHOOK_DISPATCH_CLAIM_TTL"
	cfEnvReceiptWebhookDispatchMaxAttempts = "RECEIPT_WEBHOOK_DISPATCH_MAX_ATTEMPTS"
	cfEnvReceiptWebhookDispatchRetryDelay  = "RECEIPT_WEBHOOK_DISPATCH_RETRY_DELAY"

	cfDefaultReceiptWebhookDispatchBatchSize   = 50
	cfDefaultReceiptWebhookDispatchClaimTTL    = 30 * time.Second
	cfDefaultReceiptWebhookDispatchMaxAttempts = int32(10)
	cfDefaultReceiptWebhookDispatchRetryDelay  = time.Minute

	cfReceiptWebhookMockBinding = "RECEIPT_WEBHOOK_MOCK"
	cfReceiptWebhookMockPath    = "/receipt-status"
)

func BuildCloudflareWebhookDispatcherRuntime(
	env map[string]string,
	postgresBridgeID string,
) (*scheduleradapter.WebhookDispatcherHandler, scheduleradapter.WebhookDispatcherRequest, error) {
	request, err := buildCloudflareWebhookDispatcherRequest(env)
	if err != nil {
		return nil, scheduleradapter.WebhookDispatcherRequest{}, err
	}

	notifierConfig, err := loadCloudflarePaymentReceiptWebhookNotifierConfig(env)
	if err != nil {
		return nil, scheduleradapter.WebhookDispatcherRequest{}, err
	}
	notifier, err := webhookadapter.NewCloudflarePaymentReceiptStatusWebhookNotifier(
		notifierConfig,
		webhookadapter.NewCloudflarePaymentReceiptStatusWebhookBridge(),
	)
	if err != nil {
		return nil, scheduleradapter.WebhookDispatcherRequest{}, err
	}

	unitOfWork := cloudflarepostgres.NewUnitOfWork(postgresBridgeID, cloudflarepostgres.NewJSBridge())
	clock := system.NewClock()
	useCase := usecases.NewRunReceiptWebhookDispatchCycleUseCase(unitOfWork, notifier, clock)

	return scheduleradapter.NewWebhookDispatcherHandler(scheduleradapter.WebhookDispatcherDependencies{
		RunReceiptWebhookDispatchCycleUseCase: useCase,
	}), request, nil
}

func buildCloudflareWebhookDispatcherRequest(env map[string]string) (scheduleradapter.WebhookDispatcherRequest, error) {
	batchSize, err := parsePositiveIntEnvWithDefault(env, cfEnvReceiptWebhookDispatchBatchSize, cfDefaultReceiptWebhookDispatchBatchSize)
	if err != nil {
		return scheduleradapter.WebhookDispatcherRequest{}, err
	}
	dispatchTTL, err := parseDurationMapWithDefault(env, cfEnvReceiptWebhookDispatchClaimTTL, cfDefaultReceiptWebhookDispatchClaimTTL)
	if err != nil {
		return scheduleradapter.WebhookDispatcherRequest{}, err
	}
	maxAttempts, err := parsePositiveInt32MapWithDefault(env, cfEnvReceiptWebhookDispatchMaxAttempts, cfDefaultReceiptWebhookDispatchMaxAttempts)
	if err != nil {
		return scheduleradapter.WebhookDispatcherRequest{}, err
	}
	retryDelay, err := parseDurationMapWithDefault(env, cfEnvReceiptWebhookDispatchRetryDelay, cfDefaultReceiptWebhookDispatchRetryDelay)
	if err != nil {
		return scheduleradapter.WebhookDispatcherRequest{}, err
	}

	return scheduleradapter.WebhookDispatcherRequest{
		BatchSize:   batchSize,
		DispatchTTL: dispatchTTL,
		RetryDelay:  retryDelay,
		MaxAttempts: maxAttempts,
	}, nil
}

func loadCloudflarePaymentReceiptWebhookNotifierConfig(
	env map[string]string,
) (webhookadapter.PaymentReceiptWebhookNotifierConfig, error) {
	timeout, err := parseDurationMapWithDefault(env, envPaymentReceiptWebhookTimeout, 10*time.Second)
	if err != nil {
		return webhookadapter.PaymentReceiptWebhookNotifierConfig{}, err
	}
	insecureSkipVerify, err := parseCloudflareBoolEnv(env, envPaymentReceiptWebhookInsecureSkipVerify)
	if err != nil {
		return webhookadapter.PaymentReceiptWebhookNotifierConfig{}, err
	}

	return webhookadapter.PaymentReceiptWebhookNotifierConfig{
		CloudflareBinding:  cfReceiptWebhookMockBinding,
		CloudflarePath:     cfReceiptWebhookMockPath,
		Secret:             envMapValue(env, envPaymentReceiptWebhookSecret),
		Timeout:            timeout,
		InsecureSkipVerify: insecureSkipVerify,
	}, nil
}

func parseCloudflareBoolEnv(env map[string]string, key string) (bool, error) {
	rawValue := envMapValue(env, key)
	if rawValue == "" {
		return false, nil
	}

	value, err := strconv.ParseBool(rawValue)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return value, nil
}
