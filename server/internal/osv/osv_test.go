package osv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const log4jRecord = `{
	"id": "GHSA-jfh8-c2jp-5v3q",
	"summary": "Remote code injection in Log4j",
	"aliases": ["CVE-2021-44228"],
	"published": "2021-12-10T00:40:56Z",
	"severity": [{"type": "CVSS_V3", "score": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H"}],
	"affected": [
		{"package": {"ecosystem": "Maven", "name": "org.apache.logging.log4j:log4j-core"},
		 "ranges": [{"type": "ECOSYSTEM", "events": [{"introduced": "2.0-beta9"}, {"fixed": "2.3.1"}]},
		            {"type": "ECOSYSTEM", "events": [{"introduced": "2.4"}, {"fixed": "2.12.2"}]},
		            {"type": "ECOSYSTEM", "events": [{"introduced": "2.13.0"}, {"fixed": "2.15.0"}]}]},
		{"package": {"ecosystem": "Maven", "name": "org.ops4j.pax.logging:pax-logging-log4j2"},
		 "ranges": [{"type": "ECOSYSTEM", "events": [{"introduced": "0"}, {"fixed": "1.11.10"}]}]}
	],
	"database_specific": {"severity": "CRITICAL"}
}`

const lowRecord = `{
	"id": "OSV-LOW",
	"details": "First line\nmore details",
	"severity": [{"type": "CVSS_V3", "score": "CVSS:3.1/AV:L/AC:H/PR:H/UI:R/S:U/C:L/I:N/A:N"}],
	"affected": [{"package": {"ecosystem": "Maven", "name": "org.apache.logging.log4j:log4j-core"}}]
}`

type fakeOSV struct {
	batches, details atomic.Int32
}

func (f *fakeOSV) handler(t *testing.T) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/querybatch", func(w http.ResponseWriter, r *http.Request) {
		f.batches.Add(1)
		var req struct{ Queries []query }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decoding batch: %v", err)
		}
		var results []map[string]any
		for _, q := range req.Queries {
			if q.Package.Ecosystem != "Maven" {
				t.Errorf("ecosystem = %q", q.Package.Ecosystem)
			}
			var vulns []map[string]string
			if q.Package.Name == "org.apache.logging.log4j:log4j-core" && q.Version == "2.14.1" {
				vulns = []map[string]string{{"id": "OSV-LOW"}}
				results = append(results, map[string]any{"vulns": vulns, "next_page_token": "page2"})
				continue
			}
			results = append(results, map[string]any{"vulns": vulns})
		}
		json.NewEncoder(w).Encode(map[string]any{"results": results})
	})
	mux.HandleFunc("POST /v1/query", func(w http.ResponseWriter, r *http.Request) {
		var q query
		json.NewDecoder(r.Body).Decode(&q)
		if q.PageToken != "page2" {
			t.Errorf("page token = %q", q.PageToken)
		}
		json.NewEncoder(w).Encode(map[string]any{"vulns": []map[string]string{{"id": "GHSA-jfh8-c2jp-5v3q"}}})
	})
	mux.HandleFunc("GET /v1/vulns/{id}", func(w http.ResponseWriter, r *http.Request) {
		f.details.Add(1)
		switch r.PathValue("id") {
		case "GHSA-jfh8-c2jp-5v3q":
			w.Write([]byte(log4jRecord))
		case "OSV-LOW":
			w.Write([]byte(lowRecord))
		default:
			http.NotFound(w, r)
		}
	})
	return mux
}

