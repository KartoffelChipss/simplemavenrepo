// Package osv looks up known vulnerabilities of Maven packages in an OSV database (https://osv.dev).
package osv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"
)

// DefaultURL is the public OSV API, used when no other URL is configured.
const DefaultURL = "https://api.osv.dev"

const (
	cacheTTL        = 24 * time.Hour
	maxCacheEntries = 20000
	// OSV accepts up to 1000 queries per batch
	maxBatchSize      = 1000
	maxPages          = 10
	maxDetailRequests = 8
	maxResponseSize   = 16 << 20
)

// Package is a Maven package version to check.
type Package struct {
	GroupID    string
	ArtifactID string
	Version    string
}

func (p Package) name() string { return p.GroupID + ":" + p.ArtifactID }

// Severity is a normalized severity level.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityUnknown  Severity = "unknown"
)

// Vulnerability is the part of an OSV record the UI shows.
type Vulnerability struct {
	ID        string   `json:"id"`
	Summary   string   `json:"summary,omitempty"`
	Aliases   []string `json:"aliases"`
	Severity  Severity `json:"severity"`
	CVSS      string   `json:"cvss,omitempty"`
	Published string   `json:"published,omitempty"`
	// Versions of the package that fix it, empty when none is known.
	Fixed []string `json:"fixed"`
	URL   string   `json:"url"`
}

type cacheEntry[T any] struct {
	value   T
	expires time.Time
}

// Client queries an OSV API and caches the answers for a day, since advisories change over time.
type Client struct {
	http *http.Client
	now  func() time.Time

	mu      sync.Mutex
	queries map[string]cacheEntry[[]string] // base URL, package and version -> vulnerability IDs
	vulns   map[string]cacheEntry[record]   // base URL and vulnerability ID -> record
}

func NewClient() *Client {
	return &Client{
		http:    &http.Client{Timeout: 20 * time.Second},
		now:     time.Now,
		queries: map[string]cacheEntry[[]string]{},
		vulns:   map[string]cacheEntry[record]{},
	}
}

// Reset drops all cached answers, e.g. after the URL changed.
func (c *Client) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	clear(c.queries)
	clear(c.vulns)
}

