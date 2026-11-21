package apihandler

import (
	"cmp"
	"encoding/xml"
	"io/fs"
	"path"
	"simplemavenrepo/internal/badge"
	"simplemavenrepo/internal/maven"
	"simplemavenrepo/internal/repository"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
)

const maxBadgeLabel = 64

func (h *handler) getBadge(c fiber.Ctx) error {
	label := cmp.Or(c.Query("label"), "maven")
	if len(label) > maxBadgeLabel {
		label = label[:maxBadgeLabel]
	}

	send := func(message, color string, public bool) error {
		return sendBadge(c, label, message, color, public)
	}

	artifact, ok, err := h.badgeArtifact(c)
	if err != nil {
		return err
	}
	if !ok {
		return send("not found", badge.ColorGrey, false)
	}

	version := h.latestVersion(artifact)
	switch {
	case version == "":
		return send("no version", badge.ColorGrey, artifact.Public)
	case strings.HasSuffix(version, "-SNAPSHOT"):
		return send("v"+version, badge.ColorOrange, artifact.Public)
	default:
		return send("v"+version, badge.ColorBlue, artifact.Public)
	}
}

// policyRank orders repositories for picking the artifact a badge shows: releases first, snapshots last.
var policyRank = map[repository.VersionPolicy]int{repository.PolicyRelease: 0, repository.PolicyMixed: 1, repository.PolicySnapshot: 2}

// badgeArtifact finds the artifact of a badge request that the caller can read. ?repo=<name> picks the repository,
// otherwise the artifact is taken from a release repository if possible.
func (h *handler) badgeArtifact(c fiber.Ctx) (repository.Artifact, bool, error) {
	ctx := c.Context()
	candidates, err := h.repos.Artifacts.ListByCoordinates(ctx, c.Params("groupId"), c.Params("artifactId"))
	if err != nil {
		return repository.Artifact{}, false, repoError(err, "artifact")
	}
	if name := c.Query("repo"); name != "" {
		candidates = slices.DeleteFunc(candidates, func(a repository.Artifact) bool { return a.Repository != name })
	}
	if len(candidates) == 0 {
		return repository.Artifact{}, false, nil
	}
	canRead, err := h.readFilter(c)
	if err != nil {
		return repository.Artifact{}, false, err
	}
	candidates = slices.DeleteFunc(candidates, func(a repository.Artifact) bool { return !canRead(a) })
	if len(candidates) == 0 {
		return repository.Artifact{}, false, nil
	}

	repos, err := h.repos.Repos.List(ctx)
	if err != nil {
		return repository.Artifact{}, false, repoError(err, "repository")
	}
	rank := map[string]int{}
	for _, r := range repos {
		rank[r.Name] = policyRank[r.VersionPolicy]
	}
	// Stable, so repositories with the same policy stay ordered by name
	slices.SortStableFunc(candidates, func(x, y repository.Artifact) int { return rank[x.Repository] - rank[y.Repository] })
	return candidates[0], true, nil
}

func sendBadge(c fiber.Ctx, label, message, color string, public bool) error {
	c.Set(fiber.HeaderContentType, "image/svg+xml; charset=utf-8")
	if public {
		c.Set(fiber.HeaderCacheControl, "public, max-age=300")
	} else {
		// Whether the artifact exists depends on who is asking.
		c.Set(fiber.HeaderCacheControl, "no-store")
	}
	return c.Send(badge.Render(label, message, color))
}

func (h *handler) latestVersion(artifact repository.Artifact) string {
	data, err := fs.ReadFile(h.repoFS, path.Join(artifact.Path, maven.MetadataFileName))
	if err != nil {
		return ""
	}
	var m maven.Metadata
	if err := xml.Unmarshal(data, &m); err != nil {
		return ""
	}
	return cmp.Or(m.Versioning.Release, m.Versioning.Latest)
}
