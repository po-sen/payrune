package in

import (
	"context"

	"payrune/internal/application/dto"
)

type RunReceiptWebhookDispatchCycleUseCase interface {
	Execute(ctx context.Context, input dto.RunReceiptWebhookDispatchCycleInput) (dto.RunReceiptWebhookDispatchCycleOutput, error)
}
