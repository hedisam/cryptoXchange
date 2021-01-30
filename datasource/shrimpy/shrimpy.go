package shrimpy

import (
	"fmt"
	"github.com/hedisam/CryptoXchange/config"
)

const (
	baseUrl   = "https://dev-api.shrimpy.io"
	wsBaseUrl = "wss://ws-feed.shrimpy.io"
	tokenPath = "/v1/ws/token"

	apiKeyHeader   = "DEV-SHRIMPY-API-KEY"
	apiNonceHeader = "DEV-SHRIMPY-API-NONCE"
	apiSigHeader   = "DEV-SHRIMPY-API-SIGNATURE"

	exchange       = "coinbasepro"
	ping           = "ping"
	pong           = "pong"
	shpySubsErrors = "error"
	bestBidOffer   = "bbo"
	orderBook 	   = "orderbook"
)

type Shrimpy struct {
	config *shrimpyConfig
}

func NewShrimpyDataSource(cfg *config.Config) (*Shrimpy, error) {
	if cfg.DataSource.APIKey == "" || cfg.DataSource.SecretKey == "" {
		return nil, fmt.Errorf("shrimpy: invalid data source api/secret key")
	}

	shConfig := &shrimpyConfig{
		apiKey:    cfg.DataSource.APIKey,
		secretKey: cfg.DataSource.SecretKey,
	}
	return &Shrimpy{config: shConfig}, nil
}

func (shrimpy *Shrimpy) OrderBook(done <-chan struct{}, pairs ...string) (<-chan StreamData, error) {
	subs := make([]Subscription, len(pairs))
	for i, pair := range pairs {
		subs[i] = Subscription{
			Type:     "subscribe",
			Exchange: exchange,
			Pair:     pair,
			Channel:  orderBook,
		}
	}

	return shrimpy.Subscribe(done, subs...)
}

func (shrimpy *Shrimpy) BBO(done <-chan struct{}, pairs ...string) (<-chan StreamData, error) {
	subs := make([]Subscription, len(pairs))
	for i, pair := range pairs {
		subs[i] = Subscription{
			Type:     "subscribe",
			Exchange: exchange,
			Pair:     pair,
			Channel:  bestBidOffer,
		}
	}

	return shrimpy.Subscribe(done, subs...)
}

func (shrimpy *Shrimpy) Subscribe(done <-chan struct{}, subscriptions ...Subscription) (<-chan StreamData, error) {
	if len(subscriptions) == 0 {
		return nil, fmt.Errorf("data source failed to subscribe: no subscription found")
	}

	// we use upstreamChan to send subscription requests & ping replies
	upstreamChan := make(chan interface{})

	// create a websocket connection
	stream, err := shrimpy.createStream(done, upstreamChan)
	if err != nil {
		return nil, fmt.Errorf("data source failed to create subscription stream: %w", err)
	}

	// send the subscriptions list
	for _, s := range subscriptions {
		select {
		case <-done:
			return nil, nil
		case upstreamChan <- s:
		}
	}

	// the server returns json messages of different types. we need to parse and marshal the message to its
	// appropriate struct object.
	pout, errors := parser(done, stream.data, stream.errors)
	output := filter(done, pout, errors, upstreamChan)

	return output, nil
}