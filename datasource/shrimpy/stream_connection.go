package shrimpy

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type streamConnection struct {
	client *websocket.Conn
	errors chan error
	wg sync.WaitGroup
}

func (c *streamConnection) receive(done <-chan struct{}, data chan []byte) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer close(data)

		// reading the socket
		for {
			select {
			case <-done:
				return
			default:
			}
			_, msg, err := c.client.ReadMessage()
			if err != nil {
				c.sendErr(done, fmt.Errorf("[createStream] websocket read error: %w", err))
				return
			}
			select {
			case <-done:
				return
			case data <- msg:

			}
		}
	}()
}

func (c *streamConnection) upstream(done <-chan struct{}, in <-chan interface{}) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-done:
				return
			case msg, ok := <-in:
				if !ok {
					return
				}
				err := c.client.WriteJSON(msg)
				if err != nil {
					c.sendErr(done, fmt.Errorf("[createStream] could not send message: %w", err))
					return
				}
			}
		}
	}()
}

func (c *streamConnection) dispose() {
	go func() {
		c.wg.Wait()
		close(c.errors)
		_ = c.client.Close()
	}()
}

func (c *streamConnection) sendErr(done <-chan struct{}, err error) {
	select {
	case <-done:
	case c.errors <- err:
	}
}
