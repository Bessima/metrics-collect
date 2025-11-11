package main

import (
	"context"
	"github.com/Bessima/metrics-collect/internal/handler"
	"github.com/Bessima/metrics-collect/internal/middlewares/compress"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"html/template"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	err := initLogger()
	if err != nil {
		logger.Log.Warn(err.Error())
	}

	if err := run(); err != nil {
		panic(err)
	}
}

func getMetricRouter(storage *repository.MemStorage, templates *template.Template, metricsFromFile *repository.MetricsFromFile) chi.Router {
	router := chi.NewRouter()
	router.Use(logger.RequestLogger)
	router.Use(compress.GZIPMiddleware)

	router.Get("/", handler.MainHandler(storage, templates))

	router.Post("/update/{typeMetric}/{name}/{value}", handler.SetMetricHandler(storage, metricsFromFile))
	router.Get("/value/{typeMetric}/{name}", handler.ViewMetricValue(storage))
	router.Post("/update/", handler.UpdateHandler(storage, metricsFromFile))
	router.Post("/value/", handler.ValueHandler(storage))

	return router
}

func run() error {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config := InitConfig()
	storage := repository.NewMemStorage()
	metricsFromFile := repository.MetricsFromFile{FileName: config.FileStoragePath}

	if config.Restore {
		loadMetricsFromFile(&metricsFromFile, config, &storage)
	}

	server := getServer(rootCtx, config, &storage, &metricsFromFile)

	saveCtx, saveCancel := context.WithCancel(rootCtx)
	defer saveCancel()
	go saveMetricsInFile(saveCtx, config, &storage, &metricsFromFile)

	serverErr := make(chan error, 1)
	go func() {
		logger.Log.Info("Running server on", zap.String("address", config.Address))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		} else {
			serverErr <- nil
		}
	}()

	// Ждем сигнал завершения или ошибку сервера
	var err error
	select {
	case <-rootCtx.Done():
		logger.Log.Info("Received shutdown signal, shutting down.")
	case err = <-serverErr:
		logger.Log.Error("Server error", zap.Error(err))
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
		logger.Log.Error("Server shutdown error", zap.Error(shutdownErr))
	}

	logger.Log.Info("Received shutdown signal, shutting down.")
	repository.UpdateMetricInFile(&storage, &metricsFromFile)

	saveCancel()

	return err
}

func loadMetricsFromFile(metricsFromFile *repository.MetricsFromFile, config *Config, storage *repository.MemStorage) {
	if err := metricsFromFile.Load(); err != nil {
		logger.Log.Warn(err.Error())
	} else {
		logger.Log.Info("Metrics was loaded from file", zap.String("path", config.FileStoragePath))
		storage.Load(metricsFromFile.GetMetrics())
	}
}

func saveMetricsInFile(ctx context.Context, config *Config, storage *repository.MemStorage, metricsFromFile *repository.MetricsFromFile) {
	if config.StoreInterval <= 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(config.StoreInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			repository.UpdateMetricInFile(storage, metricsFromFile)
		case <-ctx.Done():
			logger.Log.Info("Stopping metrics saver")
			return
		}
	}
}

func getServer(rootCtx context.Context, config *Config, storage *repository.MemStorage, metricsFromFile *repository.MetricsFromFile) *http.Server {
	var router chi.Router
	templates := handler.ParseAllTemplates()
	if config.StoreInterval == 0 {
		router = getMetricRouter(storage, templates, metricsFromFile)
	} else {
		router = getMetricRouter(storage, templates, nil)
	}

	server := &http.Server{
		Addr:    config.Address,
		Handler: router,
		BaseContext: func(_ net.Listener) context.Context {
			return rootCtx
		},
	}
	return server
}

func initLogger() error {
	if err := logger.Initialize("debug"); err != nil {
		return err
	}
	return nil
}
