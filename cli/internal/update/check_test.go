package update

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckForUpdateCachesLatestRelease(t *testing.T) {
	t.Parallel()
	cachePath := filepath.Join(t.TempDir(), "nested", "update-check.json")
	now := time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC)
	fetchCount := 0
	fetch := func(context.Context) (release, error) {
		fetchCount++
		return release{TagName: "v0.2.0"}, nil
	}

	latest, outdated, err := checkForUpdate(context.Background(), "0.1.0", cachePath, now, fetch)
	if err != nil || latest != "v0.2.0" || !outdated {
		t.Fatalf("unexpected first check: latest=%q outdated=%v err=%v", latest, outdated, err)
	}
	latest, outdated, err = checkForUpdate(context.Background(), "0.1.0", cachePath, now.Add(time.Hour), fetch)
	if err != nil || latest != "v0.2.0" || !outdated {
		t.Fatalf("unexpected cached check: latest=%q outdated=%v err=%v", latest, outdated, err)
	}
	if fetchCount != 1 {
		t.Fatalf("fetch count = %d, want 1", fetchCount)
	}
}

func TestCheckForUpdateUsesStaleCacheWhenRefreshFails(t *testing.T) {
	t.Parallel()
	cachePath := filepath.Join(t.TempDir(), "update-check.json")
	now := time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC)
	if err := writeCheckCache(cachePath, checkCache{CheckedAt: now.Add(-48 * time.Hour), LatestVersion: "v0.2.0"}); err != nil {
		t.Fatal(err)
	}

	latest, outdated, err := checkForUpdate(context.Background(), "0.1.0", cachePath, now, func(context.Context) (release, error) {
		return release{}, errors.New("offline")
	})
	if err != nil || latest != "v0.2.0" || !outdated {
		t.Fatalf("unexpected stale-cache result: latest=%q outdated=%v err=%v", latest, outdated, err)
	}
}

func TestCheckForUpdateThrottlesFailedRefreshWithoutExistingCache(t *testing.T) {
	t.Parallel()
	cachePath := filepath.Join(t.TempDir(), "update-check.json")
	now := time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC)
	fetchCount := 0
	fetch := func(context.Context) (release, error) {
		fetchCount++
		return release{}, errors.New("offline")
	}

	for _, checkedAt := range []time.Time{now, now.Add(time.Hour)} {
		latest, outdated, err := checkForUpdate(context.Background(), "0.1.0", cachePath, checkedAt, fetch)
		if err != nil || latest != "" || outdated {
			t.Fatalf("unexpected failed-check result: latest=%q outdated=%v err=%v", latest, outdated, err)
		}
	}
	if fetchCount != 1 {
		t.Fatalf("fetch count = %d, want 1", fetchCount)
	}
}
