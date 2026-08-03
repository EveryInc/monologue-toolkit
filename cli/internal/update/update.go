package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	latestReleaseURL = "https://api.github.com/repos/EveryInc/monologue-toolkit/releases/latest"
	maxDownloadSize  = 100 << 20
)

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type Result struct {
	CurrentVersion string
	LatestVersion  string
	ExecutablePath string
	Updated        bool
	PendingRestart bool
}

func Update(ctx context.Context, currentVersion string) (Result, error) {
	if _, err := parseVersion(currentVersion); err != nil {
		return Result{}, fmt.Errorf("cannot self-update unversioned build %q: install a released CLI first", currentVersion)
	}

	httpClient := &http.Client{Timeout: 2 * time.Minute}
	latest, err := fetchLatestRelease(ctx, httpClient, latestReleaseURL, currentVersion)
	if err != nil {
		return Result{}, err
	}

	newer, err := isNewer(latest.TagName, currentVersion)
	if err != nil {
		return Result{}, err
	}
	if !newer {
		return Result{CurrentVersion: currentVersion, LatestVersion: latest.TagName}, nil
	}

	archiveName, err := archiveName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return Result{}, err
	}
	archiveAsset, err := findAsset(latest, archiveName)
	if err != nil {
		return Result{}, err
	}
	checksumsAsset, err := findAsset(latest, "checksums.txt")
	if err != nil {
		return Result{}, err
	}

	archivePayload, err := download(ctx, httpClient, archiveAsset.URL, currentVersion)
	if err != nil {
		return Result{}, fmt.Errorf("download %s: %w", archiveName, err)
	}
	checksumsPayload, err := download(ctx, httpClient, checksumsAsset.URL, currentVersion)
	if err != nil {
		return Result{}, fmt.Errorf("download checksums.txt: %w", err)
	}
	if err := verifyChecksum(archiveName, archivePayload, checksumsPayload); err != nil {
		return Result{}, err
	}

	tempDir, err := os.MkdirTemp("", "monologue-update-*")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(tempDir)

	binaryPath := filepath.Join(tempDir, binaryName(runtime.GOOS))
	if err := extractBinary(archiveName, archivePayload, binaryPath); err != nil {
		return Result{}, err
	}
	executablePath, err := os.Executable()
	if err != nil {
		return Result{}, fmt.Errorf("locate current executable: %w", err)
	}
	executablePath, err = filepath.EvalSymlinks(executablePath)
	if err != nil {
		return Result{}, fmt.Errorf("resolve current executable: %w", err)
	}

	pendingRestart, err := installBinary(binaryPath, executablePath)
	if err != nil {
		return Result{}, fmt.Errorf("replace %s: %w", executablePath, err)
	}
	return Result{
		CurrentVersion: currentVersion,
		LatestVersion:  latest.TagName,
		ExecutablePath: executablePath,
		Updated:        true,
		PendingRestart: pendingRestart,
	}, nil
}

func fetchLatestRelease(ctx context.Context, httpClient *http.Client, endpoint string, currentVersion string) (release, error) {
	payload, err := download(ctx, httpClient, endpoint, currentVersion)
	if err != nil {
		return release{}, fmt.Errorf("check latest release: %w", err)
	}

	var latest release
	if err := json.Unmarshal(payload, &latest); err != nil {
		return release{}, fmt.Errorf("decode latest release: %w", err)
	}
	if latest.TagName == "" {
		return release{}, errors.New("latest release did not include a version")
	}
	if _, err := parseVersion(latest.TagName); err != nil {
		return release{}, fmt.Errorf("latest release version: %w", err)
	}
	return latest, nil
}

func download(ctx context.Context, httpClient *http.Client, url string, currentVersion string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "monologue-toolkit/"+currentVersion)

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d", response.StatusCode)
	}

	payload, err := io.ReadAll(io.LimitReader(response.Body, maxDownloadSize+1))
	if err != nil {
		return nil, err
	}
	if len(payload) > maxDownloadSize {
		return nil, errors.New("download exceeded 100 MiB limit")
	}
	return payload, nil
}

func archiveName(goos string, goarch string) (string, error) {
	if goarch != "amd64" && goarch != "arm64" {
		return "", fmt.Errorf("unsupported architecture: %s", goarch)
	}
	switch goos {
	case "darwin", "linux":
		return fmt.Sprintf("monologue_%s_%s.tar.gz", goos, goarch), nil
	case "windows":
		return fmt.Sprintf("monologue_windows_%s.zip", goarch), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", goos)
	}
}

func findAsset(latest release, name string) (asset, error) {
	for _, candidate := range latest.Assets {
		if candidate.Name == name && candidate.URL != "" {
			return candidate, nil
		}
	}
	return asset{}, fmt.Errorf("release %s does not include %s", latest.TagName, name)
}

func verifyChecksum(name string, payload []byte, checksums []byte) error {
	expected := ""
	for _, line := range strings.Split(string(checksums), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == name {
			expected = strings.ToLower(fields[0])
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("checksums.txt does not include %s", name)
	}
	if _, err := hex.DecodeString(expected); err != nil || len(expected) != sha256.Size*2 {
		return fmt.Errorf("invalid SHA-256 checksum for %s", name)
	}
	actual := sha256.Sum256(payload)
	if hex.EncodeToString(actual[:]) != expected {
		return fmt.Errorf("checksum verification failed for %s", name)
	}
	return nil
}

func extractBinary(archiveName string, payload []byte, destination string) error {
	if strings.HasSuffix(archiveName, ".zip") {
		return extractZipBinary(payload, destination)
	}
	return extractTarBinary(payload, destination)
}

func extractTarBinary(payload []byte, destination string) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("open release archive: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read release archive: %w", err)
		}
		if safeArchiveName(header.Name) != "monologue" || header.Typeflag != tar.TypeReg {
			continue
		}
		return writeExtractedBinary(destination, tarReader)
	}
	return errors.New("release archive does not include monologue")
}

func extractZipBinary(payload []byte, destination string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return fmt.Errorf("open release archive: %w", err)
	}
	for _, file := range zipReader.File {
		if safeArchiveName(file.Name) != "monologue.exe" || file.FileInfo().IsDir() {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return fmt.Errorf("open monologue.exe in release archive: %w", err)
		}
		writeErr := writeExtractedBinary(destination, reader)
		closeErr := reader.Close()
		if writeErr != nil {
			return writeErr
		}
		return closeErr
	}
	return errors.New("release archive does not include monologue.exe")
}

func safeArchiveName(name string) string {
	cleaned := path.Clean(strings.ReplaceAll(name, "\\", "/"))
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || path.IsAbs(cleaned) {
		return ""
	}
	return path.Base(cleaned)
}

func writeExtractedBinary(destination string, reader io.Reader) error {
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	written, err := io.Copy(file, io.LimitReader(reader, maxDownloadSize+1))
	if err != nil {
		file.Close()
		return err
	}
	if written > maxDownloadSize {
		file.Close()
		return errors.New("extracted binary exceeded 100 MiB limit")
	}
	return file.Close()
}

func binaryName(goos string) string {
	if goos == "windows" {
		return "monologue.exe"
	}
	return "monologue"
}
