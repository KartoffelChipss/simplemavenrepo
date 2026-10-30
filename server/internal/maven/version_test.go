package maven

import "testing"

func TestCompareVersionsOrder(t *testing.T) {
	ordered := []string{
		"1.0-alpha-1",
		"1.0-alpha-2",
		"1.0-beta",
		"1.0-M1",
		"1.0-rc1",
		"1.0-SNAPSHOT",
		"1.0",
		"1.0-sp",
		"1.0-xyz",
		"1.0.1",
		"1.9",
		"1.10",
		"2.0-rc1",
		"2.0",
		"10.0",
	}
	for i := 0; i < len(ordered); i++ {
		for j := 0; j < len(ordered); j++ {
			want := cmp(i, j)
			if got := CompareVersions(ordered[i], ordered[j]); got != want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", ordered[i], ordered[j], got, want)
			}
		}
	}
}

func TestCompareVersionsEqual(t *testing.T) {
	pairs := [][2]string{
		{"1", "1.0"},
		{"1.0", "1.0.0"},
		{"1.0", "1-final"},
		{"1.0", "1.0-GA"},
		{"1.0-cr1", "1.0-rc1"},
		{"1.0-a1", "1.0-alpha-1"},
		{"1.01", "1.1"},
	}
	for _, p := range pairs {
		if got := CompareVersions(p[0], p[1]); got != 0 {
			t.Errorf("CompareVersions(%q, %q) = %d, want 0", p[0], p[1], got)
		}
	}
}
