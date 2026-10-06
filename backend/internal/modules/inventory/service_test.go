package inventory

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"indico/backend/internal/platform/clock"
	"indico/backend/internal/platform/database"
	"indico/backend/internal/platform/httpx"
	"indico/backend/internal/platform/idgen"
)

type fakeItems struct {
	getByIDFn func(ctx context.Context, itemID string) (*Item, error)
	reserveFn func(ctx context.Context, itemID string, quantity int) (bool, error)
	confirmFn func(ctx context.Context, itemID string, quantity int) error
	releaseFn func(ctx context.Context, itemID string, quantity int) error
}

func (f *fakeItems) GetByID(ctx context.Context, itemID string) (*Item, error) {
	return f.getByIDFn(ctx, itemID)
}

func (f *fakeItems) ReserveStock(ctx context.Context, itemID string, quantity int) (bool, error) {
	return f.reserveFn(ctx, itemID, quantity)
}

func (f *fakeItems) ConfirmStock(ctx context.Context, itemID string, quantity int) error {
	return f.confirmFn(ctx, itemID, quantity)
}

func (f *fakeItems) ReleaseStock(ctx context.Context, itemID string, quantity int) error {
	return f.releaseFn(ctx, itemID, quantity)
}

type fakeReservations struct {
	createFn       func(ctx context.Context, reservation *Reservation) error
	getByIDFn      func(ctx context.Context, reservationID string) (*Reservation, error)
	confirmFn      func(ctx context.Context, reservationID string, now time.Time) (*Reservation, bool, error)
	claimExpiredFn func(ctx context.Context, now time.Time, limit int) ([]Reservation, error)
}

func (f *fakeReservations) Create(ctx context.Context, reservation *Reservation) error {
	return f.createFn(ctx, reservation)
}

func (f *fakeReservations) GetByID(ctx context.Context, reservationID string) (*Reservation, error) {
	return f.getByIDFn(ctx, reservationID)
}

func (f *fakeReservations) Confirm(ctx context.Context, reservationID string, now time.Time) (*Reservation, bool, error) {
	return f.confirmFn(ctx, reservationID, now)
}

func (f *fakeReservations) ClaimExpired(ctx context.Context, now time.Time, limit int) ([]Reservation, error) {
	return f.claimExpiredFn(ctx, now, limit)
}

type fakeTx struct {
	fn func(ctx context.Context, fn func(context.Context) error) error
}

func (f *fakeTx) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if f.fn != nil {
		return f.fn(ctx, fn)
	}
	return fn(ctx)
}

type fakeIDGen struct {
	calls int
	fn    func() (string, error)
}

func (f *fakeIDGen) NewReservationID() (string, error) {
	f.calls++
	return f.fn()
}

type fixedClock struct{ current time.Time }

func (c fixedClock) Now() time.Time { return c.current }

var testNow = time.Date(2026, 7, 20, 16, 30, 0, 0, time.UTC)

func asAppError(t *testing.T, err error) *httpx.AppError {
	t.Helper()
	var appErr *httpx.AppError
	require.ErrorAs(t, err, &appErr)
	return appErr
}

func TestServiceReserveSuccess(t *testing.T) {
	var created *Reservation
	items := &fakeItems{
		reserveFn: func(context.Context, string, int) (bool, error) { return true, nil },
	}
	reservations := &fakeReservations{
		createFn: func(_ context.Context, reservation *Reservation) error {
			created = reservation
			return nil
		},
	}
	ids := &fakeIDGen{fn: func() (string, error) { return "res_883291", nil }}
	service := NewService(items, reservations, &fakeTx{}, ids, fixedClock{testNow}, 5*time.Minute)

	response, err := service.Reserve(context.Background(), ReserveRequest{UserID: "usr_9981", ItemID: "item_4021", Quantity: 2})

	require.NoError(t, err)
	require.Equal(t, "success", response.Status)
	require.Equal(t, "res_883291", response.ReservationID)
	require.Equal(t, "item_4021", response.ItemID)
	require.Equal(t, 2, response.Quantity)
	require.Equal(t, "2026-07-20T16:35:00Z", response.ExpiresAt)

	require.NotNil(t, created)
	require.Equal(t, ReservationStatusActive, created.Status)
	require.Equal(t, testNow.Add(5*time.Minute), created.ExpiresAt)
	require.Equal(t, "usr_9981", created.UserID)
}

