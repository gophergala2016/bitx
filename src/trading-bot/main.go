package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bitx/bitx-go"
)

var APIKey = flag.String("api_key", "", "API key")
var APISecret = flag.String("api_secret", "", "API secret")


func main() {
	fmt.Println("Welcome to the BitX market-making trading bot!")

	if *APIKey == "" || *APISecret == "" {
		log.Fatalf("Please supply API key and secret via command flags.")
		os.Exit(1)
	}

	c := bitx.NewClient(*APIKey, *APISecret)
	if c == nil {
		log.Fatalf("Expected valid BitX client, got: %v", c)
		os.Exit(1)
	}

	// Check balance
	bal, res, err := c.Balance("ZAR")
	if err != nil {
		log.Fatalf("Error fetching balance: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Current balance: %f (Reserved: %f)\n", bal, res)

	if (bal <= 0) {
		//TODO: no funds to continue
	}

	pair := "XBTZAR"
	bids, asks, err := c.OrderBook(pair)
	if err != nil {
		log.Fatalf("Error fetching order book: %s", err)
		os.Exit(1)
	}

	bid, ask, spread, err := getBidAsSpread(bids, asks)
	if err != nil {
		log.Fatalf("Market not ripe: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Current market\n\tspread: %f\n\tbid: %f\n\task: %f\n", spread, bid, ask)
}

func getBidAsSpread(bids, asks []bitx.OrderBookEntry) (bid, ask, spread float64, err error) {
	if len(bids) == 0 || len(asks) == 0 {
		return 0, 0, 0, errors.New("Not enough liquidity on market")
	}
	bid = bids[0].Price
	ask = asks[0].Price
	return bid, ask, ask - bid, nil
}
