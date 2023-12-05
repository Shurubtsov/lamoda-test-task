package db

import (
	"fmt"

	"github.com/Shurubtsov/lamoda-test-task/internal/config"
	"github.com/Shurubtsov/lamoda-test-task/pkg/logging"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(source string, version uint) error {
	cfg := config.GetConfig()
	client := &pgx.Postgres{}

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", cfg.Storage.Username, cfg.Storage.Password, cfg.Storage.Host, cfg.Storage.Port, cfg.Storage.Database)
	d, err := client.Open(dsn)
	if err != nil {
		return err
	}
	defer func() {
		if err := d.Close(); err != nil {
			logger := logging.GetLogger()
			logger.Warn().Err(err).Msg("can't close connection with postgres client")
			return
		}
	}()

	m, err := migrate.NewWithDatabaseInstance(source, cfg.Storage.Database, d)
	if err != nil {
		return err
	}

	if err := m.Migrate(version); err != nil {
		return err
	}

	return nil
}
