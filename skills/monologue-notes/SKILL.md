---
name: "monologue-notes"
description: "Use when you need to read Monologue notes through the public API, search and list a user's notes, fetch a specific note, paginate through all notes, or pull transcripts and summaries with a Monologue Notes API token."
---

# Monologue Notes

Use this skill when an agent needs read-only access to Monologue Notes through the public API.

The current public surface is:

- `GET /v1/public-api/notes`
- `GET /v1/public-api/notes/{note_id}`

This skill is intentionally shell-first so it works across agents that can run terminal commands, including Codex and Claude Code.

## Setup

If the `monologue` CLI is missing, install it:

```bash
curl -fsSL https://raw.githubusercontent.com/EveryInc/monologue-toolkit/main/install.sh | sh
```

If credentials are not configured yet, run onboarding:

```bash
monologue onboarding
```

During onboarding, ask the user to create a Monologue Notes API token in the Monologue app and paste it into the terminal. The CLI saves it for future use.

For a non-interactive agent flow, the agent can ask the user for the token in chat and then run:

```bash
monologue onboarding --token "mono_pat_..."
```

## Commands

Use the CLI directly:

```bash
monologue notes list --limit 10
monologue notes list --q "customer interview"
monologue notes all --updated-after 2026-01-01T00:00:00Z
monologue notes get note_123
monologue notes get note_123 --field transcript
```

## Workflow

1. If the CLI is missing, install it.
2. If the user is not onboarded yet, run `monologue onboarding`.
3. Use `list` to find relevant notes.
4. Use `all` when the request spans more than one page.
5. Use `get NOTE_ID` for the full note payload.
6. Use `--field transcript` or `--field summary` when only one field is needed.

## Response style

- Prefer a short digest over raw JSON when the user asks for recent notes or summaries.
- Keep `note_id` internal unless the user explicitly asks for IDs or a follow-up action needs one.
- Surface titles, dates, summaries, and transcript highlights first.

## References

- Read `references/api.md` for the exact published schema.
- Read `references/install.md` when the CLI is missing or the environment needs setup help.

## Guardrails

- This skill is read-only. Do not invent create, update, or delete endpoints.
- Use ISO 8601 timestamps for `created-after`, `created-before`, and `updated-after`.
- Prefer the CLI over ad hoc `curl` so auth, errors, and pagination stay consistent.
- Skill installation itself should not be treated as a post-install execution hook. Install the CLI and run onboarding on first use instead.

