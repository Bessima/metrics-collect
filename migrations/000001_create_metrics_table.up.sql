
CREATE TABLE IF NOT EXISTS metrics (
                         name VARCHAR(255) NOT NULL,
                         type VARCHAR(100) NOT NULL,
                         value DOUBLE PRECISION,
                         delta INTEGER
);

-- Базовый индекс для поиска по названию
CREATE INDEX idx_metrics_name ON metrics(name);

CREATE UNIQUE INDEX idx_fields_unique ON metrics (name, type);