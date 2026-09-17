package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EveryInc/monologue-toolkit/cli/internal/monologue"
)

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

func TestRunNotesRecordingAccess(t *testing.T) {
	const recordingURL = "https://storage.example/recording?signature=test&expires=3600"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.Header.Get("Authorization") != "Bearer mono_pat_test" {
			t.Errorf("expected authenticated GET request")
		}
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/v1/public-api/notes/note_123":
			_, _ = writer.Write([]byte(`{"note_id":"note_123","recording_url":"` + recordingURL + `","recording_url_expires_at":"2026-09-17T09:00:00Z","recording_content_type":"audio/mp4","recording_bytes":123456}`))
		case "/v1/public-api/notes/note_123/audio-url":
			_, _ = writer.Write([]byte(`{"audio_url":"` + recordingURL + `","expires_in":3600}`))
		default:
			t.Errorf("unexpected path: %q", request.URL.Path)
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	t.Setenv("MONOLOGUE_API_BASE_URL", server.URL)
	t.Setenv("MONOLOGUE_API_TOKEN", "mono_pat_test")
	t.Setenv("MONOLOGUE_CONFIG_DIR", t.TempDir())

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"detail URL", []string{"get", "note_123", "--field", "recording_url"}, recordingURL + "\n"},
		{"detail expiry", []string{"get", "note_123", "--field", "recording_url_expires_at"}, "2026-09-17T09:00:00Z\n"},
		{"content type", []string{"get", "note_123", "--field", "recording_content_type"}, "audio/mp4\n"},
		{"byte count", []string{"get", "note_123", "--field", "recording_bytes"}, "123456\n"},
		{"refresh URL", []string{"audio-url", "note_123", "--field", "audio_url"}, recordingURL + "\n"},
		{"flags first", []string{"audio-url", "--field", "audio_url", "note_123"}, recordingURL + "\n"},
		{"refresh expiry", []string{"audio-url", "note_123", "--field", "expires_in"}, "3600\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(append([]string{"notes"}, test.args...), bytes.NewReader(nil), &stdout, &stderr)
			if code != 0 || stdout.String() != test.want {
				t.Fatalf("code=%d stdout=%q stderr=%q; want %q", code, stdout.String(), stderr.String(), test.want)
			}
		})
	}
	for _, command := range []string{"get", "audio-url"} {
		t.Run(command+" JSON", func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run([]string{"notes", command, "note_123"}, bytes.NewReader(nil), &stdout, &stderr)
			var response map[string]interface{}
			if err := json.Unmarshal(stdout.Bytes(), &response); err != nil || code != 0 {
				t.Fatalf("code=%d stderr=%q decode=%v", code, stderr.String(), err)
			}
			key := "recording_url"
			if command == "audio-url" {
				key = "audio_url"
			}
			if response[key] != recordingURL {
				t.Fatalf("missing recording URL: %#v", response)
			}
		})
	}
}

func TestRunNotesAudioURLFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte(`{"detail":"Note not found"}`))
	}))
	defer server.Close()
	t.Setenv("MONOLOGUE_API_BASE_URL", server.URL)
	t.Setenv("MONOLOGUE_API_TOKEN", "mono_pat_test")
	t.Setenv("MONOLOGUE_CONFIG_DIR", t.TempDir())
	for _, args := range [][]string{
		{"notes", "audio-url"},
		{"notes", "audio-url", "note_123", "extra"},
		{"notes", "audio-url", "note_123"},
	} {
		var stdout, stderr bytes.Buffer
		if code := Run(args, bytes.NewReader(nil), &stdout, &stderr); code != 1 || stdout.Len() != 0 || stderr.Len() == 0 {
			t.Fatalf("args=%v code=%d stdout=%q stderr=%q", args, code, stdout.String(), stderr.String())
		}
	}
}
