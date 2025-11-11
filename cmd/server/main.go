package main

import (
	"context"
	"github.com/Bessima/metrics-collect/internal/handler"
	"github.com/Bessima/metrics-collect/internal/middlewares/compress"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
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

func getMetricRouter(storage *repository.MemStorage, metricsFromFile *repository.MetricsFromFile) chi.Router {
	router := chi.NewRouter()
	router.Use(logger.RequestLogger)
	router.Use(compress.GZIPMiddleware)

	templates := handler.ParseAllTemplates()
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
	metricsFromFile := repository.MetricsFromFile{FileName: *config.FileStoragePath}

	if *config.Restore {
		loadMetricsFromFile(&metricsFromFile, config, &storage)
	}

	server := getServer(config, &storage, &metricsFromFile)

	go saveMetricsInFile(config, &storage, &metricsFromFile)
	go runServer(config, server)

	<-rootCtx.Done()
	stop()

	logger.Log.Info("Received shutdown signal, shutting down.")
	repository.UpdateMetricInFile(&storage, &metricsFromFile)

	return nil
}

func loadMetricsFromFile(metricsFromFile *repository.MetricsFromFile, config *Config, storage *repository.MemStorage) {
	if err := metricsFromFile.Load(); err != nil {
		logger.Log.Warn(err.Error())
	} else {
		logger.Log.Info("Metrics was loaded from file", zap.String("path", *config.FileStoragePath))
		storage.Load(metricsFromFile.GetMetrics())
	}
}

func saveMetricsInFile(config *Config, storage *repository.MemStorage, metricsFromFile *repository.MetricsFromFile) {
	if *config.StoreInterval <= 0 {
		return
	}
	for {
		repository.UpdateMetricInFile(storage, metricsFromFile)
		time.Sleep(time.Duration(*config.StoreInterval) * time.Second)
	}
}

func runServer(config *Config, server *http.Server) {
	logger.Log.Info("Running server on", zap.String("address", *config.Address))

	server.ListenAndServe()
}

func getServer(config *Config, storage *repository.MemStorage, metricsFromFile *repository.MetricsFromFile) *http.Server {
	var router chi.Router

	if *config.StoreInterval == 0 {
		router = getMetricRouter(storage, metricsFromFile)
	} else {
		router = getMetricRouter(storage, nil)
	}

	ongoingCtx := context.TODO()
	server := &http.Server{
		Addr:    *config.Address,
		Handler: router,
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
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
