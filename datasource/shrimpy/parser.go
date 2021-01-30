package shrimpy

import (
	"encoding/json"
	"fmt"
)

func parser(done <-chan struct{}, input <-chan []byte, errIn <-chan error) (<-chan interface{}, <-chan error) {
	output := make(chan interface{})
	errors := make(chan error)
	go func() {
		defer close(output)
		defer close(errors)
		for {
			select {
			case <-done:
				return
			case rawData, ok := <-input:
				if !ok {
					return
				}
				parsedData, err := parse(rawData)
				if err != nil {
					select {
					case <-done:
					case errors <- err:
					}
				} else {
					select {
					case <-done:
					case output <- parsedData:
					}
				}
			case err := <-errIn:
				select {
				case <-done:
				case errors <- err:
				}
			}
		}
	}()

	return output, errors
}

func parse(b []byte) (interface{}, error) {
	var data unknownData
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, fmt.Errorf("shrimpy parser: unable to unmarshal data: %w", err)
	}

	if data.Type != "" {
		if data.Type == ping {
			return pingPong{Type: pong, Data: data.Data}, nil
		} else if data.Type == shpySubsErrors {
			return nil, fmt.Errorf("shrimpy ws subscription: error message received: code %d - message: %s",
				data.Code, data.Message)
		} else {
			return nil, fmt.Errorf("shrimpy parser: unknown data received: %v", data)
		}
	}

	if data.Exchange == "" {
		return nil, fmt.Errorf("shrimpy parser: unknown data received: %v", data)
	}

	// we have a Price data
	var priceData PriceData
	err = json.Unmarshal(b, &priceData)
	if err != nil {
		return nil, fmt.Errorf("shrimpy parser: failed to unmarshal price data: %w", err)
	}

	return priceData, nil
}
