package maven

import (
	"strings"
)

// CompareVersions orders versions like Maven's ComparableVersion
// It returns -1, 0 or 1.
func CompareVersions(a, b string) int {
	ia, ib := parseVersion(a), parseVersion(b)
	for i := 0; i < max(len(ia), len(ib)); i++ {
		var x, y *versionItem
		if i < len(ia) {
			x = &ia[i]
		}
		if i < len(ib) {
			y = &ib[i]
		}
		if c := compareItems(x, y); c != 0 {
			return c
		}
	}
	return 0
}

type versionItem struct {
	numeric bool
	value   string
}

var qualifierAliases = map[string]string{
	"ga":      "",
	"final":   "",
	"release": "",
	"cr":      "rc",
}

var qualifierOrder = []string{"alpha", "beta", "milestone", "rc", "snapshot", "", "sp"}

func parseVersion(v string) []versionItem {
	var items []versionItem
	v = strings.ToLower(v)

	start := 0
	flush := func(end int) {
		token := v[start:end]
		if token == "" {
			return
		}
		if token[0] >= '0' && token[0] <= '9' {
			token = strings.TrimLeft(token, "0")
			items = append(items, versionItem{numeric: true, value: token})
			return
		}
		items = append(items, versionItem{value: token})
	}

	for i := 0; i < len(v); i++ {
		c := v[i]
		if c == '.' || c == '-' {
			flush(i)
			start = i + 1
			continue
		}
		if i > start && isDigit(c) != isDigit(v[i-1]) {
			flush(i)
			start = i
		}
	}
	flush(len(v))

	for i := range items {
		if items[i].numeric {
			continue
		}
		next := i+1 < len(items) && items[i+1].numeric
		switch {
		case items[i].value == "a" && next:
			items[i].value = "alpha"
		case items[i].value == "b" && next:
			items[i].value = "beta"
		case items[i].value == "m" && next:
			items[i].value = "milestone"
		default:
			if alias, ok := qualifierAliases[items[i].value]; ok {
				items[i].value = alias
			}
		}
	}

	for len(items) > 0 && items[len(items)-1].value == "" {
		items = items[:len(items)-1]
	}
	return items
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func compareItems(x, y *versionItem) int {
	switch {
	case x == nil && y == nil:
		return 0
	case x == nil:
		return -compareItems(y, nil)
	case y == nil:
		if x.numeric {
			if x.value == "" {
				return 0
			}
			return 1
		}
		return compareQualifiers(x.value, "")
	case x.numeric && y.numeric:
		if len(x.value) != len(y.value) {
			return cmp(len(x.value), len(y.value))
		}
		return strings.Compare(x.value, y.value)
	case x.numeric:
		return 1
	case y.numeric:
		return -1
	default:
		return compareQualifiers(x.value, y.value)
	}
}

func compareQualifiers(a, b string) int {
	ra, rb := qualifierRank(a), qualifierRank(b)
	if ra != rb {
		return cmp(ra, rb)
	}
	return strings.Compare(a, b)
}

func qualifierRank(q string) int {
	for i, known := range qualifierOrder {
		if q == known {
			return i
		}
	}
	return len(qualifierOrder)
}

func cmp(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}
