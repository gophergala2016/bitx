package client

import (
	"errors"

	"bitx/streamer/streamerpb"
)

// ErrUnknownOrderType indicates the type of order is neither bid nor ask.
var ErrUnknownOrderType = errors.New("Unknown order type")

// OrderType indicates the type of order.
type OrderType int64

// These constants should match the order types in streamerpb.proto.
const (
	OrderTypeUnknown OrderType = 0
	OrderTypeBid     OrderType = 1
	OrderTypeAsk     OrderType = 2
)

// Order is a bid or ask order.
type Order struct {
	typ    OrderType
	id     int64
	price  int64
	volume int64
}

func makeOrder(o *streamerpb.Order) *Order {
	order := &Order{
		typ:    OrderType(o.Type),
		id:     o.OrderId,
		price:  o.PriceE8,
		volume: o.VolumeE8,
	}
	return order
}

// OrderList is a list of orders that can be sort.Sort()ed.
type OrderList []*Order

// Len returns the number of orders in the list.
func (ol OrderList) Len() int {
	return len(ol)
}

// Swap swaps the orders at the given indices.
func (ol OrderList) Swap(i, j int) {
	ol[i], ol[j] = ol[j], ol[i]
}

// Less returns true if order i has a lower price than order j.
func (ol OrderList) Less(i, j int) bool {
	return ol[i].price < ol[j].price
}
