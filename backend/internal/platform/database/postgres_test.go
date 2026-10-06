package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOpenInvalidDSN(t *testing.T) {
	_, err := Open("://not-a-dsn", PoolConfig{})
	require.ErrorContains(t, err, "gagal membuka koneksi database")
}

func TestOpenConfigurePoolError(t *testing.T) {
	original := dbFrom
	t.Cleanup(func() { dbFrom = original })
	dbFrom = func(*gorm.DB) (*sql.DB, error) { return nil, errors.New("boom") }

	_, err := Open("postgres://user:pass@127.0.0.1:1/none?sslmode=disable", PoolConfig{})
	require.ErrorContains(t, err, "gagal mengambil koneksi database")
}

func TestOpenPingError(t *testing.T) {
	_, err := Open("postgres://user:pass@127.0.0.1:1/none?sslmode=disable&connect_timeout=1", PoolConfig{})
	require.ErrorContains(t, err, "database tidak dapat dijangkau")
}

func TestConfigurePoolDBFromError(t *testing.T) {
	original := dbFrom
	t.Cleanup(func() { dbFrom = original })
	dbFrom = func(*gorm.DB) (*sql.DB, error) { return nil, errors.New("boom") }

	err := configurePool(nil, PoolConfig{})
	require.ErrorContains(t, err, "gagal mengambil koneksi database")
}

func TestPingDBFromError(t *testing.T) {
	original := dbFrom
	t.Cleanup(func() { dbFrom = original })
	dbFrom = func(*gorm.DB) (*sql.DB, error) { return nil, errors.New("boom") }

	err := ping(nil)
	require.ErrorContains(t, err, "gagal mengambil koneksi database")
}

func unreachableSQLDB(t *testing.T) *sql.DB {
	t.Helper()
	sqlDB, err := sql.Open("pgx", "postgres://user:pass@127.0.0.1:1/none?sslmode=disable&connect_timeout=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	return sqlDB
}

func TestConfigurePoolSuccess(t *testing.T) {
	original := dbFrom
	t.Cleanup(func() { dbFrom = original })
	dbFrom = func(*gorm.DB) (*sql.DB, error) { return unreachableSQLDB(t), nil }

	err := configurePool(nil, PoolConfig{MaxOpenConns: 3, MaxIdleConns: 1, ConnMaxLifetime: time.Minute})
	require.NoError(t, err)
}

func TestPingConnectionError(t *testing.T) {
	original := dbFrom
	t.Cleanup(func() { dbFrom = original })
	dbFrom = func(*gorm.DB) (*sql.DB, error) { return unreachableSQLDB(t), nil }

	err := ping(nil)
	require.ErrorContains(t, err, "database tidak dapat dijangkau")
}

func TestOpenWithRetryExhausted(t *testing.T) {
	_, err := OpenWithRetry(context.Background(), "://not-a-dsn", PoolConfig{}, 1, time.Millisecond)
	require.ErrorContains(t, err, "setelah 1 percobaan")
}

func TestOpenWithRetryContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := OpenWithRetry(ctx, "://not-a-dsn", PoolConfig{}, 3, time.Hour)
	require.ErrorIs(t, err, context.Canceled)
}
