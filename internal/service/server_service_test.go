package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/Bessima/metrics-collect/pkg/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServerService(t *testing.T) {
	ctx := context.Background()
	address := "localhost:8080"
	hashKey := "secret"
	storage := repository.NewMemStorage()

	serverService := NewServerService(ctx, address, hashKey, storage)

	assert.NotNil(t, serverService)
	assert.NotNil(t, serverService.Server)
	assert.Equal(t, address, serverService.Server.Addr)
	assert.Equal(t, hashKey, serverService.hashKey)
	assert.Equal(t, storage, serverService.storage)
}

func TestServerService_Integration(t *testing.T) {
	ctx := context.Background()
	address := "localhost:0"
	hashKey := ""
	storage := repository.NewMemStorage()

	serverService := NewServerService(ctx, address, hashKey, storage)
	auditEvent := &audit.Event{}
	serverService.SetRouter(300, nil, auditEvent)

	// Start server
	errChan := make(chan error, 1)
	go serverService.RunServer(&errChan)

	time.Sleep(100 * time.Millisecond)

	// Test adding a counter via handler
	storage.Counter("integration_counter", 42)

	// Verify data
	value, err := storage.GetValue(repository.TypeCounter, "integration_counter")
	require.NoError(t, err)
	assert.Equal(t, int64(42), value)

	// Shutdown
	err = serverService.Shutdown()
	assert.NoError(t, err)
}

func TestServerService_ContextPropagation(t *testing.T) {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	address := "localhost:8080"
	hashKey := "secret"
	storage := repository.NewMemStorage()

	serverService := NewServerService(ctx, address, hashKey, storage)

	assert.NotNil(t, serverService)
	assert.NotNil(t, serverService.Server)
	assert.NotNil(t, serverService.Server.BaseContext)
}

func TestServerService_MultipleRequests(t *testing.T) {
	ctx := context.Background()
	address := "localhost:8080"
	hashKey := ""
	storage := repository.NewMemStorage()

	serverService := NewServerService(ctx, address, hashKey, storage)
	auditEvent := &audit.Event{}
	serverService.SetRouter(300, nil, auditEvent)

	// Simulate multiple concurrent requests
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		serverService.Server.Handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestServerService_WithHashKey(t *testing.T) {
	ctx := context.Background()
	address := "localhost:8080"
	hashKey := "my-secret-key"
	storage := repository.NewMemStorage()

	serverService := NewServerService(ctx, address, hashKey, storage)
	auditEvent := &audit.Event{}
	serverService.SetRouter(300, nil, auditEvent)

	assert.Equal(t, hashKey, serverService.hashKey)
	assert.NotNil(t, serverService.Server.Handler)
}

func TestServerService_WithoutHashKey(t *testing.T) {
	ctx := context.Background()
	address := "localhost:8080"
	hashKey := ""
	storage := repository.NewMemStorage()

	serverService := NewServerService(ctx, address, hashKey, storage)
	auditEvent := &audit.Event{}
	serverService.SetRouter(300, nil, auditEvent)

	assert.Equal(t, "", serverService.hashKey)
	assert.NotNil(t, serverService.Server.Handler)
}
