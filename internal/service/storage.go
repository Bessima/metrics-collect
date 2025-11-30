package service

import (
	"context"
	"fmt"
	"github.com/Bessima/metrics-collect/internal/config"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
)

type StorageService struct {
	config *config.Config

	repository repository.StorageRepositoryI
}

func NewStorageService(ctx context.Context, configuration *config.Config) *StorageService {
	service := &StorageService{config: configuration}
	service.setRepository(ctx)

	return service
}

func (service *StorageService) setRepository(ctx context.Context) {
	if service.config.DatabaseDNS != "" {
		service.repository = repository.NewDBRepository(ctx, service.config.DatabaseDNS)
		logger.Log.Info("Working with DB")
		return
	}

	if service.config.FileStoragePath != "" {
		service.repository = repository.NewFileStorageRepository(service.config.FileStoragePath)
		logger.Log.Info(fmt.Sprintf("Working with FILE %s", service.config.FileStoragePath))
		return
	}
	service.repository = repository.NewMemStorage()
	logger.Log.Info("Working with MemStorage")
}

func (service *StorageService) GetRepository() *repository.StorageRepositoryI {
	return &service.repository
}

func (service *StorageService) Close() {
	service.repository.Close()
}
