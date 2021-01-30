package main

import (
	"fmt"
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
		log.Println("Terminating the job")
	})

	log.Println("[!] Start streaming...")
	stream, err := ds.BBO(done, "BTC-USD")
	if err != nil {
		log.Fatal(err)
	}

	for data := range stream {
		if data.Err != nil {
			log.Println("bbo:", data.Err)
			return
		} else if data.Price.Snapshot {
			fmt.Println("[!] skipping the snapshot...")
			continue
		}

		fmt.Println(data.Price)
	}
}