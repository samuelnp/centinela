### Feature-Specialist Report: governed-project-memory
**Date:** 2026-05-29

#### Behavior Summary

When a developer completes a step in a Centinela workflow, the system automatically harvests structured knowledge from the step's artifact into a git-tracked memory ledger. Three typed sources feed the ledger: edge-case lessons (from `.workflow/<f>-edge-cases.md` at `tests` completion), gatekeeper verdicts (from `.workflow/<f>-gatekeeper.md` at `validate` completion), and decisions (one entry per bullet from the `## Decisions` section of the feature brief/plan at `plan` completion). Each entry is written as a content-hash-keyed markdown file under `.workflow/memory/entries/`, making capture idempotent and concurrent-safe. Capture failures — missing or malformed artifacts, absent `## Decisions` sections — are silent warnings that never block step advance. When the developer starts a new feature's plan step, the existing plan-advisor injection is extended to include the top-ranked entries from the ledger, scored by dependency-feature match, then shared tags, then recency, and capped by `recall_max_entries` and `recall_max_bytes`. The entire subsystem is a no-op when `[memory] enabled = false`.

#### Gherkin Scenarios

All scenarios live in `specs/governed-project-memory.feature`.

- **SC-01 — Edge-case lessons captured at tests step completion**
  Given feature "alpha" has a valid edge-cases artifact / When tests step completes / Then a `lesson` entry exists with source link
- **SC-02 — Gatekeeper verdict captured at validate step completion**
  Given feature "alpha" has a gatekeeper report / When validate step completes / Then a `verdict` entry exists
- **SC-03 — Decisions captured from plan when section present**
  Given feature "alpha" brief has `## Decisions` with N bullets / When plan step completes / Then N `decision` entries exist, one per bullet
- **SC-04 — No decisions entry when section absent**
  Given feature "alpha" brief has no `## Decisions` section / When plan step completes / Then no decision entries are created / And step advances without error
- **SC-05 — Capture is idempotent**
  Given feature "alpha" tests step already captured / When tests step completes again / Then no duplicate entries exist
- **SC-06 — Missing source artifact does not block completion**
  Given feature "alpha" has no edge-cases artifact / When tests step completes / Then step advances successfully / And no lesson entry exists for "alpha"
- **SC-07 — Malformed source artifact does not block completion**
  Given feature "alpha" has a malformed edge-cases artifact / When tests step completes / Then step advances successfully / And a warning is logged / And no lesson entry is created
- **SC-08 — Recall injects relevant memory at plan step**
  Given ledger has entries relevant to feature "beta" / When plan step for "beta" starts / Then plan-advisor context includes those entries / And slice respects recall caps
- **SC-09 — Recall respects deterministic ranking**
  Given ledger entries with varying dependency match, tags, and recency / When plan advisor builds context for "gamma" / Then entries are ranked: dependency-match > shared-tags > recency
- **SC-10 — Empty ledger produces no injection and no error**
  Given an empty ledger / When plan step for "beta" starts / Then no memory block is injected / And no error is raised
- **SC-11 — Recall is capped by configured limits**
  Given ledger has more entries than recall_max_entries / When plan step starts / Then only recall_max_entries entries are injected / And total byte size does not exceed recall_max_bytes
- **SC-12 — Memory disabled is a full no-op**
  Given memory is disabled in config / When tests step completes for "alpha" / And plan step starts for "beta" / Then no ledger entries are written / And no memory is injected
- **SC-13 — Concurrent worktree writes are safe**
  Given two features complete their tests step simultaneously in separate worktrees / When both captures run / Then both entries exist / And neither clobbers the other

#### UX States

| State   | Trigger | Surface |
|---------|---------|---------|
| loading | n/a — capture is synchronous at `centinela complete` | n/a |
| empty   | Ledger has no entries matching the planning feature | Plan-advisor injection block is omitted; no `[memory]` section appears in terminal output |
| error   | Source artifact missing or malformed during capture | Warning line printed to stderr: `[memory] warning: skipping <artifact> — <reason>`; step advance continues normally |
| success | Capture: entry written; Recall: entries found | Capture: silent (no terminal noise). Recall: compact `[memory]` block printed as part of plan-advisor context with ranked entries |

#### Out-of-Scope

- Conversation or transcript memory (D7)
- Embeddings, vector stores, or semantic recall (D7)
- Recall at non-plan steps: code, tests, validate, docs (D7)
- Cross-project memory sharing (D7)
- Garbage-collection or migration of existing `.workflow/` evidence files
- Any new hook command — recall rides the existing plan-advisor `UserPromptSubmit` hook
- Auto-tagging by keyword extraction — v1 uses explicit frontmatter tags only

#### Handoff

- Next role: senior-engineer
- Open clarifications: (1) Default values for `recall_max_entries` and `recall_max_bytes` — suggest 10 entries / 4096 bytes as sane v1 defaults; senior-engineer should confirm against the config normalizer pattern. (2) Tag extraction strategy — v1 spec mandates explicit frontmatter tags only; keyword derivation is out of scope. (3) Content-hash input: body-only hash is recommended to avoid collisions when source path changes between worktrees — confirm in entry.go implementation. (4) `## Decisions` parser granularity: one entry per bullet (each `-` line), not one entry per section — confirmed by brief D1 language "one decision per bullet". (5) Ranking tie-break: for v1, same-score entries rank by `createdAt` descending (most recent first) — no roadmap dependency graph traversal needed.
