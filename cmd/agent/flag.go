package main

import (
	"flag"
	"strings"
)

const defaultPollInterval = 2
const defaultReportInterval = 10

type AgentFlags struct {
	serverAddress  string
	pollInterval   int64
	reportInterval int64
}

func (f *AgentFlags) getServerAddressWithProtocol() string {
	http := "http://"
	https := "https://"

	if strings.Contains(f.serverAddress, https) || strings.Contains(f.serverAddress, http) {
		return f.serverAddress
	}
	return http + f.serverAddress
}

func (f *AgentFlags) Init() {
	flag.StringVar(&f.serverAddress, "a", "http://localhost:8080", "address and port server")
	flag.Int64Var(&f.pollInterval, "p", defaultPollInterval, "poll interval")
	flag.Int64Var(&f.reportInterval, "r", defaultReportInterval, "report interval")

	flag.Parse()
}
