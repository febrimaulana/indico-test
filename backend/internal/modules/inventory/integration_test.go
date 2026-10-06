package inventory

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"indico/backend/internal/platform/clock"
	"indico/backend/internal/platform/database"
	"indico/backend/internal/platform/idgen"
	"indico/backend/migrations"
)

const (
	integrationItemID = "item_test"
	integrationStock  = 10
	// Default database when TEST_DATABASE_URL is unset: matches the compose
	// PostgreSQL (host port 5433) and the Makefile.
	defaultIntegrationDSN = "postgres://indico:indico@localhost:5433/indico_test?sslmode=disable&connect_timeout=2"
)

var (
	integrationOnce sync.Once
	integrationDB   *gorm.DB
	integrationErr  error
	// integrationDSN holds the DSN actually used (env or auto-detected).
	integrationDSN string
)

// requireIntegrationDB returns the shared integration connection. It uses
// TEST_DATABASE_URL when set, otherwise it falls back to the compose
// PostgreSQL and creates the test database on the fly. When no database is
// reachable the test is skipped (not failed) so `go test ./...` stays green
// on machines without PostgreSQL.
func requireIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	explicit := dsn != ""
	if !explicit {
		dsn = defaultIntegrationDSN
		if !canOpenIntegrationDatabase(dsn) {
			createIntegrationDatabase(dsn)
			if !canOpenIntegrationDatabase(dsn) {
				t.Skipf("PostgreSQL tidak tersedia di %s - melewati integration test", dsn)
			}
		}
	}

	integrationDSN = dsn
	integrationOnce.Do(func() {
		integrationDB, integrationErr = database.OpenWithRetry(
			context.Background(),
			dsn,
			database.PoolConfig{MaxOpenConns: 20, MaxIdleConns: 5, ConnMaxLifetime: time.Minute},
			5,
			200*time.Millisecond,
		)
		if integrationErr == nil {
			integrationErr = database.RunMigrations(dsn, migrations.FS)
		}
	})
	require.NoError(t, integrationErr)
	return integrationDB
}

// canOpenIntegrationDatabase reports whether the DSN is reachable.
func canOpenIntegrationDatabase(dsn string) bool {
	db, err := database.Open(dsn, database.PoolConfig{MaxOpenConns: 1, MaxIdleConns: 1, ConnMaxLifetime: time.Minute})
	if err != nil {
		return false
	}
	if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
		_ = sqlDB.Close()
	}
	return true
}

// createIntegrationDatabase connects through the maintenance database and
// creates the configured test database, mirroring `make ensure-test-db`.
func createIntegrationDatabase(dsn string) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return
	}
	name := strings.TrimPrefix(parsed.Path, "/")
	if !validIntegrationDatabaseName(name) {
		return
	}

	adminURL := *parsed
	adminURL.Path = "/postgres"
	admin, err := database.Open(adminURL.String(), database.PoolConfig{MaxOpenConns: 1, MaxIdleConns: 1, ConnMaxLifetime: time.Minute})
	if err != nil {
		return
	}
	defer func() {
		if sqlDB, err := admin.DB(); err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	}()
	_ = admin.Exec(`CREATE DATABASE "` + name + `"`).Error
}

func validIntegrationDatabaseName(name string) bool {
	if name == "" {
		return false
	}
	for _, char := range name {
		isLower := char >= 'a' && char <= 'z'
		isDigit := char >= '0' && char <= '9'
		if !isLower && !isDigit && char != '_' {
			return false
		}
	}
	return true
}

func resetItem(t *testing.T, db *gorm.DB, totalStock int) {
	t.Helper()
	require.NoError(t, db.Exec("TRUNCATE reservations, items RESTART IDENTITY CASCADE").Error)
	require.NoError(t, db.Exec(
		"INSERT INTO items (item_id, name, total_stock, reserved_stock) VALUES (?, ?, ?, 0)",
		integrationItemID, "Integration Test Item", totalStock,
	).Error)
}

func newIntegrationService(db *gorm.DB, expiry time.Duration) (Service, ItemRepository, ReservationRepository, database.TxManager) {
	items := NewItemRepository(db)
	reservations := NewReservationRepository(db)
	txm := database.NewTxManager(db)
	service := NewService(items, reservations, txm, idgen.New(), clock.New(), expiry)
	return service, items, reservations, txm
}

func testSlog() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}

func TestIntegrationMigrationsIdempotent(t *testing.T) {
	requireIntegrationDB(t)
	require.NoError(t, database.RunMigrations(integrationDSN, migrations.FS))
}

