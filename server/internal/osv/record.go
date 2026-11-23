package osv

import (
	"math"
	"simplemavenrepo/internal/maven"
	"slices"
	"strings"
)

// record is the part of the OSV schema (https://ossf.github.io/osv-schema/) this package reads.
type record struct {
	ID        string   `json:"id"`
	Summary   string   `json:"summary"`
	Details   string   `json:"details"`
	Aliases   []string `json:"aliases"`
	Published string   `json:"published"`
	Severity  []struct {
		Type  string `json:"type"`
		Score string `json:"score"`
	} `json:"severity"`
	Affected []struct {
		Package struct {
			Ecosystem string `json:"ecosystem"`
			Name      string `json:"name"`
		} `json:"package"`
		Ranges []struct {
			Events []struct {
				Fixed string `json:"fixed"`
			} `json:"events"`
		} `json:"ranges"`
	} `json:"affected"`
	DatabaseSpecific struct {
		// Set by GitHub advisories: LOW, MODERATE, HIGH or CRITICAL
		Severity string `json:"severity"`
	} `json:"database_specific"`
}

// vulnerability converts the record for the package and version it was found in.
func (r record) vulnerability(packageName, version string) Vulnerability {
	v := Vulnerability{
		ID:        r.ID,
		Summary:   r.Summary,
		Aliases:   slices.Clone(r.Aliases),
		Severity:  SeverityUnknown,
		Published: r.Published,
		Fixed:     []string{},
		URL:       "https://osv.dev/vulnerability/" + r.ID,
	}
	if v.Aliases == nil {
		v.Aliases = []string{}
	}
	if v.Summary == "" {
		v.Summary = firstLine(r.Details)
	}

	for _, s := range r.Severity {
		if strings.HasPrefix(s.Type, "CVSS_V") && v.CVSS == "" {
			v.CVSS = s.Score
		}
	}
	switch strings.ToUpper(r.DatabaseSpecific.Severity) {
	case "CRITICAL":
		v.Severity = SeverityCritical
	case "HIGH":
		v.Severity = SeverityHigh
	case "MODERATE", "MEDIUM":
		v.Severity = SeverityMedium
	case "LOW":
		v.Severity = SeverityLow
	default:
		for _, s := range r.Severity {
			if s.Type == "CVSS_V3" {
				if score, ok := cvss3Score(s.Score); ok {
					v.Severity = severityOfScore(score)
				}
			}
		}
	}

	for _, a := range r.Affected {
		if a.Package.Ecosystem != "Maven" || a.Package.Name != packageName {
			continue
		}
		for _, rg := range a.Ranges {
			for _, e := range rg.Events {
				if e.Fixed != "" && !slices.Contains(v.Fixed, e.Fixed) {
					v.Fixed = append(v.Fixed, e.Fixed)
				}
			}
		}
	}
	// Fixes of older release lines are no upgrade path
	if newer := slices.DeleteFunc(slices.Clone(v.Fixed), func(f string) bool {
		return maven.CompareVersions(f, version) <= 0
	}); len(newer) > 0 {
		v.Fixed = newer
	}
	slices.SortFunc(v.Fixed, maven.CompareVersions)
	return v
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 200 {
		s = strings.ToValidUTF8(s[:200], "") + "…"
	}
	return s
}

func severityRank(s Severity) int {
	switch s {
	case SeverityCritical:
		return 0
	case SeverityHigh:
		return 1
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 3
	}
	return 4
}

func severityOfScore(score float64) Severity {
	switch {
	case score >= 9:
		return SeverityCritical
	case score >= 7:
		return SeverityHigh
	case score >= 4:
		return SeverityMedium
	case score > 0:
		return SeverityLow
	}
	return SeverityUnknown
}

// cvss3Score computes the base score of a CVSS v3.x vector like CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H.
func cvss3Score(vector string) (float64, bool) {
	parts := strings.Split(vector, "/")
	if len(parts) == 0 || !strings.HasPrefix(parts[0], "CVSS:3.") {
		return 0, false
	}
	m := map[string]string{}
	for _, p := range parts[1:] {
		k, v, ok := strings.Cut(p, ":")
		if !ok {
			return 0, false
		}
		m[k] = v
	}
	weight := func(metric string, weights map[string]float64) (float64, bool) {
		w, ok := weights[m[metric]]
		return w, ok
	}
	changed := m["S"] == "C"
	if m["S"] != "C" && m["S"] != "U" {
		return 0, false
	}
	prWeights := map[string]float64{"N": 0.85, "L": 0.62, "H": 0.27}
	if changed {
		prWeights = map[string]float64{"N": 0.85, "L": 0.68, "H": 0.5}
	}
	cia := map[string]float64{"H": 0.56, "L": 0.22, "N": 0}
	av, ok1 := weight("AV", map[string]float64{"N": 0.85, "A": 0.62, "L": 0.55, "P": 0.2})
	ac, ok2 := weight("AC", map[string]float64{"L": 0.77, "H": 0.44})
	pr, ok3 := weight("PR", prWeights)
	ui, ok4 := weight("UI", map[string]float64{"N": 0.85, "R": 0.62})
	c, ok5 := weight("C", cia)
	i, ok6 := weight("I", cia)
	a, ok7 := weight("A", cia)
	if !(ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7) {
		return 0, false
	}

	iss := 1 - (1-c)*(1-i)*(1-a)
	impact := 6.42 * iss
	if changed {
		impact = 7.52*(iss-0.029) - 3.25*math.Pow(iss-0.02, 15)
	}
	if impact <= 0 {
		return 0, true
	}
	exploitability := 8.22 * av * ac * pr * ui
	if changed {
		return roundUp(math.Min(1.08*(impact+exploitability), 10)), true
	}
	return roundUp(math.Min(impact+exploitability, 10)), true
}

// roundUp is the Roundup function of the CVSS v3.1 specification.
func roundUp(x float64) float64 {
	n := int64(math.Round(x * 100000))
	if n%10000 == 0 {
		return float64(n) / 100000
	}
	return float64(n/10000+1) / 10
}
