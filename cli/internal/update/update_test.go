package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFetchLatestRelease(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("User-Agent"); got != "monologue-toolkit/0.1.0" {
			t.Fatalf("unexpected user agent: %q", got)
		}
		_ = json.NewEncoder(writer).Encode(release{
			TagName: "v0.2.0",
			Assets:  []asset{{Name: "checksums.txt", URL: "https://example.com/checksums.txt"}},
		})
	}))
	defer server.Close()

	latest, err := fetchLatestRelease(context.Background(), server.Client(), server.URL, "0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if latest.TagName != "v0.2.0" || len(latest.Assets) != 1 {
		t.Fatalf("unexpected release: %#v", latest)
	}
}

func TestArchiveName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		goos   string
		goarch string
		want   string
	}{
		{goos: "darwin", goarch: "arm64", want: "monologue_darwin_arm64.tar.gz"},
		{goos: "linux", goarch: "amd64", want: "monologue_linux_amd64.tar.gz"},
		{goos: "windows", goarch: "arm64", want: "monologue_windows_arm64.zip"},
	}
	for _, test := range tests {
		got, err := archiveName(test.goos, test.goarch)
		if err != nil || got != test.want {
			t.Errorf("archiveName(%q, %q) = %q, %v; want %q", test.goos, test.goarch, got, err, test.want)
		}
	}
}

func TestVerifyChecksum(t *testing.T) {
	t.Parallel()
	payload := []byte("release archive")
	digest := sha256.Sum256(payload)
	checksums := []byte(fmt.Sprintf("%x  monologue_darwin_arm64.tar.gz\n", digest))
	if err := verifyChecksum("monologue_darwin_arm64.tar.gz", payload, checksums); err != nil {
		t.Fatal(err)
	}
	if err := verifyChecksum("monologue_darwin_arm64.tar.gz", []byte("tampered"), checksums); err == nil {
		t.Fatal("expected checksum mismatch")
	}
}

func TestExtractTarBinary(t *testing.T) {
	t.Parallel()
	var archive bytes.Buffer
	gzipWriter := gzip.NewWriter(&archive)
	tarWriter := tar.NewWriter(gzipWriter)
	payload := []byte("new binary")
	if err := tarWriter.WriteHeader(&tar.Header{Name: "monologue", Mode: 0o755, Size: int64(len(payload))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}

	destination := filepath.Join(t.TempDir(), "monologue")
	if err := extractTarBinary(archive.Bytes(), destination); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("extracted payload = %q, want %q", got, payload)
	}
}

func TestExtractZipBinary(t *testing.T) {
	t.Parallel()
	var archive bytes.Buffer
	zipWriter := zip.NewWriter(&archive)
	file, err := zipWriter.Create("monologue.exe")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("windows binary")); err != nil {
		t.Fatal(err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatal(err)
	}

	destination := filepath.Join(t.TempDir(), "monologue.exe")
	if err := extractZipBinary(archive.Bytes(), destination); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "windows binary" {
		t.Fatalf("extracted payload = %q", got)
	}
}
