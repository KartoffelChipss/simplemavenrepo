package maven

import (
	"slices"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 10, 30, 21, 15, 0, 0, time.UTC)

func TestNewArtifactMetadata(t *testing.T) {
	m := NewArtifactMetadata("org.example", "lib", []string{"1.10.0", "2.0-SNAPSHOT", "1.9.0", "1.0"}, testNow)

	if want := []string{"1.0", "1.9.0", "1.10.0", "2.0-SNAPSHOT"}; !slices.Equal(m.Versioning.Versions.Version, want) {
		t.Errorf("versions = %v, want %v", m.Versioning.Versions.Version, want)
	}
	if m.Versioning.Latest != "2.0-SNAPSHOT" {
		t.Errorf("latest = %q", m.Versioning.Latest)
	}
	if m.Versioning.Release != "1.10.0" {
		t.Errorf("release = %q", m.Versioning.Release)
	}
	if m.Versioning.LastUpdated != "20261030211500" {
		t.Errorf("lastUpdated = %q", m.Versioning.LastUpdated)
	}

	out, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"<?xml", "<metadata>", "<groupId>org.example</groupId>", "<version>1.9.0</version>", "<release>1.10.0</release>"} {
		if !strings.Contains(string(out), s) {
			t.Errorf("output missing %q:\n%s", s, out)
		}
	}
}

func TestNewArtifactMetadataOnlySnapshots(t *testing.T) {
	m := NewArtifactMetadata("org.example", "lib", []string{"1.0-SNAPSHOT"}, testNow)
	if m.Versioning.Release != "" {
		t.Errorf("release = %q, want empty", m.Versioning.Release)
	}
	out, _ := m.Marshal()
	for _, s := range []string{"<release>", "<snapshotVersions>"} {
		if strings.Contains(string(out), s) {
			t.Errorf("output must not contain %s:\n%s", s, out)
		}
	}
}

func TestNewSnapshotMetadata(t *testing.T) {
	files := []string{
		"lib-1.0-20261030.100000-1.jar",
		"lib-1.0-20261030.100000-1.pom",
		"lib-1.0-20261030.100000-1.jar.sha1",
		"lib-1.0-20261030.120000-2.jar",
		"lib-1.0-20261030.120000-2.pom",
		"lib-1.0-20261030.100000-1-sources.jar",
		"maven-metadata.xml",
		"other-1.0-20261030.130000-9.jar",
	}
	m, ok := NewSnapshotMetadata("org.example", "lib", "1.0-SNAPSHOT", files, []string{".jar", ".pom"}, testNow)
	if !ok {
		t.Fatal("expected metadata")
	}

	if s := m.Versioning.Snapshot; s.Timestamp != "20261030.120000" || s.BuildNumber != 2 {
		t.Errorf("snapshot = %+v", *s)
	}
	want := []SnapshotVersion{
		{Extension: "jar", Value: "1.0-20261030.120000-2", Updated: "20261030120000"},
		{Extension: "pom", Value: "1.0-20261030.120000-2", Updated: "20261030120000"},
		{Classifier: "sources", Extension: "jar", Value: "1.0-20261030.100000-1", Updated: "20261030100000"},
	}
	if !slices.Equal(m.Versioning.SnapshotVersions.SnapshotVersion, want) {
		t.Errorf("snapshotVersions = %+v, want %+v", m.Versioning.SnapshotVersions.SnapshotVersion, want)
	}
	if m.Version != "1.0-SNAPSHOT" || m.ModelVersion != "1.1.0" {
		t.Errorf("version = %q, modelVersion = %q", m.Version, m.ModelVersion)
	}
}

func TestNewSnapshotMetadataMultiPartExtension(t *testing.T) {
	files := []string{
		"lib-1.0-20261030.100000-1.aar",
		"lib-1.0-20261030.100000-1.module",
		"lib-1.0-20261030.100000-1.tar.gz",
		"lib-1.0-20261030.100000-1.tar.gz.sha1",
		"lib-1.0-20261030.100000-1-dist.tar.gz",
		"lib-1.0-20261030.100000-1-dist.zip",
	}
	m, ok := NewSnapshotMetadata("org.example", "lib", "1.0-SNAPSHOT", files, []string{".aar", ".module", ".zip", ".tar.gz"}, testNow)
	if !ok {
		t.Fatal("expected metadata")
	}

	want := []SnapshotVersion{
		{Extension: "aar", Value: "1.0-20261030.100000-1", Updated: "20261030100000"},
		{Extension: "module", Value: "1.0-20261030.100000-1", Updated: "20261030100000"},
		{Extension: "tar.gz", Value: "1.0-20261030.100000-1", Updated: "20261030100000"},
		{Classifier: "dist", Extension: "tar.gz", Value: "1.0-20261030.100000-1", Updated: "20261030100000"},
		{Classifier: "dist", Extension: "zip", Value: "1.0-20261030.100000-1", Updated: "20261030100000"},
	}
	if !slices.Equal(m.Versioning.SnapshotVersions.SnapshotVersion, want) {
		t.Errorf("snapshotVersions = %+v, want %+v", m.Versioning.SnapshotVersions.SnapshotVersion, want)
	}
}

func TestNewSnapshotMetadataNoTimestampedFiles(t *testing.T) {
	_, ok := NewSnapshotMetadata("org.example", "lib", "1.0-SNAPSHOT", []string{"lib-1.0-SNAPSHOT.jar"}, []string{".jar"}, testNow)
	if ok {
		t.Error("expected no metadata")
	}
}
