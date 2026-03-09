package bootstrap

import (
	"context"
	"log"
	"time"

	"payrune/internal/application/dto"
	"payrune/internal/infrastructure/di"
)

const (
	defaultPollerTickInterval = 15 * time.Second
	defaultPollerClaimTTL     = 30 * time.Second
	defaultPollerBatchSize    = 50
)

type PollerConfig struct {
	TickInterval       time.Duration
	RescheduleInterval time.Duration
	BatchSize          int
	ClaimTTL           time.Duration
	Chain              string
	Network            string
}

func RunPoller(ctx context.Context, config PollerConfig) error {
	if config.TickInterval <= 0 {
		config.TickInterval = defaultPollerTickInterval
	}
	if config.BatchSize <= 0 {
		config.BatchSize = defaultPollerBatchSize
	}
	if config.ClaimTTL <= 0 {
		config.ClaimTTL = defaultPollerClaimTTL
	}
	container, err := di.NewPollerContainer()
	if err != nil {
		return err
	}
	defer func() {
		_ = container.Close()
	}()

	runCycle := func() {
		output, err := container.RunReceiptPollingCycleUseCase.Execute(ctx, dto.RunReceiptPollingCycleInput{
			BatchSize:          config.BatchSize,
			RescheduleInterval: config.RescheduleInterval,
			ClaimTTL:           config.ClaimTTL,
			Chain:              config.Chain,
			Network:            config.Network,
		})
		if err != nil {
			log.Printf("poll cycle failed: err=%v", err)
			return
		}

		log.Printf(
			"poll cycle complete claimed=%d updated=%d terminal_failed=%d processing_errors=%d",
			output.ClaimedCount,
			output.UpdatedCount,
			output.TerminalFailedCount,
			output.ProcessingErrorCount,
		)
	}

	runCycle()

	ticker := time.NewTicker(config.TickInterval)
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
