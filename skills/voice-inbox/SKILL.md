---
name: voice-inbox
description: Process recent Monologue voice notes into a private, sourced inbox of decisions, tasks, ideas, questions, and people follow-ups. Use when a user wants to review recent recordings, clear a voice-note backlog, find commitments, or organize captured thoughts across work and personal contexts.
---

# Voice Inbox

Turn recent voice notes into a reviewable inbox without silently promoting thoughts into commitments or taking external action.

## Workflow

1. Determine the time window and the user's timezone. Resolve relative dates in that timezone and state the exact interval. If the user only says "recent," use the previous seven calendar days through now and say so.
2. Invoke `$monologue-notes` for all Monologue access. If setup is required, have the user run `monologue onboarding` in their own terminal; never request or accept an API token in chat.
3. Retrieve the complete note list with pagination. Do not use creation-time
   filters as recording-time filters; delayed uploads can be created later:

   ```bash
   monologue notes all
   ```

   Select the requested interval locally using `recorded_at` in the user's
   timezone. If it is absent, use `created_at` as a disclosed fallback. If the
   library is too large for exhaustive retrieval, ask before applying a
   creation-time optimization and disclose the resulting coverage limit.

4. Review titles, dates, and summaries to identify candidates, then fetch and verify the likely sources:

   ```bash
   monologue notes get NOTE_ID
   ```

   Treat search matches as candidates. A name or phrase appearing in a result does not establish that the note is about that subject.
5. Classify each verified item as a decision, task, idea, question, or people follow-up. Keep work and personal material in separate sections. Put genuinely ambiguous items in `Needs review` rather than guessing.
6. Deduplicate repeated items while retaining every supporting source. Where recordings conflict, preserve the latest explicit statement and flag the conflict.
7. Return a reviewable draft using the format below. Cite each item with note title and local date: `(Note title — YYYY-MM-DD)`.

## Classification Rules

- **Decision:** A choice the user or identified group explicitly made.
- **Task:** An action the user explicitly committed to doing. Do not turn another speaker's commitment, a suggestion, a possibility, or a general desire into the user's task.
- **Idea:** A possibility worth retaining that has not been committed to.
- **Question:** An unresolved question or information gap.
- **People follow-up:** A person the user explicitly said they would contact or follow up with. Otherwise label it `Suggested follow-up`, not a commitment.

When speaker identity is uncertain, do not assign the item to the user.

## Output

```markdown
# Voice inbox — DATE–DATE (TIMEZONE)

## Work
### Decisions
- Item. (Source — date)
### Tasks
- [ ] Action — timing, if explicitly stated. (Source — date)
### Ideas
- Item. (Source — date)
### Questions
- Item. (Source — date)
### People follow-ups
- Person — purpose — `Committed` or `Suggested`. (Source — date)

## Personal
[Use the same categories, omitting empty ones]

## Needs review
- Ambiguous item and what needs confirmation. (Source — date)

## Sources
- Note title — date
```

Omit empty headings. Keep sensitive details only where necessary to understand the item; do not echo unrelated private material.

## Action Boundary

- Keep note IDs internal unless the user requests them.
- Do not create tasks, contact people, send messages, or write to another system automatically.
- If the user wants follow-through, show the exact proposed action or draft first and obtain explicit approval before performing it.
