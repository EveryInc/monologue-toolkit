---
name: customer-call-to-product-work
description: "Turn a verified Monologue customer-call recording into evidence-backed product work, such as a bug report, issue draft, investigation brief, or implementation plan. Use when the user asks to connect customer feedback or a support call to a repository, issue tracker, diagnosis, or product change."
---

# Customer Call to Product Work

Use `$monologue-notes` to retrieve the call. Keep customer evidence separate from technical diagnosis, and keep every external write or code change in draft form until the user explicitly approves it.

## Workflow

1. Resolve relative dates in the user's timezone and state the explicit local interval.
2. Search with more than one discriminator: customer or company name, problem language, and feature name. Do not use creation-time filters as recording-time filters. For example:

   ```bash
   monologue notes all --q "Acme"
   monologue notes all --q "upload failed"
   ```

   Filter the results by `recorded_at` in the user's timezone. If it is absent,
   use `created_at` as a disclosed fallback; delayed uploads can have a later
   creation date than the call.

3. Verify candidates by fetching and reading their transcripts; a search hit may only mention the customer:

   ```bash
   monologue notes get NOTE_ID
   ```

   Confirm that the note is the requested call by its participants, subject, and date. If the match remains ambiguous, show candidate titles and dates and ask the user to choose.
4. Extract customer evidence before investigating:
   - observed behavior and exact error language
   - expected behavior or desired outcome
   - environment, sequence, frequency, and impact
   - workarounds already tried
   - unanswered reproduction questions

   Attribute claims to `Note title — YYYY-MM-DD`. Paraphrase by default; mark any short exact quote as a quote.
5. Inspect the repository, existing issues, telemetry, or documentation only when available and within scope. Preserve exact handles such as paths, routes, issue numbers, and versions. Do not claim a root cause from the transcript alone.
6. Separate conclusions into:
   - **Customer evidence** — directly supported by the call
   - **Technical evidence** — directly supported by code, issues, logs, or docs
   - **Diagnosis** — reasoned inference, with confidence and alternatives
   - **Unknowns** — missing evidence or reproduction steps
7. Produce the artifact requested by the user: bug report, issue draft, investigation brief, or implementation plan. If the request is unclear, provide a compact investigation brief and offer the other formats.
8. Stop at a draft. Never create an issue, edit code, open a pull request, patch, merge, deploy, or message a customer unless the user explicitly authorizes that action.

## Output

Use this structure unless the requested destination has a required template:

```text
Source
Customer evidence
Technical evidence
Diagnosis and confidence
Unknowns / reproduction gaps
Proposed product work
Draft next action
```

Keep sensitive personal details out of work artifacts unless essential and authorized. If the call includes unrelated personal conversation, omit it and note that the output was scoped to product-relevant material.

## Guardrails

- Keep Monologue retrieval read-only through `$monologue-notes`.
- Never request or accept an API token in chat; have the user run local onboarding if needed.
- Treat search results as unverified until the transcript confirms the call.
- Never present inference as customer testimony or confirmed root cause.
- Keep issue creation, repository changes, and customer communication draft-only until explicit approval.
