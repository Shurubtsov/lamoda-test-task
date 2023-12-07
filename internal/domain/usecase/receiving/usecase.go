package receiving

import (
	"context"
	"fmt"
	"time"

	"github.com/Shurubtsov/lamoda-test-task/internal/domain/models"
	"github.com/Shurubtsov/lamoda-test-task/pkg/logging"
)

type (
	Repo interface {
		ExemptProducts(ctx context.Context, products []models.Product) error
	}
	StorageService interface {
		GetAviableStorage(ctx context.Context) (*models.Storage, error)
	}
	ProductService interface {
		GetProductsInfo(ctx context.Context, products []models.Product) ([]models.Product, error)
	}
)
type usecase struct {
	storageService StorageService
	productService ProductService
	repository     Repo
}

func New(ss StorageService, ps ProductService, r Repo) *usecase {
	return &usecase{
		storageService: ss,
		productService: ps,
		repository:     r,
	}
}

func (u *usecase) CountProductReceivingFromStorage(ctx context.Context, products []models.Product) ([]models.Product, error) {
	logger := logging.GetLogger()
	logger.Trace().Msg("start ProductReservation")
	ctx, cancel := context.WithTimeout(ctx, time.Second*6)
	defer cancel()

	filledProducts, err := u.productService.GetProductsInfo(ctx, products)
	if err != nil {
		return nil, fmt.Errorf("GetProductsInfo failed: %w", err)
	}

	if err := u.repository.ExemptProducts(ctx, filledProducts); err != nil {
		return nil, fmt.Errorf("ReserveProducts failed: %w", err)
	}

	return filledProducts, nil
}