func TestServiceReserveInvalidQuantity(t *testing.T) {
	service := NewService(&fakeItems{}, &fakeReservations{}, &fakeTx{}, &fakeIDGen{}, fixedClock{testNow}, time.Minute)

	_, err := service.Reserve(context.Background(), ReserveRequest{UserID: "usr_9981", ItemID: "item_4021", Quantity: 0})

	appErr := asAppError(t, err)
	require.Equal(t, 400, appErr.Status)
	require.Equal(t, httpx.CodeInvalidRequest, appErr.Code)
	require.Len(t, appErr.Details, 1)
}

func TestServiceReserveIDGenerationError(t *testing.T) {
	ids := &fakeIDGen{fn: func() (string, error) { return "", errors.New("rand gagal") }}
	service := NewService(&fakeItems{}, &fakeReservations{}, &fakeTx{}, ids, fixedClock{testNow}, time.Minute)

	_, err := service.Reserve(context.Background(), ReserveRequest{UserID: "usr_9981", ItemID: "item_4021", Quantity: 1})

	appErr := asAppError(t, err)
	require.Equal(t, 500, appErr.Status)
	require.ErrorContains(t, appErr, "rand gagal")
}

func TestServiceReserveStockError(t *testing.T) {
	boom := errors.New("db error")
	items := &fakeItems{reserveFn: func(context.Context, string, int) (bool, error) { return false, boom }}
	ids := &fakeIDGen{fn: func() (string, error) { return "res_000001", nil }}
	service := NewService(items, &fakeReservations{}, &fakeTx{}, ids, fixedClock{testNow}, time.Minute)

	_, err := service.Reserve(context.Background(), ReserveRequest{UserID: "usr_9981", ItemID: "item_4021", Quantity: 1})
	require.ErrorIs(t, err, boom)
}

func TestServiceReserveItemNotFound(t *testing.T) {
	items := &fakeItems{
		reserveFn: func(context.Context, string, int) (bool, error) { return false, nil },
		getByIDFn: func(context.Context, string) (*Item, error) { return nil, ErrItemNotFound },
	}
	ids := &fakeIDGen{fn: func() (string, error) { return "res_000001", nil }}
	service := NewService(items, &fakeReservations{}, &fakeTx{}, ids, fixedClock{testNow}, time.Minute)

	_, err := service.Reserve(context.Background(), ReserveRequest{UserID: "usr_9981", ItemID: "item_9999", Quantity: 1})
	require.ErrorIs(t, err, ErrItemNotFound)
}

func TestServiceReserveItemLookupError(t *testing.T) {
	boom := errors.New("lookup gagal")
	items := &fakeItems{
		reserveFn: func(context.Context, string, int) (bool, error) { return false, nil },
		getByIDFn: func(context.Context, string) (*Item, error) { return nil, boom },
	}
	ids := &fakeIDGen{fn: func() (string, error) { return "res_000001", nil }}
	service := NewService(items, &fakeReservations{}, &fakeTx{}, ids, fixedClock{testNow}, time.Minute)

	_, err := service.Reserve(context.Background(), ReserveRequest{UserID: "usr_9981", ItemID: "item_4021", Quantity: 1})
	require.ErrorIs(t, err, boom)
}

