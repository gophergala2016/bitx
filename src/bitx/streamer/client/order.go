package client

import (
	"errors"

	"bitx/streamer/streamerpb"
)

var ErrUnknownOrderType = errors.New("Unknown order type")

type OrderType int64

const (
	OrderTypeUnknown OrderType = 0
	OrderTypeBid     OrderType = 1
	OrderTypeAsk     OrderType = 2
)

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

type OrderList []*Order

func (ol OrderList) Len() int {
	return len(ol)
}

func (ol OrderList) Swap(i, j int) {
	ol[i], ol[j] = ol[j], ol[i]
}

func (ol OrderList) Less(i, j int) bool {
	return ol[i].price < ol[j].price
}
