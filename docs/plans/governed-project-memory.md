# Plan — governed-project-memory

> Feature brief: [docs/features/governed-project-memory.md](../features/governed-project-memory.md)
> Roadmap: Phase 1 — Harness Capabilities. Spec:
> [specs/governed-project-memory.feature](../../specs/governed-project-memory.feature)

## Goal

A governed, git-tracked memory ledger that harvests Centinela's own step
artifacts at `centinela complete` and recalls the relevant slice into the plan
step — deterministic, reviewable, no semantic-store dependency.

## v1 scope (locked decisions)

- **Capture sources:** edge-case lessons (`tests`), gatekeeper verdicts
  (`validate`), and decisions (`plan`, from a `## Decisions` section).
- **Capture trigger:** automatic at `centinela complete <feature>` (step boundary).
- **Recall point:** plan step only — extend the existing plan-advisor injection.

Out of scope for v1: conversation memory, embeddings/vector recall, recall at
non-plan steps, cross-project memory.

## Architecture (n-tier — see PROJECT.md)

| Layer | Path | Responsibility |
|-------|------|----------------|
| Domain | `internal/memory/capture.go` | parse a step artifact → typed entries |
| Domain | `internal/memory/entry.go` | entry type, content-hash id, frontmatter (de)serialize |
| Domain | `internal/memory/dedupe.go` | idempotent write by content hash |
| Domain | `internal/memory/recall.go` | deterministic relevance ranking + caps |
| Domain | `internal/memory/index.go` | (re)generate `index.json` from entry files |
| Config | `internal/config/memory.go` | `[memory] enabled` (+ normalizers), recall count/byte caps |
| Outer | `cmd/centinela/complete.go` | call `memory.Capture` for the **just-completed** step (thin; see note) |
| Recall | `internal/planadvisor/context.go` | extend `buildBundle` with a `Memory []string` field + `contextLines` |
| UI | `internal/ui` | render recalled entries as a compact memory block in plan context |

**Integration accuracy notes (verified against current code):**

- In `cmd/centinela/complete.go`, `runComplete` captures `current := wf.CurrentStep`
  *before* `wf.Complete(cfg)` advances it. Capture must run on `current` (the step
  whose artifact now exists), so the call goes right after the successful
  `saveWorkflow(wf)` using `current`, not `wf.CurrentStep`. Map step → source:
  `plan`→brief/plan `## Decisions`, `tests`→edge-cases, `validate`→gatekeeper.
- Recall reuses the existing plan-advisor path. `planadvisor.Directive` →
  `buildBundle(feature)` already assembles `Lessons`, `Dependencies`, etc. Add a
  `Memory` slice populated via a new `internal/memory.Recall(feature, cfg)` call
  and surface it through `contextLines`. No new hook command is needed — the
  `cmd/centinela/hook_plan_advisor.go` wiring already fires on the plan step.
- `internal/config` already has the `WorkflowConfig` + normalizer pattern
  (`NormalizePlanQuestionLimit`, etc.) and a top-level `Config` struct. Add a
  `Memory MemoryConfig \`toml:"memory"\`` field and an `applyDefaults` entry so
  `enabled` defaults true and caps get sane defaults.

All new source files ≤100 lines (split as needed); `_test.go` files also ≤100
(G1 applies to tests — see prior lesson).

## Work breakdown

1. **Entry model + storage** — `entry.go`: fields, content-hash id, markdown +
   frontmatter round-trip; entries live in `.workflow/memory/entries/<id>.md`.
2. **Dedupe + write** — `dedupe.go`: write-if-absent by id; per-file writes so
   concurrent worktree completes don't clobber.
3. **Capture** — `capture.go`: one parser per source artifact (edge-cases,
   gatekeeper, decisions). Missing/malformed artifact → skip + warn, never block.
4. **Index** — `index.go`: regenerate `index.json` from entry files after capture.
5. **Recall ranking** — `recall.go`: score entries for the planning feature by
   (dependency-feature match > shared tags > recency); apply count + byte caps.
6. **Config** — `internal/config/memory.go`: `MemoryConfig` (`enabled` default
   true, `recall_max_entries`, `recall_max_bytes`) + normalizers wired into
   `applyDefaults`; mirror the existing plan-advisor config pattern.
7. **complete wiring** — in `runComplete`, after `saveWorkflow`, call
   `memory.Capture(feature, current, cfg)` (uses the pre-advance `current` step).
   Failures log a warning and return nil — never block the advance.
8. **plan-advisor injection** — add `Memory` to `planadvisor.bundle`, populate via
   `memory.Recall`, render through `contextLines`. Reuses the existing plan-step
   hook; no new hook command.
9. **UI** — compact `🛡️👁️ MEMORY` render block for recalled entries (pure render).

## Verification strategy (preview of tests step)

- **Unit:** each parser (valid/empty/malformed), content-hash stability, dedupe
  idempotence, recall ranking + caps, config gating.
- **Integration:** complete tests/validate/plan steps on a fixture feature →
  entries appear; re-complete → no duplicates; recall returns expected slice.
- **Acceptance:** Gherkin scenarios in the `.feature` executed via go test.

## Risks / mitigations

- Noise → three typed sources + capped recall.
- Index drift → entries are source of truth; index regenerable.
- Blocking completes on bad artifacts → capture failures are warnings only.
