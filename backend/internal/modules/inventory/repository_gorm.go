package inventory

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"

	"indico/backend/internal/platform/database"
)

type itemRepository struct {
	db *gorm.DB
}

// NewItemRepository returns the GORM backed item repository.
func NewItemRepository(db *gorm.DB) ItemRepository { return &itemRepository{db: db} }

func (r *itemRepository) conn(ctx context.Context) *gorm.DB {
	return database.Connection(ctx, r.db).WithContext(ctx)
}

func (r *itemRepository) GetByID(ctx context.Context, itemID string) (*Item, error) {
	var item Item
	err := r.conn(ctx).First(&item, "item_id = ?", itemID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *itemRepository) ReserveStock(ctx context.Context, itemID string, quantity int) (bool, error) {
	result := r.conn(ctx).Model(&Item{}).
		Where("item_id = ? AND total_stock - reserved_stock >= ?", itemID, quantity).
		Updates(map[string]any{
			"reserved_stock": gorm.Expr("reserved_stock + ?", quantity),
			"updated_at":     gorm.Expr("now()"),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}

func (r *itemRepository) ConfirmStock(ctx context.Context, itemID string, quantity int) error {
	result := r.conn(ctx).Model(&Item{}).
		Where("item_id = ?", itemID).
		Updates(map[string]any{
			"total_stock":    gorm.Expr("total_stock - ?", quantity),
			"reserved_stock": gorm.Expr("reserved_stock - ?", quantity),
			"updated_at":     gorm.Expr("now()"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrItemNotFound
	}
	return nil
}

func (r *itemRepository) ReleaseStock(ctx context.Context, itemID string, quantity int) error {
	result := r.conn(ctx).Model(&Item{}).
		Where("item_id = ?", itemID).
		Updates(map[string]any{
			"reserved_stock": gorm.Expr("reserved_stock - ?", quantity),
			"updated_at":     gorm.Expr("now()"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrItemNotFound
	}
	return nil
}

type reservationRepository struct {
	db *gorm.DB
}

// NewReservationRepository returns the GORM backed reservation repository.
func NewReservationRepository(db *gorm.DB) ReservationRepository {
	return &reservationRepository{db: db}
}

func (r *reservationRepository) conn(ctx context.Context) *gorm.DB {
	return database.Connection(ctx, r.db).WithContext(ctx)
}

func (r *reservationRepository) Create(ctx context.Context, reservation *Reservation) error {
	err := r.conn(ctx).Create(reservation).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrReservationIDConflict
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *reservationRepository) GetByID(ctx context.Context, reservationID string) (*Reservation, error) {
	var reservation Reservation
	err := r.conn(ctx).First(&reservation, "reservation_id = ?", reservationID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrReservationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

// Confirm performs the status transition with a single conditional UPDATE so
// two concurrent confirms can never both succeed.
func (r *reservationRepository) Confirm(ctx context.Context, reservationID string, now time.Time) (*Reservation, bool, error) {
	row := r.conn(ctx).Raw(`
		UPDATE reservations
		SET status = ?, confirmed_at = ?, updated_at = ?
		WHERE reservation_id = ? AND status = 'active' AND expires_at > ?
		RETURNING reservation_id, item_id, quantity, expires_at`,
		ReservationStatusConfirmed, now, now, reservationID, now,
	).Row()

	var (
		id        string
		itemID    string
		quantity  int
		expiresAt time.Time
	)
	if err := row.Scan(&id, &itemID, &quantity, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	confirmedAt := now
	return &Reservation{
		ReservationID: id,
		ItemID:        itemID,
		Quantity:      quantity,
		Status:        ReservationStatusConfirmed,
		ExpiresAt:     expiresAt,
		ConfirmedAt:   &confirmedAt,
	}, true, nil
}

// ClaimExpired marks one batch of expired reservations and returns them.
func (r *reservationRepository) ClaimExpired(ctx context.Context, now time.Time, limit int) ([]Reservation, error) {
	var claimed []Reservation
	err := r.conn(ctx).Raw(`
		UPDATE reservations
		SET status = 'expired', updated_at = ?
		WHERE reservation_id IN (
			SELECT reservation_id
			FROM reservations
			WHERE status = 'active' AND expires_at <= ?
			ORDER BY expires_at
			LIMIT ?
			FOR UPDATE SKIP LOCKED
		)
		RETURNING reservation_id, item_id, user_id, quantity, status, expires_at, confirmed_at, created_at, updated_at`,
		now, now, limit,
	).Scan(&claimed).Error
	if err != nil {
		return nil, err
	}
	return claimed, nil
}
