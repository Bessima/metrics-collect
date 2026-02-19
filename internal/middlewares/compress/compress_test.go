package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAllowCompressContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{
			name:        "application/json",
			contentType: "application/json",
			expected:    true,
		},
		{
			name:        "application/json with charset",
			contentType: "application/json; charset=utf-8",
			expected:    true,
		},
		{
			name:        "text/html",
			contentType: "text/html",
			expected:    true,
		},
		{
			name:        "text/html with charset",
			contentType: "text/html; charset=utf-8",
			expected:    true,
		},
		{
			name:        "text/plain",
			contentType: "text/plain",
			expected:    false,
		},
		{
			name:        "image/png",
			contentType: "image/png",
			expected:    false,
		},
		{
			name:        "empty string",
			contentType: "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAllowCompressContentType(tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewCompressWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	cw := newCompressWriter(recorder)

	assert.NotNil(t, cw)
	assert.NotNil(t, cw.w)
	assert.NotNil(t, cw.zw)
}

func TestCompressWriter_Header(t *testing.T) {
	recorder := httptest.NewRecorder()
	cw := newCompressWriter(recorder)

	cw.Header().Set("X-Test", "value")

	assert.Equal(t, "value", recorder.Header().Get("X-Test"))
}

func TestCompressWriter_WriteHeader(t *testing.T) {
	tests := []struct {
		name               string
		statusCode         int
		contentType        string
		expectGzipEncoding bool
	}{
		{
			name:               "200 with application/json",
			statusCode:         http.StatusOK,
			contentType:        "application/json",
			expectGzipEncoding: true,
		},
		{
			name:               "200 with text/html",
			statusCode:         http.StatusOK,
			contentType:        "text/html",
			expectGzipEncoding: true,
		},
		{
			name:               "200 with text/plain",
			statusCode:         http.StatusOK,
			contentType:        "text/plain",
			expectGzipEncoding: false,
		},
		{
			name:               "404 with application/json",
			statusCode:         http.StatusNotFound,
			contentType:        "application/json",
			expectGzipEncoding: false,
		},
		{
			name:               "500 with application/json",
			statusCode:         http.StatusInternalServerError,
			contentType:        "application/json",
			expectGzipEncoding: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			cw := newCompressWriter(recorder)

			cw.Header().Set("Content-Type", tt.contentType)
			cw.WriteHeader(tt.statusCode)

			if tt.expectGzipEncoding {
				assert.Equal(t, "gzip", recorder.Header().Get("Content-Encoding"))
			} else {
				assert.Empty(t, recorder.Header().Get("Content-Encoding"))
			}

			assert.Equal(t, tt.statusCode, recorder.Code)
		})
	}
}

func TestCompressWriter_Write(t *testing.T) {
	t.Run("write with compressable content type", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		cw := newCompressWriter(recorder)

		cw.Header().Set("Content-Type", "application/json")

		data := []byte(`{"key":"value"}`)
		n, err := cw.Write(data)

		require.NoError(t, err)
		assert.Equal(t, len(data), n)

		cw.Close()

		// Verify data is compressed
		assert.Greater(t, len(recorder.Body.Bytes()), 0)
	})

	t.Run("write with non-compressable content type", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		cw := newCompressWriter(recorder)

		cw.Header().Set("Content-Type", "text/plain")

		data := []byte("plain text")
		n, err := cw.Write(data)

		require.NoError(t, err)
		assert.Equal(t, len(data), n)

		// Don't close compressWriter for non-compressable content
		// Verify data is not compressed
		assert.Contains(t, recorder.Body.String(), "plain text")
	})
}

func TestCompressWriter_Close(t *testing.T) {
	recorder := httptest.NewRecorder()
	cw := newCompressWriter(recorder)

	err := cw.Close()
	assert.NoError(t, err)
}

func TestNewCompressReader(t *testing.T) {
	t.Run("valid gzip data", func(t *testing.T) {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		gw.Write([]byte("test data"))
		gw.Close()

		reader := io.NopCloser(&buf)
		cr, err := newCompressReader(reader)

		require.NoError(t, err)
		assert.NotNil(t, cr)
		assert.NotNil(t, cr.r)
		assert.NotNil(t, cr.zr)
	})

	t.Run("invalid gzip data", func(t *testing.T) {
		reader := io.NopCloser(strings.NewReader("not gzip data"))
		cr, err := newCompressReader(reader)

		assert.Error(t, err)
		assert.Nil(t, cr)
	})
}

