package cloudflarepostgres

import (
	"context"
	"database/sql"
	"errors"
)

type Result interface {
	RowsAffected() (int64, error)
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}

type Row interface {
	Scan(dest ...any) error
}

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) Row
}

type bridgeExecutor struct {
	bridge   Bridge
	bridgeID string
	txID     string
}

func NewExecutor(bridgeID string, bridge Bridge) Executor {
	return &bridgeExecutor{
		bridge:   bridge,
		bridgeID: bridgeID,
	}
}

func newTxExecutor(bridgeID string, txID string, bridge Bridge) Executor {
	return &bridgeExecutor{
		bridge:   bridge,
		bridgeID: bridgeID,
		txID:     txID,
	}
}

func (e *bridgeExecutor) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	if e.bridge == nil {
		return nil, errors.New("cloudflare postgres bridge is not configured")
	}
	rowsAffected, err := e.bridge.Exec(ctx, e.bridgeID, e.txID, query, args)
	if err != nil {
		return nil, err
	}
	return execResult{rowsAffected: rowsAffected}, nil
}

func (e *bridgeExecutor) QueryContext(ctx context.Context, query string, args ...any) (Rows, error) {
	if e.bridge == nil {
		return nil, errors.New("cloudflare postgres bridge is not configured")
	}
	rows, err := e.bridge.Query(ctx, e.bridgeID, e.txID, query, args)
	if err != nil {
		return nil, err
	}
	return &sliceRows{rows: rows}, nil
}

func (e *bridgeExecutor) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	if e.bridge == nil {
		return errorRow{err: errors.New("cloudflare postgres bridge is not configured")}
	}
	row, found, err := e.bridge.QueryRow(ctx, e.bridgeID, e.txID, query, args)
	if err != nil {
		return errorRow{err: err}
	}
	if !found {
		return errorRow{err: sql.ErrNoRows}
	}
	return valueRow{values: row}
}

type execResult struct {
	rowsAffected int64
}

func (r execResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

type sliceRows struct {
	rows [][]any
	idx  int
	err  error
}

func (r *sliceRows) Next() bool {
	if r == nil {
		return false
	}
	return r.idx < len(r.rows)
}

func (r *sliceRows) Scan(dest ...any) error {
	if r == nil {
		return errors.New("rows is nil")
	}
	if r.idx >= len(r.rows) {
		return sql.ErrNoRows
	}
	row := r.rows[r.idx]
	r.idx++
	return scanValues(row, dest...)
}

func (r *sliceRows) Err() error {
	if r == nil {
		return nil
	}
	return r.err
}

func (r *sliceRows) Close() error {
	return nil
}

type valueRow struct {
	values []any
}

func (r valueRow) Scan(dest ...any) error {
	return scanValues(r.values, dest...)
}

type errorRow struct {
	err error
}

func (r errorRow) Scan(_ ...any) error {
	return r.err
}
