package main

import (
	"fmt"
	"github.com/hedisam/CryptoXchange/config"
	"github.com/hedisam/CryptoXchange/datasource/shrimpy"
	"log"
	"time"
)

func main() {
	cfg, err := config.Read("config/config.json")
	if err != nil {
		log.Fatalf("%+v", err)
	}
	log.Println("[!] config file read")

	stream(cfg, orderBook)
}

func bbo(done <-chan struct{}, ds *shrimpy.Shrimpy) (<-chan shrimpy.StreamData, error) {
	return ds.BBO(done, "BTC-USD", "ETH-USD")
}

func orderBook(done <-chan struct{}, ds *shrimpy.Shrimpy) (<-chan shrimpy.StreamData, error) {
	return ds.OrderBook(done, "btc-usd")
}

func stream(cfg *config.Config, streamer func(done <-chan struct{}, ds *shrimpy.Shrimpy) (<-chan shrimpy.StreamData, error)) {
	ds, err := shrimpy.NewShrimpyDataSource(cfg)
	if err != nil {
		panic(err)
	}

	done := make(chan struct{})

	time.AfterFunc(5*time.Minute, func() {
		defer close(done)
		log.Println("Terminating the job")
	})

	log.Println("[!] Start streaming...")
	stream, err := streamer(done, ds)
	if err != nil {
		log.Fatal(err)
	}

	limitedStream := take(done, stream, 5) // including the snapshot

	for data := range limitedStream {
		if data.Err != nil {
			log.Println("stream:", data.Err)
			return
		} else if data.Price.Snapshot {
			fmt.Println("[!] skipping the snapshot...")
			continue
		}

		fmt.Printf("------------------------\n%v\n", data.Price)
	}
}

func take(done <-chan struct{}, in <-chan shrimpy.StreamData, count int) <-chan shrimpy.StreamData {
	out := make(chan shrimpy.StreamData)
	go func() {
		defer close(out)
		for i := 0; i < count; i++ {
			select {
			case <-done: return
			case data := <-in:
				select {
				case <-done: return
				case out <- data:
				}
			}
		}
	}()
	return out
}
