package main

import (
	"flag"
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
}

func (f *AgentFlags) Init() {
	flag.StringVar(&f.serverAddress, "a", "http://localhost:8080", "address and port server")
	flag.Int64Var(&f.poolInterval, "p", defaultPollInterval, "poll interval")
	flag.Int64Var(&f.reportInterval, "r", defaultReportInterval, "report interval")
	flag.StringVar(&f.key, "k", "", "key for hash")
	flag.IntVar(&f.rateLimit, "l", defaultRateLimit, "rate limit for pool")

	flag.Parse()
}