func TestIntegrationReserveConfirmFlow(t *testing.T) {
	db := requireIntegrationDB(t)
	resetItem(t, db, 10)
	service, _, _, _ := newIntegrationService(db, 5*time.Minute)
	ctx := context.Background()

	reserveResponse, err := service.Reserve(ctx, ReserveRequest{UserID: "usr_0001", ItemID: integrationItemID, Quantity: 2})
	require.NoError(t, err)
	require.Equal(t, "success", reserveResponse.Status)
	require.Equal(t, 2, reserveResponse.Quantity)

	stock, err := service.GetStock(ctx, integrationItemID)
	require.NoError(t, err)
	require.Equal(t, 10, stock.TotalStock)
	require.Equal(t, 2, stock.ReservedStock)
	require.Equal(t, 8, stock.AvailableStock)

	confirmResponse, err := service.Confirm(ctx, ConfirmRequest{ReservationID: reserveResponse.ReservationID})
	require.NoError(t, err)
	require.Equal(t, "success", confirmResponse.Status)

	stock, err = service.GetStock(ctx, integrationItemID)
	require.NoError(t, err)
	require.Equal(t, 8, stock.TotalStock)
	require.Equal(t, 0, stock.ReservedStock)
	require.Equal(t, 8, stock.AvailableStock)

	_, err = service.Confirm(ctx, ConfirmRequest{ReservationID: reserveResponse.ReservationID})
	require.ErrorIs(t, err, ErrReservationAlreadyConfirmed)

	_, err = service.Confirm(ctx, ConfirmRequest{ReservationID: "res_999999"})
	require.ErrorIs(t, err, ErrReservationNotFound)
}

func TestIntegrationReserveInsufficientAndMissingItem(t *testing.T) {
	db := requireIntegrationDB(t)
	resetItem(t, db, 1)
	service, _, _, _ := newIntegrationService(db, 5*time.Minute)
	ctx := context.Background()

	_, err := service.Reserve(ctx, ReserveRequest{UserID: "usr_0001", ItemID: integrationItemID, Quantity: 2})
	require.ErrorIs(t, err, ErrInsufficientStock)

	_, err = service.Reserve(ctx, ReserveRequest{UserID: "usr_0001", ItemID: "item_missing", Quantity: 1})
	require.ErrorIs(t, err, ErrItemNotFound)

	_, err = service.GetStock(ctx, "item_missing")
	require.ErrorIs(t, err, ErrItemNotFound)
}

// TestIntegrationConcurrentReserveNeverOversells is the stress test required
// by the assignment: run under `go test -race` for full value.
func TestIntegrationConcurrentReserveNeverOversells(t *testing.T) {
	db := requireIntegrationDB(t)
	resetItem(t, db, integrationStock)
	service, items, _, _ := newIntegrationService(db, 5*time.Minute)
	ctx := context.Background()

	const goroutines = 200
	var waitGroup sync.WaitGroup
	results := make([]error, goroutines)
	for index := 0; index < goroutines; index++ {
		waitGroup.Add(1)
		go func(workerIndex int) {
			defer waitGroup.Done()
			_, err := service.Reserve(ctx, ReserveRequest{
				UserID:   fmt.Sprintf("usr_%05d", workerIndex),
				ItemID:   integrationItemID,
				Quantity: 1,
			})
			results[workerIndex] = err
		}(index)
	}
	waitGroup.Wait()

	successes, insufficient, unexpected := 0, 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrInsufficientStock):
			insufficient++
		default:
			unexpected++
		}
	}
	require.Equal(t, integrationStock, successes, "hanya stok yang tersedia boleh sukses")
	require.Equal(t, goroutines-integrationStock, insufficient)
	require.Zero(t, unexpected, "tidak boleh ada error tak terduga")

	item, err := items.GetByID(ctx, integrationItemID)
	require.NoError(t, err)
	require.Equal(t, integrationStock, item.ReservedStock)
	require.LessOrEqual(t, item.ReservedStock, item.TotalStock, "stok tidak boleh negatif / oversell")
}

func TestIntegrationSweeperReturnsExpiredStock(t *testing.T) {
	db := requireIntegrationDB(t)
	resetItem(t, db, 10)
	// Negative expiry makes every reservation already expired on creation.
	service, items, reservations, txm := newIntegrationService(db, -time.Second)
	ctx := context.Background()

	reserveResponse, err := service.Reserve(ctx, ReserveRequest{UserID: "usr_0001", ItemID: integrationItemID, Quantity: 3})
	require.NoError(t, err)

	stock, err := service.GetStock(ctx, integrationItemID)
	require.NoError(t, err)
	require.Equal(t, 3, stock.ReservedStock)

	// Confirm must reject the already-expired reservation (guard clause).
	_, err = service.Confirm(ctx, ConfirmRequest{ReservationID: reserveResponse.ReservationID})
	require.ErrorIs(t, err, ErrReservationExpired)

	worker := NewExpiryWorker(reservations, items, txm, clock.New(), time.Second, 100, testSlog())
	worker.Sweep(ctx)

	stock, err = service.GetStock(ctx, integrationItemID)
	require.NoError(t, err)
	require.Equal(t, 0, stock.ReservedStock)
	require.Equal(t, 10, stock.AvailableStock)

	// Confirming after the sweep reports the expired status.
	_, err = service.Confirm(ctx, ConfirmRequest{ReservationID: reserveResponse.ReservationID})
	require.ErrorIs(t, err, ErrReservationExpired)
}

