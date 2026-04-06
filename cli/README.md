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

Run onboarding once to save your token:

```bash
monologue onboarding
```

You can also pass a token non-interactively:

```bash
monologue onboarding --token "mono_pat_..."
```

Saved config lives in your user config directory under `monologue/config.json`.

Environment variables still override saved config:

- `MONOLOGUE_API_TOKEN`
- `MONOLOGUE_API_BASE_URL`

## Local development

```bash
go test ./...
go build ./cli/cmd/monologue
```
