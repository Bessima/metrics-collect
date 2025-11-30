package config

import (
	"flag"
)

const defaultStoreInterval = 30
const metricsPath = ""
const defaultDBDNS = ""

type ServerFlags struct {
	address string

	storeInterval   int64
	fileStoragePath string
	restore         bool
	dbDNS           string
	keyHash         string
}

func (flags *ServerFlags) Init() {
	flag.StringVar(&flags.address, "a", ":8080", "Address and port to run server")

	flag.Int64Var(&flags.storeInterval, "i", defaultStoreInterval, "store interval")
	flag.StringVar(&flags.fileStoragePath, "f", metricsPath, "file storage path")
	flag.BoolVar(&flags.restore, "r", false, "restore")

	flag.StringVar(&flags.dbDNS, "d", defaultDBDNS, "db dns")
	flag.StringVar(&flags.keyHash, "k", "", "key for hash")

	flag.Parse()
}
