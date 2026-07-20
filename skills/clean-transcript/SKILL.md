---
name: clean-transcript
description: Clean a supplied transcript, transcript file, or verified Monologue note without changing its meaning. Use when a user wants to remove filler, stutters, repeated phrases, or abandoned false starts; improve punctuation and paragraphing; preserve speaker labels; or prepare a readable transcript while retaining uncertainty, technical details, numbers, and quotations.
---

# Clean Transcript

Produce a more readable transcript while treating the source as evidence. Clean speech artifacts; do not rewrite the speaker's ideas or invent missing speech.

## Choose the Source

- **Pasted text or file:** Use the supplied content directly and identify the filename when available.
- **Monologue note:** Invoke `$monologue-notes`, search by the user's description and date range, then verify the candidate's title, date, and content before cleaning it. If setup is required, have the user run `monologue onboarding` in their own terminal; never request or accept an API token in chat.

  ```bash
  monologue notes all --q "SEARCH TERMS"
  monologue notes get NOTE_ID
  ```

Resolve relative dates using the user's timezone and state the resolved date or
interval. Select by `recorded_at`, not backend `created_at`; if `recorded_at` is
absent, use `created_at` as a disclosed fallback. A search match is only a
candidate; if multiple notes plausibly match, present the shortlist or ask a
focused question rather than choosing silently.

## Choose the Cleaning Mode

Use `Standard` unless the user requests another mode:

- **Light:** Correct punctuation, capitalization, paragraphing, and exact accidental repetitions only.
- **Standard:** Also remove filler words, stutters, and abandoned false starts when the completed thought is clear.
- **Speaker-formatted:** Apply Standard cleaning and organize the result under the existing speaker labels. Do not invent speaker identities.

These modes change cleanup intensity, not substance. Summarization, copyediting, or turning the transcript into an article is a separate task and requires an explicit request.

## Preserve Meaning

- Preserve claims, intent, tone, order, and level of certainty.
- Keep meaningful hedges such as `I think`, `maybe`, `probably`, and `I'm not sure`; they are not filler when they qualify a claim.
- Preserve negation, corrections, technical terms, names, code identifiers, URLs, dates, quantities, units, and numbers exactly unless an obvious transcription error can be verified from context.
- Preserve quoted speech. Do not silently polish words attributed to another person.
- Preserve speaker labels and timestamps unless the user asks to remove timestamps.
- When a false start is followed by an unambiguous correction, keep the corrected statement. When the intended meaning is ambiguous, retain the original wording and flag it for review.
- Never fill an inaudible passage or reconstruct words absent from the source. Retain an existing marker such as `[inaudible]`; otherwise use `[unclear]` only to flag ambiguity, not as invented transcript content.

## Output

Return:

1. `Cleaned transcript` in the selected mode.
2. `Needs review` only when unresolved names, numbers, quotations, speaker identities, or ambiguous passages could materially affect meaning.
3. `Source`:
   - For Monologue: `Note title — YYYY-MM-DD` in the user's timezone.
   - For a supplied file: its filename.
   - For pasted text: `User-supplied transcript`.

Do not expose a Monologue note ID unless the user asks for it. Do not overwrite the source file or note unless the user explicitly requests that separate action.
