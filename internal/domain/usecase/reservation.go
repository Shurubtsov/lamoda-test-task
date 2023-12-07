package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Shurubtsov/lamoda-test-task/internal/domain/models"
	"github.com/Shurubtsov/lamoda-test-task/pkg/logging"
)

type (
	Repo interface {
		ReserveProducts(ctx context.Context, storage models.Storage, products []models.Product) error
	}
	StorageService interface {
		GetAviableStorage(ctx context.Context) (*models.Storage, error)
	}
	ProductService interface {
		GetProductsInfo(ctx context.Context, products []models.Product) ([]models.Product, error)
	}
)
type reservation struct {
	storageService StorageService
	productService ProductService
	repository     Repo
}

func NewReservation(ss StorageService, ps ProductService, r Repo) *reservation {
	return &reservation{
		storageService: ss,
		productService: ps,
		repository:     r,
	}
}

func (r *reservation) ProductReservation(ctx context.Context, products []models.Product) ([]models.Product, error) {
	logger := logging.GetLogger()
	logger.Trace().Msg("start ProductReservation")
	ctx, cancel := context.WithTimeout(ctx, time.Second*6)
	defer cancel()

	storage, err := r.storageService.GetAviableStorage(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAviableStorage failed: %w", err)
	}

	filledProducts, err := r.productService.GetProductsInfo(ctx, products)
	if err != nil {
		return nil, fmt.Errorf("GetProductsInfo failed: %w", err)
	}

	if err := r.repository.ReserveProducts(ctx, *storage, filledProducts); err != nil {
		return nil, fmt.Errorf("ReserveProducts failed: %w", err)
	}

	return filledProducts, nil
}
