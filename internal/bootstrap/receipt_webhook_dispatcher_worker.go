package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	scheduleradapter "payrune/internal/adapters/inbound/scheduler"
	cloudflarepostgresadapter "payrune/internal/adapters/outbound/persistence/cloudflarepostgres"
	"payrune/internal/adapters/outbound/system"
	webhookadapter "payrune/internal/adapters/outbound/webhook"
	"payrune/internal/application/usecases"
	cloudflarepostgresinfra "payrune/internal/infrastructure/cloudflarepostgres"
	cloudflarewebhookinfra "payrune/internal/infrastructure/cloudflarewebhook"
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

type receiptWebhookDispatcherWorkerRequestEnvelope struct {
	Env              map[string]string `json:"env"`
	PostgresBridgeID string            `json:"postgresBridgeId"`
	ScheduledTime    string            `json:"scheduledTime"`
	Cron             string            `json:"cron"`
}

type receiptWebhookDispatcherWorkerResponseEnvelope struct {
	Output scheduleradapter.WebhookDispatcherResponse `json:"output"`
}

func HandleCloudflareReceiptWebhookDispatcherRequestJSON(ctx context.Context, payload string) (string, error) {
	var envelope receiptWebhookDispatcherWorkerRequestEnvelope
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return "", err
	}

	output, err := handleCloudflareReceiptWebhookDispatcherRequest(ctx, envelope)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(receiptWebhookDispatcherWorkerResponseEnvelope{Output: output})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func handleCloudflareReceiptWebhookDispatcherRequest(
	ctx context.Context,
	envelope receiptWebhookDispatcherWorkerRequestEnvelope,
) (scheduleradapter.WebhookDispatcherResponse, error) {
	handler, request, err := buildCloudflareReceiptWebhookDispatcherRuntime(
		envelope.Env,
		envelope.PostgresBridgeID,
	)
	if err != nil {
		return scheduleradapter.WebhookDispatcherResponse{}, err
	}

	return handler.Handle(ctx, request)
}

func buildCloudflareReceiptWebhookDispatcherRuntime(
	env map[string]string,
	postgresBridgeID string,
) (*scheduleradapter.WebhookDispatcherHandler, scheduleradapter.WebhookDispatcherRequest, error) {
	request, err := buildCloudflareReceiptWebhookDispatcherRequest(env)
	if err != nil {
		return nil, scheduleradapter.WebhookDispatcherRequest{}, err
	}

	notifierConfig, err := loadCloudflareReceiptWebhookNotifierConfig(env)
	if err != nil {
		return nil, scheduleradapter.WebhookDispatcherRequest{}, err
	}
	notifier, err := webhookadapter.NewCloudflarePaymentReceiptStatusWebhookNotifier(
		notifierConfig,
		cloudflarewebhookinfra.NewBridge(),
	)
	if err != nil {
		return nil, scheduleradapter.WebhookDispatcherRequest{}, err
	}

	unitOfWork := cloudflarepostgresadapter.NewUnitOfWork(postgresBridgeID, cloudflarepostgresinfra.NewJSBridge())
	clock := system.NewClock()
	useCase := usecases.NewRunReceiptWebhookDispatchCycleUseCase(unitOfWork, notifier, clock)

	return scheduleradapter.NewWebhookDispatcherHandler(scheduleradapter.WebhookDispatcherDependencies{
		RunReceiptWebhookDispatchCycleUseCase: useCase,
	}), request, nil
}

