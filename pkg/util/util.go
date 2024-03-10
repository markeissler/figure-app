package util

import (
	"strconv"
	"strings"
)

// DigitCount returns the number of digits in the number provided.
func DigitCount(n int) int {
	return len(strconv.Itoa(n))
}

// FirstOrBlank returns
func FirstOrBlank(s ...string) string {
	out := ""
	if len(s) > 0 && len(strings.TrimSpace(s[0])) > 0 {
		out = s[0]
	}

	return out
}
