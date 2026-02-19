package main

import (
	"log"
	"strings"

	"github.com/caarlos0/env"
)

type Config struct {
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PoolInterval   int64  `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

func InitConfig() *Config {
	flags := AgentFlags{}
	flags.Init()

	cfg := Config{
		ServerAddress:  flags.serverAddress,
		ReportInterval: flags.reportInterval,
		PoolInterval:   flags.poolInterval,
		Key:            flags.key,
		RateLimit:      flags.rateLimit,
	}

	cfg.parseEnv()

	return &cfg
}

func (cfg *Config) parseEnv() {
	err := env.Parse(cfg)
	if err != nil {
		log.Println(err)
	}
}

func (cfg *Config) getServerAddressWithProtocol() string {
	http := "http://"
	https := "https://"

	if strings.Contains(cfg.ServerAddress, https) || strings.Contains(cfg.ServerAddress, http) {
		return cfg.ServerAddress
	}
	return http + cfg.ServerAddress
}
