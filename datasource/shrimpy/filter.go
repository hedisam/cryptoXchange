package shrimpy

import "fmt"

type StreamData struct {
	Price *PriceData
	Err   error
}

// filter filters out the actual price data and errors from ping messages.
func filter(done <-chan struct{}, in <-chan interface{}, errors <-chan error, upstream chan interface{}) <-chan StreamData {
	output := make(chan StreamData)
	go func() {
		defer close(output)
		// although the filter stage doesn't own the upstream channel, but we're closing it here since this is the
		// last place where we're writing in it.
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
					case output <- StreamData{Err: fmt.Errorf("shrimpy data source filter: unknown data type: %v", parsed)}:
					}
				}
			}
		}
	}()

	return output
}
