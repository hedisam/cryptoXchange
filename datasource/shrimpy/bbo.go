package shrimpy

import (
	"fmt"
)

func (shrimpy *Shrimpy) BBO(done <-chan struct{}, pairs ...string) (<-chan StreamData, error) {
	subs := make([]subscription, len(pairs))
	for i, pair := range pairs {
		subs[i] = subscription{
			Type:     "subscribe",
			Exchange: exchange,
			Pair:     pair,
			Channel:  "bbo",
		}
	}

	return shrimpy.Subscribe(done, subs...)
}

func (shrimpy *Shrimpy) Subscribe(done <-chan struct{}, subscriptions ...subscription) (<-chan StreamData, error) {
	if len(subscriptions) == 0 {
		return nil, fmt.Errorf("[Subscribe] no subscriptions found")
	}

	// we use upstreamChan to send subscription requests & ping replies
	upstreamChan := make(chan interface{})

	stream, err := shrimpy.createStream(done, upstreamChan)
	if err != nil {
		return nil, fmt.Errorf("[Subscribe] couldn't create the stream: %w", err)
	}

	// send the subscriptions list
	for _, s := range subscriptions {
		select {
		case <-done: return nil, nil
		case upstreamChan <- s:
		}
	}
	pout, errors := parser(done, stream.data, stream.errors)

	output := outputer(done, pout, errors, upstreamChan)

	return output, nil
}

