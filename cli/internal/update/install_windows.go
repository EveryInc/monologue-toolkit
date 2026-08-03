//go:build windows

package update

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	createNoWindow        = 0x08000000
	detachedProcess       = 0x00000008
	updateHelperFileMode  = 0o600
	updatedBinaryFileMode = 0o755
)

func installBinary(source string, target string) (bool, error) {
	stagedFile, err := os.CreateTemp(filepath.Dir(target), ".monologue-update-*.exe")
	if err != nil {
		return false, err
	}
	stagedPath := stagedFile.Name()
	cleanupStaged := true
	defer func() {
		if cleanupStaged {
			os.Remove(stagedPath)
		}
	}()

	sourceFile, err := os.Open(source)
	if err != nil {
		stagedFile.Close()
		return false, err
	}
	_, copyErr := io.Copy(stagedFile, sourceFile)
	closeSourceErr := sourceFile.Close()
	if copyErr != nil {
		stagedFile.Close()
		return false, copyErr
	}
	if closeSourceErr != nil {
		stagedFile.Close()
		return false, closeSourceErr
	}
	if err := stagedFile.Chmod(updatedBinaryFileMode); err != nil {
		stagedFile.Close()
		return false, err
	}
	if err := stagedFile.Close(); err != nil {
		return false, err
	}

	helperFile, err := os.CreateTemp("", "monologue-update-*.ps1")
	if err != nil {
		return false, err
	}
	helperPath := helperFile.Name()
	cleanupHelper := true
	defer func() {
		if cleanupHelper {
			os.Remove(helperPath)
		}
	}()

	script := fmt.Sprintf(
		"$ErrorActionPreference = 'Stop'\nGet-Process -Id %d -ErrorAction SilentlyContinue | Wait-Process\nMove-Item -LiteralPath '%s' -Destination '%s' -Force\nRemove-Item -LiteralPath '%s' -Force\n",
		os.Getpid(), powershellQuote(stagedPath), powershellQuote(target), powershellQuote(helperPath),
	)
	if err := helperFile.Chmod(updateHelperFileMode); err != nil {
		helperFile.Close()
		return false, err
	}
	if _, err := helperFile.WriteString(script); err != nil {
		helperFile.Close()
		return false, err
	}
	if err := helperFile.Close(); err != nil {
		return false, err
	}

	command := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", helperPath)
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNoWindow | detachedProcess, HideWindow: true}
	if err := command.Start(); err != nil {
		return false, fmt.Errorf("start update helper: %w", err)
	}
	if err := command.Process.Release(); err != nil {
		return false, fmt.Errorf("detach update helper: %w", err)
	}

	cleanupStaged = false
	cleanupHelper = false
	return true, nil
}

func powershellQuote(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
