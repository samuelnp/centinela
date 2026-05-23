---
feature: session-context-rehydration
summary: After /clear (or startup, compact, or resume) Centinela hands Claude a compact roadmap bootstrap so it knows the project state and the next feature to plan, and the per-prompt active-workflows panel stays small and truthful instead of ballooning with noise.
audience: end-user
status: done
---

## What it does
When a Claude Code session begins or restarts — on startup, after `/clear`, after a compact, or on resume — Centinela now injects a one-time "session rehydration" summary so the model does not have to rediscover your project from scratch. That summary contains the full roadmap with each feature's status, the next feature to plan (the first feature that is not done, scanned across every roadmap phase in order), and pointer file paths to `PROJECT.md` and the next feature's brief that Claude can read on demand. Separately, the per-prompt "active workflows" panel was fixed so it lists only genuine, in-progress workflows: it deduplicates by feature, hides done workflows and internal evidence files, shows the five most-recently-touched, and adds a "+N more" hint when there are extras. Together these mean a fresh session starts oriented, and the panel that shows on every prompt stays short and accurate.

## When you'd use it
You benefit from this every time you start or restart a Claude session — especially right after `/clear`, when you would otherwise have to re-explain the project before asking Claude to "plan the next feature." Instead of a blank slate (or, worse, a several-hundred-kilobyte wall of duplicated and finished-workflow noise that buried the roadmap), the session opens with the roadmap, the obvious next step, and the exact files to read. You do not run any command for this; the rehydration fires automatically on session entry and the cleaned-up panel appears on every prompt.

## How it behaves
- Internal evidence files in the workflow folder (for example a per-role `alpha-qa-senior.json`) are no longer mistaken for active work — only the genuine workflow for that feature shows in the panel.
- A finished workflow is hidden from the panel while a genuinely in-progress one for a different feature still appears.
- Stray bookkeeping files such as `roadmap.json` or `roadmap-quality.json` are never shown as if they were active features.
- A feature that has several related files behind it appears as a single row in the panel rather than being listed multiple times.
- When more features are active than the panel's limit, it shows the five most-recently-touched (newest first) and adds a "+2 more" style hint for the rest.
- When the number of active features is at or below the limit, every one is listed and no "+N more" hint is shown.
- On any session entry — startup, clear, compact, or resume — the rehydration summary appears with the full roadmap, names the next feature to plan, and lists the pointer paths `PROJECT.md` and `docs/features/<next>.md` (the paths only, never the file contents pasted in).
- The "next feature to plan" is the first unfinished feature found across all roadmap phases in order, so once every Phase 0 feature is done it correctly points at the first unfinished feature in a later phase.
- When every feature in every phase is already done, the summary says the roadmap is complete, names no next feature, lists no next-feature pointer, and the session still opens cleanly.
- If there is no roadmap at all, the session starts quietly with no rehydration summary and nothing breaks.
- If the roadmap file is present but corrupted, the session likewise starts quietly with no rehydration summary and nothing breaks.

## Examples
After a `/clear`, the new session opens with a block shaped like this (paths are pointers you read on demand, not inlined content):

    CENTINELA DIRECTIVE: session rehydration
    Roadmap:
      Phase 0: Bootstrap
        - docs-migration-managed-docs   [done]
      Phase 1: ...
        - next-feature                  [not started]
    Next feature to plan: next-feature
    Read on demand:
      PROJECT.md
      docs/features/next-feature.md

On a routine prompt, the active-workflows panel stays compact. With more than five in flight it caps the list and hints at the rest:

    ACTIVE WORKFLOWS
      feature-g   tests
      feature-f   code
      feature-e   plan
      feature-d   validate
      feature-c   docs
      +2 more active

When every roadmap feature is done, the session-start summary degrades gracefully instead of inventing a next step:

    CENTINELA DIRECTIVE: session rehydration
    Roadmap complete — no next feature to plan.
