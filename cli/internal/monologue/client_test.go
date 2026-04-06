package monologue

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListNotesSendsAuthAndFilters(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("Authorization"); got != "Bearer mono_pat_test" {
			t.Fatalf("unexpected auth header: %q", got)
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
		CreatedAfter:  "2026-01-01T00:00:00Z",
		CreatedBefore: "2026-02-01T00:00:00Z",
		UpdatedAfter:  "2026-01-05T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("ListNotes returned error: %v", err)
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

func writeJSON(t *testing.T, writer http.ResponseWriter, value interface{}) {
	t.Helper()

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}
