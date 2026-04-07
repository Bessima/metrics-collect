package config

import (
	"flag"
	"os"
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
	auditFile       string
	auditURL        string
	cryptoKey       string
	configFile      string
}

func (flags *ServerFlags) Init() {

	flag.StringVar(&flags.configFile, "c", "", "config json file path")

	flag.Parse()

	cfgJSON := NewConfigJSON()
	configValue, exists := os.LookupEnv("CONFIG")
	if exists {
		cfgJSON.Parse(configValue)
	} else if flags.configFile != "" {
		cfgJSON.Parse(flags.configFile)
	}

	flag.StringVar(&flags.address, "a", cfgJSON.Address, "Address and port to run server")

	flag.Int64Var(&flags.storeInterval, "i", cfgJSON.StoreInterval, "store interval")
	flag.StringVar(&flags.fileStoragePath, "f", cfgJSON.FileStoragePath, "file storage path")
	flag.BoolVar(&flags.restore, "r", cfgJSON.Restore, "restore")

	flag.StringVar(&flags.dbDNS, "d", cfgJSON.DatabaseDNS, "db dns")
	flag.StringVar(&flags.keyHash, "k", "", "key for hash")

	flag.StringVar(&flags.auditFile, "audit-file", "", "path to audit file")
	flag.StringVar(&flags.auditURL, "audit-url", "", "address for applying audit data")

	flag.StringVar(&flags.cryptoKey, "crypto_message-key", cfgJSON.CryptoKey, "crypto_message key")

	flag.Parse()
}
