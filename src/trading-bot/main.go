package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"trading-bot/market_maker_bot"
)

var APIKey = flag.String("api_key", "", "API key")
var APISecret = flag.String("api_secret", "", "API secret")
var Pair = flag.String("currency_pair", "XBTZAR", "Currency to pair trade")

func main() {
	flag.Parse()
	fmt.Println("Welcome to the BitX trading bot playground!\n")

	bot := market_maker_bot.NewBot(*APIKey, *APISecret, *Pair)
	err := bot.Execute()

	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	fmt.Println("Bot execution completed.")
}