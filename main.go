package main

import (
	"context"
	"fmt"
	"github.com/hedisam/CryptoXchange/config"
	"github.com/hedisam/shrimpygo"
	"log"
)

func main() {
	appConfig, err := config.Read("config/config.json")
	if err != nil {
		log.Fatal(err)
	}
	cli, err := shrimpygo.NewShrimpyClient(shrimpygo.Config{
		PublicKey:  appConfig.DataSource.APIKey,
		PrivateKey: appConfig.DataSource.SecretKey,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ws, err := cli.Websocket(ctx, 0)
	if err != nil {
		log.Fatal(err)
	}

	ws.Subscribe(shrimpygo.BBOSubs("binance", "btc-usdt"))

	limitedStream := take(ctx, ws.Stream(), 10)

	for update := range limitedStream {
		switch data := update.(type) {
		case *shrimpygo.OrderBook:
			if data.Snapshot {
				continue
			}
			fmt.Println("BBO:", data)
		case error:
			fmt.Println("err received:", data)
			return
		default:
			fmt.Printf("unknown data: %T\n", update)
			return
		}
	}
}

func take(ctx context.Context, in <-chan interface{}, count int) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for i := 0; i <= count; i++ {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case out <- data:
				}
			}
		}
	}()
	return out
}
