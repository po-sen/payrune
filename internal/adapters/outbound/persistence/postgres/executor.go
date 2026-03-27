package postgres

import (
	"context"
	"database/sql"
)

// executor is the shared SQL capability contract used by Postgres stores.
// Both *sql.DB and *sql.Tx satisfy this interface.
type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
