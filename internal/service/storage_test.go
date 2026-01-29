package service

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Bessima/metrics-collect/internal/config"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStorageService_MemStorage(t *testing.T) {
	cfg := &config.Config{
		Address:         "localhost:8080",
		StoreInterval:   300,
		FileStoragePath: "",
		Restore:         false,
		DatabaseDNS:     "",
	}

	service := NewStorageService(context.Background(), cfg)

	assert.NotNil(t, service)
	assert.NotNil(t, service.repository)
	assert.NotNil(t, service.config)
	assert.Equal(t, cfg, service.config)

	// Verify it's MemStorage by checking Ping error
	err := service.repository.Ping(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "memory storage")
}

func TestNewStorageService_FileStorage(t *testing.T) {
	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "storage.json")

	cfg := &config.Config{
		Address:         "localhost:8080",
		StoreInterval:   300,
		FileStoragePath: storagePath,
		Restore:         true,
		DatabaseDNS:     "",
	}

	service := NewStorageService(context.Background(), cfg)

	assert.NotNil(t, service)
	assert.NotNil(t, service.repository)

	// Verify it's FileStorage by checking Ping error
	err := service.repository.Ping(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file storage")
}

func TestStorageService_GetRepository(t *testing.T) {
	cfg := &config.Config{
		Address:         "localhost:8080",
		StoreInterval:   300,
		FileStoragePath: "",
		Restore:         false,
		DatabaseDNS:     "",
	}

	service := NewStorageService(context.Background(), cfg)

	repo := service.GetRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, *repo)
}

func TestStorageService_Close(t *testing.T) {
	cfg := &config.Config{
		Address:         "localhost:8080",
		StoreInterval:   300,
		FileStoragePath: "",
		Restore:         false,
		DatabaseDNS:     "",
	}

	service := NewStorageService(context.Background(), cfg)

	// Should not panic
	service.Close()
}

func TestStorageService_setRepository_Priority(t *testing.T) {
	t.Run("DatabaseDNS has highest priority", func(t *testing.T) {
		tempDir := t.TempDir()
		storagePath := filepath.Join(tempDir, "storage.json")

		// Invalid DB DNS to avoid actual connection, but test priority
		cfg := &config.Config{
			DatabaseDNS:     "postgres://invalid",
			FileStoragePath: storagePath,
		}

		service := &StorageService{config: cfg}
		service.setRepository(context.Background())

		// Even with FileStoragePath set, DB should be chosen
		// We can't test the exact type without reflection, but we can verify it was set
		assert.NotNil(t, service.repository)
	})

	t.Run("FileStoragePath is second priority", func(t *testing.T) {
		tempDir := t.TempDir()
		storagePath := filepath.Join(tempDir, "storage.json")

		cfg := &config.Config{
			DatabaseDNS:     "",
			FileStoragePath: storagePath,
		}

		service := &StorageService{config: cfg}
		service.setRepository(context.Background())

		assert.NotNil(t, service.repository)

		// Verify it's FileStorage
		err := service.repository.Ping(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file storage")
	})

	t.Run("MemStorage is default", func(t *testing.T) {
		cfg := &config.Config{
			DatabaseDNS:     "",
			FileStoragePath: "",
		}

		service := &StorageService{config: cfg}
		service.setRepository(context.Background())

		assert.NotNil(t, service.repository)

		// Verify it's MemStorage
		err := service.repository.Ping(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "memory storage")
	})
}

func TestStorageService_Integration_WithMemStorage(t *testing.T) {
	cfg := &config.Config{
		FileStoragePath: "",
		DatabaseDNS:     "",
	}

	service := NewStorageService(context.Background(), cfg)
	repo := *service.GetRepository()

	// Test Counter
	err := repo.Counter("test_counter", 10)
	require.NoError(t, err)

	value, err := repo.GetValue(repository.TypeCounter, "test_counter")
	require.NoError(t, err)
	assert.Equal(t, int64(10), value)

	// Test Gauge
	err = repo.ReplaceGaugeMetric("test_gauge", 3.14)
	require.NoError(t, err)

	value, err = repo.GetValue(repository.TypeGauge, "test_gauge")
	require.NoError(t, err)
	assert.Equal(t, 3.14, value)

	// Test All
	metrics, err := repo.All()
	require.NoError(t, err)
	assert.Equal(t, 2, len(metrics))

	// Test Close
	service.Close()
}

func TestStorageService_Integration_WithFileStorage(t *testing.T) {
	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "integration_storage.json")

	cfg := &config.Config{
		FileStoragePath: storagePath,
		DatabaseDNS:     "",
	}

	service := NewStorageService(context.Background(), cfg)
	repo := *service.GetRepository()

	// Test Counter
	err := repo.Counter("file_counter", 20)
	require.NoError(t, err)

	// Test Gauge
	err = repo.ReplaceGaugeMetric("file_gauge", 2.71)
	require.NoError(t, err)

	// Test All
	metrics, err := repo.All()
	require.NoError(t, err)
	assert.Equal(t, 2, len(metrics))

	// Test Close
	service.Close()
}
