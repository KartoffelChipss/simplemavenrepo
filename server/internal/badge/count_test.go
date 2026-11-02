package badge

import "testing"

func TestFormatCount(t *testing.T) {
	for n, want := range map[int64]string{
		0:             "0",
		999:           "999",
		1000:          "1k",
		1234:          "1.2k",
		9999:          "9.9k",
		12_345:        "12k",
		999_999:       "999k",
		1_500_000:     "1.5M",
		2_000_000_000: "2B",
	} {
		if got := FormatCount(n); got != want {
			t.Errorf("FormatCount(%d) = %q, want %q", n, got, want)
		}
	}
}
