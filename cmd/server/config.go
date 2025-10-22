package main

import (
	"github.com/caarlos0/env"
	"log"
)

type Config struct {
	address string `env:"ADDRESS"`
}

func InitConfig() *Config {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Println(err)
	}

	if cfg.address == "" {
		flags := ServerFlags{}
		flags.Init()

		cfg.address = flags.address
	}

	return &cfg
}
