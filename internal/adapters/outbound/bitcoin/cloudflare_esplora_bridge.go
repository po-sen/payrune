package bitcoin

import (
	"context"

	"payrune/internal/domain/value_objects"
)

type CloudflareEsploraBridge interface {
	FetchLatestBlockHeight(ctx context.Context, bridgeID string, network value_objects.NetworkID) (int64, error)
	FetchAddressChainTransactions(
		ctx context.Context,
		bridgeID string,
		network value_objects.NetworkID,
		address string,
	) ([]esploraTransaction, error)
	FetchAddressMempoolTransactions(
		ctx context.Context,
		bridgeID string,
		network value_objects.NetworkID,
		address string,
	) ([]esploraTransaction, error)
}
