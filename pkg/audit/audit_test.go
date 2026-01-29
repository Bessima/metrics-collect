package audit

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileSubscriber(t *testing.T) {
	t.Run("create with valid filename", func(t *testing.T) {
		tempDir := t.TempDir()
		filename := filepath.Join(tempDir, "audit.log")

		subscriber := NewFileSubscriber(filename)

		assert.NotNil(t, subscriber)
		assert.Equal(t, filename, subscriber.filename)

		// Check file was created
		_, err := os.Stat(filename)
		assert.NoError(t, err)
	})

	t.Run("create with empty filename", func(t *testing.T) {
		subscriber := NewFileSubscriber("")

		assert.Nil(t, subscriber)
	})
}

func TestFileSubscriber_getName(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "audit.log")

	subscriber := NewFileSubscriber(filename)
	require.NotNil(t, subscriber)

	assert.Equal(t, "file", subscriber.getName())
}

func TestFileSubscriber_notify(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "audit.log")

	subscriber := NewFileSubscriber(filename)
	require.NotNil(t, subscriber)

	metrics := []string{"metric1", "metric2", "metric3"}
	ip := "192.168.1.1"
	ts := 1234567890

	err := subscriber.notify(metrics, ip, ts)
	require.NoError(t, err)

	// Verify file contains data
	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Greater(t, len(data), 0)

	// Verify JSON content contains expected values
	assert.Contains(t, string(data), "metric1")
	assert.Contains(t, string(data), "metric2")
	assert.Contains(t, string(data), "metric3")
	assert.Contains(t, string(data), "192.168.1.1")
	assert.Contains(t, string(data), "1234567890")
}

func TestFileSubscriber_notify_MultipleWrites(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "audit_multi.log")

	subscriber := NewFileSubscriber(filename)
	require.NotNil(t, subscriber)

	// First write
	err := subscriber.notify([]string{"metric1"}, "10.0.0.1", 1000)
	require.NoError(t, err)

	// Second write
	err = subscriber.notify([]string{"metric2"}, "10.0.0.2", 2000)
	require.NoError(t, err)

	// Verify file contains both writes
	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Contains(t, string(data), "metric1")
	assert.Contains(t, string(data), "metric2")
	assert.Contains(t, string(data), "10.0.0.1")
	assert.Contains(t, string(data), "10.0.0.2")
}

func TestNewURLSubscriber(t *testing.T) {
	t.Run("create with valid URL", func(t *testing.T) {
		url := "http://example.com/audit"

		subscriber := NewURLSubscriber(url)

		assert.NotNil(t, subscriber)
		assert.Equal(t, url, subscriber.url)
		assert.NotNil(t, subscriber.HTTPClient)
	})

	t.Run("create with empty URL", func(t *testing.T) {
		subscriber := NewURLSubscriber("")

		assert.Nil(t, subscriber)
	})
}

func TestURLSubscriber_getName(t *testing.T) {
	subscriber := NewURLSubscriber("http://example.com")
	require.NotNil(t, subscriber)

	assert.Equal(t, "url", subscriber.getName())
}

func TestURLSubscriber_notify_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		// Verify JSON content
		assert.Contains(t, string(body), "test_metric")
		assert.Contains(t, string(body), "127.0.0.1")

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	subscriber := NewURLSubscriber(server.URL)
	require.NotNil(t, subscriber)

	metrics := []string{"test_metric"}
	ip := "127.0.0.1"
	ts := 1234567890

	err := subscriber.notify(metrics, ip, ts)
	assert.NoError(t, err)
}

func TestURLSubscriber_notify_ServerError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	subscriber := NewURLSubscriber(server.URL)
	require.NotNil(t, subscriber)

	metrics := []string{"test_metric"}
	ip := "127.0.0.1"
	ts := 1234567890

	err := subscriber.notify(metrics, ip, ts)
	assert.Error(t, err)
}

