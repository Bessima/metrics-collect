package service

import (
	"context"
	"github.com/Bessima/metrics-collect/internal/handler"
	"github.com/Bessima/metrics-collect/internal/middlewares/compress"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net"
	"net/http"
	"time"
)

type ServerService struct {
	Server  *http.Server
	storage *repository.MemStorage
}

func NewServerService(rootContext context.Context, address string, storage *repository.MemStorage) ServerService {
	server := &http.Server{
		Addr: address,
		BaseContext: func(_ net.Listener) context.Context {
			return rootContext
		},
	}
	return ServerService{Server: server, storage: storage}
}

func (serverService *ServerService) SetRouter(storeInterval int64, pool *pgxpool.Pool, metricsFromFile *repository.MetricsFromFile) {
	var router chi.Router

	if storeInterval == 0 {
		router = serverService.getRouter(metricsFromFile)
	} else {
		router = serverService.getRouter(nil)
	}

	router.Get("/ping", handler.PingHandler(pool))

	serverService.Server.Handler = router
}

func (serverService *ServerService) getRouter(metricsFromFile *repository.MetricsFromFile) chi.Router {
	router := chi.NewRouter()

	router.Use(logger.RequestLogger)
	router.Use(compress.GZIPMiddleware)

	templates := handler.ParseAllTemplates()
	router.Get("/", handler.MainHandler(serverService.storage, templates))

	router.Post("/update/{typeMetric}/{name}/{value}", handler.SetMetricHandler(serverService.storage, metricsFromFile))
	router.Get("/value/{typeMetric}/{name}", handler.ViewMetricValue(serverService.storage))
	router.Post("/update/", handler.UpdateHandler(serverService.storage, metricsFromFile))
	router.Post("/value/", handler.ValueHandler(serverService.storage))

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
