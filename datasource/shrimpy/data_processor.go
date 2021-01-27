package shrimpy

import (
	"encoding/json"
	"fmt"
)

const (
	ping = iota
	dataError
)

type processorStatus struct {
	Type int
	pingData int64
	err error
}

func processStream(done <-chan struct{}, in <-chan []byte) <-chan *processorStatus {
	statusChan := make(chan *processorStatus, 5)

	go func() {
		defer close(statusChan)
		
		for {
			select {
			case body := <-in:
				status := process(body)
				if status != nil {
					select {
					case statusChan <- status:
					case <-done: return
					}
				}
			case <-done: return
			}
		}
	}()

	return statusChan
}

func process(body []byte) *processorStatus {
	var dataUnknown unknownData
	err := json.Unmarshal(body, &dataUnknown)
	if err != nil {
		return &processorStatus{Type: dataError, err:  fmt.Errorf("[processStream] unable to process data: %w", err)}
	}

	if dataUnknown.Type != "" {
		if dataUnknown.Type == "ping" {
			return &processorStatus{Type: ping, pingData: dataUnknown.Data}
		} else if dataUnknown.Type == "error" {
			return &processorStatus{Type: dataError,
				err: fmt.Errorf("[processStream] received an error message from the server; code %d - message: %s",
					dataUnknown.Code, dataUnknown.Message)}
		} else {
			return &processorStatus{Type: dataError, err: fmt.Errorf("[processStream] unknown data: %v", dataUnknown)}
		}
	}

	if dataUnknown.Exchange == "" {
		return &processorStatus{Type: dataError, err: fmt.Errorf("[processStream] unknown data: %v", dataUnknown)}
	}

	// we have a price data. do whatever you want to do with it.
	var priceData PriceData
	err = json.Unmarshal(body, &priceData)
	if err != nil {
		return &processorStatus{Type: dataError, err: fmt.Errorf("[processData] failed to unmarshal price data: %v", string(body))}
	}

	if priceData.Snapshot {
		// the snapshot includes a lot of data. we just don't wanna print it.
		return nil
	}

	fmt.Printf("----------------------\n%v\n", priceData)

	return nil
}

func extractPingData(data map[string]interface{}) (int64, error) {
	iData, ok := data["data"]
	if !ok {
		return 0, fmt.Errorf("[processStream] received a ping message but no ping data: %v", data)
	}
	pingData, ok := iData.(int64)
	if !ok {
		return 0, fmt.Errorf("[processStream] unkown field type, field 'data' for ping message expected to be int64: %v", data)
	}
	return pingData, nil
}

func extractDataType(iDataType interface{}) (string, error) {
	dataType, ok := iDataType.(string)
	if !ok {
		return "", fmt.Errorf("[processStream] unkown field type, field 'type' expected to be string: %v", iDataType)
	}
	return dataType, nil
}

func unmarshalData(body []byte) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("[unmarshalData] failed to unmarshal the data: %w", err)
	}
	return data, nil
}
