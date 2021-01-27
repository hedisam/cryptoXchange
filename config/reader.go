package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DataSource DataSourceConfig `json:"data_source"`
}

type DataSourceConfig struct {
	APIKey string `json:"api_key"`
	SecretKey string `json:"secret_key"`
}

func ReadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("[ReadConfig] failed to read the config file: %w", err)
	}

	var config *Config
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("[ReadConfig] failed to decode the config json data: %w", err)
	}

	return config, nil
}

