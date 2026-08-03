package monologue

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/EveryInc/monologue-toolkit/cli/internal/version"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type APIError struct {
	StatusCode int
	Detail     string
	Body       string
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("monologue API returned HTTP %d: %s", e.StatusCode, e.Detail)
	}
	if e.Body != "" {
		return fmt.Sprintf("monologue API returned HTTP %d: %s", e.StatusCode, e.Body)
	}

	return fmt.Sprintf("monologue API returned HTTP %d", e.StatusCode)
}

func NewClient(baseURL string, token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: httpClient,
	}
}

func (c *Client) ListNotes(ctx context.Context, params ListNotesParams) (NoteListResponse, error) {
	query := url.Values{}
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		query.Set("cursor", params.Cursor)
	}
	if params.Query != "" {
		query.Set("q", params.Query)
	}
	for _, tagID := range params.TagIDs {
		if tagID != "" {
			query.Add("tag_id", tagID)
		}
	}
	if params.CreatedAfter != "" {
		query.Set("created_after", params.CreatedAfter)
	}
	if params.CreatedBefore != "" {
		query.Set("created_before", params.CreatedBefore)
	}
	if params.UpdatedAfter != "" {
		query.Set("updated_after", params.UpdatedAfter)
	}

	endpoint := fmt.Sprintf("%s/v1/public-api/notes", c.baseURL)
	if encoded := query.Encode(); encoded != "" {
		endpoint = endpoint + "?" + encoded
	}

	var response NoteListResponse
	if err := c.doJSON(ctx, http.MethodGet, endpoint, &response); err != nil {
		return NoteListResponse{}, err
	}

	return response, nil
}

func (c *Client) ListAllNotes(ctx context.Context, params ListNotesParams) (NoteListAllResponse, error) {
	if params.Limit <= 0 {
		params.Limit = 100
	}

	allItems := make([]NoteListItem, 0)
	nextCursor := params.Cursor

	for {
		pageParams := params
		pageParams.Cursor = nextCursor

		response, err := c.ListNotes(ctx, pageParams)
		if err != nil {
			return NoteListAllResponse{}, err
		}

		allItems = append(allItems, response.Items...)
		if response.NextCursor == nil || *response.NextCursor == "" {
			break
		}
		nextCursor = *response.NextCursor
	}

	return NoteListAllResponse{
		Items: allItems,
		Count: len(allItems),
	}, nil
}

func (c *Client) GetNote(ctx context.Context, noteID string) (Note, error) {
	endpoint := fmt.Sprintf("%s/v1/public-api/notes/%s", c.baseURL, url.PathEscape(noteID))

	var response Note
	if err := c.doJSON(ctx, http.MethodGet, endpoint, &response); err != nil {
		return Note{}, err
	}

	return response, nil
}

func (c *Client) doJSON(ctx context.Context, method string, endpoint string, out interface{}) error {
	request, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return err
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("User-Agent", "monologue-toolkit/"+version.Current())

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &APIError{
			StatusCode: response.StatusCode,
			Detail:     parseErrorDetail(body),
			Body:       strings.TrimSpace(string(body)),
		}
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode monologue API response: %w", err)
	}

	return nil
}

func parseErrorDetail(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var envelope struct {
		Detail interface{} `json:"detail"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return ""
	}
	if envelope.Detail == nil {
		return ""
	}

	switch value := envelope.Detail.(type) {
	case string:
		return value
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			return ""
		}
		return string(encoded)
	}
}
