package main

import (
	"context"
	"github.com/Bessima/metrics-collect/internal/config/db"
	"github.com/Bessima/metrics-collect/internal/handler"
	"github.com/Bessima/metrics-collect/internal/middlewares/compress"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"html/template"
	"net"
	"net/http"
	"time"
)

type App struct {
	config          *Config
	storage         *repository.MemStorage
	metricsFromFile repository.MetricsFromFile
	rootContext     context.Context
}

func NewApp(ctx context.Context, storage *repository.MemStorage) *App {
	app := &App{rootContext: ctx}

	app.config = InitConfig()
	if storage != nil {
		app.storage = storage
	} else {
		newStorage := repository.NewMemStorage()
		app.storage = &newStorage
	}
	app.metricsFromFile = repository.MetricsFromFile{FileName: app.config.FileStoragePath}

	return app
}

func (app *App) loadMetricsFromFile() {
	if err := app.metricsFromFile.Load(); err != nil {
		logger.Log.Warn(err.Error())
	} else {
		logger.Log.Info("Metrics was loaded from file", zap.String("path", app.config.FileStoragePath))
		app.storage.Load(app.metricsFromFile.GetMetrics())
	}
}

func (app *App) initDB() *db.DB {
	dbObj, errDB := db.NewDB(app.rootContext, app.config.DatabaseDNS)

	if errDB != nil {

		logger.Log.Panic(
			"Unable to connect to database",
			zap.String("path", app.config.DatabaseDNS),
			zap.String("error", errDB.Error()),
		)
	}

	return dbObj
}

func (app *App) saveMetricsInFile(ctx context.Context) {
	if app.config.StoreInterval <= 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(app.config.StoreInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			repository.UpdateMetricInFile(app.storage, &app.metricsFromFile)
		case <-ctx.Done():
			logger.Log.Info("Stopping metrics saver")
			return
		}
	}
}

func (app *App) getServer(pool *pgxpool.Pool) *http.Server {
	var router chi.Router
	templates := handler.ParseAllTemplates()
	if app.config.StoreInterval == 0 {
		router = app.getMetricRouter(templates, &app.metricsFromFile)
	} else {
		router = app.getMetricRouter(templates, nil)
	}

	router.Get("/ping", handler.PingHandler(pool))

	server := &http.Server{
		Addr:    app.config.Address,
		Handler: router,
		BaseContext: func(_ net.Listener) context.Context {
			return app.rootContext
		},
	}
	return server
}

func (app *App) getMetricRouter(templates *template.Template, metricsFromFile *repository.MetricsFromFile) chi.Router {
	router := chi.NewRouter()
	router.Use(logger.RequestLogger)
	router.Use(compress.GZIPMiddleware)

	router.Get("/", handler.MainHandler(app.storage, templates))

	router.Post("/update/{typeMetric}/{name}/{value}", handler.SetMetricHandler(app.storage, metricsFromFile))
	router.Get("/value/{typeMetric}/{name}", handler.ViewMetricValue(app.storage))
	router.Post("/update/", handler.UpdateHandler(app.storage, metricsFromFile))
	router.Post("/value/", handler.ValueHandler(app.storage))

	return router
}
