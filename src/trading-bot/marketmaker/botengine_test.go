package marketmaker

import (
	"github.com/bitx/bitx-go"
	"testing"
)

func TestIsYesString(t *testing.T) {
	input := map[string]bool{
		"":     true,
		"y":    true,
		"Y":    true,
		"yes":  true,
		"Yes":  true,
		"YES":  true,
		"yebo": true,
		"\n":   true,

		"no":     false,
		"NO":     false,
		"random": false,
	}

	for text, expected := range input {
		if expected != isYesString(text) {
			t.Errorf("Expected %t, got: %t for '%s'", expected, isYesString(text), text)
		}
	}
}

func TestShouldPlaceNextOrder(t *testing.T) {
	sNotComplete := marketState{
		lastOrder: &bitx.Order{State: bitx.Pending},
	}
	if shouldPlaceNextOrder(sNotComplete) {
		t.Errorf("Expected not to place next order for Pending lastOrder")
	}

	sComplete := marketState{
		lastOrder: &bitx.Order{State: bitx.Complete},
	}
	if !shouldPlaceNextOrder(sComplete) {
		t.Errorf("Expected to place next order for Complete lastOrder")
	}
}

func TestGetNextOrderParamsForAsk(t *testing.T) {
	orderType, price := getNextOrderParams(marketState{
		bid:       100,
		lastOrder: &bitx.Order{Type: bitx.ASK},
	})
	if orderType != bitx.BID {
		t.Errorf("Expected OrderType of BID, got %s", orderType)
	}
	if price != 101 {
		t.Errorf("Expected price of 101, got %f", price)
	}
}

func TestGetNextOrderParamsForBid(t *testing.T) {
	orderType, price := getNextOrderParams(marketState{
		ask:       100,
		lastOrder: &bitx.Order{Type: bitx.BID},
	})
	if orderType != bitx.ASK {
		t.Errorf("Expected OrderType of ASK, got %s", orderType)
	}
	if price != 99 {
		t.Errorf("Expected price of 99, got %f", price)
	}
}
