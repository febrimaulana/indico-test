package app

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"indico/backend/internal/platform/config"
	"indico/backend/internal/platform/database"
)

// The app lifecycle test uses its own database so it never shares tables with
// the inventory integration tests (packages may run in parallel).
const defaultIntegrationDSN = "postgres://indico:indico@localhost:5433/indico_test_app?sslmode=disable&connect_timeout=2"

// TestIntegrationAppLifecycle boots the real application (database,
// migrations, module, router, server) and exercises the full graceful
// shutdown path. It uses TEST_DATABASE_URL when set and otherwise falls back
// to the compose PostgreSQL, creating the database on the fly; when no
// database is reachable the test skips instead of failing.
func TestIntegrationAppLifecycle(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	explicit := dsn != ""
	if !explicit {
		dsn = defaultIntegrationDSN
		if !canOpenIntegrationDatabase(dsn) {
			createIntegrationDatabase(dsn)
			if !canOpenIntegrationDatabase(dsn) {
				t.Skipf("PostgreSQL tidak tersedia di %s - melewati integration test", dsn)
			}
		}
	}

	cfg := config.Config{
		Port:               "0",
		DatabaseURL:        dsn,
		ExpiryDuration:     time.Minute,
		SweepInterval:      time.Hour,
		SweepBatchSize:     10,
		ShutdownTimeout:    2 * time.Second,
		LogLevel:           "error",
		CORSAllowedOrigins: []string{"*"},
		DBMaxOpenConns:     5,
		DBMaxIdleConns:     2,
		DBConnMaxLifetime:  time.Minute,
		DBConnectAttempts:  5,
		DBConnectDelay:     200 * time.Millisecond,
	}

	application, err := New(cfg, testLog())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	require.NoError(t, application.Run(ctx))
}

func canOpenIntegrationDatabase(dsn string) bool {
	db, err := database.Open(dsn, database.PoolConfig{MaxOpenConns: 1, MaxIdleConns: 1, ConnMaxLifetime: time.Minute})
	if err != nil {
		return false
	}
	if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
		_ = sqlDB.Close()
	}
	return true
}

func createIntegrationDatabase(dsn string) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return
	}
	name := strings.TrimPrefix(parsed.Path, "/")
	if !validIntegrationDatabaseName(name) {
		return
	}

	adminURL := *parsed
	adminURL.Path = "/postgres"
	admin, err := database.Open(adminURL.String(), database.PoolConfig{MaxOpenConns: 1, MaxIdleConns: 1, ConnMaxLifetime: time.Minute})
	if err != nil {
		return
	}
	defer func() {
		if sqlDB, err := admin.DB(); err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	}()
	_ = admin.Exec(`CREATE DATABASE "` + name + `"`).Error
}

func validIntegrationDatabaseName(name string) bool {
	if name == "" {
		return false
	}
	for _, char := range name {
		isLower := char >= 'a' && char <= 'z'
		isDigit := char >= '0' && char <= '9'
		if !isLower && !isDigit && char != '_' {
			return false
		}
	}
	return true
}
