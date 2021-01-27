package shrimpy

import (
	"fmt"
	"github.com/gorilla/websocket"
)

type dataStream struct {
	data <-chan []byte
	errors <-chan error
	cmdErrors <-chan error
}

func createStream(cfg *shrimpyConfig, cmdChan <-chan interface{}, done <-chan struct{}) (*dataStream, error) {
	// the subscription requires a valid wsToken
	token, err := getToken(cfg)
	if err != nil {
		return nil, fmt.Errorf("[createStream] couldn't get a stream wsToken: %w", err)
	}

	// connecting to the server
	url := fmt.Sprintf("%s?token=%s", wsBaseUrl, token)
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("[createStream] websocket dial failed: %w", err)
	}

	data := make(chan []byte, 10)
	errors := make(chan error, 1)
	cmdErrors := make(chan error, 2)

	go func() {
		defer close(cmdErrors)
		for {
			select {
			case <-done: return
			case cmd := <-cmdChan:
				wErr := client.WriteJSON(cmd)
				if wErr != nil {
					cmdErrors <- fmt.Errorf("[createStream] could not send message: %w", wErr)
				}
			}
		}
	}()

	go func() {
		defer client.Close()
		defer close(errors)
		defer close(data)

		// reading the socket
		for {
			select {
			case <-done: return
			default:
			}
			_, msg, rErr := client.ReadMessage()
			if rErr != nil {
				errors <- fmt.Errorf("[createStream] websocket read error: %w", rErr)
				return
			}
			data <- msg
		}
	}()

	return &dataStream{data: data, errors: errors, cmdErrors: cmdErrors}, nil
}
