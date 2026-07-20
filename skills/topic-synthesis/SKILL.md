---
name: topic-synthesis
description: "Synthesize what a user has said about a topic across multiple Monologue notes, including the central thesis, examples, evolution, contradictions, gaps, and a dated source trail. Use when the user asks to find a theme across recordings, compare their thinking over time, or assemble multi-note context for writing or decisions."
---

# Topic Synthesis

Use `$monologue-notes` for read-only discovery and retrieval. Build a traceable synthesis across verified notes; never hide conflicting evidence or silently drop relevant candidates.

## Workflow

1. Define the topic and resolve any relative date range in the user's timezone. State the interpreted range. If none is supplied, choose a reasonable range, label it, and offer to expand it.
2. Generate several search queries before retrieving notes:
   - the user's phrase
   - synonyms and abbreviations
   - named people, projects, products, or problems tied to the topic
   - likely opposing or earlier terminology
3. Run each query without treating creation time as recording time:

   ```bash
   monologue notes all --q "voice guide"
   monologue notes all --q "dictation guide"
   ```

   Apply the requested interval locally to `recorded_at` in the user's timezone.
   If it is absent, use `created_at` as a disclosed fallback. The API creation
   filters can omit delayed uploads and are not a substitute for this step.

4. Deduplicate candidates, then fetch and inspect each plausible transcript:

   ```bash
   monologue notes get NOTE_ID
   ```

   A title or keyword match is insufficient. Confirm substantive discussion of the topic and record the note title and date.
5. Maintain an inclusion log: included notes, excluded candidates with a short reason, and any candidate that could not be fetched. If the candidate set is too large, disclose the sampling rule and ask whether the user wants exhaustive retrieval.
6. Protect boundaries. Do not mix personal recordings into a work synthesis unless requested or substantively necessary. When necessary, use only the relevant point and avoid unrelated personal detail.
7. Synthesize in chronological as well as thematic order:
   - **Central thesis** — the best-supported through-line
   - **Supporting examples** — concrete recurring evidence
   - **Evolution** — what changed and when
   - **Contradictions or tensions** — unresolved disagreement, reversals, or competing priorities
   - **Gaps** — questions not answered by the recordings
8. Distinguish direct evidence from inference. Attribute each material direct claim to `Note title — YYYY-MM-DD`; label cross-note interpretations as synthesis or inference.
9. Report coverage and the source trail. Do not imply completeness when searches, access, date range, or sampling limited it.

## Output

Use this structure:

```text
Scope and coverage
Central thesis
Supporting examples
How the thinking evolved
Contradictions and tensions
Gaps and open questions
Source trail
Excluded or unavailable candidates
```

Keep quotes short and necessary. Prefer faithful paraphrase with source attribution. If no note is verified, report the queries and date range tried and ask for another phrase or known recording instead of manufacturing a synthesis.

## Guardrails

- Keep Monologue access read-only through `$monologue-notes`.
- Never request or accept an API token in chat; have the user run local onboarding if needed.
- Never treat keyword matches as verified evidence.
- Never omit contradictions, failed fetches, or sampling limits silently.
- Do not create documents, tasks, messages, or other external records from the synthesis without explicit approval.
