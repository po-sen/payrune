package bootstrap

import (
	"context"
	"encoding/json"
	"time"

	scheduleradapter "payrune/internal/adapters/inbound/scheduler"
	cloudflarepostgresadapter "payrune/internal/adapters/outbound/persistence/cloudflarepostgres"
	policyadapter "payrune/internal/adapters/outbound/policy"
	"payrune/internal/adapters/outbound/system"
	webhookadapter "payrune/internal/adapters/outbound/webhook"
	"payrune/internal/application/usecases"
	cloudflarepostgresinfra "payrune/internal/infrastructure/cloudflarepostgres"
	cloudflarewebhookinfra "payrune/internal/infrastructure/cloudflarewebhook"
)

const (
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
	addressPolicyReader := policyadapter.NewAddressPolicyReader(
		buildAddressIssuancePolicies(func(key string) string {
			return cloudflareAPIEnvValue(env, key)
		}, nil),
	)

	unitOfWork := cloudflarepostgresadapter.NewUnitOfWork(postgresBridgeID, cloudflarepostgresinfra.NewJSBridge())
	clock := system.NewClock()
	useCase := usecases.NewRunReceiptWebhookDispatchCycleUseCase(unitOfWork, addressPolicyReader, notifier, clock)

	return scheduleradapter.NewWebhookDispatcherHandler(scheduleradapter.WebhookDispatcherDependencies{
		RunReceiptWebhookDispatchCycleUseCase: useCase,
	}), request, nil
}

func buildCloudflareReceiptWebhookDispatcherRequest(
	env map[string]string,
) (scheduleradapter.WebhookDispatcherRequest, error) {
	settings, err := loadReceiptWebhookDispatchSettingsFromLookup(func(key string) string {
		return env[key]
	}, receiptWebhookDispatchDefaults{
		BatchSize:   cfDefaultReceiptWebhookDispatchBatchSize,
		DispatchTTL: cfDefaultReceiptWebhookDispatchClaimTTL,
		MaxAttempts: cfDefaultReceiptWebhookDispatchMaxAttempts,
		RetryDelay:  cfDefaultReceiptWebhookDispatchRetryDelay,
	}, false)
	if err != nil {
		return scheduleradapter.WebhookDispatcherRequest{}, err
	}

	return scheduleradapter.WebhookDispatcherRequest{
		BatchSize:   settings.BatchSize,
		DispatchTTL: settings.DispatchTTL,
		RetryDelay:  settings.RetryDelay,
		MaxAttempts: settings.MaxAttempts,
	}, nil
}

func loadCloudflareReceiptWebhookNotifierConfig(
	env map[string]string,
) (webhookadapter.PaymentReceiptWebhookNotifierConfig, error) {
	settings, err := loadReceiptWebhookNotifierSettingsFromLookup(func(key string) string {
		return env[key]
	}, 10*time.Second)
	if err != nil {
		return webhookadapter.PaymentReceiptWebhookNotifierConfig{}, err
	}

	return webhookadapter.PaymentReceiptWebhookNotifierConfig{
		CloudflareBinding:  cfReceiptWebhookMockBinding,
		CloudflarePath:     cfReceiptWebhookMockPath,
		Secret:             settings.Secret,
		Timeout:            settings.Timeout,
		InsecureSkipVerify: settings.InsecureSkipVerify,
	}, nil
}
