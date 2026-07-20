---
name: "monologue-notes"
description: "Use when you need to read Monologue notes through the public API, search and list a user's notes, fetch a specific note, paginate through all notes, or pull transcripts and summaries through the Monologue CLI."
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

Onboarding is a human-only step. Ask the user to open their own interactive
terminal, create a Monologue Notes API token in the Monologue app, and run
`monologue onboarding` themselves. Pause until they confirm it completed.

Never ask the user to paste a token into chat. Never place a token in a tool
call, command argument, transcript, log, or response. Do not attempt to automate
onboarding or work around the interactive-terminal check.

## Commands

Use the CLI directly:

```bash
monologue notes list --limit 10
monologue notes list --q "customer interview"
monologue notes list --tag-id TAG_UUID
monologue notes all --updated-after 2026-01-01T00:00:00Z
monologue notes get note_123
monologue notes get note_123 --field transcript
monologue notes get --field summary note_123
```

## Retrieval protocol

1. If the CLI is missing, install it.
2. If credentials are missing, stop and ask the user to run `monologue onboarding`
   in their own terminal. Continue only after they confirm completion.
3. Establish scope before searching: subject, people, project, requested date
   range, and the user's timezone. Resolve relative dates such as "last week"
   in that timezone, convert the boundaries to ISO 8601, and state the resolved
   range when it affects the answer.
4. Match human recording ranges against `recorded_at`, not `created_at`.
   `created-after` and `created-before` filter when the backend created a note,
   so they can omit delayed uploads. For an exhaustive recording-time request,
   retrieve all matching candidates and filter their `recorded_at` values
   locally. If `recorded_at` is absent, use `created_at` only as an explicit,
   disclosed fallback. Use creation filters only when the user asks about note
   creation or accepts a stated coverage optimization.
5. Search with more than one focused query. Try topic, person, project, and
   distinctive phrase variants rather than trusting one broad query. Use `all`
   when relevant results may span multiple pages.
6. Treat list/search matches only as candidates. A result may merely mention a
   person or topic. Fetch each plausible candidate with `get NOTE_ID`, inspect
   its transcript and metadata, and verify that the date, participants, and
   subject actually match the request.
7. Use `--field transcript` or `--field summary` only after a candidate has been
   identified. Fetch the full payload when title, timestamps, tags, or transcript
   segments are needed for verification or provenance.
8. If the first search finds nothing, broaden it once or twice by relaxing one
   dimension at a time. Report the range and queries checked; do not turn an
   incomplete search into a claim that no note exists.
9. Synthesize only verified sources. Keep work and personal material separate,
   and follow the privacy and attribution rules below.

## Response style

- Prefer a short digest over raw JSON when the user asks for recent notes or summaries.
- Keep `note_id` internal unless the user explicitly asks for IDs or a follow-up action needs one.
- Give every source a human-readable provenance label: note title and date in
  the user's timezone. Add the recorded time when it disambiguates similarly
  named notes.
- Clearly label direct transcript evidence versus your synthesis or inference.
  Use quotation marks only for words verified in the transcript; otherwise
  paraphrase.
- For multi-note answers, finish with a compact source trail listing the notes
  actually used. Do not cite candidates that were inspected and rejected.

## References

- Read `references/api.md` for the exact published schema.
- Read `references/install.md` when the CLI is missing or the environment needs setup help.

## Guardrails

- This skill is read-only. Do not invent create, update, or delete endpoints.
- Use ISO 8601 timestamps for `created-after`, `created-before`, and `updated-after`.
- Prefer the CLI over ad hoc `curl` so auth, errors, and pagination stay consistent.
- Skill installation itself should not be treated as a post-install execution hook. Install the CLI and run onboarding on first use instead.
- Never request, display, echo, log, or pass an API token in chat or tool
  arguments. Only the user may enter it during interactive onboarding.
- Do not silently mix personal notes into a work synthesis. Exclude clearly
  personal material unless the user explicitly includes it. If classification
  is ambiguous and inclusion would materially change the answer, ask first.
- Attribute decisions, promises, and action items to a speaker only when the
  transcript supports both the speaker identity and the commitment. Do not turn
  a suggestion, question, collective "we", or another person's plan into the
  user's commitment.
- Treat speaker labels as unverified unless the note or surrounding transcript
  identifies them. Say "an unidentified speaker" when identity is uncertain.
