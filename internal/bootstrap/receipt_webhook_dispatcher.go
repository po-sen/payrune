package bootstrap

import (
	"context"
	"log"
	"time"

	scheduleradapter "payrune/internal/adapters/inbound/scheduler"
	"payrune/internal/infrastructure/di"
)

const (
	defaultReceiptWebhookDispatcherInterval    = 15 * time.Second
	defaultReceiptWebhookDispatcherBatchSize   = 50
	defaultReceiptWebhookDispatcherClaimTTL    = 30 * time.Second
	defaultReceiptWebhookDispatcherMaxAttempts = int32(10)
	defaultReceiptWebhookDispatcherRetryDelay  = time.Minute
)

type ReceiptWebhookDispatcherConfig struct {
	Interval    time.Duration
	BatchSize   int
	ClaimTTL    time.Duration
	MaxAttempts int32
	RetryDelay  time.Duration
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

	container, err := di.NewReceiptWebhookDispatcherContainer()
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
