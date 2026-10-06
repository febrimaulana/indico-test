package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"PORT", "DATABASE_URL", "LOG_LEVEL", "CORS_ALLOWED_ORIGINS",
		"EXPIRY_DURATION", "SWEEP_INTERVAL", "SWEEP_BATCH_SIZE", "SHUTDOWN_TIMEOUT",
		"DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME",
		"DB_CONNECT_ATTEMPTS", "DB_CONNECT_RETRY_DELAY",
	} {
		t.Setenv(key, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/indico?sslmode=disable")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "8080", cfg.Port)
	require.Equal(t, "info", cfg.LogLevel)
	require.Equal(t, []string{"*"}, cfg.CORSAllowedOrigins)
	require.Equal(t, 5*time.Minute, cfg.ExpiryDuration)
	require.Equal(t, 15*time.Second, cfg.SweepInterval)
	require.Equal(t, 500, cfg.SweepBatchSize)
	require.Equal(t, 15*time.Second, cfg.ShutdownTimeout)
	require.Equal(t, 25, cfg.DBMaxOpenConns)
	require.Equal(t, 5, cfg.DBMaxIdleConns)
	require.Equal(t, 5*time.Minute, cfg.DBConnMaxLifetime)
	require.Equal(t, 10, cfg.DBConnectAttempts)
	require.Equal(t, time.Second, cfg.DBConnectDelay)
}

func TestLoadOverrides(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/indico?sslmode=disable")
	t.Setenv("PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	t.Setenv("EXPIRY_DURATION", "1m")
	t.Setenv("SWEEP_INTERVAL", "2s")
	t.Setenv("SWEEP_BATCH_SIZE", "10")
	t.Setenv("SHUTDOWN_TIMEOUT", "3s")
	t.Setenv("DB_MAX_OPEN_CONNS", "7")
	t.Setenv("DB_MAX_IDLE_CONNS", "2")
	t.Setenv("DB_CONN_MAX_LIFETIME", "1m")
	t.Setenv("DB_CONNECT_ATTEMPTS", "2")
	t.Setenv("DB_CONNECT_RETRY_DELAY", "50ms")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "9090", cfg.Port)
	require.Equal(t, "debug", cfg.LogLevel)
	require.Equal(t, []string{"http://localhost:3000"}, cfg.CORSAllowedOrigins)
	require.Equal(t, time.Minute, cfg.ExpiryDuration)
	require.Equal(t, 2*time.Second, cfg.SweepInterval)
	require.Equal(t, 10, cfg.SweepBatchSize)
	require.Equal(t, 3*time.Second, cfg.ShutdownTimeout)
	require.Equal(t, 7, cfg.DBMaxOpenConns)
	require.Equal(t, 2, cfg.DBMaxIdleConns)
	require.Equal(t, time.Minute, cfg.DBConnMaxLifetime)
	require.Equal(t, 2, cfg.DBConnectAttempts)
	require.Equal(t, 50*time.Millisecond, cfg.DBConnectDelay)
}

func TestLoadMissingDatabaseURL(t *testing.T) {
	clearEnv(t)
	_, err := Load()
	require.ErrorContains(t, err, "DATABASE_URL")
}

func TestLoadInvalidValues(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		value string
	}{
		{"invalid expiry duration", "EXPIRY_DURATION", "abc"},
		{"invalid sweep interval", "SWEEP_INTERVAL", "abc"},
		{"invalid sweep batch size", "SWEEP_BATCH_SIZE", "abc"},
		{"invalid shutdown timeout", "SHUTDOWN_TIMEOUT", "abc"},
		{"invalid db max open conns", "DB_MAX_OPEN_CONNS", "abc"},
		{"invalid db max idle conns", "DB_MAX_IDLE_CONNS", "abc"},
		{"invalid db conn max lifetime", "DB_CONN_MAX_LIFETIME", "abc"},
		{"invalid db connect attempts", "DB_CONNECT_ATTEMPTS", "abc"},
		{"invalid db connect retry delay", "DB_CONNECT_RETRY_DELAY", "abc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/indico?sslmode=disable")
			t.Setenv(tc.key, tc.value)
			_, err := Load()
			require.ErrorContains(t, err, tc.key)
		})
	}
}

func TestLoadZeroValuesRejected(t *testing.T) {
	t.Run("sweep batch size must be positive", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/indico?sslmode=disable")
		t.Setenv("SWEEP_BATCH_SIZE", "0")
		_, err := Load()
		require.ErrorContains(t, err, "SWEEP_BATCH_SIZE")
	})

	t.Run("db connect attempts must be positive", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/indico?sslmode=disable")
		t.Setenv("DB_CONNECT_ATTEMPTS", "0")
		_, err := Load()
		require.ErrorContains(t, err, "DB_CONNECT_ATTEMPTS")
	})
}
