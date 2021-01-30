package shrimpy

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type dataStream struct {
	data <-chan []byte
	errors <-chan error
}

// createStream connects to the ws server and returns a dataStream object containing the outbound channel.
func (shrimpy *Shrimpy) createStream(done <-chan struct{}, upstreamChan <-chan interface{}) (*dataStream, error) {
	client, err := shrimpy.setupWSConnection()
	if err != nil {
		return nil, fmt.Errorf("[createStream] couldn't connect to the ws server: %w", err)
	}

	dataChan := make(chan []byte) // incoming messages will be pushed to the data channel.
	connection := &streamConnection{
		client: client,
		errors: make(chan error),
		wg:     sync.WaitGroup{},
	}

	// listen for and send subscription requests & pong replies
	connection.upstream(done, upstreamChan)

	// listen from incoming messages and push them to the stream channel
	connection.receive(done, dataChan)

	// dispose starts a new goroutine and waits for the upstream and receive goroutines to get done. we need it
	// synchronized since upstream and receiver both write to the same error channel.
	connection.dispose()

	return &dataStream{data: dataChan, errors: connection.errors}, nil
}

func (shrimpy *Shrimpy) setupWSConnection() (*websocket.Conn, error) {
	// the ws server requires a valid token
	token, err := getToken(shrimpy.config)
	if err != nil {
		return nil, fmt.Errorf("[setupWSConnection] couldn't get a websocket token: %w", err)
	}

	// connecting to the server
	url := fmt.Sprintf("%s?token=%s", wsBaseUrl, token)
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("[setupWSConnection] websocket dial failed: %w", err)
	}
	return client, nil
}
