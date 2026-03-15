package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

var testDriverCounter uint64

type stubDriver struct {
	pingErr error
}

type stubConn struct {
	pingErr error
}

func (d stubDriver) Open(string) (driver.Conn, error) {
	return stubConn(d), nil
}

func (c stubConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func (c stubConn) Close() error {
	return nil
}

func (c stubConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not implemented")
}

func (c stubConn) Ping(context.Context) error {
	return c.pingErr
}

func TestOpenFromEnvMissingDatabaseURL(t *testing.T) {
	t.Setenv(envDatabaseURL, " ")

	_, err := OpenFromEnv()
	if err == nil {
		t.Fatal("expected missing DATABASE_URL error")
	}
	if got := err.Error(); got != "DATABASE_URL is required" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestOpenInvalidDriver(t *testing.T) {
	_, err := open(context.Background(), "postgres-driver-does-not-exist", "postgres://example")
	if err == nil {
		t.Fatal("expected invalid driver error")
	}
	if !strings.Contains(err.Error(), "open database connection:") {
		t.Fatalf("expected open error, got %q", err)
	}
}

func TestOpenPingError(t *testing.T) {
	driverName := registerTestDriver(stubDriver{pingErr: errors.New("ping failed")})

	_, err := open(context.Background(), driverName, "postgres://example")
	if err == nil {
		t.Fatal("expected ping error")
	}
	if !strings.Contains(err.Error(), "ping database connection:") {
		t.Fatalf("expected ping error, got %q", err)
	}
}

func TestOpenSuccess(t *testing.T) {
	driverName := registerTestDriver(stubDriver{})

	db, err := open(context.Background(), driverName, "postgres://example")
	if err != nil {
		t.Fatalf("open returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if db == nil {
		t.Fatal("expected non-nil db")
	}
	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("ping returned error: %v", err)
	}
}

func registerTestDriver(driverInstance driver.Driver) string {
	driverName := "postgres-driver-test-" + strconv.FormatUint(atomic.AddUint64(&testDriverCounter, 1), 10)
	sql.Register(driverName, driverInstance)
	return driverName
}
