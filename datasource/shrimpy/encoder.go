package shrimpy

import "fmt"

type StreamData struct {
	Price *PriceData
	Err   error
}

func outputer(done <-chan struct{}, in <-chan interface{}, errors <-chan error, upstream chan interface{}) <-chan StreamData {
	output := make(chan StreamData)
	go func() {
		defer close(output)
		defer close(upstream)
		for {
			select {
			case <-done:
				return
			case err := <-errors:
				select {
				case <-done:
				case output <- StreamData{Err: err}:
				}
			case data := <-in:
				switch parsed := data.(type) {
				case pingPong:
					select {
					case <-done:
					case upstream <- parsed:
					}
				case PriceData:
					select {
					case <-done:
					case output <- StreamData{Price: &parsed}:
					}
				default:
					select {
					case <-done:
					case output <- StreamData{Err: fmt.Errorf("unknown parsed data type: %v", parsed)}:
					}
				}
			}
		}
	}()

	return output
}

