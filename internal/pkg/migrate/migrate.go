package migrate

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func New(databaseURL, migrationsPath string) (*migrate.Migrate, error) {
	return migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
}

func Up(m *migrate.Migrate) error {
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func Down(m *migrate.Migrate) error {
	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func Steps(m *migrate.Migrate, n int) error {
	if err := m.Steps(n); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func Version(m *migrate.Migrate) (uint, bool, error) {
	return m.Version()
}

func Force(m *migrate.Migrate, version int) error {
	return m.Force(version)
}
