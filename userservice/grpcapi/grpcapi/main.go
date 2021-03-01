package main

import (
	"fmt"
	"github.com/hedisam/CryptoXchange/msgbroker"
	"github.com/hedisam/CryptoXchange/msgbroker/amqpbroker"
	"github.com/hedisam/CryptoXchange/userservice/grpcapi"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	port := 7672
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	connChan := amqpbroker.ConnectFromEnv(3, 3)
	amqpConnection := <-connChan
	// this way we can reuse the amqp connection without making our grpc handler dependent on the concrete amqp connection
	consumerBuilder := func(exchange string) (msgbroker.MsgConsumer, error) {
		return amqpbroker.NewAMQPConsumer(amqpConnection, exchange)
	}
	defer amqpConnection.Close()

	grpcServer := grpc.NewServer()
	grpcapi.RegisterPriceServiceServer(grpcServer, grpcapi.NewGrpcApiServer(consumerBuilder))

	fmt.Println("listening on", port)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}
