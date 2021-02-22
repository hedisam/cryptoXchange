package shrimpymock

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hedisam/CryptoXchange/models"
	"github.com/hedisam/shrimpygo"
	"io/ioutil"
	"strconv"
	"sync"
)

func Streamer(ctx context.Context, path string) (<-chan *models.PriceData, <-chan error, error) {
	done := make(chan struct{})

	fileStream, err := fileStreamer(done, path)
	if err != nil {
		return nil, nil, err
	}

	priceStream, errors := parser(ctx, fileStream)
	return priceStream, errors, nil
}

func fileStreamer(done <-chan struct{}, path string) (<-chan shrimpygo.OrderBook, error) {
	data, err := readData(path)
	if err != nil {
		return nil, err
	}

	rawStream := make(chan shrimpygo.OrderBook)

	// sending raw data to the rawStream chan
	go func() {
		defer close(rawStream)
	LOOP:
		for i := 0; i < len(data); i++ {
			select {
			case rawStream <- data[i]:
			case <-done: return
			}
		}
		goto LOOP
	}()

	return rawStream, nil
}

func parser(ctx context.Context, in <-chan shrimpygo.OrderBook) (<-chan *models.PriceData, <-chan error) {
	out := make(chan *models.PriceData)
	errors := make(chan error)

	go func() {
		defer close(out)
		defer close(errors)

		for {
			select {
			case <-ctx.Done(): return
			case data := <-in:
				parsed, err := parse(data)
				if err != nil {
					select {
					case <-ctx.Done(): return
					case errors <- err: return
					}
				}
				select {
				case <-ctx.Done(): return
				case out <- parsed:
				}
			}
		}
	}()

	return out, errors
}

func parse(data shrimpygo.OrderBook) (priceData *models.PriceData, err error) {
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}

	priceData = &models.PriceData{
		Exchange: data.Exchange,
		Pair:     data.Pair,
		Snapshot: data.Snapshot,
		Sequence: data.Sequence,
		Asks: make([]models.Quote, len(data.Content.Asks)),
		Bids: make([]models.Quote, len(data.Content.Bids)),
	}

	quotesParser := func(src []shrimpygo.OrderBookItem, dest []models.Quote) {
		defer wg.Done()
		for i, item := range src {
			quote, quoteErr := parseQuoteData(&item)
			if quoteErr != nil {
				lock.Lock()
				err = quoteErr
				lock.Unlock()
				return
			}
			dest[i] = *quote
		}
	}

	wg.Add(2)
	go quotesParser(data.Content.Asks, priceData.Asks)
	go quotesParser(data.Content.Bids, priceData.Bids)
	wg.Wait()

	return
}

func parseQuoteData(data *shrimpygo.OrderBookItem) (*models.Quote, error) {
	quote := &models.Quote{}
	var err error

	quote.Price, err = strconv.ParseFloat(data.Price, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to convert price from string to float: %w", err)
	}

	quote.Volume, err = strconv.ParseFloat(data.Quantity, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to convert quantity from string to float: %w", err)
	}

	return quote, nil
}

func readData(path string) ([]shrimpygo.OrderBook, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open data file: %w", err)
	}
	var data []shrimpygo.OrderBook
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal json data: %w", err)
	}
	return data, nil
}