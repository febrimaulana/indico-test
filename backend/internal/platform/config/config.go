// Package config loads every runtime configuration value from environment
// variables, applying safe defaults for local development.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for the API server.
type Config struct {
	Port        string
	DatabaseURL string

	ExpiryDuration time.Duration
	SweepInterval  time.Duration
	SweepBatchSize int

	ShutdownTimeout time.Duration
	LogLevel        string

	CORSAllowedOrigins []string

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnectAttempts int
	DBConnectDelay    time.Duration
}

// Load reads configuration from the environment and validates it.
// DATABASE_URL is the only required value; everything else has defaults.
func Load() (Config, error) {
	cfg := Config{
		Port:               envString("PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		LogLevel:           envString("LOG_LEVEL", "info"),
		CORSAllowedOrigins: []string{envString("CORS_ALLOWED_ORIGINS", "*")},
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("konfigurasi DATABASE_URL wajib diisi")
	}

	var err error
	if cfg.ExpiryDuration, err = envDuration("EXPIRY_DURATION", 5*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.SweepInterval, err = envDuration("SWEEP_INTERVAL", 15*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.SweepBatchSize, err = envInt("SWEEP_BATCH_SIZE", 500); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = envDuration("SHUTDOWN_TIMEOUT", 15*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.DBMaxOpenConns, err = envInt("DB_MAX_OPEN_CONNS", 25); err != nil {
		return Config{}, err
	}
	if cfg.DBMaxIdleConns, err = envInt("DB_MAX_IDLE_CONNS", 5); err != nil {
		return Config{}, err
	}
	if cfg.DBConnMaxLifetime, err = envDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.DBConnectAttempts, err = envInt("DB_CONNECT_ATTEMPTS", 10); err != nil {
		return Config{}, err
	}
	if cfg.DBConnectDelay, err = envDuration("DB_CONNECT_RETRY_DELAY", time.Second); err != nil {
		return Config{}, err
	}

	if cfg.SweepBatchSize <= 0 {
		return Config{}, fmt.Errorf("konfigurasi SWEEP_BATCH_SIZE harus lebih dari 0")
	}
	if cfg.DBConnectAttempts <= 0 {
		return Config{}, fmt.Errorf("konfigurasi DB_CONNECT_ATTEMPTS harus lebih dari 0")
	}
	return cfg, nil
}

func envString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func envInt(key string, defaultValue int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("konfigurasi %s tidak valid: %q bukan angka", key, raw)
	}
	return value, nil
}

func envDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("konfigurasi %s tidak valid: %q bukan durasi", key, raw)
	}
	return value, nil
}