func TestServiceReserveInsufficientStock(t *testing.T) {
	items := &fakeItems{
		reserveFn: func(context.Context, string, int) (bool, error) { return false, nil },
		getByIDFn: func(context.Context, string) (*Item, error) {
			return &Item{ItemID: "item_4021", TotalStock: 1, ReservedStock: 1}, nil
		},
	}
	ids := &fakeIDGen{fn: func() (string, error) { return "res_000001", nil }}
	service := NewService(items, &fakeReservations{}, &fakeTx{}, ids, fixedClock{testNow}, time.Minute)

	_, err := service.Reserve(context.Background(), ReserveRequest{UserID: "usr_9981", ItemID: "item_4021", Quantity: 2})
	require.ErrorIs(t, err, ErrInsufficientStock)
}

func TestServiceReserveRetriesOnIDConflict(t *testing.T) {
	attempts := 0
	ids := &fakeIDGen{fn: func() (string, error) {
		attempts++
		return fmt.Sprintf("res_%06d", attempts), nil
	}}
	createCalls := 0
	reservations := &fakeReservations{createFn: func(context.Context, *Reservation) error {
		createCalls++
		if createCalls == 1 {
			return ErrReservationIDConflict
		}
		return nil
	}}
	items := &fakeItems{reserveFn: func(context.Context, string, int) (bool, error) { return true, nil }}
	service := NewService(items, reservations, &fakeTx{}, ids, fixedClock{testNow}, time.Minute)

	response, err := service.Reserve(context.Background(), ReserveRequest{UserID: "usr_9981", ItemID: "item_4021", Quantity: 1})

	require.NoError(t, err)
	require.Equal(t, "res_000002", response.ReservationID)
	require.Equal(t, 2, attempts)
}

func TestServiceReserveIDConflictExhausted(t *testing.T) {
	ids := &fakeIDGen{fn: func() (string, error) { return "res_000001", nil }}
	reservations := &fakeReservations{createFn: func(context.Context, *Reservation) error {
		return ErrReservationIDConflict
	}}
	items := &fakeItems{reserveFn: func(context.Context, string, int) (bool, error) { return true, nil }}
	service := NewService(items, reservations, &fakeTx{}, ids, fixedClock{testNow}, time.Minute)

	_, err := service.Reserve(context.Background(), ReserveRequest{UserID: "usr_9981", ItemID: "item_4021", Quantity: 1})

	appErr := asAppError(t, err)
	require.Equal(t, 500, appErr.Status)
	require.Contains(t, appErr.Error(), "id reservasi unik")
	require.Equal(t, maxReservationIDAttempts, ids.calls)
}

func TestServiceReserveCreateError(t *testing.T) {
	boom := errors.New("insert gagal")
	ids := &fakeIDGen{fn: func() (string, error) { return "res_000001", nil }}
	reservations := &fakeReservations{createFn: func(context.Context, *Reservation) error { return boom }}
	items := &fakeItems{reserveFn: func(context.Context, string, int) (bool, error) { return true, nil }}
	service := NewService(items, reservations, &fakeTx{}, ids, fixedClock{testNow}, time.Minute)

	_, err := service.Reserve(context.Background(), ReserveRequest{UserID: "usr_9981", ItemID: "item_4021", Quantity: 1})
	require.ErrorIs(t, err, boom)
}

func TestServiceReserveTxError(t *testing.T) {
	boom := errors.New("tx gagal")
	ids := &fakeIDGen{fn: func() (string, error) { return "res_000001", nil }}
	tx := &fakeTx{fn: func(context.Context, func(context.Context) error) error { return boom }}
	service := NewService(&fakeItems{}, &fakeReservations{}, tx, ids, fixedClock{testNow}, time.Minute)

	_, err := service.Reserve(context.Background(), ReserveRequest{UserID: "usr_9981", ItemID: "item_4021", Quantity: 1})
	require.ErrorIs(t, err, boom)
}

