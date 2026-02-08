package handler

import (
	"net/http"

	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
)

func PingHandler(storage repository.StorageRepositorier) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		if err := storage.Ping(request.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		logger.Log.Info("Successfully pinged the database.")
		w.Write([]byte("OK"))
	}
}
