package main

import (
	"flag"
)

const defaultStoreInterval = 30
const metricsPath = "metrics.json"
const defaultDBDNS = "postgresql://username:password@localhost:5432/metrics"

type ServerFlags struct {
	address string

	storeInterval   int64
	fileStoragePath string
	restore         bool
	dbDNS           string
}

func (flags *ServerFlags) Init() {
	flag.StringVar(&flags.address, "a", ":8080", "Address and port to run server")

	flag.Int64Var(&flags.storeInterval, "i", defaultStoreInterval, "store interval")
	flag.StringVar(&flags.fileStoragePath, "f", metricsPath, "file storage path")
	flag.BoolVar(&flags.restore, "r", false, "restore")

	flag.StringVar(&flags.dbDNS, "d", defaultDBDNS, "db dns")

	flag.Parse()
}
