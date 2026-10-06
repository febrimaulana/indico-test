// Package app is the composition root: it wires configuration, database,
// migrations, modules, the Gin router and the HTTP server, then owns the
// lifecycle (start worker, serve, graceful shutdown).
package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"

	"indico/backend/internal/modules/inventory"
	"indico/backend/internal/platform/config"
	"indico/backend/internal/platform/database"
	"indico/backend/internal/platform/server"
	"indico/backend/migrations"
)

// Runnable runs until its context is canceled.
type Runnable interface {
	Run(ctx context.Context) error
}

// Closer releases resources (the database pool in production).
type Closer interface {
	Close() error
}

// openDB and runMigrations are seams so the wiring can be unit tested.
var (
	openDB = func(ctx context.Context, cfg config.Config) (*gorm.DB, error) {
		pool := database.PoolConfig{
			MaxOpenConns:    cfg.DBMaxOpenConns,
			MaxIdleConns:    cfg.DBMaxIdleConns,
			ConnMaxLifetime: cfg.DBConnMaxLifetime,
		}
		return database.OpenWithRetry(ctx, cfg.DatabaseURL, pool, cfg.DBConnectAttempts, cfg.DBConnectDelay)
	}
	runMigrations = func(databaseURL string) error {
		return database.RunMigrations(databaseURL, migrations.FS)
	}
)

type noopCloser struct{}

func (noopCloser) Close() error { return nil }

// App owns the server lifecycle.
type App struct {
	server Runnable
	worker Runnable
	closer Closer
	log    *slog.Logger
}

func newApp(server Runnable, worker Runnable, closer Closer, log *slog.Logger) *App {
	return &App{server: server, worker: worker, closer: closer, log: log}
}

// New connects to the database, runs migrations and builds the application.
func New(cfg config.Config, log *slog.Logger) (*App, error) {
	db, err := openDB(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	return newWithDB(cfg, log, db)
}

func newWithDB(cfg config.Config, log *slog.Logger, db *gorm.DB) (*App, error) {
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		return nil, err
	}

	inventoryModule := inventory.NewModule(db, log, inventory.ModuleConfig{
		ExpiryDuration: cfg.ExpiryDuration,
		SweepInterval:  cfg.SweepInterval,
		SweepBatchSize: cfg.SweepBatchSize,
	})

	engine := BuildRouter(cfg, log, inventoryModule)
	httpServer := server.New(engine, ":"+cfg.Port, server.Timeouts{
		ReadHeader: 5 * time.Second,
		Read:       10 * time.Second,
		Write:      15 * time.Second,
		Idle:       60 * time.Second,
	}, cfg.ShutdownTimeout)

	var closer Closer = noopCloser{}
	if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
		closer = sqlDB
	}

	return newApp(httpServer, inventoryModule.ExpiryWorker, closer, log), nil
}

// Run starts the expiry worker, serves HTTP until ctx is canceled, then stops
// the worker and closes the database pool. Order matters: requests drain
// before the pool closes.
func (a *App) Run(ctx context.Context) error {
	workerCtx, cancelWorker := context.WithCancel(ctx)
	defer cancelWorker()

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		if err := a.worker.Run(workerCtx); err != nil {
			a.log.Error("worker berhenti dengan error", "error", err)
		}
	}()

	runErr := a.server.Run(ctx)
	cancelWorker()
	waitGroup.Wait()

	if closeErr := a.closer.Close(); closeErr != nil && runErr == nil {
		return fmt.Errorf("gagal menutup koneksi database: %w", closeErr)
	}
	return runErr
}
