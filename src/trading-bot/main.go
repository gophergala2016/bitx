package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bitx/bitx-go"
)

var APIKey = flag.String("api_key", "", "API key")
var APISecret = flag.String("api_secret", "", "API secret")
var Pair = flag.String("currency_pair", "XBTZAR", "Currency to pair trade")

func main() {
	flag.Parse()
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
	bal, res, err := c.Balance(strings.Replace(*Pair, "XBT", "", 1))
	if err != nil {
		log.Fatalf("Error fetching balance: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Current balance: %f (Reserved: %f)\n", bal, res)

	if (bal <= 0) {
		//TODO: no funds to continue
	}

	bid, ask, spread, err := getMarketData(c)
	if err != nil {
		log.Fatalf("Market not ripe: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Current market\n\tspread: %f\n\tbid: %f\n\task: %f\n", spread, bid, ask)

	doOrder, err := promptYesNo("Place trade?")
	if err != nil {
		log.Fatalf("Could not get user confirmation: %s", err)
		os.Exit(1)
	}
	if doOrder {
		err := placeNextOrder(c, bid, ask, spread, 0.0001)
		if err != nil {
			log.Fatalf("Could not place next order: %s", err)
			os.Exit(1)
		}
	}

	fmt.Println("\nBot finished working. Bye.")
}

func getMarketData(c *bitx.Client) (bid, ask, spread float64, err error) {
	bids, asks, err := c.OrderBook(*Pair)
	if err != nil {
		log.Fatalf("Error fetching order book: %s", err)
		os.Exit(1)
	}

	if len(bids) == 0 || len(asks) == 0 {
		return 0, 0, 0, errors.New("Not enough liquidity on market")
	}
	bid = bids[0].Price
	ask = asks[0].Price
	return bid, ask, ask - bid, nil
}

func promptYesNo(question string) (yes bool, err error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [Y/n] ", question)
	text, _ := reader.ReadString('\n')

	firstChr := strings.ToLower(text)[0]
//	fmt.Printf("Text: '%s', first char: '%v'\n", text, firstChr)
	if text == "" || firstChr == 'y' || firstChr == 10 {
		return true, nil
	}
	return false, nil
}

func placeNextOrder(c *bitx.Client, bid, ask, spread, volume float64) (error) {
	fmt.Println("Fetching last order...")
	orders, err := c.ListOrders(*Pair)
	if err != nil {
		log.Fatalf("Error fetching open orders: %s", err)
		os.Exit(1)
	}

	orderType := bitx.BID
	price := bid +1;
	if len(orders) > 0 {
		// First order in this run
		if orders[0].Type == bitx.BID {
			orderType = bitx.ASK
			price = ask -1;
		}
		fmt.Printf("Last order was of type %s (%+v)\n", orders[0].Type, orders[0])
	} else {
		fmt.Println("No previous orders")
	}
	fmt.Printf("Placing %s order at %f %s...\n", orderType, price, *Pair)
	return placeOrder(c, orderType, price, volume)
}

func placeOrder(c *bitx.Client, orderType bitx.OrderType, price, volume float64) (error) {
	// TODO: place an order
	fmt.Printf("Placing order of type: %s, price: %f, volume: %f", orderType, price, volume)
	return nil
}