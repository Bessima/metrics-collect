package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/retry"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Observer interface {
	notify(metrics []string, ip string, ts int) error
	getName() string
}

type FileSubscriber struct {
	filename string
}

func NewFileSubscriber(filename string) *FileSubscriber {
	if filename == "" {
		return nil
	}
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Log.Error(
			"Unable to open file",
			zap.String("filename", filename),
			zap.String("error", err.Error()),
		)
	}
	defer file.Close()
	return &FileSubscriber{filename: filename}
}

func (observer *FileSubscriber) notify(metrics []string, ip string, ts int) error {
	event := AuditEventDTO{TS: ts, Metrics: metrics, IPAddress: ip}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(observer.filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}

	logger.Log.Info(
		"audit data was successfully written",
		zap.String("filename", observer.filename),
		zap.ByteString("data", data),
	)

	return nil
}

func (observer *FileSubscriber) getName() string {
	return "file"
}

type URLSubscriber struct {
	url        string
	HTTPClient *http.Client
}

func NewURLSubscriber(url string) *URLSubscriber {
	if url == "" {
		return nil
	}
	return &URLSubscriber{url: url, HTTPClient: &http.Client{}}
}

func (observer *URLSubscriber) getName() string {
	return "url"
}

func (observer *URLSubscriber) notify(metrics []string, ip string, ts int) error {
	event := AuditEventDTO{TS: ts, Metrics: metrics, IPAddress: ip}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return retry.DoRetry(context.Background(), func() error {
		req, err := http.NewRequest(http.MethodPost, observer.url, bytes.NewBuffer(data))
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
		}
		req.Header.Add("Content-Type", "application/json")

		var response *http.Response

		response, err = observer.HTTPClient.Do(req)
		if err != nil {
			logger.Log.Error(
				"Failed sending audit to url",
				zap.String("url", observer.url),
				zap.String("err", err.Error()),
			)
			return err
		}

		defer func() {
			if err := response.Body.Close(); err != nil {
				logger.Log.Error("Error closing response body", zap.String("err", err.Error()))
			}
		}()
		if response.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(response.Body)
			logger.Log.Error(
				"Audit server returned non-OK status",
				zap.Int("status", response.StatusCode),
				zap.String("body", string(body)),
			)
			return fmt.Errorf("server returned status: %d", response.StatusCode)
		}

		logger.Log.Info("Successful sending data")
		return nil
	}, retry.AgentRetryConfig)

}

type Event struct {
	observers map[string]Observer
}

func (e *Event) Register(o Observer) {
	if e.observers == nil {
		e.observers = make(map[string]Observer)
	}
	e.observers[o.getName()] = o
}

func (e *Event) Notify(metrics []string, ip string) {
	ts := int(time.Now().Unix())
	for _, observer := range e.observers {
		err := observer.notify(metrics, ip, ts)
		if err != nil {
			logger.Log.Error(err.Error())
			continue
		}
	}
}
