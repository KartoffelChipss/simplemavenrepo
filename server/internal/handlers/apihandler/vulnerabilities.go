package apihandler

import (
	"log"
	"simplemavenrepo/internal/osv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

type checkedDependency struct {
	GroupID         string              `json:"groupId"`
	ArtifactID      string              `json:"artifactId"`
	Version         string              `json:"version"`
	Vulnerabilities []osv.Vulnerability `json:"vulnerabilities"`
}

type skippedDependency struct {
	GroupID    string `json:"groupId"`
	ArtifactID string `json:"artifactId"`
	Version    string `json:"version,omitempty"`
	// hosted: the artifact is on this server, OSV doesn't know it and its coordinates must not leak.
	// unresolved: the version is missing, a range or contains a placeholder.
	Reason string `json:"reason"`
}

type vulnerabilitiesResponse struct {
	Dependencies []checkedDependency `json:"dependencies"`
	Skipped      []skippedDependency `json:"skipped"`
	CheckedAt    time.Time           `json:"checkedAt"`
}

// getVulnerabilities checks the direct dependencies of a version against OSV. The path is that of getPom.
func (h *handler) getVulnerabilities(c fiber.Ctx) error {
	settings, err := h.repos.Settings.Get(c.Context())
	if err != nil {
		return repoError(err, "settings")
	}
	if !settings.OSVEnabled {
		return fiber.NewError(fiber.StatusNotFound, "vulnerability checks are disabled")
	}
	baseURL := osv.DefaultURL
	if settings.OSVURL != nil {
		baseURL = *settings.OSVURL
	}

	project, registered, err := h.resolvePom(c)
	if err != nil {
		return err
	}

	res := vulnerabilitiesResponse{Dependencies: []checkedDependency{}, Skipped: []skippedDependency{}}
	var pkgs []osv.Package
	seen := map[osv.Package]bool{}
	for _, d := range project.Dependencies {
		pkg := osv.Package{GroupID: d.GroupID, ArtifactID: d.ArtifactID, Version: d.Version}
		if seen[pkg] {
			continue
		}
		seen[pkg] = true
		switch {
		case registered[d.GroupID+":"+d.ArtifactID]:
			res.Skipped = append(res.Skipped, skippedDependency{d.GroupID, d.ArtifactID, d.Version, "hosted"})
		case !concreteVersion(d.Version):
			res.Skipped = append(res.Skipped, skippedDependency{d.GroupID, d.ArtifactID, d.Version, "unresolved"})
		default:
			pkgs = append(pkgs, pkg)
		}
	}

	found, err := h.osv.Check(c.Context(), baseURL, pkgs)
	if err != nil {
		log.Printf("Checking vulnerabilities failed: %v", err)
		return fiber.NewError(fiber.StatusBadGateway, "the vulnerability database can't be reached")
	}
	for i, p := range pkgs {
		vulns := found[i]
		if vulns == nil {
			vulns = []osv.Vulnerability{}
		}
		res.Dependencies = append(res.Dependencies, checkedDependency{p.GroupID, p.ArtifactID, p.Version, vulns})
	}
	res.CheckedAt = time.Now().UTC()
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(res)
}

// concreteVersion reports whether v names a single version OSV can match, not a range or an unexpanded property.
func concreteVersion(v string) bool {
	return v != "" && !strings.ContainsAny(v, "$[](),")
}
