package app

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"indico/backend/internal/platform/config"
)

// TestIntegrationAppLifecycle boots the real application (database,
// migrations, module, router, server) and exercises the full graceful
// shutdown path. Skipped without TEST_DATABASE_URL.
func TestIntegrationAppLifecycle(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL tidak diset - melewati integration test")
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
