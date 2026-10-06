// Package database owns every connection detail: opening GORM, tuning the
// pool, running migrations and transaction management.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// PoolConfig tunes the underlying database/sql connection pool.
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// dbFrom is a seam so tests can force the (practically unreachable) error of
// gorm.DB.DB().
var dbFrom = func(db *gorm.DB) (*sql.DB, error) { return db.DB() }

// Open opens a GORM connection, tunes the pool and verifies connectivity.
func Open(dsn string, pool PoolConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:               gormlogger.Default.LogMode(gormlogger.Silent),
		TranslateError:       true,
		DisableAutomaticPing: true, // ping is done explicitly after pool tuning
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi database: %w", err)
	}
	if err := configurePool(db, pool); err != nil {
		return nil, err
	}
	if err := ping(db); err != nil {
		return nil, err
	}
	return db, nil
}

func configurePool(db *gorm.DB, pool PoolConfig) error {
	sqlDB, err := dbFrom(db)
	if err != nil {
		return fmt.Errorf("gagal mengambil koneksi database: %w", err)
	}
	sqlDB.SetMaxOpenConns(pool.MaxOpenConns)
	sqlDB.SetMaxIdleConns(pool.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(pool.ConnMaxLifetime)
	return nil
}

func ping(db *gorm.DB) error {
	sqlDB, err := dbFrom(db)
	if err != nil {
		return fmt.Errorf("gagal mengambil koneksi database: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database tidak dapat dijangkau: %w", err)
	}
	return nil
}

// OpenWithRetry keeps trying to open the database until the attempt budget is
// exhausted or the context is canceled. This lets the backend start
// concurrently with a cold PostgreSQL container.
func OpenWithRetry(ctx context.Context, dsn string, pool PoolConfig, attempts int, delay time.Duration) (*gorm.DB, error) {
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		db, err := Open(dsn, pool)
		if err == nil {
			return db, nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("gagal menyambung ke database: %w", ctx.Err())
		case <-time.After(delay):
		}
	}
	return nil, fmt.Errorf("gagal menyambung ke database setelah %d percobaan: %w", attempts, lastErr)
}
