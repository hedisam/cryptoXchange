package amqp

import (
	"github.com/streadway/amqp"
	"log"
	"time"
)

func Connect(url string, retryInterval time.Duration, maxRetries int) <-chan *amqp.Connection {
	connChan := make(chan *amqp.Connection)
	go func() {
		defer close(connChan)
		var conn *amqp.Connection
		var err error
		for i := 0; i < maxRetries; i++ {
			conn, err = amqp.Dial(url)
			if err == nil {
				log.Println("AMQP connection established successfully")
				connChan <- conn
				return
			}
			log.Printf("failed to establish AMQP connection. retrying in %v, %v", retryInterval, err)
			time.Sleep(retryInterval)
		}
		panic("AMQP connect reached max retires")
	}()
	return connChan
}
