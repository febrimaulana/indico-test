package inventory

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testLogger(buffer *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buffer, nil))
}

type releaseCall struct {
	itemID   string
	quantity int
}

func TestWorkerSweepAggregatesAndSorts(t *testing.T) {
	var releases []releaseCall
	items := &fakeItems{releaseFn: func(_ context.Context, itemID string, quantity int) error {
		releases = append(releases, releaseCall{itemID: itemID, quantity: quantity})
		return nil
	}}
	reservations := &fakeReservations{claimExpiredFn: func(_ context.Context, _ time.Time, limit int) ([]Reservation, error) {
		require.Equal(t, 3, limit)
		return []Reservation{
			{ReservationID: "res_2", ItemID: "item_b", Quantity: 2},
			{ReservationID: "res_1", ItemID: "item_a", Quantity: 1},
			{ReservationID: "res_3", ItemID: "item_a", Quantity: 3},
		}, nil
	}}

	var buffer bytes.Buffer
	worker := NewExpiryWorker(reservations, items, &fakeTx{}, fixedClock{testNow}, time.Minute, 3, testLogger(&buffer))

	worker.Sweep(context.Background())

	require.Equal(t, []releaseCall{{"item_a", 4}, {"item_b", 2}}, releases)
	require.Contains(t, buffer.String(), "reservasi kedaluwarsa dikembalikan")
}

func TestWorkerSweepNothingExpired(t *testing.T) {
	items := &fakeItems{releaseFn: func(context.Context, string, int) error {
		t.Fatal("ReleaseStock tidak boleh dipanggil")
		return nil
	}}
	reservations := &fakeReservations{claimExpiredFn: func(context.Context, time.Time, int) ([]Reservation, error) {
		return nil, nil
	}}

	var buffer bytes.Buffer
	worker := NewExpiryWorker(reservations, items, &fakeTx{}, fixedClock{testNow}, time.Minute, 10, testLogger(&buffer))

	worker.Sweep(context.Background())

	require.NotContains(t, buffer.String(), "reservasi kedaluwarsa dikembalikan")
	require.Empty(t, buffer.String())
}

func TestWorkerSweepClaimError(t *testing.T) {
	reservations := &fakeReservations{claimExpiredFn: func(context.Context, time.Time, int) ([]Reservation, error) {
		return nil, errors.New("claim gagal")
	}}

	var buffer bytes.Buffer
	worker := NewExpiryWorker(reservations, &fakeItems{}, &fakeTx{}, fixedClock{testNow}, time.Minute, 10, testLogger(&buffer))

	worker.Sweep(context.Background())

	require.Contains(t, buffer.String(), "sweep reservasi kedaluwarsa gagal")
	require.Contains(t, buffer.String(), "claim gagal")
}

func TestWorkerSweepReleaseError(t *testing.T) {
	items := &fakeItems{releaseFn: func(context.Context, string, int) error {
		return errors.New("release gagal")
	}}
	reservations := &fakeReservations{claimExpiredFn: func(context.Context, time.Time, int) ([]Reservation, error) {
		return []Reservation{{ReservationID: "res_1", ItemID: "item_a", Quantity: 1}}, nil
	}}

	var buffer bytes.Buffer
	worker := NewExpiryWorker(reservations, items, &fakeTx{}, fixedClock{testNow}, time.Minute, 10, testLogger(&buffer))

	worker.Sweep(context.Background())

	require.Contains(t, buffer.String(), "release gagal")
}

func TestWorkerSweepTxError(t *testing.T) {
	tx := &fakeTx{fn: func(context.Context, func(context.Context) error) error {
		return errors.New("tx gagal")
	}}

	var buffer bytes.Buffer
	worker := NewExpiryWorker(&fakeReservations{}, &fakeItems{}, tx, fixedClock{testNow}, time.Minute, 10, testLogger(&buffer))

	worker.Sweep(context.Background())

	require.Contains(t, buffer.String(), "tx gagal")
}

func TestWorkerRunSweepsUntilCanceled(t *testing.T) {
	var claimCalls atomic.Int32
	reservations := &fakeReservations{claimExpiredFn: func(context.Context, time.Time, int) ([]Reservation, error) {
		claimCalls.Add(1)
		return nil, nil
	}}

	var buffer bytes.Buffer
	worker := NewExpiryWorker(reservations, &fakeItems{}, &fakeTx{}, fixedClock{testNow}, 5*time.Millisecond, 10, testLogger(&buffer))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()

	require.Eventually(t, func() bool { return claimCalls.Load() >= 3 }, time.Second, 5*time.Millisecond)
	cancel()

	require.NoError(t, <-done)
	require.Contains(t, buffer.String(), "worker expiry berhenti")
}
