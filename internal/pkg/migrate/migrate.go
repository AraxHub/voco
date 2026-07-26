package migrate

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"voco/migrations"
)

func New(databaseURL, migrationsPath string) (*migrate.Migrate, error) {
	return migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
}

// NewFromFS создаёт migrator из SQL, вшитых в бинарь (go:embed).
func NewFromFS(databaseURL string) (*migrate.Migrate, error) {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("migrate source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("migrate init: %w", err)
	}
	return m, nil
}

// UpEmbedded накатывает все неприменённые миграции из embed.
func UpEmbedded(databaseURL string) error {
	m, err := NewFromFS(databaseURL)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()
	return Up(m)
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
