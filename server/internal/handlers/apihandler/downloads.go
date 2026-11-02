package apihandler

import (
	"cmp"
	"errors"
	"simplemavenrepo/internal/badge"
	"simplemavenrepo/internal/repository"
	"slices"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	defaultDownloadDays = 30
	defaultTopArtifacts = 10
	maxTopArtifacts     = 100
)

var downloadDays = []int{7, 30, 90, 365}

// downloadPeriod parses the days query parameter into the first day of the period, today being its last.
func downloadPeriod(c fiber.Ctx) (days int, from, to time.Time, err error) {
	days = fiber.Query[int](c, "days", defaultDownloadDays)
	if !slices.Contains(downloadDays, days) {
		return 0, time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "days must be 7, 30, 90 or 365")
	}
	to = repository.Day(time.Now())
	return days, to.AddDate(0, 0, 1-days), to, nil
}

func (h *handler) getArtifactDownloads(c fiber.Ctx) error {
	a, err := h.artifactParam(c)
	if err != nil {
		return err
	}
	days, from, to, err := downloadPeriod(c)
	if err != nil {
		return err
	}

	ctx := c.Context()
	total, err := h.repos.Downloads.Total(ctx, a.Path, time.Time{})
	if err != nil {
		return repoError(err, "downloads")
	}
	daily, err := h.repos.Downloads.Daily(ctx, a.Path, from, to)
	if err != nil {
		return repoError(err, "downloads")
	}
	versions, err := h.repos.Downloads.ByVersion(ctx, a.Path, from)
	if err != nil {
		return repoError(err, "downloads")
	}
	return c.JSON(fiber.Map{
		"total":       total,
		"periodTotal": sumDownloads(daily),
		"days":        days,
		"daily":       daily,
		"versions":    versions,
	})
}

// artifactDownloads is an artifact with its downloads in a period.
type artifactDownloads struct {
	repository.Artifact
	Downloads int64 `json:"downloads"`
}

func (h *handler) getDownloadStats(c fiber.Ctx) error {
	days, from, to, err := downloadPeriod(c)
	if err != nil {
		return err
	}
	limit := fiber.Query[int](c, "limit", defaultTopArtifacts)
	if limit < 1 || limit > maxTopArtifacts {
		return fiber.NewError(fiber.StatusBadRequest, "invalid limit")
	}

	ctx := c.Context()
	total, err := h.repos.Downloads.Total(ctx, "", time.Time{})
	if err != nil {
		return repoError(err, "downloads")
	}
	daily, err := h.repos.Downloads.Daily(ctx, "", from, to)
	if err != nil {
		return repoError(err, "downloads")
	}
	byPath, err := h.repos.Downloads.ByArtifact(ctx, from)
	if err != nil {
		return repoError(err, "downloads")
	}
	artifacts, err := h.repos.Artifacts.List(ctx)
	if err != nil {
		return repoError(err, "artifact")
	}
	registered := make(map[string]repository.Artifact, len(artifacts))
	for _, a := range artifacts {
		registered[a.Path] = a
	}

	// Downloads of artifacts that are no longer registered still count towards the totals, but aren't listed
	top := []artifactDownloads{}
	for _, p := range byPath {
		if a, ok := registered[p.ArtifactPath]; ok && len(top) < limit {
			top = append(top, artifactDownloads{Artifact: a, Downloads: p.Count})
		}
	}
	return c.JSON(fiber.Map{
		"total":        total,
		"periodTotal":  sumDownloads(daily),
		"days":         days,
		"daily":        daily,
		"topArtifacts": top,
	})
}

func sumDownloads(daily []repository.DayDownloads) int64 {
	var sum int64
	for _, d := range daily {
		sum += d.Count
	}
	return sum
}

// badgePeriods maps the period query parameter of the downloads badge to its length in days, 0 for all time, and the
// suffix of the count.
var badgePeriods = map[string]struct {
	days   int
	suffix string
}{
	"week":  {7, "/week"},
	"month": {30, "/month"},
	"year":  {365, "/year"},
	"all":   {0, ""},
}

func (h *handler) getDownloadsBadge(c fiber.Ctx) error {
	label := cmp.Or(c.Query("label"), "downloads")
	if len(label) > maxBadgeLabel {
		label = label[:maxBadgeLabel]
	}
	period, ok := badgePeriods[cmp.Or(c.Query("period"), "all")]
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "period must be week, month, year or all")
	}

	artifact, err := h.repos.Artifacts.Get(c.Context(), c.Params("groupId"), c.Params("artifactId"))
	if errors.Is(err, repository.ErrNotFound) {
		return sendBadge(c, label, "not found", badge.ColorGrey, false)
	}
	if err != nil {
		return repoError(err, "artifact")
	}
	canRead, err := h.readFilter(c)
	if err != nil {
		return err
	}
	if !canRead(artifact) {
		return sendBadge(c, label, "not found", badge.ColorGrey, false)
	}

	var since time.Time
	if period.days > 0 {
		since = repository.Day(time.Now()).AddDate(0, 0, 1-period.days)
	}
	n, err := h.repos.Downloads.Total(c.Context(), artifact.Path, since)
	if err != nil {
		return repoError(err, "downloads")
	}
	return sendBadge(c, label, badge.FormatCount(n)+period.suffix, badge.ColorBlue, artifact.Public)
}
