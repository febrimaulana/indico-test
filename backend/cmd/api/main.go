// Package main is the entry point of the inventory reservation API.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"indico/backend/internal/app"
	"indico/backend/internal/platform/config"
	"indico/backend/internal/platform/logger"
)

//	@title			INDICO Inventory Reservation API
//	@version		1.0
//	@description	High-throughput inventory reservation system. Reservasi stok otomatis kedaluwarsa dalam 5 menit jika tidak dikonfirmasi.
//	@host			localhost:8080
//	@BasePath		/
func main() {
	if err := run(); err != nil {
		logger.New("error").Error("server berhenti dengan error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.New(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.New(cfg, log)
	if err != nil {
		return err
	}

	log.Info("server berjalan",
		"port", cfg.Port,
		"expiry", cfg.ExpiryDuration.String(),
		"swagger", "/swagger/index.html",
	)
	return application.Run(ctx)
}
