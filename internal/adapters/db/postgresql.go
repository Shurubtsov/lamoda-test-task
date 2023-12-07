package db

import (
	"context"
	"errors"
	"time"

	"github.com/Shurubtsov/lamoda-test-task/internal/domain/models"
	"github.com/Shurubtsov/lamoda-test-task/pkg/client/postgresql"
	"github.com/Shurubtsov/lamoda-test-task/pkg/logging"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
)

type repository struct {
	client postgresql.Client
}

func New(cl postgresql.Client) *repository {
	return &repository{client: cl}
}

func (r *repository) FindAviableStorage(ctx context.Context) (*models.Storage, error) {
	logger := logging.GetLogger()
	logger.Trace().Msg("start FindAviableStorage")
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	storage := &models.Storage{}
	q := `SELECT storage_id, storage_aviable, storage_name FROM storages WHERE storage_aviable = true LIMIT 1;`
	if err := r.client.QueryRow(ctx, q).Scan(&storage.ID, &storage.Aviable, &storage.Name); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			logger.Error().Msgf("SQL Error message (%s), Details: %s where -> %s", pgErr.Message, pgErr.Detail, pgErr.Where)
			return nil, nil
		}

		return nil, err
	}

	logger.Debug().Any("storage", *storage).Msg("storage info")

	return storage, nil
}

func (r *repository) FindProductsViaCode(ctx context.Context, products []models.Product) ([]models.Product, error) {
	logger := logging.GetLogger()
	logger.Trace().Msg("start FindProductsViaCode")
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	q := `SELECT product_id, product_name, product_size, product_count FROM products WHERE product_code = @product_code;`
	batch := &pgx.Batch{}
	for _, product := range products {
		args := pgx.NamedArgs{
			"product_code": product.Code,
		}
		batch.Queue(q, args)
	}
	results := r.client.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(products); i++ {
		product := &products[i]
		if err := results.QueryRow().Scan(&product.ID, &product.Name, &product.Size, &product.Count); err != nil {
			var (
				pgErr *pgconn.PgError
			)
			if errors.As(err, &pgErr) {
				logger.Error().Msgf("SQL Error message (%s), Details: %s where -> %s", pgErr.Message, pgErr.Detail, pgErr.Where)
			} else if errors.Is(err, pgx.ErrNoRows) {
				// было бы хорошо отсюда доставать те товары, которые не заведены в таблице
				logger.Warn().Err(err).Msgf("product with code: %s not exists", product.Code)
				continue
			}

			return nil, err
		}
	}

	return products, nil
}

func (r *repository) ReserveProducts(ctx context.Context, storage models.Storage, products []models.Product) error {
	logger := logging.GetLogger()
	logger.Trace().Msg("start ReserveProducts")
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	q := `INSERT INTO reservation (storage_id, product_id) VALUES (@storageID, @productID)`
	batch := &pgx.Batch{}
	for _, product := range products {
		args := pgx.NamedArgs{
			"productID": product.ID,
			"storageID": storage.ID,
		}
		batch.Queue(q, args)
	}
	results := r.client.SendBatch(ctx, batch)
	defer results.Close()
	for _, product := range products {
		_, err := results.Exec()
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			logger.Error().Msgf("SQL Error message (%s), Details: %s where -> %s", pgErr.Message, pgErr.Detail, pgErr.Where)
			logger.Warn().Msgf("product %s already exists", product.Name)
			continue
		}
	}

	return nil
}

func (r *repository) ExemptProducts(ctx context.Context, products []models.Product) error {
	logger := logging.GetLogger()
	logger.Trace().Msg("start ExemptProducts")
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	q := `DELETE FROM reservation USING products WHERE reservation.product_id = products.product_id AND products.product_id = @productID`
	batch := &pgx.Batch{}
	for _, product := range products {
		args := pgx.NamedArgs{
			"productID": product.ID,
		}
		batch.Queue(q, args)
	}
	results := r.client.SendBatch(ctx, batch)
	defer results.Close()
	for _, product := range products {
		_, err := results.Exec()
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			logger.Error().Msgf("SQL Error message (%s), Details: %s where -> %s", pgErr.Message, pgErr.Detail, pgErr.Where)
			logger.Warn().Msgf("product: %s", product.Name)
			continue
		}
	}

	return nil
}

func (r *repository) FindProductsViaStorageID(ctx context.Context, storageID uint) ([]models.Product, error) {
	logger := logging.GetLogger()
	logger.Trace().Msg("start FindProductsViaStorageID")
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	tx, err := r.client.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	firstQ := `SELECT product_id FROM reservation WHERE storage_id = $1`
	rows, err := tx.Query(ctx, firstQ, storageID)
	if err != nil {
		logger.Debug().Msg("error after tx.Query()")
		return nil, err
	}
	products := make([]models.Product, 0)
	for rows.Next() {
		var product models.Product
		err = rows.Scan(&product.ID)
		if err != nil {
			logger.Debug().Msg("error after rows.Scan()")
			return nil, err
		}

		products = append(products, product)
	}
	if err = rows.Err(); err != nil {
		logger.Debug().Msg("error after rows.Err()")
		return nil, err
	}
	rows.Close()

	secondQ := `SELECT product_code, product_name, product_size, product_count FROM products WHERE product_id = @productID`
	batch := &pgx.Batch{}
	for _, product := range products {
		args := pgx.NamedArgs{
			"productID": product.ID,
		}
		batch.Queue(secondQ, args)
	}
	results := tx.SendBatch(ctx, batch)
	for i := 0; i < len(products); i++ {
		product := &products[i]
		if err := results.QueryRow().Scan(&product.Code, &product.Name, &product.Size, &product.Count); err != nil {
			logger.Debug().Msg("error after results.QueryRow()")
			var (
				pgErr *pgconn.PgError
			)
			if errors.As(err, &pgErr) {
				logger.Error().Msgf("SQL Error message (%s), Details: %s where -> %s", pgErr.Message, pgErr.Detail, pgErr.Where)
			} else if errors.Is(err, pgx.ErrNoRows) {
				// было бы хорошо отсюда доставать те товары, которые не заведены в таблице
				logger.Warn().Err(err).Msgf("product with code: %s not exists", product.Code)
				continue
			}

			return nil, err
		}
	}
	if err := results.Close(); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Debug().Msg("error after tx.Commit()")
		return nil, err
	}

	return products, nil
}
