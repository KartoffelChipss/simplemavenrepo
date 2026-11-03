package badge

import (
	"strconv"
	"strings"
)

// FormatCount shortens n like 999, 1.2k, 12k, 3.4M, the way download badges usually show counts.
func FormatCount(n int64) string {
	for _, u := range []struct {
		size   int64
		suffix string
	}{{1_000_000_000, "B"}, {1_000_000, "M"}, {1_000, "k"}} {
		if n < u.size {
			continue
		}
		// Rounded down so 999,999 doesn't show as 1000k. One decimal below 10, without a trailing .0
		if n < 10*u.size {
			tenths := n * 10 / u.size
			return strings.TrimSuffix(strconv.FormatInt(tenths/10, 10)+"."+strconv.FormatInt(tenths%10, 10), ".0") + u.suffix
		}
		return strconv.FormatInt(n/u.size, 10) + u.suffix
	}
	return strconv.FormatInt(n, 10)
}
