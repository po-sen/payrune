package bitcoin

import "context"

type cloudflareEsploraBridge interface {
	FetchLatestBlockHeight(ctx context.Context, bridgeID string, network string) (int64, error)
	FetchAddressChainTransactions(
		ctx context.Context,
		bridgeID string,
		network string,
		address string,
	) ([]esploraTransaction, error)
	FetchAddressMempoolTransactions(
		ctx context.Context,
		bridgeID string,
		network string,
		address string,
	) ([]esploraTransaction, error)
}
