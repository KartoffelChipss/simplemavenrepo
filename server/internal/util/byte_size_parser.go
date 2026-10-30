package util

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var sizeUnits = map[string]float64{
	"":  1,
	"b": 1,
	"k": 1 << 10, "kb": 1 << 10, "kib": 1 << 10,
	"m": 1 << 20, "mb": 1 << 20, "mib": 1 << 20,
	"g": 1 << 30, "gb": 1 << 30, "gib": 1 << 30,
	"t": 1 << 40, "tb": 1 << 40, "tib": 1 << 40,
}

// ParseByteSize converts strings like "512MB", "1.5 GiB" or "4096" into bytes.
func ParseByteSize(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}

	i := 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '.') {
		i++
	}
	numStr, unitStr := s[:i], strings.ToLower(strings.TrimSpace(s[i:]))

	if numStr == "" {
		return 0, fmt.Errorf("invalid size %q: missing number", s)
	}
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q: %w", s, err)
	}

	mult, ok := sizeUnits[unitStr]
	if !ok {
		return 0, fmt.Errorf("invalid size %q: unknown unit %q", s, unitStr)
	}

	bytes := num * mult
	if bytes > math.MaxInt {
		return 0, fmt.Errorf("size %q is too large", s)
	}
	return int(bytes), nil
}
