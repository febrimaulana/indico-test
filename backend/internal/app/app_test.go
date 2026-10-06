package app

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"indico/backend/internal/platform/config"
)

type fakeRunnable struct {
	fn func(ctx context.Context) error
}

func (f *fakeRunnable) Run(ctx context.Context) error { return f.fn(ctx) }

type fakeCloser struct {
	err error
}

func (f fakeCloser) Close() error { return f.err }

func testLog() *slog.Logger { return slog.New(slog.NewTextHandler(os.Stderr, nil)) }

func TestRunHappyPath(t *testing.T) {
	application := newApp(
		&fakeRunnable{fn: func(context.Context) error { return nil }},
		&fakeRunnable{fn: func(ctx context.Context) error { <-ctx.Done(); return nil }},
		fakeCloser{},
		testLog(),
	)
	require.NoError(t, application.Run(context.Background()))
}

func TestRunWorkerErrorStillShutsDownCleanly(t *testing.T) {
	application := newApp(
		&fakeRunnable{fn: func(context.Context) error { return nil }},
		&fakeRunnable{fn: func(context.Context) error { return errors.New("worker gagal") }},
		fakeCloser{},
		testLog(),
	)
	require.NoError(t, application.Run(context.Background()))
}

func TestRunCloserError(t *testing.T) {
	application := newApp(
		&fakeRunnable{fn: func(context.Context) error { return nil }},
		&fakeRunnable{fn: func(ctx context.Context) error { <-ctx.Done(); return nil }},
		fakeCloser{err: errors.New("close gagal")},
		testLog(),
	)
	err := application.Run(context.Background())
	require.ErrorContains(t, err, "gagal menutup koneksi database")
}

func TestRunServerErrorTakesPrecedence(t *testing.T) {
	application := newApp(
		&fakeRunnable{fn: func(context.Context) error { return errors.New("server gagal") }},
		&fakeRunnable{fn: func(context.Context) error { return nil }},
		fakeCloser{err: errors.New("close gagal")},
		testLog(),
	)
	err := application.Run(context.Background())
	require.ErrorContains(t, err, "server gagal")
}

func TestNoopCloser(t *testing.T) {
	require.NoError(t, noopCloser{}.Close())
}

func TestNewOpenDBError(t *testing.T) {
	original := openDB
	t.Cleanup(func() { openDB = original })
	openDB = func(context.Context, config.Config) (*gorm.DB, error) {
		return nil, errors.New("db gagal")
	}

	_, err := New(config.Config{}, testLog())
	require.ErrorContains(t, err, "db gagal")
}

func TestNewWithDBMigrationError(t *testing.T) {
	original := runMigrations
	t.Cleanup(func() { runMigrations = original })
	runMigrations = func(string) error { return errors.New("migrasi gagal") }

	_, err := newWithDB(config.Config{}, testLog(), nil)
	require.ErrorContains(t, err, "migrasi gagal")
}
