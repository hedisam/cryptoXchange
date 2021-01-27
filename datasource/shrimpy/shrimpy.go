package shrimpy

import (
	"fmt"
	"github.com/hedisam/CryptoXchange/config"
)

const (
	baseUrl   = "https://dev-api.shrimpy.io"
	wsBaseUrl = "wss://ws-feed.shrimpy.io"
	tokenPath = "/v1/ws/token"

	apiKeyHeader	= "DEV-SHRIMPY-API-KEY"
	apiNonceHeader  = "DEV-SHRIMPY-API-NONCE"
	apiSigHeader	= "DEV-SHRIMPY-API-SIGNATURE"

	exchange	= "coinbasepro"
)

type Shrimpy struct {
	config *shrimpyConfig
}

func NewShrimpyDataSource(cfg *config.Config) (*Shrimpy, error) {
	if cfg.DataSource.APIKey == "" || cfg.DataSource.SecretKey == "" {
		return nil, fmt.Errorf("[ShrimpyDataSource] invalid api_key/secret_key is provided")
	}

	shConfig := &shrimpyConfig{
		apiKey:    cfg.DataSource.APIKey,
		secretKey: cfg.DataSource.SecretKey,
	}
	return &Shrimpy{config: shConfig}, nil
}