package marketmaker

import (
	"github.com/bitx/bitx-go"
	"testing"
)

func TestShouldPlaceNextOrderPending(t *testing.T) {
	if shouldPlaceNextOrder(marketState{
		bid:       100,
		ask:       110,
		lastOrder: &bitx.Order{State: bitx.Pending},
	}) {
		t.Errorf("Expected not to place next order for Pending lastOrder and decent spread.")
	}
}

func TestShouldPlaceNextOrderComplete(t *testing.T) {
	if !shouldPlaceNextOrder(marketState{
		bid:       100,
		ask:       110,
		lastOrder: &bitx.Order{State: bitx.Complete},
	}) {
		t.Errorf("Expected to place next order for Complete lastOrder and decent spread.")
	}

	if shouldPlaceNextOrder(marketState{
		bid:       100,
		ask:       101,
		lastOrder: &bitx.Order{State: bitx.Complete},
	}) {
		t.Errorf("Expected to not place next order for Complete lastOrder and spread of 1.")
	}
}

func TestGetNextOrderParamsForAsk(t *testing.T) {
	orderType, price := getNextOrderParams(marketState{
		bid:       100,
		lastOrder: &bitx.Order{Type: bitx.ASK},
	})
	if orderType != bitx.BID {
		t.Errorf("Expected OrderType of BID, got %s.", orderType)
	}
	if price != 101 {
		t.Errorf("Expected price of 101, got %f.", price)
	}
}

func TestGetNextOrderParamsForBid(t *testing.T) {
	orderType, price := getNextOrderParams(marketState{
		ask:       100,
		lastOrder: &bitx.Order{Type: bitx.BID},
	})
	if orderType != bitx.ASK {
		t.Errorf("Expected OrderType of ASK, got %s.", orderType)
	}
	if price != 99 {
		t.Errorf("Expected price of 99, got %f.", price)
	}
}
