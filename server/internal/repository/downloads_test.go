package repository

import (
	"context"
	"testing"
	"time"
)

func TestDownloads(t *testing.T) {
	ctx := context.Background()
	r := newTestRepos(t)
	day := func(d int) time.Time { return time.Date(2026, 10, d, 15, 0, 0, 0, time.UTC) }

	for _, d := range []struct {
		path, version string
		at            time.Time
	}{
		{"com/example/demo", "1.0", day(1)},
		{"com/example/demo", "1.0", day(1)},
		{"com/example/demo", "1.1", day(3)},
		{"com/example/demo", "1.1", day(3)},
		{"com/example/demo", "1.1", day(3)},
		{"com/example/other", "2.0", day(2)},
	} {
		if err := r.Downloads.Record(ctx, d.path, d.version, d.at); err != nil {
			t.Fatal(err)
		}
	}

	if n, _ := r.Downloads.Total(ctx, "com/example/demo", time.Time{}); n != 5 {
		t.Errorf("demo total = %d, want 5", n)
	}
	if n, _ := r.Downloads.Total(ctx, "com/example/demo", day(2)); n != 3 {
		t.Errorf("demo since day 2 = %d, want 3", n)
	}
	if n, _ := r.Downloads.Total(ctx, "", time.Time{}); n != 6 {
		t.Errorf("all total = %d, want 6", n)
	}

	daily, err := r.Downloads.Daily(ctx, "com/example/demo", day(1), day(4))
	if err != nil {
		t.Fatal(err)
	}
	want := []DayDownloads{{"2026-10-01", 2}, {"2026-10-02", 0}, {"2026-10-03", 3}, {"2026-10-04", 0}}
	if len(daily) != len(want) {
		t.Fatalf("daily = %v, want %v", daily, want)
	}
	for i := range want {
		if daily[i] != want[i] {
			t.Errorf("daily[%d] = %v, want %v", i, daily[i], want[i])
		}
	}

	versions, err := r.Downloads.ByVersion(ctx, "com/example/demo", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 || versions[0] != (VersionDownloads{"1.1", 3}) || versions[1] != (VersionDownloads{"1.0", 2}) {
		t.Errorf("by version = %v", versions)
	}

	paths, err := r.Downloads.ByArtifact(ctx, day(2))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 || paths[0] != (PathDownloads{"com/example/demo", 3}) || paths[1] != (PathDownloads{"com/example/other", 1}) {
		t.Errorf("by artifact = %v", paths)
	}
}
