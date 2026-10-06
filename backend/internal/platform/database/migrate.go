package database

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // registers the postgres driver for golang-migrate
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// migrator is the subset of *migrate.Migrate used by RunMigrations.
type migrator interface {
	Up() error
	Close() (source error, database error)
}

// newMigrator is a seam so unit tests can inject migration failures.
var newMigrator = func(fsys fs.FS, databaseURL string) (migrator, error) {
	source, err := iofs.New(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file migrasi: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat instance migrasi: %w", err)
	}
	return m, nil
}

// RunMigrations applies every pending migration from the embedded SQL files.
// Returns nil when the database is already up to date.
func RunMigrations(databaseURL string, fsys fs.FS) error {
	m, err := newMigrator(fsys, databaseURL)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("gagal menjalankan migrasi: %w", err)
	}
	return nil
}