// Check returns the known vulnerabilities of each package, keyed by the index of the package in pkgs. baseURL is the
// OSV API to ask, DefaultURL when empty.
func (c *Client) Check(ctx context.Context, baseURL string, pkgs []Package) (map[int][]Vulnerability, error) {
	if baseURL == "" {
		baseURL = DefaultURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	ids := make([][]string, len(pkgs))
	var missing []int
	c.mu.Lock()
	for i, p := range pkgs {
		if e, ok := c.queries[queryKey(baseURL, p)]; ok && c.now().Before(e.expires) {
			ids[i] = e.value
		} else {
			missing = append(missing, i)
		}
	}
	c.mu.Unlock()

	for chunk := range slices.Chunk(missing, maxBatchSize) {
		batch := make([]Package, len(chunk))
		for j, i := range chunk {
			batch[j] = pkgs[i]
		}
		found, err := c.queryBatch(ctx, baseURL, batch)
		if err != nil {
			return nil, err
		}
		c.mu.Lock()
		c.pruneLocked()
		for j, i := range chunk {
			ids[i] = found[j]
			c.queries[queryKey(baseURL, pkgs[i])] = cacheEntry[[]string]{found[j], c.now().Add(cacheTTL)}
		}
		c.mu.Unlock()
	}

	var allIDs []string
	for _, list := range ids {
		allIDs = append(allIDs, list...)
	}
	slices.Sort(allIDs)
	records, err := c.fetchRecords(ctx, baseURL, slices.Compact(allIDs))
	if err != nil {
		return nil, err
	}

	result := map[int][]Vulnerability{}
	for i, list := range ids {
		for _, id := range list {
			if r, ok := records[id]; ok {
				result[i] = append(result[i], r.vulnerability(pkgs[i].name(), pkgs[i].Version))
			}
		}
		slices.SortStableFunc(result[i], func(a, b Vulnerability) int {
			return severityRank(a.Severity) - severityRank(b.Severity)
		})
	}
	return result, nil
}

func queryKey(baseURL string, p Package) string {
	return baseURL + "\x00" + p.name() + "\x00" + p.Version
}

// pruneLocked drops expired entries once the cache grew large, and everything if that is not enough.
func (c *Client) pruneLocked() {
	if len(c.queries)+len(c.vulns) < maxCacheEntries {
		return
	}
	now := c.now()
	for k, e := range c.queries {
		if !now.Before(e.expires) {
			delete(c.queries, k)
		}
	}
	for k, e := range c.vulns {
		if !now.Before(e.expires) {
			delete(c.vulns, k)
		}
	}
	if len(c.queries)+len(c.vulns) >= maxCacheEntries {
		clear(c.queries)
		clear(c.vulns)
	}
}

type query struct {
	Package struct {
		Ecosystem string `json:"ecosystem"`
		Name      string `json:"name"`
	} `json:"package"`
	Version   string `json:"version"`
	PageToken string `json:"page_token,omitempty"`
}

type queryResult struct {
	Vulns []struct {
		ID string `json:"id"`
	} `json:"vulns"`
	NextPageToken string `json:"next_page_token"`
}

func newQuery(p Package) query {
	var q query
	q.Package.Ecosystem = "Maven"
	q.Package.Name = p.name()
	q.Version = p.Version
	return q
}

// queryBatch returns the IDs of the vulnerabilities affecting each package. Results that OSV splits into pages are
// completed with single queries.
func (c *Client) queryBatch(ctx context.Context, baseURL string, pkgs []Package) ([][]string, error) {
	queries := make([]query, len(pkgs))
	for i, p := range pkgs {
		queries[i] = newQuery(p)
	}
	var res struct {
		Results []queryResult `json:"results"`
	}
	if err := c.post(ctx, baseURL+"/v1/querybatch", map[string]any{"queries": queries}, &res); err != nil {
		return nil, err
	}
	if len(res.Results) != len(pkgs) {
		return nil, fmt.Errorf("osv: got %d results for %d queries", len(res.Results), len(pkgs))
	}

	ids := make([][]string, len(pkgs))
	for i, r := range res.Results {
		ids[i] = []string{}
		for page := 0; ; page++ {
			for _, v := range r.Vulns {
				ids[i] = append(ids[i], v.ID)
			}
			if r.NextPageToken == "" || page >= maxPages {
				break
			}
			q := queries[i]
			q.PageToken = r.NextPageToken
			r = queryResult{}
			if err := c.post(ctx, baseURL+"/v1/query", q, &r); err != nil {
				return nil, err
			}
		}
		slices.Sort(ids[i])
		ids[i] = slices.Compact(ids[i])
	}
	return ids, nil
}

func (c *Client) fetchRecords(ctx context.Context, baseURL string, ids []string) (map[string]record, error) {
	result := map[string]record{}
	var missing []string
	c.mu.Lock()
	for _, id := range ids {
		if e, ok := c.vulns[baseURL+"\x00"+id]; ok && c.now().Before(e.expires) {
			result[id] = e.value
		} else {
			missing = append(missing, id)
		}
	}
	c.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		firstErr error
		sem      = make(chan struct{}, maxDetailRequests)
	)
	for _, id := range missing {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			var r record
			err := c.get(ctx, baseURL+"/v1/vulns/"+url.PathEscape(id), &r)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
					cancel()
				}
				return
			}
			result[id] = r
		})
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	c.mu.Lock()
	c.pruneLocked()
	for _, id := range missing {
		c.vulns[baseURL+"\x00"+id] = cacheEntry[record]{result[id], c.now().Add(cacheTTL)}
	}
	c.mu.Unlock()
	return result, nil
}

func (c *Client) post(ctx context.Context, u string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *Client) get(ctx context.Context, u string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "simplemavenrepo")
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("osv: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("osv: %s %s returned %s", req.Method, req.URL.Path, res.Status)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseSize)).Decode(out); err != nil {
		return fmt.Errorf("osv: decoding %s: %w", req.URL.Path, err)
	}
	return nil
}
