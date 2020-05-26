package pkg

import (
	"strconv"
	"strings"
)

func FormatFloat(v float64, precision int) string {
	s := strconv.FormatFloat(v, 'f', precision, 64)

	for len(s) > 0 && strings.Contains(s, ".") && s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}

	for len(s) > 0 && s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	return s
}