func TestIntegrationRepositoryBranches(t *testing.T) {
	db := requireIntegrationDB(t)
	resetItem(t, db, 5)

	items := NewItemRepository(db)
	reservations := NewReservationRepository(db)
	ctx := context.Background()
	canceled, cancel := context.WithCancel(ctx)
	cancel()

	// Canceled contexts surface raw database errors.
	_, err := items.GetByID(canceled, integrationItemID)
	require.Error(t, err)
	_, err = items.ReserveStock(canceled, integrationItemID, 1)
	require.Error(t, err)
	require.Error(t, items.ConfirmStock(canceled, integrationItemID, 1))
	require.Error(t, items.ReleaseStock(canceled, integrationItemID, 1))
	require.Error(t, reservations.Create(canceled, &Reservation{
		ReservationID: "res_canceled",
		UserID:        "usr_0001",
		ItemID:        integrationItemID,
		Quantity:      1,
		Status:        ReservationStatusActive,
		ExpiresAt:     time.Now().UTC().Add(time.Minute),
	}))
	_, err = reservations.GetByID(canceled, "res_canceled")
	require.Error(t, err)
	_, _, err = reservations.Confirm(canceled, "res_canceled", time.Now().UTC())
	require.Error(t, err)
	_, err = reservations.ClaimExpired(canceled, time.Now().UTC(), 10)
	require.Error(t, err)

	// Zero rows affected branches.
	require.ErrorIs(t, items.ConfirmStock(ctx, "item_missing", 1), ErrItemNotFound)
	require.ErrorIs(t, items.ReleaseStock(ctx, "item_missing", 1), ErrItemNotFound)

	// Not-found branches.
	_, err = items.GetByID(ctx, "item_missing")
	require.ErrorIs(t, err, ErrItemNotFound)
	_, err = reservations.GetByID(ctx, "res_missing")
	require.ErrorIs(t, err, ErrReservationNotFound)

	// Conditional confirm with an unknown id: no error, not matched.
	confirmed, matched, err := reservations.Confirm(ctx, "res_missing", time.Now().UTC())
	require.NoError(t, err)
	require.False(t, matched)
	require.Nil(t, confirmed)

	// Empty claim returns no rows.
	claimed, err := reservations.ClaimExpired(ctx, time.Now().UTC(), 10)
	require.NoError(t, err)
	require.Empty(t, claimed)

	// Duplicate reservation id maps to the retry sentinel.
	reservation := &Reservation{
		ReservationID: "res_dupe",
		UserID:        "usr_0001",
		ItemID:        integrationItemID,
		Quantity:      1,
		Status:        ReservationStatusActive,
		ExpiresAt:     time.Now().UTC().Add(time.Minute),
	}
	require.NoError(t, reservations.Create(ctx, reservation))
	require.ErrorIs(t, reservations.Create(ctx, reservation), ErrReservationIDConflict)

	// ReserveStock returns false when stock is exhausted.
	updated, err := items.ReserveStock(ctx, integrationItemID, 5)
	require.NoError(t, err)
	require.True(t, updated)
	updated, err = items.ReserveStock(ctx, integrationItemID, 1)
	require.NoError(t, err)
	require.False(t, updated)
}

func TestIntegrationTxManagerCommitAndRollback(t *testing.T) {
	db := requireIntegrationDB(t)
	resetItem(t, db, 10)
	reservations := NewReservationRepository(db)
	txm := database.NewTxManager(db)
	ctx := context.Background()

	newReservation := func(id string) *Reservation {
		return &Reservation{
			ReservationID: id,
			UserID:        "usr_0001",
			ItemID:        integrationItemID,
			Quantity:      1,
			Status:        ReservationStatusActive,
			ExpiresAt:     time.Now().UTC().Add(time.Minute),
		}
	}

	// Commit path.
	require.NoError(t, txm.WithTx(ctx, func(txCtx context.Context) error {
		return reservations.Create(txCtx, newReservation("res_commit"))
	}))
	_, err := reservations.GetByID(ctx, "res_commit")
	require.NoError(t, err)

	// Rollback path.
	rollbackErr := errors.New("rollback disengaja")
	err = txm.WithTx(ctx, func(txCtx context.Context) error {
		if createErr := reservations.Create(txCtx, newReservation("res_rollback")); createErr != nil {
			return createErr
		}
		return rollbackErr
	})
	require.ErrorIs(t, err, rollbackErr)
	_, err = reservations.GetByID(ctx, "res_rollback")
	require.ErrorIs(t, err, ErrReservationNotFound)
}
