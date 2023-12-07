package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Shurubtsov/lamoda-test-task/internal/domain/models"
	"github.com/Shurubtsov/lamoda-test-task/pkg/logging"
)

var (
	ErrNilProducts   = errors.New("array of products is null")
	ErrEmptyProducts = errors.New("lenght array of products is zero")
)

type ProductRepo interface {
	FindProductsViaCode(ctx context.Context, products []models.Product) ([]models.Product, error)
	ExemptProducts(ctx context.Context, products []models.Product) error
}
type productService struct {
	repository ProductRepo
}

func NewProductService(pr ProductRepo) *productService {
	return &productService{repository: pr}
}

func (ps *productService) GetProductsInfo(ctx context.Context, products []models.Product) ([]models.Product, error) {
	logger := logging.GetLogger()
	logger.Trace().Msg("start GetProductsInfo")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	products, err := ps.repository.FindProductsViaCode(ctx, products)
	if err != nil {
		return nil, fmt.Errorf("FindProductsViaCode failed: %w", err)
	}

	if products == nil {
		return nil, ErrNilProducts
	}

	if len(products) == 0 {
		return nil, ErrEmptyProducts
	}

	return products, nil
}

func (ps *productService) ProductExemption(ctx context.Context, products []models.Product) ([]models.Product, error) {
	logger := logging.GetLogger()
	logger.Trace().Msg("start ProductReservation")
	ctx, cancel := context.WithTimeout(ctx, time.Second*6)
	defer cancel()

	filledProducts, err := ps.GetProductsInfo(ctx, products)
	if err != nil {
		return nil, fmt.Errorf("GetProductsInfo failed: %w", err)
	}

	if err := ps.repository.ExemptProducts(ctx, filledProducts); err != nil {
		return nil, fmt.Errorf("ExemptProducts failed: %w", err)
	}

	return filledProducts, nil
}
