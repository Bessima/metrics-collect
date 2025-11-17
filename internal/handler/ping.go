package handler

import (
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
)

func PingHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		if pool == nil {
			http.Error(w, "Database pool not initialized", http.StatusInternalServerError)
			return
		}

		if err := pool.Ping(request.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		logger.Log.Info("Successfully pinged the database.")
		w.Write([]byte("OK"))
	}
}
