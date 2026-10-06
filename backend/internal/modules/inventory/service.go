package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"indico/backend/internal/platform/clock"
	"indico/backend/internal/platform/database"
	"indico/backend/internal/platform/httpx"
	"indico/backend/internal/platform/idgen"
)

const maxReservationIDAttempts = 3

// Service is the public API of the inventory module.
type Service interface {
	Reserve(ctx context.Context, req ReserveRequest) (*ReserveResponse, error)
	Confirm(ctx context.Context, req ConfirmRequest) (*ConfirmResponse, error)
	GetStock(ctx context.Context, itemID string) (*StockResponse, error)
}

type service struct {
	items        ItemRepository
	reservations ReservationRepository
	txm          database.TxManager
	ids          idgen.Generator
	clock        clock.Clock
	expiry       time.Duration
}

// NewService builds the inventory service. It depends only on interfaces, so
// it can be unit tested without PostgreSQL.
func NewService(
	items ItemRepository,
	reservations ReservationRepository,
	txm database.TxManager,
	ids idgen.Generator,
	clk clock.Clock,
	expiry time.Duration,
) Service {
	return &service{
		items:        items,
		reservations: reservations,
		txm:          txm,
		ids:          ids,
		clock:        clk,
		expiry:       expiry,
	}
}

// Reserve atomically holds stock and creates an expiring reservation.
func (s *service) Reserve(ctx context.Context, req ReserveRequest) (*ReserveResponse, error) {
	if req.Quantity <= 0 {
		return nil, httpx.ErrInvalidRequest.WithDetails(httpx.ErrorDetail{
			Field:   "quantity",
			Rule:    "gt",
			Message: "quantity harus lebih dari 0",
		})
	}

	for attempt := 1; attempt <= maxReservationIDAttempts; attempt++ {
		id, err := s.ids.NewReservationID()
		if err != nil {
			return nil, httpx.Internal(err)
		}
		now := s.clock.Now()
		expiresAt := now.Add(s.expiry)

		var response *ReserveResponse
		err = s.txm.WithTx(ctx, func(txCtx context.Context) error {
			reserved, err := s.items.ReserveStock(txCtx, req.ItemID, req.Quantity)
			if err != nil {
				return err
			}
			if !reserved {
				// Distinguish "item does not exist" from "not enough stock".
				if _, err := s.items.GetByID(txCtx, req.ItemID); err != nil {
					return err
				}
				return ErrInsufficientStock
			}
			if err := s.reservations.Create(txCtx, &Reservation{
				ReservationID: id,
				UserID:        req.UserID,
				ItemID:        req.ItemID,
				Quantity:      req.Quantity,
				Status:        ReservationStatusActive,
				ExpiresAt:     expiresAt,
			}); err != nil {
				return err
			}
			response = &ReserveResponse{
				Status:        "success",
				ReservationID: id,
				ItemID:        req.ItemID,
				Quantity:      req.Quantity,
				ExpiresAt:     formatTime(expiresAt),
			}
			return nil
		})

		if errors.Is(err, ErrReservationIDConflict) {
			continue // regenerate the id and retry
		}
		if err != nil {
			return nil, err
		}
		return response, nil
	}

	return nil, httpx.Internal(errors.New("gagal menghasilkan id reservasi unik setelah beberapa percobaan"))
}

// Confirm permanently commits an active reservation and decrements physical
// stock in the same transaction.
func (s *service) Confirm(ctx context.Context, req ConfirmRequest) (*ConfirmResponse, error) {
	var response *ConfirmResponse
	err := s.txm.WithTx(ctx, func(txCtx context.Context) error {
		now := s.clock.Now()
		confirmed, ok, err := s.reservations.Confirm(txCtx, req.ReservationID, now)
		if err != nil {
			return err
		}
		if !ok {
			return classifyConfirmFailure(txCtx, s.reservations, req.ReservationID)
		}
		if err := s.items.ConfirmStock(txCtx, confirmed.ItemID, confirmed.Quantity); err != nil {
			return err
		}
		response = &ConfirmResponse{
			Status:        "success",
			ReservationID: req.ReservationID,
			ConfirmedAt:   formatTime(confirmed.ConfirmedAt.UTC()),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetStock returns the real-time stock breakdown for one item.
func (s *service) GetStock(ctx context.Context, itemID string) (*StockResponse, error) {
	item, err := s.items.GetByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	return &StockResponse{
		ItemID:         item.ItemID,
		TotalStock:     item.TotalStock,
		ReservedStock:  item.ReservedStock,
		AvailableStock: item.TotalStock - item.ReservedStock,
	}, nil
}

// classifyConfirmFailure inspects the reservation to return the most precise
// error: not found, already confirmed, or expired.
func classifyConfirmFailure(ctx context.Context, reservations ReservationRepository, reservationID string) error {
	reservation, err := reservations.GetByID(ctx, reservationID)
	if err != nil {
		return err
	}
	switch reservation.Status {
	case ReservationStatusConfirmed:
		return ErrReservationAlreadyConfirmed
	case ReservationStatusExpired:
		return ErrReservationExpired
	case ReservationStatusActive:
		// Still active but expires_at <= now: the conditional UPDATE guard
		// rejected it, so from the client's perspective it is expired.
		return ErrReservationExpired
	default:
		return httpx.Internal(fmt.Errorf("status reservasi tidak dikenal: %q", reservation.Status))
	}
}

// formatTime renders a UTC timestamp truncated to seconds (RFC3339), matching
// the API contract examples such as 2026-07-20T16:35:00Z.
func formatTime(t time.Time) string {
	return t.UTC().Truncate(time.Second).Format(time.RFC3339)
}
