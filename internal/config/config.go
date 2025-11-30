package config

import (
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
)

type Config struct {
	Address string `env:"ADDRESS"`

	StoreInterval   int64  `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDNS     string `env:"DATABASE_DSN"`
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
