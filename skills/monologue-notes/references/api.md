# Monologue Public Notes API

Source: `https://api.monologue.to/public-openapi.json`
Fetched: 2026-04-06

## Authentication

- Security scheme: HTTP bearer token
- Recommended base URL: `https://api.monologue.to`

Header:

```http
Authorization: Bearer YOUR_PERSONAL_API_TOKEN
```

## Endpoints

### `GET /v1/public-api/notes`

Query parameters:

- `limit` integer, default `20`, min `1`, max `100`
- `cursor` string, opaque pagination cursor
- `created_after` ISO 8601 date-time
- `created_before` ISO 8601 date-time
- `updated_after` ISO 8601 date-time
- `q` string, searches note titles, summaries, and transcripts

Success shape:

```json
{
  "items": [
    {
      "note_id": "string",
      "title": "string or null",
      "summary": "string or null",
      "created_at": "2026-04-06T00:00:00Z",
      "updated_at": "2026-04-06T00:00:00Z"
    }
  ],
  "next_cursor": "string or null"
}
```

### `GET /v1/public-api/notes/{note_id}`

Returns the full note payload. The public response adds:

- `transcript`
- `transcript_segments`

## Notes

- The public Notes API is currently read-only.
- Pagination is cursor based.
- `transcript_segments` should be treated as loosely typed JSON.