func TestServiceConfirmSuccess(t *testing.T) {
	confirmedAt := testNow
	var confirmedItem string
	var confirmedQuantity int

	items := &fakeItems{confirmFn: func(_ context.Context, itemID string, quantity int) error {
		confirmedItem = itemID
		confirmedQuantity = quantity
		return nil
	}}
	reservations := &fakeReservations{confirmFn: func(context.Context, string, time.Time) (*Reservation, bool, error) {
		return &Reservation{
			ReservationID: "res_883291",
			ItemID:        "item_4021",
			Quantity:      2,
			ConfirmedAt:   &confirmedAt,
		}, true, nil
	}}
	service := NewService(items, reservations, &fakeTx{}, &fakeIDGen{}, fixedClock{testNow}, time.Minute)

	response, err := service.Confirm(context.Background(), ConfirmRequest{ReservationID: "res_883291"})

	require.NoError(t, err)
	require.Equal(t, "success", response.Status)
	require.Equal(t, "res_883291", response.ReservationID)
	require.Equal(t, "2026-07-20T16:30:00Z", response.ConfirmedAt)
	require.Equal(t, "item_4021", confirmedItem)
	require.Equal(t, 2, confirmedQuantity)
}

func TestServiceConfirmRepositoryError(t *testing.T) {
	boom := errors.New("update gagal")
	reservations := &fakeReservations{confirmFn: func(context.Context, string, time.Time) (*Reservation, bool, error) {
		return nil, false, boom
	}}
	service := NewService(&fakeItems{}, reservations, &fakeTx{}, &fakeIDGen{}, fixedClock{testNow}, time.Minute)

	_, err := service.Confirm(context.Background(), ConfirmRequest{ReservationID: "res_1"})
	require.ErrorIs(t, err, boom)
}

func TestServiceConfirmNotFound(t *testing.T) {
	reservations := &fakeReservations{
		confirmFn:  func(context.Context, string, time.Time) (*Reservation, bool, error) { return nil, false, nil },
		getByIDFn: func(context.Context, string) (*Reservation, error) { return nil, ErrReservationNotFound },
	}
	service := NewService(&fakeItems{}, reservations, &fakeTx{}, &fakeIDGen{}, fixedClock{testNow}, time.Minute)

	_, err := service.Confirm(context.Background(), ConfirmRequest{ReservationID: "res_404"})
	require.ErrorIs(t, err, ErrReservationNotFound)
}

func TestServiceConfirmAlreadyConfirmed(t *testing.T) {
	reservations := &fakeReservations{
		confirmFn:  func(context.Context, string, time.Time) (*Reservation, bool, error) { return nil, false, nil },
		getByIDFn: func(context.Context, string) (*Reservation, error) {
			return &Reservation{Status: ReservationStatusConfirmed}, nil
		},
	}
	service := NewService(&fakeItems{}, reservations, &fakeTx{}, &fakeIDGen{}, fixedClock{testNow}, time.Minute)

	_, err := service.Confirm(context.Background(), ConfirmRequest{ReservationID: "res_1"})
	require.ErrorIs(t, err, ErrReservationAlreadyConfirmed)
}

func TestServiceConfirmExpiredStatus(t *testing.T) {
	reservations := &fakeReservations{
		confirmFn:  func(context.Context, string, time.Time) (*Reservation, bool, error) { return nil, false, nil },
		getByIDFn: func(context.Context, string) (*Reservation, error) {
			return &Reservation{Status: ReservationStatusExpired}, nil
		},
	}
	service := NewService(&fakeItems{}, reservations, &fakeTx{}, &fakeIDGen{}, fixedClock{testNow}, time.Minute)

	_, err := service.Confirm(context.Background(), ConfirmRequest{ReservationID: "res_1"})
	require.ErrorIs(t, err, ErrReservationExpired)
}

