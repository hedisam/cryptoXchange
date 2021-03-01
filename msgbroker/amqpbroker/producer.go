package amqp

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
	"time"
)

type amqpProducer struct {
	connection *amqp.Connection
	exchange string
	channel *amqp.Channel
}

func NewAMQPProducerFromEnv(exchange string) (*amqpProducer, error) {
	url := os.Getenv("AMQP_URL")
	if url == "" {
		url = "amqp://localhost:5672"
	}

	conn := <- Connect(url, 3 * time.Second, 5)
	return NewAMQPProducer(conn, exchange)
}

func NewAMQPProducer(conn *amqp.Connection, exchange string) (*amqpProducer, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("producer: failed to open amqp channel: %w", err)
	}

	err = channel.ExchangeDeclare(exchange, "topic", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("producer: failed to declare amqp exchange %s: %w", exchange, err)
	}

	return &amqpProducer{connection: conn, exchange: exchange, channel: channel}, nil
}

func (p *amqpProducer) Publish(msg []byte, msgType string) error {
	amqpMessage := amqp.Publishing{
		Headers:         amqp.Table{"x-msg-type": msgType},
		Body:            msg,
	}
	return p.channel.Publish(p.exchange, msgType, false, false, amqpMessage)
}

func (p *amqpProducer) Dispose() {
	if err := p.channel.Close(); err != nil {
		log.Printf("amqp producer on dispose: failed to close the channel: %v", err)
	}
}