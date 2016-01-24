package market_maker_bot

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bitx/bitx-go"
)

type MarketMakerBot struct {
	Name string
	apiKey string
	apiSecret string
	pair string
	client *bitx.Client
}

func NewBot(apiKey, apiSecret, pair string) *MarketMakerBot {
	return &MarketMakerBot{
		Name: "Sexy bot",
		apiKey: apiKey,
		apiSecret: apiSecret,
		pair: pair,
	}
}

func(bot *MarketMakerBot) Execute() error {
	fmt.Printf("%s is initialising...\n", bot.Name)

	if bot.apiKey == "" || bot.apiSecret == "" {
		return errors.New("Please supply API key and secret via command flags.")
	}

	bot.client = bitx.NewClient(bot.apiKey, bot.apiSecret)
	if bot.client == nil {
		return errors.New(fmt.Sprintf("Expected valid BitX client, got: %v", bot.client))
	}

	// Check balance
	bal, res, err := bot.client.Balance(strings.Replace(bot.pair, "XBT", "", 1))
	if err != nil {
		return errors.New(fmt.Sprintf("Error fetching balance: %s", err))
	}
	fmt.Printf("Current balance: %f (Reserved: %f)\n", bal, res)

	if (bal <= 0.005) {
		return errors.New("Insuficcient balance to place an order.")
	}

	bid, ask, spread, err := getMarketData(bot.client, bot.pair)
	if err != nil {
		return errors.New(fmt.Sprintf("Market not ripe: %s", err))
	}
	fmt.Printf("Current market\n\tspread: %f\n\tbid: %f\n\task: %f\n", spread, bid, ask)

	doOrder, err := promptYesNo("Place trade?")
	if err != nil {
		return errors.New(fmt.Sprintf("Could not get user confirmation: %s", err))
	}

	var lastOrder *bitx.Order;
	for doOrder {
		lastOrder, err = bot.placeNextOrder(lastOrder, bid, ask, spread, 0.0005)
		if err != nil {
			return errors.New(fmt.Sprintf("Could not place next order: %s", err))
		}

		doOrder, err = promptYesNo("Place another trade if ready?")
		if err != nil {
			return errors.New(fmt.Sprintf("Could not get user confirmation: %s", err))
		}

		bid, ask, spread, err = getMarketData(bot.client, bot.pair)
		if err != nil {
			return errors.New(fmt.Sprintf("Market not ripe: %s", err))
		}
		fmt.Printf("Current market\n\tspread: %f\n\tbid: %f\n\task: %f\n", spread, bid, ask)
	}

	fmt.Printf("\n%s has finished working. Bye.\n")
	return nil
}

func getMarketData(c *bitx.Client, pair string) (bid, ask, spread float64, err error) {
	bids, asks, err := c.OrderBook(pair)
	if err != nil {
		return 0, 0, 0, err
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
	if text == "" || firstChr == 'y' || firstChr == 10 {
		return true, nil
	}
	return false, nil
}

func(bot *MarketMakerBot) placeNextOrder(lastOrder *bitx.Order, bid, ask, spread, volume float64) (order *bitx.Order, err error) {
	// Fetch or refresh order
	if lastOrder == nil {
		fmt.Println("Fetching NEW last order...")
		orders, err := bot.client.ListOrders(bot.pair)
		if err != nil {
			return lastOrder, err
		}
		if len(orders) > 0 {
			// First order in this run
			lastOrder = &orders[0]
		}
	} else {
		// Refresh order
		fmt.Printf("Refreshing last order (%s)...\n", lastOrder.Id)
		lastOrder, err = bot.client.GetOrder(lastOrder.Id)
		if err != nil {
			return lastOrder, err
		}
	}

	// Check if last order has executed
	fmt.Printf("Last order: %+v\n", lastOrder)
	if lastOrder.State != bitx.Complete {
		fmt.Println("Order has not completed yet.")
		return lastOrder, nil
	}

	// Time to place a new one
	orderType := bitx.BID
	price := bid + 1;
	if lastOrder != nil && lastOrder.Type == bitx.BID {
		orderType = bitx.ASK
		price = ask - 1;
	}
	return bot.placeOrder(orderType, price, volume)
}

func(bot *MarketMakerBot) placeOrder(orderType bitx.OrderType, price, volume float64) (*bitx.Order, error) {
	fmt.Printf("Placing order of type: %s, price: %f, volume: %f\n", orderType, price, volume)
	orderId, err := bot.client.PostOrder(bot.pair, orderType, volume, price)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Order placed! Fetching order details: %s\n", orderId)
	return bot.client.GetOrder(orderId)
}