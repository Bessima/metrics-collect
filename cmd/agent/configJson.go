package main

import (
	"encoding/json"
	"log"
	"os"
)

type ConfigAgentJSON struct {
	ServerAddress  string `json:"address"`
	ReportInterval int64  `json:"report_interval"`
	PoolInterval   int64  `json:"pool_interval"`
	CryptoKey      string `json:"crypto_key"`
}

func NewConfigJSON() *ConfigAgentJSON {
	return &ConfigAgentJSON{
		ServerAddress:  "http://localhost:8080",
		ReportInterval: defaultReportInterval,
		PoolInterval:   defaultPollInterval,
		CryptoKey:      "",
	}
}

func (cfg *ConfigAgentJSON) Parse(filename string) error {
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
