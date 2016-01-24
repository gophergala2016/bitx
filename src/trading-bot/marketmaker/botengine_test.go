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
