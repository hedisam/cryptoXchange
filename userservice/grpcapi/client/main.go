package main

import (
	"context"
	"fmt"
	"github.com/hedisam/CryptoXchange/userservice/grpcapi"
	"google.golang.org/grpc"
	"io"
	"log"
)

func main() {
	conn, err := grpc.Dial("localhost:7672", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := grpcapi.NewPriceServiceClient(conn)

	fmt.Println("Prices....")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.StreamPrice(ctx, &grpcapi.QuoteRequest{})
	if err != nil {
		log.Fatal(err)
	}

	i := 0
	for {
		data, err := stream.Recv()
		if err == io.EOF {return}
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("[%d] %v\n", i, data)
		i++
	}
}
