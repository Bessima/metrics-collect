package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/Bessima/metrics-collect/internal/config/db"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	models "github.com/Bessima/metrics-collect/internal/model"
	"go.uber.org/zap"
)

type DBRepository struct {
	db *db.DB
}

func NewDBRepository(rootContext context.Context, databaseDNS string) *DBRepository {
	dbObj, errDB := db.NewDB(rootContext, databaseDNS)

	if errDB != nil {

		logger.Log.Error(
			"Unable to connect to database",
			zap.String("path", databaseDNS),
			zap.String("error", errDB.Error()),
		)
	}

	return &DBRepository{db: dbObj}
}

func (repository *DBRepository) Counter(name string, value int64) error {
	query := `INSERT INTO metrics (name, type, delta) VALUES ($1, $2, $3) ON CONFLICT (name, type) DO UPDATE SET delta = metrics.delta + EXCLUDED.delta`
	result, err := repository.db.Pool.Exec(context.Background(), query, name, TypeCounter, value)
	if err != nil {
		return err
	}
	affected := result.RowsAffected()
	if affected == 0 {
		return errors.New("counter metric is not changed")
	}
	return nil
}

func (repository *DBRepository) ReplaceGaugeMetric(name string, value float64) error {
	query := "INSERT INTO metrics (name, type, value) VALUES ($1, $2, $3) ON CONFLICT (name, type) DO UPDATE SET value = EXCLUDED.value"
	result, err := repository.db.Pool.Exec(context.Background(), query, name, TypeGauge, value)
	if err != nil {
		return err
	}
	affected := result.RowsAffected()

	if affected == 0 {
		return errors.New("gauge metric is not changed")
	}
	return nil
}

func (repository *DBRepository) GetValue(typeMetric TypeMetric, name string) (interface{}, error) {
	metric, err := repository.GetMetric(typeMetric, name)
	if err != nil {
		return nil, err
	}
	switch {
	case typeMetric == TypeCounter:
		return *metric.Delta, err

	case typeMetric == TypeGauge:
		return *metric.Value, err
	default:
		err = fmt.Errorf("unknown metric type: %s", typeMetric)
	}

	return nil, err
}

func (repository *DBRepository) GetMetric(typeMetric TypeMetric, name string) (models.Metrics, error) {
	row := repository.db.Pool.QueryRow(
		context.Background(),
		"SELECT name, type, value, delta FROM metrics WHERE name = $1 AND type = $2 LIMIT 1",
		name,
		typeMetric,
	)

	elem := models.Metrics{}

	err := row.Scan(&elem.ID, &elem.MType, &elem.Value, &elem.Delta)

	return elem, err
}

func (repository *DBRepository) Load(metrics []models.Metrics) error {
	ctx := context.Background()
	tx, err := repository.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		ctx,
		"insert or update metric",
		"INSERT INTO metrics (name, type, value, delta) VALUES($1,$2,$3,$4)"+
			" ON CONFLICT (name, type) DO UPDATE SET value = EXCLUDED.value",
	)
	if err != nil {
		return err
	}

	for _, m := range metrics {
		_, err = tx.Exec(ctx, stmt.SQL, m.ID, m.MType, m.Value, m.Delta)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	}
	return tx.Commit(ctx)

}
func (repository *DBRepository) All() ([]models.Metrics, error) {
	rows, err := repository.db.Pool.Query(context.Background(), "SELECT name, type, value, delta FROM metrics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := []models.Metrics{}

	for rows.Next() {
		var metric models.Metrics
		err = rows.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, metric)
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (repository *DBRepository) Ping(ctx context.Context) error {
	err := repository.db.Pool.Ping(ctx)
	return err
}

func (repository *DBRepository) Close() error {
	repository.db.Close()
	return nil
}
