package main

import (
	"github.com/caarlos0/env"
	"log"
)

type Config struct {
	Address string `env:"ADDRESS"`
}

func InitConfig() *Config {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Println(err)
	}

	if cfg.Address == "" {
		flags := ServerFlags{}
		flags.Init()

		cfg.Address = flags.address
	}

	return &cfg
}
