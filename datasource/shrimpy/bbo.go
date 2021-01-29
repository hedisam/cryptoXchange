package shrimpy

import (
	"fmt"
	"log"
)

func (source *Shrimpy) BBOStream(done <-chan struct{}, pairs ...string) error {
	if len(pairs) == 0 {
		return fmt.Errorf("[BBOStream] no pairs found. at least one pair is needed")
	}

	cmdChan := make(chan interface{}, 3)

	stream, err := createStream(source.config, cmdChan, done)
	if err != nil {
		return fmt.Errorf("[BBOStream] couldn't create the stream: %w", err)
	}

	// request a subscription for each pair
	for _, pair := range pairs {
		msg := subscribeMessage("bbo", pair)
		ok, err := sendMessage(done, cmdChan, stream, msg, fmt.Sprintf("BBO %s subscription", pair))
		if !ok {
			return err
		}
	}

	parserChan := make(chan []byte, 20)
	defer close(parserChan)

	parserStatusChan := processStream(done, parserChan)

	for {
		select {
		case <-done: return nil
		case body := <-stream.data:
			parserChan <- body
		case err = <-stream.errors:
			return fmt.Errorf("[BBOStream] stream error: %w", err)
		case s, active := <-parserStatusChan:
			if !active {return nil} // the processor's goroutine has exited due to the done chan been closed
			ok, err := handleStatus(done, cmdChan, stream, s)
			if !ok {
				return err
			}

		}
	}
}

func handleStatus(done <-chan struct{}, cmdChan chan interface{}, stream *dataStream, s *processorStatus) (bool, error) {
	switch s.Type {
	case ping:
		pong := pingPong{Type: "pong", Data: s.pingData}
		fmt.Printf("================\n%v: %v\n================\n", "pong message", pong)
		return sendMessage(done, cmdChan, stream, pong, fmt.Sprintf("pong message: %v", pong))
	case dataError:
		return false, fmt.Errorf("[BBOStream] unable to process stream data: %w", s.err)
	default:
		log.Println("[BBOStream] unknown processor status:", s)
		return false, nil
	}
}

func sendMessage(done <-chan struct{}, cmdChan chan interface{}, stream *dataStream,
		msg interface{}, errTag string) (bool, error) {

	select {
	case <-done: return false, nil
	case cmdChan <- msg: return true, nil
	case err := <-stream.errors: return false, fmt.Errorf("[BBOStream] failed to send message: %s, err: %w", errTag, err)
	}
}

func subscribeMessage(channel, pair string) subUnsub {
	return subUnsub{
		Type:     "subscribe",
		Exchange: exchange,
		Pair:     pair,
		Channel:  channel,
	}
}