// Модуль config предназначен для настроек конфигураций сервера
package config

import (
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
)

// Config хранит основные пареметры для запуска сервера
type Config struct {
	// Address адрес и порт для запуска сервера
	Address string `env:"ADDRESS"`

	// StoreInterval интервал записи в хранилище
	StoreInterval int64 `env:"STORE_INTERVAL"`
	// FileStoragePath путь для сохранения данных
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// Restore перезапись
	Restore bool `env:"RESTORE"`
	// DatabaseDNS Адрес доступа к БД
	DatabaseDNS string `env:"DATABASE_DSN"`
	// KeyHash Хэш-ключ
	KeyHash string `env:"KEY"`
	//AuditFile путь для сохранения аудит данных в файл
	AuditFile string `env:"AUDIT_FILE"`
	//AuditURL аддрес сервера для сохранения аудит данных в файл
	AuditURL string `env:"AUDIT_URL"`
}

func InitConfig() *Config {
	flags := ServerFlags{}
	flags.Init()

	cfg := Config{
		Address:         flags.address,
		StoreInterval:   flags.storeInterval,
		FileStoragePath: flags.fileStoragePath,
		Restore:         flags.restore,
		DatabaseDNS:     flags.dbDNS,
		KeyHash:         flags.keyHash,
		AuditFile:       flags.auditFile,
		AuditURL:        flags.auditURL,
	}
	cfg.parseEnv()

	return &cfg
}

func (cfg *Config) parseEnv() {
	err := env.Parse(cfg)
	if err != nil {
		logger.Log.Warn("Getting an error while parsing the configuration", zap.String("err", err.Error()))
	}
}
