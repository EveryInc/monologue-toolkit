package update

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const checkInterval = 24 * time.Hour

type checkCache struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
}

func CheckForUpdate(ctx context.Context, currentVersion string, cachePath string) (string, bool, error) {
	if _, err := parseVersion(currentVersion); err != nil {
		return "", false, nil
	}

	httpClient := &http.Client{Timeout: 2 * time.Second}
	return checkForUpdate(ctx, currentVersion, cachePath, time.Now(), func(ctx context.Context) (release, error) {
		return fetchLatestRelease(ctx, httpClient, latestReleaseURL, currentVersion)
	})
}

func checkForUpdate(
	ctx context.Context,
	currentVersion string,
	cachePath string,
	now time.Time,
	fetch func(context.Context) (release, error),
) (string, bool, error) {
	cached, cacheErr := readCheckCache(cachePath)
	if cacheErr == nil && now.Sub(cached.CheckedAt) >= 0 && now.Sub(cached.CheckedAt) < checkInterval {
		if cached.LatestVersion == "" {
			return "", false, nil
		}
		newer, err := isNewer(cached.LatestVersion, currentVersion)
		return cached.LatestVersion, newer, err
	}

	latest, err := fetch(ctx)
	if err != nil {
		failedCheck := checkCache{CheckedAt: now.UTC()}
		if cacheErr == nil {
			failedCheck.LatestVersion = cached.LatestVersion
			_ = writeCheckCache(cachePath, failedCheck)
			if cached.LatestVersion == "" {
				return "", false, nil
			}
			newer, compareErr := isNewer(cached.LatestVersion, currentVersion)
			return cached.LatestVersion, newer, compareErr
		}
		_ = writeCheckCache(cachePath, failedCheck)
		return "", false, nil
	}

	_ = writeCheckCache(cachePath, checkCache{CheckedAt: now.UTC(), LatestVersion: latest.TagName})
	newer, err := isNewer(latest.TagName, currentVersion)
	return latest.TagName, newer, err
}

func readCheckCache(path string) (checkCache, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return checkCache{}, err
	}

	var cached checkCache
	if err := json.Unmarshal(payload, &cached); err != nil {
		return checkCache{}, err
	}
	if cached.CheckedAt.IsZero() {
		return checkCache{}, errors.New("incomplete update check cache")
	}
	return cached, nil
}

func writeCheckCache(path string, cached checkCache) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')

	tempFile, err := os.CreateTemp(filepath.Dir(path), ".update-check-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if err := tempFile.Chmod(0o600); err != nil {
		tempFile.Close()
		return err
	}
	if _, err := tempFile.Write(payload); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	err = os.Rename(tempPath, path)
	if err == nil || runtime.GOOS != "windows" {
		return err
	}
	if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		return err
	}
	return os.Rename(tempPath, path)
}
