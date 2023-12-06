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
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				logger.Error().Msgf("SQL Error message (%s), Details: %s where -> %s", pgErr.Message, pgErr.Detail, pgErr.Where)
			}
			return nil, err
		}
	}

	return products, nil
}

func (r *repository) ReserveProducts(ctx context.Context, storage models.Storage, products []models.Product) error {
	logger := logging.GetLogger()
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	q := `INSERT INTO storage_product (storage_id, product_id) VALUES (@storageID, @productID)`
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
			logger.Warn().Msgf("user %s already exists", product.Name)
			continue
		}
	}

	return nil
}
