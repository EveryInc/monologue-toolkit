package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EveryInc/monologue-toolkit/cli/internal/monologue"
	cliupdate "github.com/EveryInc/monologue-toolkit/cli/internal/update"
	"github.com/EveryInc/monologue-toolkit/cli/internal/version"
)

func TestRunWarnsOnStaleVersionWithoutChangingStdout(t *testing.T) {
	originalVersion := version.Version
	originalCheck := checkForUpdate
	version.Version = "0.1.0"
	checkForUpdate = func(context.Context, string, string) (string, bool, error) {
		return "v0.2.0", true, nil
	}
	t.Cleanup(func() {
		version.Version = originalVersion
		checkForUpdate = originalCheck
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := Run([]string{"version"}, bytes.NewReader(nil), &stdout, &stderr); exitCode != 0 {
		t.Fatalf("Run returned %d", exitCode)
	}
	if got, want := stdout.String(), "monologue 0.1.0 (commit none, built unknown)\n"; got != want {
		t.Fatalf("unexpected stdout: got %q, want %q", got, want)
	}
	if got, want := stderr.String(), "A newer Monologue CLI version is available (v0.2.0; you are using 0.1.0). Run `monologue update` to update.\n"; got != want {
		t.Fatalf("unexpected stderr: got %q, want %q", got, want)
	}
}

func TestRunUpdateReportsSuccess(t *testing.T) {
	originalUpdate := updateCLI
	updateCLI = func(context.Context, string) (cliupdate.Result, error) {
		return cliupdate.Result{CurrentVersion: "0.1.0", LatestVersion: "v0.2.0", Updated: true}, nil
	}
	t.Cleanup(func() { updateCLI = originalUpdate })

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := Run([]string{"update"}, bytes.NewReader(nil), &stdout, &stderr); exitCode != 0 {
		t.Fatalf("Run returned %d, stderr: %s", exitCode, stderr.String())
	}
	if got, want := stdout.String(), "Updated Monologue CLI from 0.1.0 to v0.2.0.\n"; got != want {
		t.Fatalf("unexpected stdout: got %q, want %q", got, want)
	}
}

func TestRunNotesGetAcceptsFieldBeforeOrAfterNoteID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if got := request.URL.Path; got != "/v1/public-api/notes/note_123" {
			t.Fatalf("unexpected path: %q", got)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer mono_pat_test" {
			t.Fatalf("unexpected auth header: %q", got)
		}

		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(monologue.Note{Transcript: stringPointer("hello from the transcript")}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	t.Setenv("MONOLOGUE_API_BASE_URL", server.URL)
	t.Setenv("MONOLOGUE_API_TOKEN", "mono_pat_test")
	t.Setenv("MONOLOGUE_CONFIG_DIR", t.TempDir())

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "field after note id",
			args: []string{"notes", "get", "note_123", "--field", "transcript"},
		},
		{
			name: "field before note id",
			args: []string{"notes", "get", "--field", "transcript", "note_123"},
		},
		{
			name: "flags after note id",
			args: []string{"notes", "get", "note_123", "--field", "transcript", "--token", "mono_pat_test"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			if exitCode := Run(test.args, bytes.NewReader(nil), &stdout, &stderr); exitCode != 0 {
				t.Fatalf("Run returned %d, stderr: %s", exitCode, stderr.String())
			}
			if got, want := stdout.String(), "hello from the transcript\n"; got != want {
				t.Fatalf("unexpected stdout: got %q, want %q", got, want)
			}
		})
	}
}

func TestRunOnboardingAcceptsTokenFlagForTrustedAutomation(t *testing.T) {
	t.Setenv("MONOLOGUE_API_TOKEN", "")
	t.Setenv("MONOLOGUE_CONFIG_DIR", t.TempDir())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(
		[]string{"onboarding", "--token", "mono_pat_test", "--skip-verify"},
		bytes.NewReader(nil),
		&stdout,
		&stderr,
	)
	if exitCode != 0 {
		t.Fatalf("Run returned %d, stderr: %s", exitCode, stderr.String())
	}
}

func TestRunNotesListAndAllPassRepeatedTagIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if got := request.URL.Path; got != "/v1/public-api/notes" {
			t.Fatalf("unexpected path: %q", got)
		}
		if got := request.URL.Query()["tag_id"]; len(got) != 2 || got[0] != "tag_1" || got[1] != "tag_2" {
			t.Fatalf("unexpected tag_id values: %#v", got)
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err := writer.Write([]byte(`{"items":[]}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	t.Setenv("MONOLOGUE_API_BASE_URL", server.URL)
	t.Setenv("MONOLOGUE_API_TOKEN", "mono_pat_test")
	t.Setenv("MONOLOGUE_CONFIG_DIR", t.TempDir())

	for _, command := range []string{"list", "all"} {
		t.Run(command, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := Run(
				[]string{"notes", command, "--tag-id", "tag_1", "--tag-id", "tag_2"},
				bytes.NewReader(nil),
				&stdout,
				&stderr,
			)
			if exitCode != 0 {
				t.Fatalf("Run returned %d, stderr: %s", exitCode, stderr.String())
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}
