# Installing and onboarding the Monologue CLI

## Preferred

Install from GitHub Releases:

```bash
curl -fsSL https://raw.githubusercontent.com/EveryInc/monologue-toolkit/main/install.sh | sh
```

## Alternative

Install with Go:

```bash
go install github.com/EveryInc/monologue-toolkit/cli/cmd/monologue@latest
```

## Onboarding

The user must run this command in their own interactive terminal:

```bash
monologue onboarding
```

The CLI stores credentials in your user config directory under `monologue/config.json`.

An agent must never ask for the token in chat or put it in a command, tool call,
log, or response. If onboarding is required, pause and wait for the user to
confirm that the interactive command completed. There is intentionally no
agent-driven or non-interactive onboarding flow.

For development against another environment, the user may override only the
base URL without exposing credentials:

```bash
export MONOLOGUE_API_BASE_URL="https://api.monologue.to"
```
