package inventory

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"indico/backend/internal/platform/clock"
	"indico/backend/internal/platform/database"
)

// ExpiryWorker periodically returns expired reservations to available stock.
type ExpiryWorker struct {
	reservations ReservationRepository
	items        ItemRepository
	txm          database.TxManager
	clock        clock.Clock
	interval     time.Duration
	batchSize    int
	log          *slog.Logger
}

// NewExpiryWorker builds the sweeper.
func NewExpiryWorker(
	reservations ReservationRepository,
	items ItemRepository,
	txm database.TxManager,
	clk clock.Clock,
	interval time.Duration,
	batchSize int,
	log *slog.Logger,
) *ExpiryWorker {
	return &ExpiryWorker{
		reservations: reservations,
		items:        items,
		txm:          txm,
		clock:        clk,
		interval:     interval,
		batchSize:    batchSize,
		log:          log,
	}
}

// Run sweeps expired reservations immediately (covering restarts) and then
// every interval, until the context is canceled.
func (w *ExpiryWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.Sweep(ctx)
	for {
		select {
		case <-ctx.Done():
			w.log.Info("worker expiry berhenti")
			return nil
		case <-ticker.C:
			w.Sweep(ctx)
		}
	}
}

// Sweep returns exactly one batch of expired reservations to available stock.
// It aggregates per item and updates items in sorted order to avoid deadlocks.
func (w *ExpiryWorker) Sweep(ctx context.Context) {
	err := w.txm.WithTx(ctx, func(txCtx context.Context) error {
		expired, err := w.reservations.ClaimExpired(txCtx, w.clock.Now(), w.batchSize)
		if err != nil {
			return err
		}
		if len(expired) == 0 {
			return nil
		}

		quantities := make(map[string]int)
		for _, reservation := range expired {
			quantities[reservation.ItemID] += reservation.Quantity
		}

		itemIDs := make([]string, 0, len(quantities))
		for itemID := range quantities {
			itemIDs = append(itemIDs, itemID)
		}
		sort.Strings(itemIDs)

		for _, itemID := range itemIDs {
			if err := w.items.ReleaseStock(txCtx, itemID, quantities[itemID]); err != nil {
				return err
			}
		}

		w.log.Info("reservasi kedaluwarsa dikembalikan",
			"jumlah_reservasi", len(expired),
			"jumlah_item", len(itemIDs),
		)
		return nil
	})
	if err != nil {
		w.log.Error("sweep reservasi kedaluwarsa gagal", "error", err)
	}
}
