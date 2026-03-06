package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"payrune/internal/bootstrap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config, err := loadReceiptWebhookDispatchConfigFromEnv()
	if err != nil {
		log.Fatalf("invalid webhook dispatcher config: %v", err)
	}

	if err := bootstrap.RunReceiptWebhookDispatcher(ctx, config); err != nil {
		log.Fatalf("webhook dispatcher exited with error: %v", err)
	}
}

func loadReceiptWebhookDispatchConfigFromEnv() (bootstrap.ReceiptWebhookDispatchConfig, error) {
	interval, err := parseRequiredPositiveDurationEnv("RECEIPT_WEBHOOK_DISPATCH_INTERVAL")
	if err != nil {
		return bootstrap.ReceiptWebhookDispatchConfig{}, err
	}
	batchSize, err := parseRequiredPositiveIntEnv("RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE")
	if err != nil {
		return bootstrap.ReceiptWebhookDispatchConfig{}, err
	}
	claimTTL, err := parseRequiredPositiveDurationEnv("RECEIPT_WEBHOOK_DISPATCH_CLAIM_TTL")
	if err != nil {
		return bootstrap.ReceiptWebhookDispatchConfig{}, err
	}
	maxAttempts, err := parseRequiredPositiveInt32Env("RECEIPT_WEBHOOK_DISPATCH_MAX_ATTEMPTS")
	if err != nil {
		return bootstrap.ReceiptWebhookDispatchConfig{}, err
	}
	retryDelay, err := parseRequiredPositiveDurationEnv("RECEIPT_WEBHOOK_DISPATCH_RETRY_DELAY")
	if err != nil {
		return bootstrap.ReceiptWebhookDispatchConfig{}, err
	}

	return bootstrap.ReceiptWebhookDispatchConfig{
		Interval:    interval,
		BatchSize:   batchSize,
		ClaimTTL:    claimTTL,
		MaxAttempts: maxAttempts,
		RetryDelay:  retryDelay,
	}, nil
}

func parseRequiredPositiveDurationEnv(key string) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, logConfigRequiredError(key)
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, logConfigInvalidError(key, "must be a valid duration", err)
	}
	if value <= 0 {
		return 0, logConfigSimpleError(key, "must be greater than zero")
	}
	return value, nil
}

func parseRequiredPositiveIntEnv(key string) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, logConfigRequiredError(key)
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, logConfigInvalidError(key, "must be an integer", err)
	}
	if value <= 0 {
		return 0, logConfigSimpleError(key, "must be greater than zero")
	}
	return value, nil
}

func parseRequiredPositiveInt32Env(key string) (int32, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, logConfigRequiredError(key)
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, logConfigInvalidError(key, "must be a 32-bit integer", err)
	}
	if value <= 0 {
		return 0, logConfigSimpleError(key, "must be greater than zero")
	}
	return int32(value), nil
}

func logConfigRequiredError(key string) error {
	return logConfigSimpleError(key, "is required")
}

func logConfigSimpleError(key string, message string) error {
	return &configError{message: key + " " + message}
}

func logConfigInvalidError(key string, message string, err error) error {
	return &configError{message: key + " " + message + ": " + err.Error()}
}

type configError struct {
	message string
}

func (e *configError) Error() string {
	return e.message
}
