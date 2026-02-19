package service

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/Bessima/metrics-collect/internal/handler"
	"github.com/Bessima/metrics-collect/internal/middlewares/compress"
	hashMiddleware "github.com/Bessima/metrics-collect/internal/middlewares/hash"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/Bessima/metrics-collect/pkg/audit"
	"github.com/go-chi/chi/v5"
)

type ServerService struct {
	Server  *http.Server
	storage repository.StorageRepositorier
	hashKey string
}

func NewServerService(rootContext context.Context, address string, hashKey string, storage repository.StorageRepositorier) ServerService {
	server := &http.Server{
		Addr: address,
		BaseContext: func(_ net.Listener) context.Context {
			return rootContext
		},
	}
	return ServerService{Server: server, storage: storage, hashKey: hashKey}
}

func (serverService *ServerService) SetRouter(storeInterval int64, metricsFromFile *repository.MetricsFromFile, auditEvent *audit.Event) {
	var router chi.Router

	if storeInterval == 0 {
		router = serverService.getRouter(metricsFromFile, auditEvent)
	} else {
		router = serverService.getRouter(nil, auditEvent)
	}

	serverService.Server.Handler = router
}

func (serverService *ServerService) getRouter(metricsFromFile *repository.MetricsFromFile, auditEvent *audit.Event) chi.Router {
	router := chi.NewRouter()

	router.Use(logger.RequestLogger)
	router.Use(compress.GZIPMiddleware)
	router.Use(hashMiddleware.HashCheckerMiddleware(serverService.hashKey))

	templates := handler.ParseAllTemplates()
	router.Get("/", handler.MainHandler(serverService.storage, templates))

	router.Post("/update/{typeMetric}/{name}/{value}", handler.SetMetricHandler(serverService.storage, metricsFromFile))
	router.Get("/value/{typeMetric}/{name}", handler.ViewMetricValue(serverService.storage))

	router.Post("/update/", handler.UpdateHandler(serverService.storage, metricsFromFile))
	router.Post("/value/", handler.ValueHandler(serverService.storage))

	router.Post("/updates/", handler.UpdatesHandler(serverService.storage, metricsFromFile, auditEvent))

	router.Get("/ping", handler.PingHandler(serverService.storage))

	return router
}

func (serverService *ServerService) RunServer(serverErr *chan error) {
	if err := serverService.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		*serverErr <- err
	} else {
		*serverErr <- nil
	}
}

func (serverService *ServerService) Shutdown() error {
	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if shutdownErr := serverService.Server.Shutdown(shutdownCtx); shutdownErr != nil {
		return shutdownErr
	}

	return nil
}
