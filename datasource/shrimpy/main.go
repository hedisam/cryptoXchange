package main

import (
	"context"
	"fmt"
	"github.com/hedisam/CryptoXchange/datasource/shrimpy/shrimpymock"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, errors, err := shrimpymock.Streamer(ctx, "datasource/shrimpy/shrimpymock/price_data.json")
	if err != nil {
		log.Fatalf("failed to setup the stream: %v", err)
	}

	for i := 0; i < 1000; i++ {
		select {
		case data, ok := <-stream:
			if !ok {return}
			fmt.Printf("[%d]: %v\n", i, data)
		case err = <-errors:
			log.Printf("Error from shrimpy data source: %v\n", err)
			return
		}
	}
}
