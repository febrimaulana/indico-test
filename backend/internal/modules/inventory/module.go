package inventory

import (
	"log/slog"
	"time"

	"gorm.io/gorm"

	"indico/backend/internal/platform/clock"
	"indico/backend/internal/platform/database"
	"indico/backend/internal/platform/idgen"
)

// ModuleConfig carries the runtime knobs owned by the inventory module.
type ModuleConfig struct {
	ExpiryDuration time.Duration
	SweepInterval  time.Duration
	SweepBatchSize int
}

// Module is the composition root of the inventory domain: it wires
// repository -> service -> handler and exposes the public service API.
type Module struct {
	Handler      *Handler
	Service      Service
	ExpiryWorker *ExpiryWorker
}

// NewModule assembles the inventory module.
func NewModule(db *gorm.DB, log *slog.Logger, cfg ModuleConfig) *Module {
	items := NewItemRepository(db)
	reservations := NewReservationRepository(db)
	txm := database.NewTxManager(db)
	clk := clock.New()
	ids := idgen.New()

	service := NewService(items, reservations, txm, ids, clk, cfg.ExpiryDuration)
	worker := NewExpiryWorker(reservations, items, txm, clk, cfg.SweepInterval, cfg.SweepBatchSize, log)

	return &Module{
		Handler:      NewHandler(service),
		Service:      service,
		ExpiryWorker: worker,
	}
}
