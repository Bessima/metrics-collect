package config

import (
	"encoding/json"
	"log"
	"os"
)

type ConfigServerJson struct {
	Address         string `json:"address"`
	StoreInterval   int64  `json:"store_interval"`
	FileStoragePath string `json:"store_file"`
	Restore         bool   `json:"restore"`
	DatabaseDNS     string `json:"database_dns"`
	CryptoKey       string `json:"crypto_key"`
}

func NewConfigJson() *ConfigServerJson {
	return &ConfigServerJson{
		Address:         ":8080",
		StoreInterval:   defaultStoreInterval,
		FileStoragePath: metricsPath,
		Restore:         false,
		DatabaseDNS:     defaultDBDNS,
		CryptoKey:       "",
	}
}

func (cfg *ConfigServerJson) Parse(filename string) error {
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
