package msgbroker

import "context"

// MsgProducer will be implemented by components that emit data messages (e.g price data by data source)
type MsgProducer interface {
	Publish(msg []byte, topic string) error
}

// MsgConsumer will be implemented by those components interested in data messages. (e.g. users can subscribe to
// price data through the grpc api)
type MsgConsumer interface {
	Consume(ctx context.Context, topics ...string) (<-chan []byte, error)
}