func TestCheck(t *testing.T) {
	fake := &fakeOSV{}
	srv := httptest.NewServer(fake.handler(t))
	defer srv.Close()

	c := NewClient()
	pkgs := []Package{
		{"org.apache.logging.log4j", "log4j-core", "2.14.1"},
		{"com.google.guava", "guava", "33.0.0-jre"},
	}
	found, err := c.Check(context.Background(), srv.URL+"/", pkgs)
	if err != nil {
		t.Fatal(err)
	}
	if len(found[1]) != 0 {
		t.Errorf("guava: %+v", found[1])
	}
	vulns := found[0]
	if len(vulns) != 2 {
		t.Fatalf("log4j: got %+v", vulns)
	}
	// Sorted by severity, critical first
	crit, low := vulns[0], vulns[1]
	if crit.ID != "GHSA-jfh8-c2jp-5v3q" || crit.Severity != SeverityCritical || !slices.Equal(crit.Aliases, []string{"CVE-2021-44228"}) {
		t.Errorf("critical = %+v", crit)
	}
	if !slices.Equal(crit.Fixed, []string{"2.15.0"}) {
		t.Errorf("fixed versions must be newer and only come from the queried package: %v", crit.Fixed)
	}
	if crit.URL != "https://osv.dev/vulnerability/GHSA-jfh8-c2jp-5v3q" || !strings.HasPrefix(crit.CVSS, "CVSS:3.1/") {
		t.Errorf("critical = %+v", crit)
	}
	if low.Severity != SeverityLow || low.Summary != "First line" || low.Aliases == nil || low.Fixed == nil {
		t.Errorf("low = %+v", low)
	}

	// Answers are cached
	if _, err := c.Check(context.Background(), srv.URL, pkgs); err != nil {
		t.Fatal(err)
	}
	if fake.batches.Load() != 1 || fake.details.Load() != 2 {
		t.Errorf("batches = %d, details = %d, want cached answers", fake.batches.Load(), fake.details.Load())
	}

	// Until they expire
	c.now = func() time.Time { return time.Now().Add(cacheTTL + time.Minute) }
	if _, err := c.Check(context.Background(), srv.URL, pkgs); err != nil {
		t.Fatal(err)
	}
	if fake.batches.Load() != 2 {
		t.Errorf("batches = %d after expiry", fake.batches.Load())
	}

	// Or the cache is reset
	c.Reset()
	if _, err := c.Check(context.Background(), srv.URL, pkgs); err != nil {
		t.Fatal(err)
	}
	if fake.batches.Load() != 3 {
		t.Errorf("batches = %d after reset", fake.batches.Load())
	}
}

func TestCheckNothing(t *testing.T) {
	c := NewClient()
	// No packages must not contact the server at all
	found, err := c.Check(context.Background(), "http://127.0.0.1:1", nil)
	if err != nil || len(found) != 0 {
		t.Fatalf("got %v, %v", found, err)
	}
}

func TestCheckError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	if _, err := NewClient().Check(context.Background(), srv.URL, []Package{{"a", "b", "1"}}); err == nil {
		t.Fatal("want an error")
	}
}

func TestFixedVersions(t *testing.T) {
	var r record
	if err := json.Unmarshal([]byte(log4jRecord), &r); err != nil {
		t.Fatal(err)
	}
	for version, want := range map[string][]string{
		"2.14.1": {"2.15.0"},
		"2.5":    {"2.12.2", "2.15.0"},
		"2.0":    {"2.3.1", "2.12.2", "2.15.0"},
		// No newer fix, all are kept
		"3.0": {"2.3.1", "2.12.2", "2.15.0"},
	} {
		if got := r.vulnerability("org.apache.logging.log4j:log4j-core", version).Fixed; !slices.Equal(got, want) {
			t.Errorf("%s: fixed = %v, want %v", version, got, want)
		}
	}
}

func TestCVSS3Score(t *testing.T) {
	for vector, want := range map[string]float64{
		"CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H": 10.0,
		"CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H": 9.8,
		"CVSS:3.0/AV:N/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N": 6.5,
		"CVSS:3.1/AV:N/AC:H/PR:N/UI:R/S:U/C:L/I:N/A:N": 3.1,
		"CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N": 6.1,
		"CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:N": 0,
	} {
		if got, ok := cvss3Score(vector); !ok || got != want {
			t.Errorf("%s = %v, %v, want %v", vector, got, ok, want)
		}
	}
	for _, invalid := range []string{"", "CVSS:2.0/AV:N", "CVSS:3.1/AV:N/AC:L", "CVSS:3.1/AV:X/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"} {
		if _, ok := cvss3Score(invalid); ok {
			t.Errorf("%q must be invalid", invalid)
		}
	}
}
