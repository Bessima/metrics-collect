package repository

import "errors"

// Предопределенные ошибки для избежания аллокаций через errors.New()
var (
	ErrMetricNotFound             = errors.New("metric not found")
	ErrCounterNotChanged          = errors.New("counter metric is not changed")
	ErrGaugeNotChanged            = errors.New("gauge metric is not changed")
	ErrUnknownMetricType          = errors.New("unknown metric type")
	ErrNotSupportedForMemStorage  = errors.New("current command only for DB. Server is working with memory storage now")
	ErrNotSupportedForFileStorage = errors.New("current command only for DB. Server is working with file storage now")
)
