# CLI

The `monologue` binary is a thin Go client for Monologue's public Notes API.

## Commands

- `monologue version`
- `monologue onboarding`
- `monologue notes list`
- `monologue notes all`
- `monologue notes get NOTE_ID`
- `monologue notes audio-url NOTE_ID`

## Setup

Install from GitHub Releases:

```bash
curl -fsSL https://raw.githubusercontent.com/EveryInc/monologue-toolkit/main/install.sh | sh
```

Or install with Go:

```bash
go install github.com/EveryInc/monologue-toolkit/cli/cmd/monologue@latest
```

In your own interactive terminal, run onboarding once to save your token:

```bash
monologue onboarding
```

Do not paste an API token into an agent chat. If an agent discovers that
credentials are missing, it must pause and ask you to run `monologue onboarding`
yourself.

Saved config lives in your user config directory under `monologue/config.json`.

Environment overrides remain available for trusted automation:

- `MONOLOGUE_API_BASE_URL`
- `MONOLOGUE_API_TOKEN`

## Reading notes

Both placements of `--field` are supported:

```bash
monologue notes get NOTE_ID --field transcript
monologue notes get --field transcript NOTE_ID
```

Prefer the saved onboarding configuration. The existing `--token` flag remains
available for trusted automation, but avoid placing secrets in shell history or
agent tool calls.

Filter by one or more tags by repeating `--tag-id`:

```bash
monologue notes list --tag-id TAG_UUID
monologue notes all --tag-id TAG_UUID_1 --tag-id TAG_UUID_2
```

## Local development

```bash
go test ./...
go build ./cli/cmd/monologue
```

## Original recordings

Requires CLI v0.3.0 or later. `notes get NOTE_ID` includes `recording_url`,
`recording_url_expires_at` (UTC), `recording_content_type`, and `recording_bytes`.
The content type and byte count may be `null`. List and pagination output stay
unchanged. Existing personal API keys with `notes:read` work for recordings;
only notes owned by the key's user are accessible.

```bash
monologue notes get NOTE_ID --field recording_url
monologue notes get NOTE_ID --field recording_url_expires_at
monologue notes audio-url NOTE_ID
monologue notes audio-url NOTE_ID --field audio_url
```

`audio-url` requests a fresh signed link and returns `audio_url` and `expires_in`
(currently 3600 seconds). It accepts the same `--base-url`, `--token`, and `--field`
flags as `get`, with flags before or after the note ID. Prefer saved credentials.

Download the original uploaded file after onboarding:

```bash
recording_url="$(monologue notes audio-url NOTE_ID --field audio_url)" &&
  curl --fail --location --output recording.audio "$recording_url"
```

The file keeps its original format; choose an extension based on
`recording_content_type` if needed. The CLI returns links and metadata; `curl`
downloads the bytes. Do not send your API token to the download URL.

Links expire after one hour. Request a fresh link after expiry. Anyone with a
signed link can download until it expires, even if the API key is revoked in the
meantime, so keep links private and avoid saving them in logs or shared output.
A successful link response does not guarantee the stored file is still available;
handle download failures separately from API errors.
