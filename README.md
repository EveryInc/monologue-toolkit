# monologue-toolkit

`monologue-toolkit` gives Monologue users two ways to work with their notes outside the app:

- `monologue`, a CLI for the Monologue Notes public API
- installable agent skills for tools like Codex, Claude Code, and other terminal-capable agents

The CLI and `monologue-notes` foundation skill provide read-only access to the public Notes API:

- list notes
- search notes
- fetch a specific note
- pull summaries and transcripts

Six workflow skills turn that source material into focused work: starting the day, processing a voice inbox, synthesizing a topic, drafting a weekly update, cleaning a transcript, and translating a customer call into product work.

## Skill catalog

| Skill | Scope | Use it to |
| --- | --- | --- |
| `monologue-notes` | Access | Find and read notes, summaries, and transcripts |
| `morning-note-to-work-session` | Single note | Turn the latest morning note into a focused work session |
| `customer-call-to-product-work` | Single note + tools | Turn verified customer evidence into a bug report, issue draft, or implementation plan |
| `topic-synthesis` | Multiple notes | Synthesize an idea across multiple notes with a source trail |
| `weekly-work-update` | Multiple notes | Draft a sourced weekly update without pulling in personal notes |
| `voice-inbox` | Multiple notes | Sort recent notes into decisions, tasks, ideas, questions, and follow-ups |
| `clean-transcript` | Supplied text or single note | Remove verbal noise while preserving meaning, uncertainty, and speaker voice |

## Quick start

1. Install the CLI.
2. Create a Monologue Notes API key in the Monologue app.
3. Run `monologue onboarding`.
4. Start using `monologue notes ...`.
5. If you use an agent, install one workflow skill or the complete collection.

## Get a Monologue API key

Before the CLI or skill can access your notes, you need a Monologue Notes API key.

In Monologue, go to the API settings screen and create a new key.

- On macOS, this appears under the Notes settings API section.
- In the iOS app, this appears in the API screen.
- The apps describe these as personal keys for the public Notes API.

Keep the generated token somewhere safe. Enter it only into the local CLI onboarding prompt. Never paste it into an agent chat.

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
monologue notes get --field transcript NOTE_ID
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

## Install the skills

The current Skills CLI requires Node.js 22.20 or newer.

Install the complete collection globally for all supported agents:

```bash
npx skills add https://github.com/EveryInc/monologue-toolkit --all -g
```

To choose skills and target agents interactively:

```bash
npx skills add https://github.com/EveryInc/monologue-toolkit -g
```

Install one workflow together with its `monologue-notes` foundation:

```bash
npx skills add https://github.com/EveryInc/monologue-toolkit \
  --skill monologue-notes \
  --skill topic-synthesis \
  -g -y
```

These commands use the open [Skills CLI](https://github.com/vercel-labs/skills). Workflow skills that retrieve Monologue content expect `monologue-notes` to be installed too, so installing the complete collection is the simplest path.

After installing the skill:

1. Make sure the `monologue` CLI is already installed.
2. Run `monologue onboarding` once.
3. Restart your agent if it does not auto-refresh installed skills.

### Use the skills

Prompt your agent naturally:

- "Use morning-note-to-work-session to turn my latest morning note into a plan for today."
- "Use customer-call-to-product-work to draft a bug report from yesterday's Acme call."
- "Use topic-synthesis to show how my thinking about onboarding changed this month."
- "Use weekly-work-update to draft my update for this week."
- "Use voice-inbox to process my notes from the last three days."
- "Use clean-transcript to polish this interview transcript without rewriting it."

Retrieval skills cite the note title and date, verify likely search matches before relying on them, and separate direct evidence from inference. They treat Monologue as read-only. Any action in another system remains a draft until the user explicitly approves the write.

### Update the skill

Use the Skills CLI:

```bash
npx skills check
npx skills update
```

### Alternative skill: lightweight, direct API instructions

For standalone note retrieval without the CLI, the third-party [`monologue-notes-api`](https://skills.sh/intellectronica/agent-skills/monologue-notes-api) skill gives an agent instructions for calling the API directly with `curl` or an ad hoc script:

```bash
npx skills add https://github.com/intellectronica/agent-skills --skill monologue-notes-api
```

Follow that skill's credential setup in your own terminal and never paste the key into agent chat. The workflow skills in this repository use the `monologue` CLI and `$monologue-notes` instead.

## Build or remix a workflow

Create a skill when a voice workflow is repeated, has a stable outcome, or needs a specific retrieval procedure. Keep one-off instructions as prompts and simple repeated text as snippets.

Start from the closest skill in `skills/`, then change its name, triggering description, workflow, output contract, and `agents/openai.yaml`. A Monologue workflow skill should:

1. Establish the requested scope, date range, and timezone.
2. Select human date ranges by `recorded_at`, with a disclosed `created_at`
   fallback when recording time is absent.
3. Search with more than one useful phrase when recall may be incomplete.
4. Fetch and verify candidate notes instead of trusting a title or search hit.
5. Separate direct evidence, inference, and unresolved questions.
6. Cite each source by note title and date.
7. Protect work/personal boundaries and attribute commitments to the right speaker.
8. Draft external actions first and wait for explicit approval before writing.

Check local skill discovery before opening a pull request:

```bash
scripts/validate-skills.sh
npx skills add . --list
```

## Repository layout

```text
monologue-toolkit/
├── cli/                     # Go source for the monologue CLI
└── skills/                  # base access and voice workflow skills
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

- The public Notes API and Monologue CLI are currently read-only.
- The current public API surface is notes list and note detail retrieval.
- Workflow skills can inspect other tools already available to an agent, but must ask before creating or changing anything outside Monologue.
- The no-Go installer depends on GitHub Releases existing for the requested version.
