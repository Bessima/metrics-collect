package compress

import (
	"bytes"
	"crypto/hmac"
	"github.com/Bessima/metrics-collect/internal/common"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"io"
	"net/http"
)

type HashResponseWriter struct {
	http.ResponseWriter
	keyHash string
}

func (hw HashResponseWriter) Write(data []byte) (int, error) {
	if hw.keyHash != "" {
		hash := common.GetHashData(data, hw.keyHash)
		hw.ResponseWriter.Header().Set(common.HashHeader, hash)
	}
	return hw.ResponseWriter.Write(data)
}

func HashCheckerMiddleware(keyHash string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hash := r.Header.Get(common.HashHeader)
			if keyHash == "" || hash == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Читаем тело запроса
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}

			// Восстанавливаем тело запроса для дальнейшего использования
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			serverHash := common.GetHashData(body, keyHash)

			if hmac.Equal([]byte(hash), []byte(serverHash)) {
				http.Error(w, "Data was not transferred fully", http.StatusBadRequest)
				return
			}
			logger.Log.Info("Hashes was equal")
			hw := HashResponseWriter{ResponseWriter: w, keyHash: keyHash}

			next.ServeHTTP(hw, r)
		})
	}
}
