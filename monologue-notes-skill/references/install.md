# Installing and onboarding the Monologue CLI

## Preferred

Install with Go:

```bash
go install github.com/EveryInc/monologue-toolkit/cli/cmd/monologue@latest
```

## Fallback

Build from a local checkout:

```bash
go build ./cli/cmd/monologue
```

## Onboarding

```bash
monologue onboarding
```

For agent-driven or non-interactive setup:

```bash
monologue onboarding --token "mono_pat_..."
```

The CLI stores credentials in your user config directory under `monologue/config.json`.

Environment overrides still work when needed:

```bash
export MONOLOGUE_API_TOKEN="mono_pat_..."
export MONOLOGUE_API_BASE_URL="https://api.monologue.to"
```
