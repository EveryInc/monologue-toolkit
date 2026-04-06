# monologue-toolkit

`monologue-toolkit` gives Monologue users two ways to work with their notes outside the app:

- `monologue`, a CLI for the Monologue Notes public API
- `monologue-notes`, an installable agent skill that uses the CLI from tools like Codex and Claude Code

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

## What you can do

With the CLI:

```bash
monologue notes list --limit 10
monologue notes list --q "customer interview"
monologue notes all --updated-after 2026-01-01T00:00:00Z
monologue notes get NOTE_ID
monologue notes get NOTE_ID --field transcript
```

With the skill, an agent can do things like:

- pull your latest notes
- search notes about a topic
- fetch a transcript for a specific note
- summarize notes without manually copying JSON around

## Get a Monologue API key

Before the CLI or skill can access your notes, you need a Monologue Notes API key.

In Monologue, go to the API settings screen and create a new key.

- On macOS, this appears under the Notes settings API section.
- In the iOS app, this appears in the API screen.
- The apps describe these as personal keys for the Notes API and MCP server.

Keep the generated token somewhere safe. You will paste it into the CLI during onboarding.

## Install the CLI

### Recommended today

The supported install path today is Go:

```bash
go install github.com/EveryInc/monologue-toolkit/cli/cmd/monologue@latest
```

Make sure your Go bin directory is on `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Add that to `~/.zshrc` or your shell profile if needed.

### Build from a local checkout

```bash
git clone https://github.com/EveryInc/monologue-toolkit.git
cd monologue-toolkit
go build ./cli/cmd/monologue
```

### If you do not have Go

Prebuilt releases are not wired up yet in this repo. For now, the CLI assumes Go is available.

The next distribution step for this repo should be GitHub Releases with prebuilt binaries for macOS, Linux, and Windows. Until then, `go install` is the supported path.

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

### List notes

```bash
monologue notes list --limit 10
monologue notes list --q "design review"
monologue notes list --created-after 2026-01-01T00:00:00Z
```

### Fetch every matching note across pagination

```bash
monologue notes all --q "customer"
```

### Fetch one note

```bash
monologue notes get NOTE_ID
monologue notes get NOTE_ID --field summary
monologue notes get NOTE_ID --field transcript
```

### Help

```bash
monologue --help
monologue notes --help
```

## Install the skill

The skill lives in [`monologue-notes-skill/`](./monologue-notes-skill).

The skill is intentionally simple:

- it does not bundle a separate wrapper CLI anymore
- it expects the `monologue` binary to already be installed
- it tells the agent to run `monologue onboarding` if credentials are missing

### Install in Codex locally

Copy the skill into your Codex skills directory:

```bash
mkdir -p ~/.codex/skills
cp -R ./monologue-notes-skill ~/.codex/skills/monologue-notes
```

Then restart Codex.

### Use the skill

After restarting Codex, prompt it naturally:

- "Use the monologue-notes skill and pull my latest notes"
- "Use monologue-notes to find notes about onboarding"
- "Use monologue-notes to fetch the transcript from my latest customer interview note"

### Other agents

The skill is written to be terminal-first, so the same pattern works in other agents that can:

- read a `SKILL.md`-style skill
- run shell commands
- call the installed `monologue` binary

When this repo is published and indexed by a skill directory such as [skills.sh](https://skills.sh), use the `monologue-notes-skill/` folder as the installable skill path.

## Current limitations

- The public Notes API is currently read-only.
- The current public API surface is notes list and note detail retrieval.
- The nicest install path today still requires Go.
- Prebuilt binaries and package-manager installs are not set up yet.

## Repository layout

```text
monologue-toolkit/
├── cli/                     # Go source for the monologue CLI
└── monologue-notes-skill/   # installable skill for terminal-capable agents
```

## Development notes

This repo includes a minimal [`.goreleaser.yaml`](./.goreleaser.yaml) so GitHub Releases can be added later without restructuring the project.
