package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const (
	envDatabaseURL     = "DATABASE_URL"
	defaultPingTimeout = 5 * time.Second
)

func OpenFromEnv() (*sql.DB, error) {
	databaseURL := strings.TrimSpace(os.Getenv(envDatabaseURL))
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
	defer cancel()

	return open(ctx, "postgres", databaseURL)
}

func open(ctx context.Context, driverName string, databaseURL string) (*sql.DB, error) {
	db, err := sql.Open(driverName, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database connection: %w", err)
	}

	return db, nil
}
