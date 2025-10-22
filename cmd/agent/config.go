package main

import (
	"github.com/caarlos0/env"
	"log"
	"strings"
)

type Config struct {
	serverAddress  string `env:"ADDRESS"`
	reportInterval int64  `env:"REPORT_INTERVAL"`
	poolInterval   int64  `env:"POLL_INTERVAL"`
}

func InitConfig() *Config {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Println(err)
	}

	flags := AgentFlags{}
	flags.Init()

	if cfg.serverAddress == "" {
		cfg.serverAddress = flags.serverAddress
	}
	if cfg.reportInterval == 0 {
		cfg.reportInterval = flags.reportInterval
	}
	if cfg.poolInterval == 0 {
		cfg.poolInterval = flags.poolInterval
	}
	return &cfg
}

func (cfg *Config) getServerAddressWithProtocol() string {
	http := "http://"
	https := "https://"

	if strings.Contains(cfg.serverAddress, https) || strings.Contains(cfg.serverAddress, http) {
		return cfg.serverAddress
	}
	return http + cfg.serverAddress
}
