package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hedisam/CryptoXchange/datasource/shrimpy/shrimpymock"
	"github.com/hedisam/CryptoXchange/models"
	"github.com/hedisam/CryptoXchange/msgbroker"
	"github.com/hedisam/CryptoXchange/msgbroker/amqpbroker"
	"log"
	"strings"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, errors, err := shrimpymock.Streamer(ctx, "shrimpymock/price_data.json")
	if err != nil {
		log.Fatal(err)
	}

	producer, err := amqpbroker.NewAMQPProducerFromEnv("price_data")
	defer producer.Dispose()
	if err != nil {
		log.Fatal("data source failed to get amqp producer:", err)
	}

	fmt.Println("[!] Data source streaming price data")
	for {
		select {
		case err = <- errors:
			log.Println("data source streamer:", err)
			return
		case data := <-stream:
			err = publish(producer, data)
			if err != nil {
				log.Println("data source failed to publish price data:", err)
				return
			}
		}
	}
}

func publish(producer msgbroker.MsgProducer, msg *models.PriceData) error {
	pair := strings.Replace(msg.Pair, "-", ".", 1)
	topic := msg.Channel + "." + pair
	if msg.Snapshot {
		topic += ".snapshot"
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("publish failed to marshal price data: %w", err)
	}
	err = producer.Publish(data, topic)
	if err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}

	return nil
}
