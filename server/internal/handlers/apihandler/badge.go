package apihandler

import (
	"cmp"
	"encoding/xml"
	"errors"
	"io/fs"
	"path"
	"simplemavenrepo/internal/badge"
	"simplemavenrepo/internal/maven"
	"simplemavenrepo/internal/repository"
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

	artifact, err := h.repos.Artifacts.Get(c.Context(), c.Params("groupId"), c.Params("artifactId"))
	if errors.Is(err, repository.ErrNotFound) {
		return send("not found", badge.ColorGrey, false)
	}
	if err != nil {
		return repoError(err, "artifact")
	}
	canRead, err := h.readFilter(c)
	if err != nil {
		return err
	}
	if !canRead(artifact) {
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
