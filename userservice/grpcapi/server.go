package grpcapi

import (
	"encoding/json"
	"fmt"
	"github.com/hedisam/CryptoXchange/models"
	"github.com/hedisam/CryptoXchange/msgbroker"
	"log"
)

type consumerFn func(exchange string) (msgbroker.MsgConsumer, error)

type server struct {
	UnimplementedPriceServiceServer
	getConsumer consumerFn
}

func NewGrpcApiServer(getConsumer consumerFn) *server {
	return &server{
		getConsumer: getConsumer,
	}
}

func (s *server) StreamPrice(req *QuoteRequest, stream PriceService_StreamPriceServer) error {
	consumer, err := s.getConsumer("price_data")
	if err != nil {
		log.Println("userservice: grpc api: failed to acquire a price consumer:", err)
		return fmt.Errorf("internal error")
	}

	// subscribing to those pairs in the sample data
	base := "bbo."
	dataChan, err := consumer.Consume(stream.Context(), base+"eth.usd", base+"btc.usd")
	if err != nil {
		log.Println("userservice: grpc api: failed to subscribe to price data", err)
		return fmt.Errorf("internal error")
	}

	for {
		data, ok := <-dataChan
		if !ok {break}

		priceData := &models.PriceData{}
		err = json.Unmarshal(data, &priceData)
		if err != nil {
			log.Println("userservice: grpc api: failed to decode data price:", err)
			return fmt.Errorf("internal error")
		}

		err = stream.Send(&QuoteReply{
			Info: &QuoteInfo{
				Pair:           priceData.Pair,
				Exchange:       priceData.Exchange,
				Snapshot:       priceData.Snapshot,
				SequenceNumber: uint64(priceData.Sequence),
			},
			AskPrice:      priceData.Asks[0].Price,
			BidPrice:      priceData.Bids[0].Price,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
