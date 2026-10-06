package inventory

import (
	"context"
	"time"
)

// ItemRepository is the data contract for items.
type ItemRepository interface {
	GetByID(ctx context.Context, itemID string) (*Item, error)

	// ReserveStock atomically adds quantity to reserved_stock only when
	// enough stock is still available. It reports whether the row was
	// updated.
	ReserveStock(ctx context.Context, itemID string, quantity int) (bool, error)

	// ConfirmStock permanently decrements physical stock and the hold.
	ConfirmStock(ctx context.Context, itemID string, quantity int) error

	// ReleaseStock returns a hold to available stock.
	ReleaseStock(ctx context.Context, itemID string, quantity int) error
}

// ReservationRepository is the data contract for reservations.
type ReservationRepository interface {
	Create(ctx context.Context, reservation *Reservation) error
	GetByID(ctx context.Context, reservationID string) (*Reservation, error)

	// Confirm atomically transitions an active, unexpired reservation to
	// confirmed via one conditional UPDATE. It reports whether it matched.
	Confirm(ctx context.Context, reservationID string, now time.Time) (*Reservation, bool, error)

	// ClaimExpired atomically marks up to limit expired reservations as
	// expired and returns them. Safe for concurrent sweeper instances thanks
	// to FOR UPDATE SKIP LOCKED.
	ClaimExpired(ctx context.Context, now time.Time, limit int) ([]Reservation, error)
}
