package cloudflarepostgres

import "context"

type Bridge interface {
	BeginTx(ctx context.Context, bridgeID string) (string, error)
	CommitTx(ctx context.Context, bridgeID string, txID string) error
	RollbackTx(ctx context.Context, bridgeID string, txID string) error
	Exec(ctx context.Context, bridgeID string, txID string, query string, args []any) (int64, error)
	Query(ctx context.Context, bridgeID string, txID string, query string, args []any) ([][]any, error)
	QueryRow(ctx context.Context, bridgeID string, txID string, query string, args []any) ([]any, bool, error)
}
