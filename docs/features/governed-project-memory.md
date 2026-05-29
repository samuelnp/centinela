# Feature Brief — governed-project-memory

## Problem — what pain does this solve? Who is the user?

Centinela produces a stream of high-value, structured knowledge as a side effect
of its workflow — edge-case lessons, gatekeeper verdicts, and the decisions
recorded in feature briefs and plans — but that knowledge dies at the end of each
feature. The next feature starts cold. Lessons learned the hard way are
re-learned; resolved gate failures recur; architectural decisions are silently
contradicted.

**User:** the developer (and the AI agent acting on their behalf) starting a new
feature in a Centinela-governed project, who wants the project's accumulated
hard-won knowledge surfaced automatically instead of having to remember or
re-discover it.

This is the **memory & state management** subsystem of harness engineering,
expressed the Centinela way: a *governed* ledger — structured, git-tracked,
human-reviewable, deterministic — rather than a fuzzy semantic store.

## Decisions

> Locked v1 scope. This section is itself a capture source — completing the
> **plan** step harvests these decisions into the ledger, so it must stay a
> first-level `## Decisions` heading with one decision per bullet.

- **D1 — Capture sources (3, typed):** edge-case lessons (from
  `.workflow/<f>-edge-cases.md`), gatekeeper verdicts (from
  `.workflow/<f>-gatekeeper.md`), and decisions (from the `## Decisions` section
  of the feature brief / plan). No other sources in v1.
- **D2 — Capture trigger:** automatic at `centinela complete <feature>`, keyed to
  the step being completed (plan → decisions, tests → lessons, validate →
  verdicts). One source artifact per step boundary.
- **D3 — Recall point:** the **plan** step only. Recall extends the existing
  plan-advisor `UserPromptSubmit` injection; no recall at code/tests/validate/docs.
- **D4 — Deterministic ranking:** relevance is computed by explicit signals
  (dependency-feature match > shared tags > recency). No embeddings, no semantic
  / vector store.
- **D5 — Governed storage:** one git-tracked markdown file per fact under
  `.workflow/memory/entries/`, with frontmatter linking to the source artifact;
  the `index.json` is a regenerable cache, not the source of truth.
- **D6 — Non-blocking & opt-out:** capture failures (missing/malformed artifact)
  are warnings, never block `centinela complete`; the whole subsystem is gated by
  a `[memory] enabled` config flag (default on).
- **D7 — Out of scope for v1:** conversation memory, vector recall, recall at
  non-plan steps, cross-project memory.

## User Stories

- As a developer, when I start planning a feature, I want relevant prior lessons,
  gate verdicts, and decisions surfaced automatically so I don't repeat past
  mistakes.
- As an AI agent, when I advance a step, I want durable facts harvested from the
  artifacts I just produced so a future session inherits them.
- As a reviewer, I want every remembered fact to be a reviewable diff tied to its
  source artifact, so memory can be audited and corrected in a PR.

## Acceptance Criteria (→ Gherkin)

1. Completing the **tests** step harvests edge-case lessons from
   `.workflow/<f>-edge-cases.md` into the ledger.
2. Completing the **validate** step harvests the gatekeeper verdict + findings
   from `.workflow/<f>-gatekeeper.md` into the ledger.
3. Completing the **plan** step harvests decision entries from a `## Decisions`
   section in the feature brief / plan, when present.
4. Each captured entry is written as a reviewable, git-tracked file with
   frontmatter linking back to its source artifact.
5. Capture is **idempotent** — re-completing a step does not duplicate entries.
6. Starting the **plan** step injects the relevant slice of the ledger into the
   plan advisor context, ranked deterministically (no embeddings).
7. Memory can be disabled via config; when disabled, capture and recall are
   no-ops.

## Edge Cases

- Empty ledger (first feature ever) → recall injects nothing, no error.
- Source artifact missing or malformed → capture skips it with a warning, never
  blocks `centinela complete`.
- Feature brief has no `## Decisions` section → plan capture is a no-op.
- Duplicate facts across re-runs → dedupe by stable content hash.
- Very large ledger → recall caps the injected slice (count + byte budget).
- Concurrent completes across worktrees → capture writes are per-entry files to
  avoid clobbering a shared index.
- Memory disabled in config → all capture/recall short-circuits.

## Data Model

**Ledger entry** (one file per fact, `.workflow/memory/entries/<id>.md`):

| Field | Meaning |
|-------|---------|
| `id` | stable content hash (dedupe key) |
| `feature` | source feature slug |
| `step` | producing step (`plan` / `tests` / `validate`) |
| `type` | `lesson` \| `verdict` \| `decision` |
| `title` | one-line summary (used in recall) |
| `tags` | keywords for deterministic relevance matching |
| `sourceArtifact` | path the fact was harvested from |
| `createdAt` | ISO timestamp |
| body | the fact itself (markdown) |

Plus `.workflow/memory/index.json` — a generated index for fast recall
(regenerable from the entry files; entry files are the source of truth).

## Integration Points

- `centinela complete` — capture hook at step advance (domain logic in
  `internal/memory`, thin wiring in `cmd/`).
- UserPromptSubmit plan-advisor hook — recall injection (extends existing
  edge-case-lesson injection).
- `internal/config` — `[memory] enabled` flag + recall caps.

## Risks

- **Noise:** capturing too much makes recall useless → v1 limits to three typed
  sources and a capped injection slice.
- **Drift:** entry files vs. index diverge → index is always regenerable; entries
  are source of truth.
- **Performance:** large ledgers slow recall → deterministic match + byte cap, no
  embeddings.
- **Scope creep:** must not become a semantic memory store — explicitly out of
  scope for v1.

## Decomposition

Single feature. Internal slices: (a) `internal/memory` capture + dedupe,
(b) `internal/memory` recall ranking, (c) `cmd/` complete-step wiring,
(d) plan-advisor injection, (e) config flag + caps. Each kept ≤100 lines per the
file-size gate.