func TestCompressReader_Read(t *testing.T) {
	// Prepare compressed data
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	testData := "test data for reading"
	gw.Write([]byte(testData))
	gw.Close()

	reader := io.NopCloser(&buf)
	cr, err := newCompressReader(reader)
	require.NoError(t, err)

	// Read all data
	result, err := io.ReadAll(cr)

	require.NoError(t, err)
	assert.Equal(t, testData, string(result))
}

func TestCompressReader_Close(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte("test"))
	gw.Close()

	reader := io.NopCloser(&buf)
	cr, err := newCompressReader(reader)
	require.NoError(t, err)

	err = cr.Close()
	assert.NoError(t, err)
}

func TestGZIPMiddleware_NoCompression(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"hello"}`))
	})

	middleware := GZIPMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Accept-Encoding header
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `{"message":"hello"}`, rec.Body.String())
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
}

func TestGZIPMiddleware_WithCompression(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"hello"}`))
	})

	middleware := GZIPMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

	// Decompress and verify
	gr, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, `{"message":"hello"}`, string(decompressed))
}

func TestGZIPMiddleware_CompressedRequest(t *testing.T) {
	receivedData := ""

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		receivedData = string(body)
		w.WriteHeader(http.StatusOK)
	})

	middleware := GZIPMiddleware(handler)

	// Compress request body
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	testData := `{"key":"value"}`
	gw.Write([]byte(testData))
	gw.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, testData, receivedData)
}

func TestGZIPMiddleware_CompressedRequestAndResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"received":"` + string(body) + `"}`))
	})

	middleware := GZIPMiddleware(handler)

	// Compress request body
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	requestData := "test data"
	gw.Write([]byte(requestData))
	gw.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

	// Decompress response
	gr, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Contains(t, string(decompressed), requestData)
}

func TestGZIPMiddleware_InvalidCompressedRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := GZIPMiddleware(handler)

	// Send invalid gzip data
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not gzip"))
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGZIPMiddleware_NonCompressableContentType(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("binary data"))
	})

	middleware := GZIPMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Contains(t, rec.Body.String(), "binary data")
}

func TestGZIPMiddleware_ContentTypeWithCharset(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test":"data"}`))
	})

	middleware := GZIPMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
}

// TestGZIPMiddleware_HandlerWithoutExplicitWriteHeader воспроизводит проблему
// когда handler НЕ вызывает WriteHeader() явно (типичный случай!)
func TestGZIPMiddleware_HandlerWithoutExplicitWriteHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// НЕ вызываем w.WriteHeader() - это типичный случай!
		// Go автоматически вызовет WriteHeader(200) при первом Write()
		w.Write([]byte(`{"message":"hello world"}`))
	})

	middleware := GZIPMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	// КРИТИЧНО: Должен быть установлен Content-Encoding
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"),
		"Content-Encoding MUST be set even when handler doesn't call WriteHeader explicitly")

	// КРИТИЧНО: Content-Length должен быть удален
	assert.Empty(t, rec.Header().Get("Content-Length"),
		"Content-Length MUST be removed when using gzip")

	// Проверяем, что данные действительно сжаты
	gr, err := gzip.NewReader(rec.Body)
	require.NoError(t, err, "Response body should be valid gzip")
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, `{"message":"hello world"}`, string(decompressed))
}

// TestGZIPMiddleware_HTMLResponse проверяет отдачу HTML в браузер
func TestGZIPMiddleware_HTMLResponse(t *testing.T) {
	htmlContent := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body><h1>Hello, мир! 🌍</h1></body>
</html>`

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Типичный случай: НЕ вызываем WriteHeader
		w.Write([]byte(htmlContent))
	})

	middleware := GZIPMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
	assert.Empty(t, rec.Header().Get("Content-Length"))

	// Декодируем gzip
	gr, err := gzip.NewReader(rec.Body)
	require.NoError(t, err, "HTML should be properly gzipped")
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, htmlContent, string(decompressed))
	assert.Contains(t, string(decompressed), "мир", "UTF-8 должен корректно декодироваться")
}

// TestGZIPMiddleware_MultipleWrites проверяет несколько вызовов Write()
func TestGZIPMiddleware_MultipleWrites(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Несколько вызовов Write - типичный случай для streaming
		w.Write([]byte(`{"part1":`))
		w.Write([]byte(`"value1",`))
		w.Write([]byte(`"part2":"value2"}`))
	})

	middleware := GZIPMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

	gr, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, `{"part1":"value1","part2":"value2"}`, string(decompressed))
}
