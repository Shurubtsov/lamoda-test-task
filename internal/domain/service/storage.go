package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Shurubtsov/lamoda-test-task/internal/domain/models"
	"github.com/Shurubtsov/lamoda-test-task/pkg/logging"
)

var (
	ErrNilStorageObj = errors.New("object of storage is null")
)

type StorageRepo interface {
	FindAviableStorage(ctx context.Context) (*models.Storage, error)
}

type storageService struct {
	repository StorageRepo
}

func NewStorageService(sr StorageRepo) *storageService {
	return &storageService{repository: sr}
}

func (s *storageService) GetAviableStorage(ctx context.Context) (*models.Storage, error) {
	logger := logging.GetLogger()
	logger.Trace().Msg("start GetAviableStorage")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	storage, err := s.repository.FindAviableStorage(ctx)
	if err != nil {
		return nil, fmt.Errorf("FindAviableStorage failed: %w", err)
	}

	if storage == nil {
		return nil, ErrNilStorageObj
	}

	if err := storage.Validate(); err != nil {
		return nil, err
	}

	return storage, nil
}
