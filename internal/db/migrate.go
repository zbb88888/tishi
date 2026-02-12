package db

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations executes database migrations.
// direction must be "up" or "down".
func RunMigrations(dsn string, direction string) error {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("creating migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil || dbErr != nil {
			// close errors are non-fatal during migration
			_ = srcErr
			_ = dbErr
		}
	}()

	switch direction {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("running migrations up: %w", err)
		}
	case "down":
		if err := m.Steps(-1); err != nil {
			return fmt.Errorf("running migration down: %w", err)
		}
	default:
		return fmt.Errorf("invalid migration direction: %s", direction)
	}

	return nil
}
