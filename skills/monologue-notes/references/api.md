# Monologue Public Notes API

Source: `https://api.monologue.to/public-openapi.json`
Live specification verified: 2026-07-20
OpenAPI version: 3.1.0
API version: 0.1.0

This reference covers the read-only Notes endpoints used by this skill. The
same public OpenAPI document also describes incoming webhook events, which are
outside this skill's read-only retrieval workflow.

## Authentication

- Security scheme: HTTP bearer token
- Required token scope: `notes:read`
- HTTPS base URL used by the CLI: `https://api.monologue.to`

Credentials must come from interactive `monologue onboarding`. Never ask for a
token in chat or include it in an agent command or tool argument.

## `GET /v1/public-api/notes`

Returns one cursor-paginated page of the authenticated user's notes.

Query parameters:

- `limit`: integer, default `20`, minimum `1`, maximum `100`
- `cursor`: opaque string returned by the previous page
- `created_after`: only return notes created after this ISO 8601 date-time
- `created_before`: only return notes created before this ISO 8601 date-time
- `updated_after`: only return notes updated after this ISO 8601 date-time
- `q`: string search across titles, summaries, and transcripts
- `tag_id`: repeatable UUID parameter; returns notes assigned to any supplied
  tag ID. In the CLI, repeat `--tag-id TAG_UUID` for multiple tags.

Successful response:

```json
{
  "items": [
    {
      "note_id": "8f14e45f-ceea-467f-a34e-4a3d1e1a89b2",
      "title": "Standup notes",
      "summary": "Daily standup covering the launch.",
      "tags": [
        {
          "tag_id": "string",
          "name": "Work",
          "position": 0,
          "created_at": "2026-07-20T00:00:00Z",
          "updated_at": "2026-07-20T00:00:00Z",
          "source": "user",
          "confidence": null
        }
      ],
      "recorded_at": "2026-07-20T00:00:00Z",
      "created_at": "2026-07-20T00:00:05Z",
      "updated_at": "2026-07-20T00:01:00Z"
    }
  ],
  "next_cursor": "string or null"
}
```

`title`, `summary`, and `recorded_at` are optional and may be `null`. `tags` is
an optional array. `next_cursor` may be absent or `null` when there is no next
page.

Documented errors:

- `400`: invalid filter or cursor
- `401`: missing or invalid personal API token
- `403`: token is missing the `notes:read` scope
- `422`: request validation error

## `GET /v1/public-api/notes/{note_id}`

Returns the full details for one note owned by the authenticated user.
`note_id` is a UUID path parameter.

The response contains the list-item fields plus:

- `transcript`: full transcript string or `null`
- `transcript_segments`: array of loosely structured JSON objects or `null`

Documented errors:

- `401`: missing or invalid personal API token
- `403`: token is missing the `notes:read` scope
- `404`: note does not exist for the authenticated user
- `422`: request validation error

## Retrieval notes

- The Notes API is read-only; there are no public create, update, or delete note
  endpoints in the verified specification.
- Pagination is cursor based. Treat cursors as opaque.
- Search is candidate discovery, not participant verification: `q` searches
  transcript text and can match a note that only mentions a person or topic.
- `recorded_at` is the recording start time supplied by the client; `created_at`
  is when the backend created the note. Prefer `recorded_at` for human-facing
  chronology when present, while retaining `created_at` for provenance.
- The API does not expose a `recorded_at` filter. For exhaustive human recording
  ranges, retrieve matching candidates without creation-time bounds and filter
  their `recorded_at` values locally. Fall back to `created_at` only when
  `recorded_at` is absent, and disclose that fallback.
