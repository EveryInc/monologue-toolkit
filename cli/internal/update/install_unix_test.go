//go:build !windows

package update

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallBinaryAtomicallyReplacesTarget(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	source := filepath.Join(directory, "source")
	target := filepath.Join(directory, "monologue")
	if err := os.WriteFile(source, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	pending, err := installBinary(source, target)
	if err != nil {
		t.Fatal(err)
	}
	if pending {
		t.Fatal("unix install should not be pending")
	}
	payload, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) != "new" {
		t.Fatalf("target payload = %q, want new", payload)
	}
}
