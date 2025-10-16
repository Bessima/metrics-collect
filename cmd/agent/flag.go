package main

import (
	"flag"
	"strconv"
	"strings"
	"time"
)

type Flags struct {
	serverAddress  string
	pollInterval   string
	reportInterval string
}

func (f *Flags) getServerAddressWithProtocol() string {
	http := "http://"
	https := "https://"

	if strings.Contains(f.serverAddress, https) || strings.Contains(f.serverAddress, http) {
		return f.serverAddress
	}
	return http + f.serverAddress
}

func (f *Flags) getPollInterval() time.Duration {
	value, err := strconv.Atoi(f.pollInterval)
	if err != nil {
		panic(err)
	}
	return time.Duration(value)
}

func (f *Flags) getReportInterval() float64 {
	value, err := strconv.Atoi(f.reportInterval)
	if err != nil {
		panic(err)
	}
	return float64(value)
}

func (f *Flags) Init() {
	flag.StringVar(&f.serverAddress, "a", "http://localhost:8080", "address and port server")
	flag.StringVar(&f.pollInterval, "p", "2", "poll interval")
	flag.StringVar(&f.reportInterval, "r", "10", "report interval")

	flag.Parse()
}
