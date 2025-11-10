package main

import (
	"github.com/caarlos0/env"
	"log"
)

type Config struct {
	Address *string `env:"ADDRESS"`

	StoreInterval   *int64  `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
}

func InitConfig() *Config {
	var cfg Config
	cfg.parseEnv()
	flags := cfg.parseFlag()
	cfg.mergeConfig(flags)

	return &cfg
}

func (cfg *Config) parseEnv() {
	err := env.Parse(&cfg)
	if err != nil {
		log.Println(err)
	}
}

func (cfg *Config) parseFlag() *ServerFlags {
	flags := ServerFlags{}
	flags.Init()
	return &flags
}

func (cfg *Config) mergeConfig(flags *ServerFlags) {
	if cfg.Address == nil {
		cfg.Address = &flags.address
	}

	if cfg.StoreInterval == nil {
		cfg.StoreInterval = &flags.storeInterval
	}

	if cfg.FileStoragePath == nil {
		cfg.FileStoragePath = &flags.fileStoragePath
	}

	if cfg.Restore == nil {
		cfg.Restore = &flags.restore
	}
}
