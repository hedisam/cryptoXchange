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
func createStream(cfg *shrimpyConfig, cmdChan <-chan interface{}, done <-chan struct{}) (*dataStream, error) {
	// the subscription requires a valid ws token
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

	data := make(chan []byte)
	errors := make(chan error)
	var wg sync.WaitGroup

	go func() {
		defer func() {
			// signalling the receiver goroutine to return by disposing the ws client; in case of the done channel
			// not been closed so the receiver won't get stuck on listening the ws socket.
			_ = client.Close()
			wg.Wait()
			close(errors)
		}()
		for {
			select {
			case <-done: return
			case cmd, ok := <-cmdChan:
				if !ok {return}
				wErr := client.WriteJSON(cmd)
				if wErr != nil {
					errors <- fmt.Errorf("[createStream] could not send message: %w", wErr)
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
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
			select {
			case <-done: return
			case data <- msg:

			}
		}
	}()

	return &dataStream{data: data, errors: errors}, nil
}
