package hash

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bessima/metrics-collect/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashResponseWriter_Write(t *testing.T) {
	t.Run("with hash key", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		keyHash := "secret-key"
		hw := HashResponseWriter{
			ResponseWriter: recorder,
			keyHash:        keyHash,
		}

		data := []byte(`{"test":"data"}`)
		n, err := hw.Write(data)

		require.NoError(t, err)
		assert.Equal(t, len(data), n)

		// Verify hash header is set
		hashHeader := recorder.Header().Get(common.HashHeader)
		assert.NotEmpty(t, hashHeader)

		// Verify hash is correct
		expectedHash := common.GetHashData(data, keyHash)
		assert.Equal(t, expectedHash, hashHeader)

		// Verify data is written
		assert.Equal(t, string(data), recorder.Body.String())
	})

	t.Run("without hash key", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		hw := HashResponseWriter{
			ResponseWriter: recorder,
			keyHash:        "",
		}

		data := []byte(`{"test":"data"}`)
		n, err := hw.Write(data)

		require.NoError(t, err)
		assert.Equal(t, len(data), n)

		// Verify hash header is NOT set
		hashHeader := recorder.Header().Get(common.HashHeader)
		assert.Empty(t, hashHeader)

		// Verify data is written
		assert.Equal(t, string(data), recorder.Body.String())
	})

	t.Run("multiple writes with hash", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		keyHash := "test-key"
		hw := HashResponseWriter{
			ResponseWriter: recorder,
			keyHash:        keyHash,
		}

		data1 := []byte("part1")
		data2 := []byte("part2")

		hw.Write(data1)
		hw.Write(data2)

		// Each write sets its own hash
		hashHeader := recorder.Header().Get(common.HashHeader)
		assert.NotEmpty(t, hashHeader)
	})
}

func TestHashCheckerMiddleware_NoHashKey(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := HashCheckerMiddleware("")
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"data":"test"}`))
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

func TestHashCheckerMiddleware_NoHashHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := HashCheckerMiddleware("secret-key")
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"data":"test"}`))
	// No hash header
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

func TestHashCheckerMiddleware_ValidHash(t *testing.T) {
	receivedBody := ""

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	keyHash := "secret-key"
	middleware := HashCheckerMiddleware(keyHash)
	wrappedHandler := middleware(handler)

	requestBody := []byte(`{"data":"test"}`)
	hash := common.GetHashData(requestBody, keyHash)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(requestBody))
	req.Header.Set(common.HashHeader, hash)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
	assert.Equal(t, string(requestBody), receivedBody)
}

func TestHashCheckerMiddleware_InvalidHash(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	keyHash := "secret-key"
	middleware := HashCheckerMiddleware(keyHash)
	wrappedHandler := middleware(handler)

	requestBody := []byte(`{"data":"test"}`)
	invalidHash := "invalid-hash-value"

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(requestBody))
	req.Header.Set(common.HashHeader, invalidHash)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Data was not transferred fully")
}

func TestHashCheckerMiddleware_WrongKeyHash(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	keyHash := "secret-key"
	middleware := HashCheckerMiddleware(keyHash)
	wrappedHandler := middleware(handler)

	requestBody := []byte(`{"data":"test"}`)
	// Generate hash with different key
	wrongHash := common.GetHashData(requestBody, "different-key")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(requestBody))
	req.Header.Set(common.HashHeader, wrongHash)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Data was not transferred fully")
}

func TestHashCheckerMiddleware_ResponseWithHash(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"response":"data"}`))
	})

	keyHash := "secret-key"
	middleware := HashCheckerMiddleware(keyHash)
	wrappedHandler := middleware(handler)

	requestBody := []byte(`{"data":"test"}`)
	hash := common.GetHashData(requestBody, keyHash)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(requestBody))
	req.Header.Set(common.HashHeader, hash)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify response has hash header
	responseHash := rec.Header().Get(common.HashHeader)
	assert.NotEmpty(t, responseHash)

	// Verify response hash is correct
	responseBody := rec.Body.Bytes()
	expectedHash := common.GetHashData(responseBody, keyHash)
	assert.Equal(t, expectedHash, responseHash)
}

func TestHashCheckerMiddleware_EmptyBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	keyHash := "secret-key"
	middleware := HashCheckerMiddleware(keyHash)
	wrappedHandler := middleware(handler)

	requestBody := []byte("")
	hash := common.GetHashData(requestBody, keyHash)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(requestBody))
	req.Header.Set(common.HashHeader, hash)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

func TestHashCheckerMiddleware_LargeBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	keyHash := "secret-key"
	middleware := HashCheckerMiddleware(keyHash)
	wrappedHandler := middleware(handler)

	// Create large body
	largeBody := bytes.Repeat([]byte("x"), 10000)
	hash := common.GetHashData(largeBody, keyHash)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(largeBody))
	req.Header.Set(common.HashHeader, hash)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

func TestHashCheckerMiddleware_BodyCanBeReadMultipleTimes(t *testing.T) {
	callCount := 0
	var bodies []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
		w.WriteHeader(http.StatusOK)
	})

	keyHash := "secret-key"
	middleware := HashCheckerMiddleware(keyHash)
	wrappedHandler := middleware(handler)

	requestBody := []byte(`{"data":"test"}`)
	hash := common.GetHashData(requestBody, keyHash)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(requestBody))
	req.Header.Set(common.HashHeader, hash)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, callCount)
	// Body should be readable in handler after being read in middleware
	assert.Equal(t, string(requestBody), bodies[0])
}

func TestHashCheckerMiddleware_DifferentHashAlgorithms(t *testing.T) {
	// Test that hash is consistent
	keyHash := "test-key-12345"
	data := []byte(`{"key":"value"}`)

	hash1 := common.GetHashData(data, keyHash)
	hash2 := common.GetHashData(data, keyHash)

	assert.Equal(t, hash1, hash2, "Same data and key should produce same hash")
}

func TestHashCheckerMiddleware_Integration(t *testing.T) {
	// Test full request-response cycle with hash
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Received: " + string(body)))
	})

	keyHash := "integration-test-key"
	middleware := HashCheckerMiddleware(keyHash)
	wrappedHandler := middleware(handler)

	requestBody := []byte(`{"test":"integration"}`)
	requestHash := common.GetHashData(requestBody, keyHash)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(requestBody))
	req.Header.Set(common.HashHeader, requestHash)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Verify request was successful
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "integration")

	// Verify response hash is present
	responseHash := rec.Header().Get(common.HashHeader)
	assert.NotEmpty(t, responseHash)

	// Verify response hash is correct
	expectedResponseHash := common.GetHashData(rec.Body.Bytes(), keyHash)
	assert.Equal(t, expectedResponseHash, responseHash)
}
