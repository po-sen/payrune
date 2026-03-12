package bitcoin

import (
	"context"

	"payrune/internal/domain/valueobjects"
)

type CloudflareEsploraBridge interface {
	FetchLatestBlockHeight(ctx context.Context, bridgeID string, network valueobjects.NetworkID) (int64, error)
	FetchAddressChainTransactions(
		ctx context.Context,
		bridgeID string,
		network valueobjects.NetworkID,
		address string,
	) ([]esploraTransaction, error)
	FetchAddressMempoolTransactions(
		ctx context.Context,
		bridgeID string,
		network valueobjects.NetworkID,
		address string,
	) ([]esploraTransaction, error)
}
