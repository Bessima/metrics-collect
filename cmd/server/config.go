package main

import (
	"github.com/caarlos0/env"
	"log"
	"os"
)

type Config struct {
	Address string `env:"ADDRESS"`

	StoreInterval   int64  `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func InitConfig() *Config {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Println(err)
	}

	flags := ServerFlags{}
	flags.Init()

	if cfg.Address == "" {
		cfg.Address = flags.address
	}

	if _, ok := os.LookupEnv("STORE_INTERVAL"); !ok {
		cfg.StoreInterval = flags.storeInterval
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = flags.fileStoragePath
	}

	if _, ok := os.LookupEnv("RESTORE"); !ok {
		cfg.Restore = flags.restore
	}

	return &cfg
}
