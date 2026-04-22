# monologue-toolkit

`monologue-toolkit` gives Monologue users two ways to work with their notes outside the app:

- `monologue`, a CLI for the Monologue Notes public API
- `monologue-notes`, an installable agent skill for tools like Codex, Claude Code, and other terminal-capable agents

Today, both are read-only and focused on the public Notes API:

- list notes
- search notes
- fetch a specific note
- pull summaries and transcripts

## Quick start

1. Install the CLI.
2. Create a Monologue Notes API key in the Monologue app.
3. Run `monologue onboarding`.
4. Start using `monologue notes ...`.
5. If you use an agent, install the `monologue-notes` skill too.

## Get a Monologue API key

Before the CLI or skill can access your notes, you need a Monologue Notes API key.

In Monologue, go to the API settings screen and create a new key.

- On macOS, this appears under the Notes settings API section.
- In the iOS app, this appears in the API screen.
- The apps describe these as personal keys for the Notes API and MCP server.

Keep the generated token somewhere safe. You will paste it into the CLI during onboarding.

## Install the CLI

### macOS and Linux without Go

Once the repo has a tagged GitHub release, the easiest install path is:

```bash
curl -fsSL https://raw.githubusercontent.com/EveryInc/monologue-toolkit/main/install.sh | sh
```

Install to a custom location:

```bash
curl -fsSL https://raw.githubusercontent.com/EveryInc/monologue-toolkit/main/install.sh | sh -s -- --install-dir /usr/local/bin
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/EveryInc/monologue-toolkit/main/install.sh | sh -s -- --version v0.1.0
```

### Windows without Go

Once the repo has a tagged GitHub release:

```powershell
irm https://raw.githubusercontent.com/EveryInc/monologue-toolkit/main/install.ps1 | iex
```

### With Go

If you already have Go installed, you can still use:

```bash
go install github.com/EveryInc/monologue-toolkit/cli/cmd/monologue@latest
```

If you use `go install`, make sure your Go bin directory is on `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### If you are reading this before the first release

The no-Go install scripts rely on GitHub Releases. Until the first release is published, use `go install` or build from source.

## Onboard the CLI

Once `monologue` is installed, run:

```bash
monologue onboarding
```

The CLI will:

- ask you for your Monologue Notes API token
- optionally verify it against the API
- save it locally for future commands

You can also pass the token directly:

```bash
monologue onboarding --token "mono_pat_..."
```

After onboarding, test it with:

```bash
monologue notes list --limit 5
```

## CLI commands

```bash
monologue version
monologue notes list --limit 10
monologue notes list --q "customer interview"
monologue notes all --updated-after 2026-01-01T00:00:00Z
monologue notes get NOTE_ID
monologue notes get NOTE_ID --field transcript
```

Use `monologue --help` and `monologue notes --help` for the full command list.

## Update the CLI

If you installed with the shell or PowerShell installer, rerun the same install command to get the latest release.

If you installed with Go, rerun:

```bash
go install github.com/EveryInc/monologue-toolkit/cli/cmd/monologue@latest
```

Check the installed version with:

```bash
monologue version
```

## Install the skill

The easiest `skills.sh` install path is:

```bash
npx skills add https://github.com/EveryInc/monologue-toolkit --skill monologue-notes -g -y
```

That installs the skill globally for supported agents.

After installing the skill:

1. Make sure the `monologue` CLI is already installed.
2. Run `monologue onboarding` once.
3. Restart your agent if it does not auto-refresh installed skills.

### Use the skill

Prompt your agent naturally:

- "Use the monologue-notes skill and pull my latest notes"
- "Use monologue-notes to find notes about onboarding"
- "Use monologue-notes to fetch the transcript from my latest customer interview note"

### Update the skill

Use the Skills CLI:

```bash
npx skills check
npx skills update
```

### Alternative skill: lightweight, direct API instructions

As an alternative to installing the CLI and the matching skill, you can install [this skill](https://skills.sh/intellectronica/agent-skills/monologue-notes-api) which only includes directions
for an agent to call the API directly. You can use this skill to get an agent to make calls to the API using `curl` or by
writing an ad-hoc script.

```bash
npx skills add https://github.com/intellectronica/agent-skills --skill monologue-notes-api
export MONOLOGUE_API_KEY="..."
```

## Repository layout

```text
monologue-toolkit/
├── cli/                     # Go source for the monologue CLI
└── skills/monologue-notes/  # installable skill package
```

## Maintainer release flow

This repo now includes:

- `.github/workflows/ci.yml` for tests
- `.github/workflows/release.yml` for tagged releases
- `.goreleaser.yaml` for cross-platform CLI binaries
- `install.sh` and `install.ps1` for no-Go installs

To publish a release:

1. Push a semver tag such as `v0.1.0`.
2. GitHub Actions will build archives for macOS, Linux, and Windows.
3. The workflow will publish a GitHub Release with checksums.

After that, the no-Go install commands above will work for end users.

## Current limitations

- The public Notes API is currently read-only.
- The current public API surface is notes list and note detail retrieval.
- The no-Go installer depends on GitHub Releases existing for the requested version.