func TestURLSubscriber_notify_NetworkError(t *testing.T) {
	// Use invalid URL to simulate network error
	subscriber := NewURLSubscriber("http://invalid-domain-that-does-not-exist-12345.com")
	require.NotNil(t, subscriber)

	metrics := []string{"test_metric"}
	ip := "127.0.0.1"
	ts := 1234567890

	err := subscriber.notify(metrics, ip, ts)
	assert.Error(t, err)
}

func TestEvent_Register(t *testing.T) {
	event := &Event{}

	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "audit.log")
	fileSubscriber := NewFileSubscriber(filename)

	event.Register(fileSubscriber)

	assert.NotNil(t, event.observers)
	assert.Equal(t, 1, len(event.observers))
	assert.NotNil(t, event.observers["file"])
}

func TestEvent_Register_Multiple(t *testing.T) {
	event := &Event{}

	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "audit.log")
	fileSubscriber := NewFileSubscriber(filename)

	urlSubscriber := NewURLSubscriber("http://example.com")

	event.Register(fileSubscriber)
	event.Register(urlSubscriber)

	assert.Equal(t, 2, len(event.observers))
	assert.NotNil(t, event.observers["file"])
	assert.NotNil(t, event.observers["url"])
}

func TestEvent_Register_Replace(t *testing.T) {
	event := &Event{}

	tempDir := t.TempDir()
	filename1 := filepath.Join(tempDir, "audit1.log")
	filename2 := filepath.Join(tempDir, "audit2.log")

	fileSubscriber1 := NewFileSubscriber(filename1)
	fileSubscriber2 := NewFileSubscriber(filename2)

	event.Register(fileSubscriber1)
	event.Register(fileSubscriber2)

	// Should have only one file observer (replaced)
	assert.Equal(t, 1, len(event.observers))
	assert.Equal(t, filename2, event.observers["file"].(*FileSubscriber).filename)
}

func TestEvent_Notify(t *testing.T) {
	event := &Event{}

	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "audit_notify.log")
	fileSubscriber := NewFileSubscriber(filename)

	event.Register(fileSubscriber)

	metrics := []string{"counter1", "gauge1"}
	ip := "192.168.1.100"

	event.Notify(metrics, ip)

	// Verify file was written
	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Greater(t, len(data), 0)
	assert.Contains(t, string(data), "counter1")
	assert.Contains(t, string(data), "gauge1")
	assert.Contains(t, string(data), "192.168.1.100")
}

func TestEvent_Notify_MultipleObservers(t *testing.T) {
	event := &Event{}

	// Setup file observer
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "audit_multi_observers.log")
	fileSubscriber := NewFileSubscriber(filename)

	// Setup URL observer with test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	urlSubscriber := NewURLSubscriber(server.URL)

	event.Register(fileSubscriber)
	event.Register(urlSubscriber)

	metrics := []string{"metric_test"}
	ip := "10.20.30.40"

	event.Notify(metrics, ip)

	// Verify file was written
	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Greater(t, len(data), 0)
	assert.Contains(t, string(data), "metric_test")
}

func TestEvent_Notify_NoObservers(t *testing.T) {
	event := &Event{}

	metrics := []string{"metric1"}
	ip := "192.168.1.1"

	// Should not panic with no observers
	event.Notify(metrics, ip)
}

func TestEvent_Notify_EmptyMetrics(t *testing.T) {
	event := &Event{}

	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "audit_empty.log")
	fileSubscriber := NewFileSubscriber(filename)

	event.Register(fileSubscriber)

	metrics := []string{}
	ip := "192.168.1.1"

	event.Notify(metrics, ip)

	// Verify file was written even with empty metrics
	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Greater(t, len(data), 0)
	assert.Contains(t, string(data), "[]")
}

func TestAuditEventDTO_Structure(t *testing.T) {
	event := AuditEventDTO{
		Ts:        1234567890,
		Metrics:   []string{"metric1", "metric2"},
		IPAddress: "192.168.1.1",
	}

	assert.Equal(t, 1234567890, event.Ts)
	assert.Equal(t, 2, len(event.Metrics))
	assert.Equal(t, "metric1", event.Metrics[0])
	assert.Equal(t, "metric2", event.Metrics[1])
	assert.Equal(t, "192.168.1.1", event.IPAddress)
}
