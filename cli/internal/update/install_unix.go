//go:build !windows

package update

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func installBinary(source string, target string) (bool, error) {
	info, err := os.Stat(target)
	if err != nil {
		return false, err
	}

	tempFile, err := os.CreateTemp(filepath.Dir(target), ".monologue-update-*")
	if err != nil {
		return false, err
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	sourceFile, err := os.Open(source)
	if err != nil {
		tempFile.Close()
		return false, err
	}
	_, copyErr := io.Copy(tempFile, sourceFile)
	closeSourceErr := sourceFile.Close()
	if copyErr != nil {
		tempFile.Close()
		return false, copyErr
	}
	if closeSourceErr != nil {
		tempFile.Close()
		return false, closeSourceErr
	}
	if err := tempFile.Chmod(info.Mode().Perm() | 0o111); err != nil {
		tempFile.Close()
		return false, err
	}
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return false, err
	}
	if err := tempFile.Close(); err != nil {
		return false, err
	}
	if err := os.Rename(tempPath, target); err != nil {
		return false, fmt.Errorf("install updated executable: %w", err)
	}
	return false, nil
}
