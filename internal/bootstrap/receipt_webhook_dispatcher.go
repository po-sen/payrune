package bootstrap

import (
	"context"
	"log"
	"time"

	"payrune/internal/application/dto"
	"payrune/internal/infrastructure/di"
)

const (
	defaultReceiptWebhookDispatchInterval    = 15 * time.Second
	defaultReceiptWebhookDispatchBatchSize   = 50
	defaultReceiptWebhookDispatchClaimTTL    = 30 * time.Second
	defaultReceiptWebhookDispatchMaxAttempts = int32(10)
	defaultReceiptWebhookDispatchRetryDelay  = time.Minute
)

type ReceiptWebhookDispatchConfig struct {
	Interval    time.Duration
	BatchSize   int
	ClaimTTL    time.Duration
	MaxAttempts int32
	RetryDelay  time.Duration
}

func RunReceiptWebhookDispatcher(ctx context.Context, config ReceiptWebhookDispatchConfig) error {
	if config.Interval <= 0 {
		config.Interval = defaultReceiptWebhookDispatchInterval
	}
	if config.BatchSize <= 0 {
		config.BatchSize = defaultReceiptWebhookDispatchBatchSize
	}
	if config.ClaimTTL <= 0 {
		config.ClaimTTL = defaultReceiptWebhookDispatchClaimTTL
	}
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = defaultReceiptWebhookDispatchMaxAttempts
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = defaultReceiptWebhookDispatchRetryDelay
	}

	container, err := di.NewReceiptWebhookDispatcherContainer()
	if err != nil {
		return err
	}
	defer func() {
		_ = container.Close()
	}()

	runCycle := func() {
		output, err := container.RunReceiptWebhookDispatchCycleUseCase.Execute(ctx, dto.RunReceiptWebhookDispatchCycleInput{
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
