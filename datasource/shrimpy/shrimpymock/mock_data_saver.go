package shrimpymock

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hedisam/CryptoXchange/config"
	"github.com/hedisam/shrimpygo"
	"io/ioutil"
	"log"
	"os"
)

func UpdateMockData() {
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

	ws.Subscribe(shrimpygo.BBOSubs("coinbasepro", "eth-usd"))
	ws.Subscribe(shrimpygo.BBOSubs("coinbasepro", "btc-usd"))

	n := 1000
	limitedStream := take(ctx, ws.Stream(), n)
	dataArray := make([]*shrimpygo.OrderBook, n-1)
	i := 0

	defer func() {
		// write data to a json file
		fmt.Println("writing data")
		b, err := json.Marshal(dataArray)
		if err != nil {
			log.Println("error writing data to json file:", err)
			return
		}
		if err = ioutil.WriteFile("datasource/shrimpy/shrimpymock/price_data.json", b, os.ModePerm); err != nil {
			log.Println("Failed to write price_data:", err)
		}
	}()

	for update := range limitedStream {
		switch data := update.(type) {
		case *shrimpygo.OrderBook:
			if data.Snapshot {
				continue
			}
			dataArray[i] = data
			i++
			fmt.Printf("BBO[%d]: %v\n", i, data)
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
