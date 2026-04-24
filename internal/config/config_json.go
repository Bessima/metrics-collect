package config

import (
	"encoding/json"
	"log"
	"os"
)

type ConfigServerJSON struct {
	Address         string `json:"address"`
	StoreInterval   int64  `json:"store_interval"`
	FileStoragePath string `json:"store_file"`
	Restore         bool   `json:"restore"`
	DatabaseDNS     string `json:"database_dns"`
	CryptoKey       string `json:"crypto_key"`
}

func NewConfigJSON() *ConfigServerJSON {
	return &ConfigServerJSON{
		Address:         ":8080",
		StoreInterval:   defaultStoreInterval,
		FileStoragePath: metricsPath,
		Restore:         false,
		DatabaseDNS:     defaultDBDNS,
		CryptoKey:       "",
	}
}

func (cfg *ConfigServerJSON) Parse(filename string) error {
	log.Printf("Parsing JSON config file %s", filename)

	settingsFile, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(settingsFile, &cfg)

	if err != nil {
		return err
	}
	return nil
}