func TestServiceConfirmActiveButExpired(t *testing.T) {
	reservations := &fakeReservations{
		confirmFn:  func(context.Context, string, time.Time) (*Reservation, bool, error) { return nil, false, nil },
		getByIDFn: func(context.Context, string) (*Reservation, error) {
			return &Reservation{Status: ReservationStatusActive}, nil
		},
	}
	service := NewService(&fakeItems{}, reservations, &fakeTx{}, &fakeIDGen{}, fixedClock{testNow}, time.Minute)

	_, err := service.Confirm(context.Background(), ConfirmRequest{ReservationID: "res_1"})
	require.ErrorIs(t, err, ErrReservationExpired)
}

func TestServiceConfirmUnknownStatus(t *testing.T) {
	reservations := &fakeReservations{
		confirmFn:  func(context.Context, string, time.Time) (*Reservation, bool, error) { return nil, false, nil },
		getByIDFn: func(context.Context, string) (*Reservation, error) {
			return &Reservation{Status: "status_aneh"}, nil
		},
	}
	service := NewService(&fakeItems{}, reservations, &fakeTx{}, &fakeIDGen{}, fixedClock{testNow}, time.Minute)

	_, err := service.Confirm(context.Background(), ConfirmRequest{ReservationID: "res_1"})
	appErr := asAppError(t, err)
	require.Equal(t, 500, appErr.Status)
	require.Contains(t, appErr.Error(), "status_aneh")
}

func TestServiceConfirmStockError(t *testing.T) {
	boom := errors.New("stok gagal")
	items := &fakeItems{confirmFn: func(context.Context, string, int) error { return boom }}
	confirmedAt := testNow
	reservations := &fakeReservations{confirmFn: func(context.Context, string, time.Time) (*Reservation, bool, error) {
		return &Reservation{ItemID: "item_4021", Quantity: 1, ConfirmedAt: &confirmedAt}, true, nil
	}}
	service := NewService(items, reservations, &fakeTx{}, &fakeIDGen{}, fixedClock{testNow}, time.Minute)

	_, err := service.Confirm(context.Background(), ConfirmRequest{ReservationID: "res_1"})
	require.ErrorIs(t, err, boom)
}

func TestServiceConfirmTxError(t *testing.T) {
	boom := errors.New("tx gagal")
	tx := &fakeTx{fn: func(context.Context, func(context.Context) error) error { return boom }}
	service := NewService(&fakeItems{}, &fakeReservations{}, tx, &fakeIDGen{}, fixedClock{testNow}, time.Minute)

	_, err := service.Confirm(context.Background(), ConfirmRequest{ReservationID: "res_1"})
	require.ErrorIs(t, err, boom)
}

func TestServiceGetStock(t *testing.T) {
	items := &fakeItems{getByIDFn: func(context.Context, string) (*Item, error) {
		return &Item{ItemID: "item_4021", TotalStock: 100, ReservedStock: 15}, nil
	}}
	service := NewService(items, &fakeReservations{}, &fakeTx{}, &fakeIDGen{}, fixedClock{testNow}, time.Minute)

	response, err := service.GetStock(context.Background(), "item_4021")

	require.NoError(t, err)
	require.Equal(t, "item_4021", response.ItemID)
	require.Equal(t, 100, response.TotalStock)
	require.Equal(t, 15, response.ReservedStock)
	require.Equal(t, 85, response.AvailableStock)
}

func TestServiceGetStockNotFound(t *testing.T) {
	items := &fakeItems{getByIDFn: func(context.Context, string) (*Item, error) { return nil, ErrItemNotFound }}
	service := NewService(items, &fakeReservations{}, &fakeTx{}, &fakeIDGen{}, fixedClock{testNow}, time.Minute)

	_, err := service.GetStock(context.Background(), "item_9999")
	require.ErrorIs(t, err, ErrItemNotFound)
}

// Compile-time interface checks keep the fakes honest.
var (
	_ ItemRepository        = (*fakeItems)(nil)
	_ ReservationRepository = (*fakeReservations)(nil)
	_ database.TxManager    = (*fakeTx)(nil)
	_ idgen.Generator       = (*fakeIDGen)(nil)
	_ clock.Clock           = fixedClock{}
)
