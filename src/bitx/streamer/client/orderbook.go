package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"sort"
	"sync"

	"bitx/streamer/streamerpb"
)

type OrderBook struct {
	mu       sync.Mutex
	sequence int64
	Asks     map[int64]*Order
	Bids     map[int64]*Order
}

func makeOrderBook(ob *streamerpb.OrderBook) (*OrderBook, error) {
	orderBook := &OrderBook{
		sequence: ob.Sequence,
		Asks:     make(map[int64]*Order, 0),
		Bids:     make(map[int64]*Order, 0),
	}

	for _, o := range ob.Asks {
		orderBook.Asks[o.OrderId] = makeOrder(o)
	}

	for _, o := range ob.Bids {
		orderBook.Bids[o.OrderId] = makeOrder(o)
	}

	return orderBook, nil
}

var ErrOutOfSequence = errors.New("Received out-of-sequence update")
var ErrOrderNotFound = errors.New("Order not found")

func (ob *OrderBook) handleUpdate(upd *streamerpb.Update) error {
	if upd.Sequence < ob.sequence {
		log.Printf("bitx/streamer/client.OrderBook.handleUpdate(): Ignoring "+
			"update with lower sequence number (order book = %d, update = %d)",
			ob.sequence, upd.Sequence)
		return nil
	}
	if upd.Sequence != ob.sequence+1 {
		log.Printf("bitx/streamer/client.OrderBook.handleUpdate(): Out of "+
			"sequence update (order book = %d, update = %d)",
			ob.sequence, upd.Sequence)
		return ErrOutOfSequence
	}
	ob.sequence = upd.Sequence

	log.Printf("bitx/streamer/client.OrderBook.handleUpdate(): Received "+
		"update: %#v", upd)

	// Process trades
	if tradesRequest := upd.GetTradeUpdate(); tradesRequest != nil &&
		len(tradesRequest) > 0 {
		for _, t := range tradesRequest {
			if err := ob.handleTrade(t); err != nil {
				return err
			}
		}
	}

	// Process create
	if createRequest := upd.GetCreateUpdate(); createRequest != nil {
		if err := ob.addOrder(createRequest); err != nil {
			return err
		}
	}

	// Process delete
	if deleteRequest := upd.GetDeleteUpdate(); deleteRequest != nil {
		ob.removeOrder(deleteRequest.OrderId)
	}

	return nil
}

func (ob *OrderBook) handleTrade(t *streamerpb.TradeUpdate) error {
	log.Printf("bitx/streamer/client.OrderBook.handleTrade(): Received trade "+
		"%#v", t)

	o, ok := ob.Asks[t.OrderId]
	if !ok {
		o, ok = ob.Bids[t.OrderId]
		if !ok {
			log.Printf("bitx/streamer/client.OrderBook.handleTrade(): Order "+
				"%d not found", t.OrderId)
			return ErrOrderNotFound
		}
	}

	o.volume = o.volume - t.BaseE8
	if o.volume <= 0 {
		ob.removeOrder(t.OrderId)
	}

	return nil
}

func (ob *OrderBook) addOrder(o *streamerpb.CreateUpdate) error {
	order := makeOrder(o.Order)

	switch order.typ {
	case OrderTypeAsk:
		ob.Asks[order.id] = order
	case OrderTypeBid:
		ob.Bids[order.id] = order
	default:
		return ErrUnknownOrderType
	}

	return nil
}

func (ob *OrderBook) removeOrder(id int64) {
	delete(ob.Asks, id)
	delete(ob.Bids, id)
}

// String prints all the order book asks and bids.
func (ob *OrderBook) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "\n")
	printSorted(&buf, ob.Asks)
	fmt.Fprintf(&buf, "\n")
	printSorted(&buf, ob.Bids)

	return string(buf.Bytes())
}

func printSorted(w io.Writer, orders map[int64]*Order) {
	orderList := make(OrderList, 0)
	for _, o := range orders {
		orderList = append(orderList, o)
	}

	sort.Sort(sort.Reverse(orderList))

	for _, o := range orderList {
		fmt.Fprintf(w, "%.2f %.6f\n", float64(o.price)/1e8,
			float64(o.volume)/1e8)
	}
}
