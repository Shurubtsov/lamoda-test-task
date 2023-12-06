package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/Shurubtsov/lamoda-test-task/internal/config"
	"github.com/Shurubtsov/lamoda-test-task/pkg/logging"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client Интерфейс для имплементации методов из драйвера PGX для быстродействия.
type Client interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

// NewClient создаёт новый клиент пула соединений pgxpool от PGX драйвера для PostgreSQL.
func NewClient(ctx context.Context, maxAttemts int) (pool *pgxpool.Pool, err error) {
	cfg := config.GetConfig()
	logger := logging.GetLogger().Logger

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", cfg.Storage.Username, cfg.Storage.Password, cfg.Storage.Host, cfg.Storage.Port, cfg.Storage.Database)
	logger.Info().Str("dsn", dsn).Msg("Start connect to database")

	err = doWithTries(func() error {
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		pool, err = pgxpool.New(ctx, dsn)
		if err != nil {
			return err
		}

		// пингуем базу данных для подтверждения соединения
		err = pool.Ping(ctx)
		if err != nil {
			return err
		}

		return nil

	}, maxAttemts, 5*time.Second)
	if err != nil {
		logger.Fatal().Err(err).Msg("Can't connect to database")
	}

	logger.Info().Msg("Connected to database")
	return pool, nil
}

// doWithTries функция для воспроизведения нескольких попыток создать пул соединений.
func doWithTries(fn func() error, attemts int, delay time.Duration) (err error) {
	for attemts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			attemts--

			continue
		}
		return nil
	}

	return
}
