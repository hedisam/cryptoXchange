package amqpbroker

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
	"time"
)

type amqpConsumer struct {
	conn *amqp.Connection
	channel *amqp.Channel
	exchange string
	queue string
}

func NewAMQPConsumerFromEnv(exchange string) (*amqpConsumer, error) {
	url := os.Getenv("AMQP_URL")
	if url == "" {
		url = "amqp://localhost:5672"
	}

	conn := <-Connect(url, 3 * time.Second, 5)
	return NewAMQPConsumer(conn, exchange)
}

func NewAMQPConsumer(conn *amqp.Connection, exchange string) (*amqpConsumer, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("consumer: failed to open amqp channel: %w", err)
	}

	err = channel.ExchangeDeclare(exchange, "topic", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("consumer: failed to declare amqp exchange %s: %w", exchange, err)
	}

	q, err := channel.QueueDeclare("", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("consumer: failed to declare amqp queue: %w", err)
	}

	return &amqpConsumer{conn: conn, channel: channel, exchange: exchange, queue: q.Name}, nil
}

func (c *amqpConsumer) Consume(ctx context.Context, topics ...string) (<-chan []byte, error) {
	for _, topic := range topics {
		err := c.channel.QueueBind(c.queue, topic, c.exchange, false, nil)
		if err != nil {
			return nil, fmt.Errorf("consumer: amqp failed to bind topic %s to the queue %s on exchange %s: %w",
				topic, c.queue, c.exchange, err)
		}
	}

	delivery, err := c.channel.Consume(c.queue, "", true, false, false, false,
		nil)
	if err != nil {
		return nil, fmt.Errorf("consumer: amqp failed to consume %s on exchange %s: %w", c.queue, c.exchange, err)
	}

	messages := make(chan []byte)

	go func() {
		defer close(messages)

		for msg := range delivery {
			select {
			case <-ctx.Done(): return
			case messages <- msg.Body:
			}
		}
	}()

	return messages, nil
}

func (c *amqpConsumer) Dispose() {
	if err := c.channel.Close(); err != nil {
		log.Printf("amqp consumer on dispose: failed to close the channel: %v", err)
	}
}
