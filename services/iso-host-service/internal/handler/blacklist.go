package handler

import (
	"fmt"
	"strings"
	"unicode"
)

func isBlacklistedPAN(pan string) bool {
	var b strings.Builder
	for _, r := range pan {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	p := b.String()

	switch p {
	case "5555555555554444":
		return true
	case "4000000000000002":
		return true
	default:
		return false
	}
}

func errBlacklistedPAN() error {
	return fmt.Errorf("pan blacklisted")
}
