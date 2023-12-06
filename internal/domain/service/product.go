package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Shurubtsov/lamoda-test-task/internal/domain/models"
	"github.com/Shurubtsov/lamoda-test-task/pkg/logging"
)

var (
	ErrNilProducts   = errors.New("array of products is null")
	ErrEmptyProducts = errors.New("lenght array of products is zero")
)

type ProductRepo interface {
	FindProductsViaCode(ctx context.Context, products []models.Product) ([]models.Product, error)
}
type productService struct {
	repository ProductRepo
}

func NewProductService(pr ProductRepo) *productService {
	return &productService{repository: pr}
}

func (p *productService) GetProductsInfo(ctx context.Context, products []models.Product) ([]models.Product, error) {
	logger := logging.GetLogger()
	logger.Trace().Msg("start GetProductsInfo")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	products, err := p.repository.FindProductsViaCode(ctx, products)
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
