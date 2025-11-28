package services

import (
	"fmt"
	"strconv"
	"strings"
)

func F(v float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.4f", v), "0"), ".")
}

func IntComma(v int64) string {
	s := strconv.FormatInt(v, 10)
	n := len(s)
	if n <= 3 {
		return s
	}
	var parts []string
	for n > 3 {
		parts = append([]string{s[n-3 : n]}, parts...)
		n -= 3
	}
	parts = append([]string{s[:n]}, parts...)
	return strings.Join(parts, ",")
}
