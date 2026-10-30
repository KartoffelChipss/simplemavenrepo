package maven

import (
	"encoding/xml"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const MetadataFileName = "maven-metadata.xml"

// Metadata is the root of a maven-metadata.xml file, both on artifact and on SNAPSHOT version level.
type Metadata struct {
	XMLName      xml.Name   `xml:"metadata"`
	ModelVersion string     `xml:"modelVersion,attr,omitempty"`
	GroupID      string     `xml:"groupId"`
	ArtifactID   string     `xml:"artifactId"`
	Version      string     `xml:"version,omitempty"`
	Versioning   Versioning `xml:"versioning"`
}

type Versioning struct {
	Latest           string            `xml:"latest,omitempty"`
	Release          string            `xml:"release,omitempty"`
	Snapshot         *Snapshot         `xml:"snapshot,omitempty"`
	Versions         *Versions         `xml:"versions,omitempty"`
	LastUpdated      string            `xml:"lastUpdated"`
	SnapshotVersions *SnapshotVersions `xml:"snapshotVersions,omitempty"`
}

// Versions and SnapshotVersions are pointers so that the wrapper element is left out entirely when absent.
type Versions struct {
	Version []string `xml:"version"`
}

type SnapshotVersions struct {
	SnapshotVersion []SnapshotVersion `xml:"snapshotVersion"`
}

type Snapshot struct {
	Timestamp   string `xml:"timestamp"`
	BuildNumber int    `xml:"buildNumber"`
}

type SnapshotVersion struct {
	Classifier string `xml:"classifier,omitempty"`
	Extension  string `xml:"extension"`
	Value      string `xml:"value"`
	Updated    string `xml:"updated"`
}

func (m Metadata) Marshal() ([]byte, error) {
	out, err := xml.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), append(out, '\n')...), nil
}

func formatLastUpdated(t time.Time) string {
	return t.UTC().Format("20060102150405")
}

// NewArtifactMetadata builds the <group>/<artifactId>/maven-metadata.xml listing all versions.
func NewArtifactMetadata(groupID, artifactID string, versions []string, now time.Time) Metadata {
	sorted := slices.Clone(versions)
	slices.SortFunc(sorted, CompareVersions)

	v := Versioning{
		Versions:    &Versions{Version: sorted},
		LastUpdated: formatLastUpdated(now),
	}
	if len(sorted) > 0 {
		v.Latest = sorted[len(sorted)-1]
	}
	for i := len(sorted) - 1; i >= 0; i-- {
		if !strings.HasSuffix(sorted[i], "-SNAPSHOT") {
			v.Release = sorted[i]
			break
		}
	}

	return Metadata{GroupID: groupID, ArtifactID: artifactID, Versioning: v}
}

type snapshotFile struct {
	timestamp   string // yyyyMMdd.HHmmss
	buildNumber int
	classifier  string
	extension   string
}

func (f snapshotFile) newerThan(o snapshotFile) bool {
	if f.timestamp != o.timestamp {
		return f.timestamp > o.timestamp
	}
	return f.buildNumber > o.buildNumber
}

// NewSnapshotMetadata builds the <group>/<artifactId>/<version>-SNAPSHOT/maven-metadata.xml from
// the timestamped file names in that directory. ok is false if there are no timestamped files.
func NewSnapshotMetadata(groupID, artifactID, version string, fileNames []string, extensions []string, now time.Time) (m Metadata, ok bool) {
	base := strings.TrimSuffix(version, "-SNAPSHOT")
	exts := make([]string, len(extensions))
	for i, e := range extensions {
		exts[i] = regexp.QuoteMeta(strings.TrimPrefix(e, "."))
	}
	pattern := regexp.MustCompile(`^` + regexp.QuoteMeta(artifactID+"-"+base) +
		`-(\d{8}\.\d{6})-(\d+)(?:-(.+?))?\.(` + strings.Join(exts, "|") + `)$`)

	type key struct{ classifier, extension string }
	latestPerKey := map[key]snapshotFile{}
	var latest *snapshotFile

	for _, name := range fileNames {
		match := pattern.FindStringSubmatch(name)
		if match == nil {
			continue
		}
		build, err := strconv.Atoi(match[2])
		if err != nil {
			continue
		}
		f := snapshotFile{timestamp: match[1], buildNumber: build, classifier: match[3], extension: match[4]}

		k := key{f.classifier, f.extension}
		if prev, exists := latestPerKey[k]; !exists || f.newerThan(prev) {
			latestPerKey[k] = f
		}
		if latest == nil || f.newerThan(*latest) {
			latest = &f
		}
	}
	if latest == nil {
		return Metadata{}, false
	}

	var snapshotVersions []SnapshotVersion
	for _, f := range latestPerKey {
		snapshotVersions = append(snapshotVersions, SnapshotVersion{
			Classifier: f.classifier,
			Extension:  f.extension,
			Value:      base + "-" + f.timestamp + "-" + strconv.Itoa(f.buildNumber),
			Updated:    strings.Replace(f.timestamp, ".", "", 1),
		})
	}
	slices.SortFunc(snapshotVersions, func(a, b SnapshotVersion) int {
		if c := strings.Compare(a.Classifier, b.Classifier); c != 0 {
			return c
		}
		return strings.Compare(a.Extension, b.Extension)
	})

	return Metadata{
		ModelVersion: "1.1.0",
		GroupID:      groupID,
		ArtifactID:   artifactID,
		Version:      version,
		Versioning: Versioning{
			Snapshot:         &Snapshot{Timestamp: latest.timestamp, BuildNumber: latest.buildNumber},
			LastUpdated:      formatLastUpdated(now),
			SnapshotVersions: &SnapshotVersions{SnapshotVersion: snapshotVersions},
		},
	}, true
}
