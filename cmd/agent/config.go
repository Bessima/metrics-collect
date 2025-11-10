package main

import (
	"github.com/caarlos0/env"
	"log"
	"strings"
)

type Config struct {
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PoolInterval   int64  `env:"POLL_INTERVAL"`
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

func (cfg *Config) parseFlag() *AgentFlags {
	flags := AgentFlags{}
	flags.Init()
	return &flags
}

func (cfg *Config) mergeConfig(flags *AgentFlags) {
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = flags.serverAddress
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = flags.reportInterval
	}
	if cfg.PoolInterval == 0 {
		cfg.PoolInterval = flags.poolInterval
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
