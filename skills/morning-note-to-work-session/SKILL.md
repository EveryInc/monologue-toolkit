---
name: morning-note-to-work-session
description: "Turn a recent Monologue morning note into a focused, evidence-backed work-session brief. Use when the user asks what to work on today, wants to start from a morning voice note, or wants decisions, questions, today's tasks, later ideas, and a recommended first task extracted from recent notes."
---

# Morning Note to Work Session

Use `$monologue-notes` for read-only note discovery and retrieval. Produce a brief first; do not edit files, create tasks, message people, or start implementation until the user explicitly approves a next action.

## Workflow

1. Resolve relative dates such as "today" in the user's timezone. State the resulting date when it affects the search.
2. Search using several likely phrases, not only `morning note`. Do not use
   `created-after` or `created-before` to represent the recording day: those
   flags filter backend creation time and can omit a delayed upload. For example:

   ```bash
   monologue notes all --q "morning"
   monologue notes all --q "today"
   ```

   Filter candidates into the requested local day using `recorded_at`. If it is
   absent, use `created_at` as a disclosed fallback. If no candidate appears,
   expand to the previous seven recording days and say that the range changed.
3. Treat search results as candidates, not proof. Fetch the most plausible notes and inspect the full payload and transcript:

   ```bash
   monologue notes get NOTE_ID
   ```

4. Select the latest note whose content is actually a morning planning or reflection note. Prefer content relevance over a matching title. If two candidates are similarly plausible, present their titles and dates and ask the user which one to use.
5. Protect the work/personal boundary. Include personal material only when the user requested it or when it directly changes today's availability; otherwise summarize it as a constraint without sensitive detail.
6. Extract without inventing commitments:
   - **Decisions** — choices the speaker made
   - **Open questions** — unresolved decisions or missing information
   - **Today** — explicit work intended for today
   - **Later** — ideas, follow-ups, and non-today work
7. Recommend one first task based on explicit urgency, dependencies, and leverage. Label this recommendation as an inference and explain it in one sentence. Do not convert vague reflection into a firm task.
8. Return the brief and wait for the user before modifying anything.

## Output

Lead with the source as `Note title — YYYY-MM-DD` and then use:

```text
Decisions
Open questions
Today
Later
Recommended first task (inference)
```

Attach the source note title and date to material claims when more than one note is used. Clearly label direct statements as evidence and prioritization or interpretation as inference. If no verified morning note is found, report the searches tried and ask for a title, date, or phrase rather than substituting an unrelated note.

## Guardrails

- Keep Monologue access read-only through `$monologue-notes`.
- Never request or accept an API token in chat; have the user run local onboarding if needed.
- Do not expose note IDs unless needed for troubleshooting or explicitly requested.
- Do not silently include personal recordings in a work brief.
- Do not perform the recommended task until the user explicitly asks.
