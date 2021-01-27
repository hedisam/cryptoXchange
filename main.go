package main

import (
	"github.com/hedisam/CryptoXchange/config"
	"github.com/hedisam/CryptoXchange/datasource/shrimpy"
	"log"
	"time"
)

func main() {
	cfg, err := config.ReadConfig("config/config.json")
	if err != nil {
		panic(err)
	}
	log.Println("[!] config file read")

	bbo(cfg)
}

func bbo(cfg *config.Config) {
	ds, err := shrimpy.NewShrimpyDataSource(cfg)
	if err != nil {
		panic(err)
	}

	done := make(chan struct{})

	time.AfterFunc(5 * time.Minute, func() {
		defer close(done)
		select {
		case <-time.After(100 * time.Second):
			log.Println("Terminating the job")
		}
	})

	log.Println("[!] Start streaming...")
	err = ds.BBOStream(done, "BTC-USD")
	if err != nil {
		log.Fatal(err)
	}
}