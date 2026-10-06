package database

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fakeMigrator struct {
	upErr  error
	closed bool
}

func (f *fakeMigrator) Up() error { return f.upErr }
func (f *fakeMigrator) Close() (error, error) {
	f.closed = true
	return nil, nil
}

type failingFS struct{}

func (failingFS) Open(string) (fs.File, error) { return nil, errors.New("boom") }

func TestRunMigrationsSuccess(t *testing.T) {
	original := newMigrator
	t.Cleanup(func() { newMigrator = original })

	fake := &fakeMigrator{}
	newMigrator = func(fs.FS, string) (migrator, error) { return fake, nil }

	require.NoError(t, RunMigrations("dsn", fstest.MapFS{}))
	require.True(t, fake.closed)
}

func TestRunMigrationsNoChange(t *testing.T) {
	original := newMigrator
	t.Cleanup(func() { newMigrator = original })

	newMigrator = func(fs.FS, string) (migrator, error) {
		return &fakeMigrator{upErr: migrate.ErrNoChange}, nil
	}
	require.NoError(t, RunMigrations("dsn", fstest.MapFS{}))
}

func TestRunMigrationsUpError(t *testing.T) {
	original := newMigrator
	t.Cleanup(func() { newMigrator = original })

	newMigrator = func(fs.FS, string) (migrator, error) {
		return &fakeMigrator{upErr: errors.New("boom")}, nil
	}
	err := RunMigrations("dsn", fstest.MapFS{})
	require.ErrorContains(t, err, "gagal menjalankan migrasi")
}

func TestRunMigrationsInitError(t *testing.T) {
	original := newMigrator
	t.Cleanup(func() { newMigrator = original })

	newMigrator = func(fs.FS, string) (migrator, error) { return nil, errors.New("init gagal") }
	err := RunMigrations("dsn", fstest.MapFS{})
	require.ErrorContains(t, err, "init gagal")
}

func TestNewMigratorRealIOFSError(t *testing.T) {
	// A filesystem whose Open always fails makes iofs.New return an error.
	_, err := newMigratorReal(failingFS{}, "postgres://user:pass@127.0.0.1:1/none?sslmode=disable")
	require.ErrorContains(t, err, "gagal membaca file migrasi")
}

func TestNewMigratorRealDatabaseError(t *testing.T) {
	// A valid filesystem with an invalid database URL fails when the migrate
	// instance is created.
	files := fstest.MapFS{
		"000001_test.up.sql":   {Data: []byte("SELECT 1;")},
		"000001_test.down.sql": {Data: []byte("SELECT 1;")},
	}
	_, err := newMigratorReal(files, "://not-a-dsn")
	require.ErrorContains(t, err, "gagal membuat instance migrasi")
}

func TestConnectionFallback(t *testing.T) {
	fallback := &gorm.DB{}
	require.Same(t, fallback, Connection(context.Background(), fallback))

	tx := &gorm.DB{}
	ctx := context.WithValue(context.Background(), txKey{}, tx)
	require.Same(t, tx, Connection(ctx, fallback))
}
