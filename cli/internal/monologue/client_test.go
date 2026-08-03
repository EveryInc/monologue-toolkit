package monologue

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListNotesSendsAuthAndFilters(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("Authorization"); got != "Bearer mono_pat_test" {
			t.Fatalf("unexpected auth header: %q", got)
		}
		if got := request.Header.Get("User-Agent"); got != "monologue-toolkit/dev" {
			t.Fatalf("unexpected user agent: %q", got)
		}
		if got := request.URL.Path; got != "/v1/public-api/notes" {
			t.Fatalf("unexpected path: %q", got)
		}
		query := request.URL.Query()
		if got := query.Get("limit"); got != "10" {
			t.Fatalf("unexpected limit: %q", got)
		}
		if got := query.Get("cursor"); got != "cursor_1" {
			t.Fatalf("unexpected cursor: %q", got)
		}
		if got := query.Get("q"); got != "customer interview" {
			t.Fatalf("unexpected q: %q", got)
		}
		if got := query["tag_id"]; len(got) != 2 || got[0] != "tag_1" || got[1] != "tag_2" {
			t.Fatalf("unexpected tag_id values: %#v", got)
		}
		if got := query.Get("created_after"); got != "2026-01-01T00:00:00Z" {
			t.Fatalf("unexpected created_after: %q", got)
		}
		if got := query.Get("created_before"); got != "2026-02-01T00:00:00Z" {
			t.Fatalf("unexpected created_before: %q", got)
		}
		if got := query.Get("updated_after"); got != "2026-01-05T00:00:00Z" {
			t.Fatalf("unexpected updated_after: %q", got)
		}

		writeJSON(t, writer, NoteListResponse{
			Items: []NoteListItem{},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "mono_pat_test", server.Client())
	_, err := client.ListNotes(context.Background(), ListNotesParams{
		Limit:         10,
		Cursor:        "cursor_1",
		Query:         "customer interview",
		TagIDs:        []string{"tag_1", "tag_2"},
		CreatedAfter:  "2026-01-01T00:00:00Z",
		CreatedBefore: "2026-02-01T00:00:00Z",
		UpdatedAfter:  "2026-01-05T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("ListNotes returned error: %v", err)
	}
}

func TestListAndGetNotesDecodeRecordedAt(t *testing.T) {
	t.Parallel()

	const recordedAt = "2026-07-20T08:15:30Z"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/v1/public-api/notes":
			fmt.Fprint(writer, `{"items":[{"note_id":"note_1","recorded_at":"`+recordedAt+`","created_at":"2026-07-20T08:16:00Z","updated_at":"2026-07-20T08:17:00Z"}]}`)
		case "/v1/public-api/notes/note_1":
			fmt.Fprint(writer, `{"note_id":"note_1","recorded_at":"`+recordedAt+`","created_at":"2026-07-20T08:16:00Z","updated_at":"2026-07-20T08:17:00Z"}`)
		default:
			t.Fatalf("unexpected path: %q", request.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "mono_pat_test", server.Client())
	list, err := client.ListNotes(context.Background(), ListNotesParams{})
	if err != nil {
		t.Fatalf("ListNotes returned error: %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].RecordedAt == nil || list.Items[0].RecordedAt.Format(time.RFC3339) != recordedAt {
		t.Fatalf("unexpected list recorded_at: %#v", list.Items)
	}

	note, err := client.GetNote(context.Background(), "note_1")
	if err != nil {
		t.Fatalf("GetNote returned error: %v", err)
	}
	if note.RecordedAt == nil || note.RecordedAt.Format(time.RFC3339) != recordedAt {
		t.Fatalf("unexpected detail recorded_at: %#v", note.RecordedAt)
	}
}

func TestListNotesDecodesNullRecordedAt(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, `{"items":[{"note_id":"note_1","recorded_at":null,"created_at":"2026-07-20T08:16:00Z","updated_at":"2026-07-20T08:17:00Z"}]}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "mono_pat_test", server.Client())
	list, err := client.ListNotes(context.Background(), ListNotesParams{})
	if err != nil {
		t.Fatalf("ListNotes returned error: %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].RecordedAt != nil {
		t.Fatalf("unexpected list recorded_at: %#v", list.Items)
	}
}

func TestListAllNotesFollowsCursorPagination(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cursor := request.URL.Query().Get("cursor")
		switch cursor {
		case "":
			next := "cursor_2"
			writeJSON(t, writer, NoteListResponse{
				Items:      []NoteListItem{{NoteID: "note_1"}},
				NextCursor: &next,
			})
		case "cursor_2":
			writeJSON(t, writer, NoteListResponse{
				Items: []NoteListItem{{NoteID: "note_2"}},
			})
		default:
			t.Fatalf("unexpected cursor: %q", cursor)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "mono_pat_test", server.Client())
	response, err := client.ListAllNotes(context.Background(), ListNotesParams{Limit: 1})
	if err != nil {
		t.Fatalf("ListAllNotes returned error: %v", err)
	}
	if response.Count != 2 {
		t.Fatalf("unexpected count: %d", response.Count)
	}
	if len(response.Items) != 2 {
		t.Fatalf("unexpected number of items: %d", len(response.Items))
	}
	if response.Items[0].NoteID != "note_1" || response.Items[1].NoteID != "note_2" {
		t.Fatalf("unexpected items: %#v", response.Items)
	}
}

func TestListNotesDecodesTags(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
			"items": [{
				"note_id": "note_1",
				"title": "Planning note",
				"summary": null,
				"tags": [{"tag_id": "tag_1", "name": "Planning", "source": "auto", "confidence": 0.9}],
				"created_at": "2026-07-13T00:00:00Z",
				"updated_at": "2026-07-13T00:00:00Z"
			}],
			"next_cursor": null
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "mono_pat_test", server.Client())
	response, err := client.ListNotes(context.Background(), ListNotesParams{})
	if err != nil {
		t.Fatalf("ListNotes returned error: %v", err)
	}
	if len(response.Items) != 1 || len(response.Items[0].Tags) != 1 {
		t.Fatalf("unexpected items: %#v", response.Items)
	}
	if got := response.Items[0].Tags[0]; got.TagID != "tag_1" || got.Name != "Planning" {
		t.Fatalf("unexpected tag: %#v", got)
	}
}

func TestGetNoteDecodesTags(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
			"note_id": "note_1",
			"title": "Meeting note",
			"summary": null,
			"transcript": null,
			"transcript_segments": null,
			"tags": [{"tag_id": "tag_2", "name": "Meetings", "source": "user", "confidence": null}],
			"created_at": "2026-07-13T00:00:00Z",
			"updated_at": "2026-07-13T00:00:00Z"
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "mono_pat_test", server.Client())
	note, err := client.GetNote(context.Background(), "note_1")
	if err != nil {
		t.Fatalf("GetNote returned error: %v", err)
	}
	if len(note.Tags) != 1 {
		t.Fatalf("unexpected tags: %#v", note.Tags)
	}
	if got := note.Tags[0]; got.TagID != "tag_2" || got.Name != "Meetings" {
		t.Fatalf("unexpected tag: %#v", got)
	}
}

func writeJSON(t *testing.T, writer http.ResponseWriter, value interface{}) {
	t.Helper()

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}
