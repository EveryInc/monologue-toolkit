# CLI

The `monologue` binary is a thin Go client for Monologue's public Notes API.

## Commands

- `monologue version`
- `monologue onboarding`
- `monologue notes list`
- `monologue notes all`
- `monologue notes get NOTE_ID`

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
