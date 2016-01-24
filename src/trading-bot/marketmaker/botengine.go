package marketmaker

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bitx/bitx-go"
)

const minVolume = 0.005

type MarketMakerBot struct {
	Name      string
	apiKey    string
	apiSecret string
	pair      string
	client    *bitx.Client
}

func NewBot(apiKey, apiSecret, pair string) *MarketMakerBot {
	return &MarketMakerBot{
		Name:      "Sexy bot",
		apiKey:    apiKey,
		apiSecret: apiSecret,
		pair:      pair,
	}
}

type marketState struct {
	bid       float64
	ask       float64
	lastOrder *bitx.Order
}

func (state marketState) spread() float64 {
	return state.ask - state.bid
}

func (bot *MarketMakerBot) Execute() error {
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

	if bal <= minVolume {
		return errors.New("Insuficcient balance to place an order.")
	}

	marketState, err := getMarketState(bot.client, nil, bot.pair)
	if err != nil {
		return errors.New(fmt.Sprintf("Market not ripe: %s", err))
	}
	fmt.Printf("Current market\n\tspread: %f\n\tbid: %f\n\task: %f\n", marketState.spread(), marketState.bid, marketState.ask)

	doOrder, err := promptYesNo("Place trade?")
	if err != nil {
		return errors.New(fmt.Sprintf("Could not get user confirmation: %s", err))
	}

	var lastOrder *bitx.Order
	for doOrder {
		lastOrder, err = bot.placeNextOrder(marketState, minVolume)
		if err != nil {
			return errors.New(fmt.Sprintf("Could not place next order: %s", err))
		}

		doOrder, err = promptYesNo("Place another trade if ready?")
		if err != nil {
			return errors.New(fmt.Sprintf("Could not get user confirmation: %s", err))
		}

		marketState, err = getMarketState(bot.client, lastOrder, bot.pair)
		if err != nil {
			return errors.New(fmt.Sprintf("Market not ripe: %s", err))
		}
		fmt.Printf("Current market\n\tspread: %f\n\tbid: %f\n\task: %f\n", marketState.spread(), marketState.bid, marketState.ask)
	}

	fmt.Printf("\n%s has finished working. Bye.\n")
	return nil
}

func getMarketState(c *bitx.Client, lastOrder *bitx.Order, pair string) (state marketState, err error) {
	bids, asks, err := c.OrderBook(pair)
	if err != nil {
		return marketState{}, err
	}

	if len(bids) == 0 || len(asks) == 0 {
		return marketState{}, errors.New("Not enough liquidity on market")
	}
	state = marketState{
		bid: bids[0].Price,
		ask: asks[0].Price,
	}

	lastOrder, err = fetchOrRefreshLastOrder(c, lastOrder, pair)
	state.lastOrder = lastOrder

	return state, err
}

func fetchOrRefreshLastOrder(c *bitx.Client, lastOrder *bitx.Order, pair string) (*bitx.Order, error) {
	if lastOrder == nil {
		fmt.Println("Fetching NEW last order...")
		orders, err := c.ListOrders(pair)
		if err != nil {
			return nil, err
		}
		if len(orders) > 0 {
			// First order in this run
			return &orders[0], nil
		}
	}

	// Refresh order
	fmt.Printf("Refreshing last order (%s)...\n", lastOrder.Id)
	return c.GetOrder(lastOrder.Id)
}

func promptYesNo(question string) (yes bool, err error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [Y/n] ", question)
	text, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	return isYesString(text), nil
}

func isYesString(text string) bool {
	if text == "" {
		return true
	}
	firstChr := strings.ToLower(text)[0]
	if firstChr == 'y' || firstChr == 10 {
		return true
	}
	return false
}

func (bot *MarketMakerBot) placeNextOrder(state marketState, volume float64) (order *bitx.Order, err error) {
	fmt.Printf("Last order: %+v\n", state.lastOrder)

	// Check if last order has executed
	if !shouldPlaceNextOrder(state) {
		fmt.Println("Order has not completed yet.")
		return state.lastOrder, nil
	}

	// Time to place a new one
	orderType, price := getNextOrderParams(state)
	return bot.placeOrder(orderType, price, volume)
}

func shouldPlaceNextOrder(state marketState) bool {
	// Check if last order has executed
	return state.lastOrder.State == bitx.Complete && state.spread() > 1
}

func getNextOrderParams(state marketState) (orderType bitx.OrderType, price float64) {
	orderType = bitx.BID
	price = state.bid + 1
	if state.lastOrder != nil && state.lastOrder.Type == bitx.BID {
		orderType = bitx.ASK
		price = state.ask - 1
	}
	return orderType, price
}

func (bot *MarketMakerBot) placeOrder(orderType bitx.OrderType, price, volume float64) (*bitx.Order, error) {
	fmt.Printf("Placing order of type: %s, price: %f, volume: %f\n", orderType, price, volume)
	orderId, err := bot.client.PostOrder(bot.pair, orderType, volume, price)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Order placed! Fetching order details: %s\n", orderId)
	return bot.client.GetOrder(orderId)
}
