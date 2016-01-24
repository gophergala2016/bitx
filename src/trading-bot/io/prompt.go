package io

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptYesNo asks user for confirmation before continuing.
func PromptYesNo(question string) (yes bool, err error) {
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
