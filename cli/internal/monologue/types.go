package monologue

import "time"

type ListNotesParams struct {
	Limit         int
	Cursor        string
	Query         string
	CreatedAfter  string
	CreatedBefore string
	UpdatedAfter  string
}

type NoteTag struct {
	TagID string `json:"tag_id"`
	Name  string `json:"name"`
}

type NoteListItem struct {
	NoteID    string    `json:"note_id"`
	Title     *string   `json:"title"`
	Summary   *string   `json:"summary"`
	Tags      []NoteTag `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NoteListResponse struct {
	Items      []NoteListItem `json:"items"`
	NextCursor *string        `json:"next_cursor"`
}

type Note struct {
	NoteID             string      `json:"note_id"`
	Title              *string     `json:"title"`
	Summary            *string     `json:"summary"`
	Transcript         *string     `json:"transcript"`
	TranscriptSegments interface{} `json:"transcript_segments"`
	Tags               []NoteTag   `json:"tags"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

type NoteListAllResponse struct {
	Items []NoteListItem `json:"items"`
	Count int            `json:"count"`
}
