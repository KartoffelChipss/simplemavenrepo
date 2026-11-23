package apihandler

import "testing"

func TestConcreteVersion(t *testing.T) {
	for v, want := range map[string]bool{
		"1.0":              true,
		"2.14.1":           true,
		"33.0.0-jre":       true,
		"1.0-SNAPSHOT":     true,
		"":                 false,
		"${log4j.version}": false,
		"[1.0,2.0)":        false,
		"[1.5]":            false,
	} {
		if got := concreteVersion(v); got != want {
			t.Errorf("concreteVersion(%q) = %v, want %v", v, got, want)
		}
	}
}

func TestValidateOSVURL(t *testing.T) {
	for u, valid := range map[string]bool{
		"https://api.osv.dev":           true,
		"http://osv.internal:8080/api":  true,
		"ftp://api.osv.dev":             false,
		"https://user:pass@api.osv.dev": false,
		"https://api.osv.dev?x=1":       false,
		"https://api.osv.dev#frag":      false,
		"api.osv.dev":                   false,
		"https://":                      false,
	} {
		if err := validateOSVURL(u); (err == nil) != valid {
			t.Errorf("validateOSVURL(%q) = %v, want valid = %v", u, err, valid)
		}
	}
}
