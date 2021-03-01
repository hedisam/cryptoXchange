package models

type Quote struct {
	Price  float64 `json:"price"`
	Volume float64 `json:"size"`
}

type PriceData struct {
	Exchange string `json:"exchange"`
	Pair     string `json:"pair"`
	Channel  string	`json:"channel"`
	Snapshot bool    `json:"snapshot"`
	Sequence int64   `json:"sequence"`
	Asks     []Quote `json:"asks"`
	Bids     []Quote `json:"bids"`
}
