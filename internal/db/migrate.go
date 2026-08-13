package db

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// RunMigrations applies all pending embedded SQL migrations to the database.
// If all migrations have already been applied, it returns nil (no-op).
func RunMigrations(dsn string, srcFS embed.FS) error {
	d, err := iofs.New(srcFS, ".")
	if err != nil {
		return fmt.Errorf("db: failed to initialize migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("db: failed to initialize migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("database schema is up to date, no migrations applied")
			return nil
		}
		return fmt.Errorf("db: migration failed: %w", err)
	}

	slog.Info("database migrations applied successfully")
	return nil
}
