package main

import (
	"flag"
	"os"
)

const defaultPollInterval = 2
const defaultReportInterval = 10
const defaultRateLimit = 10

type AgentFlags struct {
	serverAddress  string
	poolInterval   int64
	reportInterval int64
	key            string
	rateLimit      int
	cryptoKey      string
	config         string
}

func (f *AgentFlags) Init() {
	flag.StringVar(&f.config, "c", "", "config json file path")

	flag.Parse()

	cfgJSON := NewConfigJSON()
	configValue, exists := os.LookupEnv("CONFIG")
	if exists {
		cfgJSON.Parse(configValue)
	} else if f.config != "" {
		cfgJSON.Parse(f.config)
	}

	flag.StringVar(&f.serverAddress, "a", cfgJSON.ServerAddress, "address and port server")
	flag.Int64Var(&f.poolInterval, "p", cfgJSON.PoolInterval, "poll interval")
	flag.Int64Var(&f.reportInterval, "r", cfgJSON.ReportInterval, "report interval")
	flag.StringVar(&f.key, "k", "", "key for hash")
	flag.IntVar(&f.rateLimit, "l", defaultRateLimit, "rate limit for pool")
	flag.StringVar(&f.cryptoKey, "crypto_message-key", cfgJSON.CryptoKey, "crypto_message key")

	flag.Parse()
}