func buildCloudflareReceiptWebhookDispatcherRequest(
	env map[string]string,
) (scheduleradapter.WebhookDispatcherRequest, error) {
	batchSize, err := parseCloudflareReceiptWebhookDispatcherPositiveIntEnvWithDefault(
		env,
		cfEnvReceiptWebhookDispatchBatchSize,
		cfDefaultReceiptWebhookDispatchBatchSize,
	)
	if err != nil {
		return scheduleradapter.WebhookDispatcherRequest{}, err
	}
	dispatchTTL, err := parseCloudflareReceiptWebhookDispatcherDurationMapWithDefault(
		env,
		cfEnvReceiptWebhookDispatchClaimTTL,
		cfDefaultReceiptWebhookDispatchClaimTTL,
	)
	if err != nil {
		return scheduleradapter.WebhookDispatcherRequest{}, err
	}
	maxAttempts, err := parseCloudflareReceiptWebhookDispatcherPositiveInt32MapWithDefault(
		env,
		cfEnvReceiptWebhookDispatchMaxAttempts,
		cfDefaultReceiptWebhookDispatchMaxAttempts,
	)
	if err != nil {
		return scheduleradapter.WebhookDispatcherRequest{}, err
	}
	retryDelay, err := parseCloudflareReceiptWebhookDispatcherDurationMapWithDefault(
		env,
		cfEnvReceiptWebhookDispatchRetryDelay,
		cfDefaultReceiptWebhookDispatchRetryDelay,
	)
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

func loadCloudflareReceiptWebhookNotifierConfig(
	env map[string]string,
) (webhookadapter.PaymentReceiptWebhookNotifierConfig, error) {
	timeout, err := parseCloudflareReceiptWebhookDispatcherDurationMapWithDefault(
		env,
		envPaymentReceiptWebhookTimeout,
		10*time.Second,
	)
	if err != nil {
		return webhookadapter.PaymentReceiptWebhookNotifierConfig{}, err
	}
	insecureSkipVerify, err := parseCloudflareReceiptWebhookDispatcherBoolEnv(
		env,
		envPaymentReceiptWebhookInsecureSkipVerify,
	)
	if err != nil {
		return webhookadapter.PaymentReceiptWebhookNotifierConfig{}, err
	}

	return webhookadapter.PaymentReceiptWebhookNotifierConfig{
		CloudflareBinding:  cfReceiptWebhookMockBinding,
		CloudflarePath:     cfReceiptWebhookMockPath,
		Secret:             cloudflareReceiptWebhookDispatcherEnvValue(env, envPaymentReceiptWebhookSecret),
		Timeout:            timeout,
		InsecureSkipVerify: insecureSkipVerify,
	}, nil
}

func parseCloudflareReceiptWebhookDispatcherBoolEnv(
	env map[string]string,
	key string,
) (bool, error) {
	rawValue := cloudflareReceiptWebhookDispatcherEnvValue(env, key)
	if rawValue == "" {
		return false, nil
	}

	value, err := strconv.ParseBool(rawValue)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return value, nil
}

func parseCloudflareReceiptWebhookDispatcherPositiveIntEnvWithDefault(
	env map[string]string,
	key string,
	fallback int,
) (int, error) {
	rawValue := cloudflareReceiptWebhookDispatcherEnvValue(env, key)
	if rawValue == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(rawValue)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return value, nil
}

func parseCloudflareReceiptWebhookDispatcherPositiveInt32MapWithDefault(
	env map[string]string,
	key string,
	fallback int32,
) (int32, error) {
	rawValue := cloudflareReceiptWebhookDispatcherEnvValue(env, key)
	if rawValue == "" {
		return fallback, nil
	}

	parsedValue, err := strconv.ParseInt(rawValue, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be a positive integer: %w", key, err)
	}
	if parsedValue <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}

	return int32(parsedValue), nil
}

func parseCloudflareReceiptWebhookDispatcherDurationMapWithDefault(
	env map[string]string,
	key string,
	fallback time.Duration,
) (time.Duration, error) {
	rawValue := cloudflareReceiptWebhookDispatcherEnvValue(env, key)
	if rawValue == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(rawValue)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return duration, nil
}

func cloudflareReceiptWebhookDispatcherEnvValue(env map[string]string, key string) string {
	return strings.TrimSpace(env[key])
}
