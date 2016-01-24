package marketmaker

import "testing"

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
