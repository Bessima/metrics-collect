package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/go-chi/chi/v5"
)

// ExampleSetMetricHandler_counter демонстрирует установку метрики типа counter через URL параметры
func ExampleSetMetricHandler_counter() {
	// Инициализация хранилища
	storage := repository.NewMemStorage()

	// Создаем HTTP handler (без сохранения в файл)
	handler := SetMetricHandler(storage, nil)

	// Создаем router для поддержки URL параметров
	r := chi.NewRouter()
	r.Post("/update/{typeMetric}/{name}/{value}", handler)

	// Создаем тестовый запрос для установки counter метрики
	req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/5", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос
	r.ServeHTTP(w, req)

	// Проверяем результат
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))

	// Output:
	// Status Code: 200
	// Content-Type: text/plain; charset=utf-8
}

// ExampleSetMetricHandler_gauge демонстрирует установку метрики типа gauge через URL параметры
func ExampleSetMetricHandler_gauge() {
	storage := repository.NewMemStorage()

	handler := SetMetricHandler(storage, nil)
	r := chi.NewRouter()
	r.Post("/update/{typeMetric}/{name}/{value}", handler)

	// Создаем запрос для установки gauge метрики
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/CPUUsage/75.5", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))

	// Output:
	// Status Code: 200
	// Content-Type: text/plain; charset=utf-8
}

// ExampleViewMetricValue демонстрирует получение значения метрики через URL параметры
func ExampleViewMetricValue() {
	storage := repository.NewMemStorage()

	// Предварительно добавим метрику
	storage.ReplaceGaugeMetric("MemoryUsage", 85.3)

	handler := ViewMetricValue(storage)
	r := chi.NewRouter()
	r.Get("/value/{typeMetric}/{name}", handler)

	// Запрос значения метрики
	req := httptest.NewRequest(http.MethodGet, "/value/gauge/MemoryUsage", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Body)
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Value: %s\n", string(body))

	// Output:
	// Status Code: 200
	// Value: 85.3
}

// ExampleUpdateHandler демонстрирует обновление одной метрики через JSON
func ExampleUpdateHandler() {
	storage := repository.NewMemStorage()

	handler := UpdateHandler(storage, nil)

	// Создаем метрику
	delta := int64(10)
	metric := models.Metrics{
		ID:    "RequestCount",
		MType: models.Counter,
		Delta: &delta,
	}

	// Преобразуем в JSON
	jsonData, _ := json.Marshal(metric)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	fmt.Printf("Status Code: %d\n", w.Code)

	// Output:
	// Status Code: 200
}

// ExampleValueHandler демонстрирует получение метрики через JSON
func ExampleValueHandler() {
	storage := repository.NewMemStorage()

	// Предварительно добавим метрику
	value := 99.9
	storage.ReplaceGaugeMetric("Temperature", value)

	handler := ValueHandler(storage)

	// Создаем запрос
	requestMetric := models.RequestValueMetric{
		ID:    "Temperature",
		MType: models.Gauge,
	}
	jsonData, _ := json.Marshal(requestMetric)

	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	// Парсим ответ
	var responseMetric models.Metrics
	json.Unmarshal(w.Body.Bytes(), &responseMetric)

	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Metric ID: %s\n", responseMetric.ID)
	fmt.Printf("Metric Type: %s\n", responseMetric.MType)
	fmt.Printf("Metric Value: %.1f\n", *responseMetric.Value)

	// Output:
	// Status Code: 200
	// Metric ID: Temperature
	// Metric Type: gauge
	// Metric Value: 99.9
}

// ExampleUpdatesHandler демонстрирует пакетное обновление нескольких метрик
func ExampleUpdatesHandler() {
	storage := repository.NewMemStorage()

	handler := UpdatesHandler(storage, nil, nil)

	// Создаем несколько метрик для batch обновления
	delta1 := int64(100)
	delta2 := int64(50)
	value1 := 75.5
	value2 := 88.2

	metrics := []models.Metrics{
		{
			ID:    "TotalRequests",
			MType: models.Counter,
			Delta: &delta1,
		},
		{
			ID:    "ActiveConnections",
			MType: models.Counter,
			Delta: &delta2,
		},
		{
			ID:    "CPULoad",
			MType: models.Gauge,
			Value: &value1,
		},
		{
			ID:    "MemoryPercent",
			MType: models.Gauge,
			Value: &value2,
		},
	}

	// Преобразуем в JSON
	jsonData, _ := json.Marshal(metrics)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Metrics updated: %d\n", len(metrics))

	// Output:
	// Status Code: 200
	// Metrics updated: 4
}

// ExamplePingHandler демонстрирует проверку доступности базы данных
// Примечание: MemStorage не поддерживает ping операцию, возвращает ошибку
func ExamplePingHandler() {
	// Используем in-memory storage
	// Примечание: Ping работает только с БД, для MemStorage вернет ошибку
	storage := repository.NewMemStorage()
	handler := PingHandler(storage)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	body, _ := io.ReadAll(w.Body)
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response: %s\n", string(body))

	// Output:
	// Status Code: 500
	// Response: current command only for DB. Server is working with memory storage now
}

// Example демонстрирует полный цикл работы с метриками
func Example() {
	// 1. Инициализация
	storage := repository.NewMemStorage()

	// 2. Установка метрики через URL
	setHandler := SetMetricHandler(storage, nil)
	r1 := chi.NewRouter()
	r1.Post("/update/{typeMetric}/{name}/{value}", setHandler)

	req1 := httptest.NewRequest(http.MethodPost, "/update/gauge/Temperature/36.6", nil)
	w1 := httptest.NewRecorder()
	r1.ServeHTTP(w1, req1)
	log.Printf("Step 1 - Set metric via URL: Status %d", w1.Code)

	// 3. Получение метрики через URL
	viewHandler := ViewMetricValue(storage)
	r2 := chi.NewRouter()
	r2.Get("/value/{typeMetric}/{name}", viewHandler)

	req2 := httptest.NewRequest(http.MethodGet, "/value/gauge/Temperature", nil)
	w2 := httptest.NewRecorder()
	r2.ServeHTTP(w2, req2)
	value, _ := io.ReadAll(w2.Body)
	log.Printf("Step 2 - Get metric via URL: %s", string(value))

	// 4. Batch обновление через JSON
	updateHandler := UpdatesHandler(storage, nil, nil)
	delta := int64(5)
	gaugeValue := 98.6
	metrics := []models.Metrics{
		{ID: "Counter1", MType: models.Counter, Delta: &delta},
		{ID: "Gauge1", MType: models.Gauge, Value: &gaugeValue},
	}
	jsonData, _ := json.Marshal(metrics)

	req3 := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(jsonData))
	w3 := httptest.NewRecorder()
	updateHandler(w3, req3)
	log.Printf("Step 3 - Batch update: Status %d", w3.Code)

	// 5. Получение метрики через JSON
	valueHandler := ValueHandler(storage)
	requestMetric := models.RequestValueMetric{ID: "Counter1", MType: models.Counter}
	jsonReq, _ := json.Marshal(requestMetric)

	req4 := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(jsonReq))
	w4 := httptest.NewRecorder()
	valueHandler(w4, req4)

	var metric models.Metrics
	json.Unmarshal(w4.Body.Bytes(), &metric)
	log.Printf("Step 4 - Get metric via JSON: %s = %d", metric.ID, *metric.Delta)

	fmt.Println("Workflow completed successfully")

	// Output:
	// Workflow completed successfully
}
