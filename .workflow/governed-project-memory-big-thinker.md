### Big-Thinker Report: governed-project-memory
**Date:** 2026-05-29

#### Problem
Centinela emits a stream of high-value, structured knowledge as a byproduct of its own workflow — edge-case lessons at the tests step, gatekeeper verdicts at the validate step, and the decisions recorded in feature briefs/plans at the plan step — but today that knowledge dies at feature completion. The next feature starts cold: hard-won lessons get re-learned, resolved gate failures recur, and prior architectural decisions are silently contradicted. The person who hurts is the developer (and the AI agent acting for them) starting a new feature in a governed project; today they must remember or re-discover that context manually. This matters now because Phase 1 ("Harness Capabilities") is precisely about turning Centinela's byproducts into durable harness leverage, and the artifacts already exist in a predictable, parseable shape — the gap is purely that nothing harvests or recalls them.

#### Scope
- In: A governed, git-tracked memory ledger (one markdown file per fact under `.workflow/memory/entries/`, frontmatter linking to source) plus a regenerable `index.json` cache. Automatic capture at `centinela complete` keyed to the just-completed step, from exactly three typed sources (edge-cases → lessons, gatekeeper → verdicts, plan `## Decisions` → decisions). Idempotent capture (dedupe by content-hash id). Deterministic recall (dependency match > shared tags > recency, with count + byte caps) injected ONLY at the plan step by extending the existing plan-advisor `UserPromptSubmit` path. A `[memory] enabled` config flag (default on) gating the whole subsystem.
- Out: Conversation/transcript memory; embeddings, vector or semantic stores; recall at non-plan steps (code/tests/validate/docs); cross-project memory; garbage-collection or relocation of existing `.workflow` evidence JSONs; any new hook command (recall rides the existing plan-advisor hook).

#### Dependencies & Assumptions
- Reuses existing artifact production end-to-end: edge-case reports (`edge-case-subagent-tests-phase`), gatekeeper reports, and the brief/plan `## Decisions` section are already produced at the relevant step boundaries — this feature only reads them.
- Recall extends the existing `internal/planadvisor` injection: `planadvisor.Directive` → `buildBundle(feature)` → `contextLines` already runs on the plan step via `cmd/centinela/hook_plan_advisor.go`. No new hook, no `settings.json` change.
- Capture hooks into `cmd/centinela/complete.go` `runComplete`, which already captures `current := wf.CurrentStep` BEFORE advancing — the capture call uses that pre-advance `current`, placed after `saveWorkflow`.
- Config follows the established `internal/config` `WorkflowConfig` + normalizer + `applyDefaults` pattern.
- No external dependencies, no network, no new third-party libraries. Persistence stays JSON/markdown files in `.workflow/`, consistent with the project's chosen persistence model.
- N-tier layer rules hold: capture/dedupe/recall/index logic lives in a new `internal/memory` domain package; `cmd/` stays thin; `internal/ui` only renders.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Recall noise — too many low-value facts injected, drowning the signal | Medium | Medium | Limit to 3 typed sources; deterministic ranking + `recall_max_entries` and `recall_max_bytes` caps; render as a compact block |
| Index drift — `index.json` diverges from entry files | Medium | Medium | Entry files are the single source of truth; `index.json` is always regenerable from them after capture |
| Capture blocking step advance on a missing/malformed artifact | High | Low | Capture failures log a warning and return nil; `complete` never fails on capture; missing `## Decisions` is a clean no-op |
| Concurrent completes across worktrees clobbering shared state | Medium | Low | Per-entry files keyed by content hash (write-if-absent); no shared mutable index during capture |
| Scope creep into a semantic/conversation memory store | High | Medium | v1 decisions D1/D4/D7 explicitly forbid embeddings, vector recall, and conversation memory; gatekeeper enforces |
| Capturing wrong step's artifact (off-by-one on advance) | High | Medium | Documented integration note: capture uses pre-advance `current`, not `wf.CurrentStep`; covered by integration test |
| File-size gate (G1, ≤100 lines incl. tests) violated by parsers | Low | Medium | One small file per concern (entry/dedupe/capture/recall/index); split parsers per source |

#### Rollout
- Step 1 (smallest correct slice): entry model + content-hash id + frontmatter round-trip (`internal/memory/entry.go`) and idempotent per-file write (`dedupe.go`). This is the storage substrate everything else builds on and is independently testable.
- Step 2: capture parsers (`capture.go`, one per source: edge-cases, gatekeeper, decisions) with missing/malformed → skip+warn, plus `index.go` regeneration.
- Step 3: config (`internal/config/memory.go`, `enabled` default on + caps) and `complete.go` wiring on the pre-advance `current` step.
- Step 4: deterministic recall (`recall.go`: dependency > tags > recency, count + byte caps).
- Step 5: plan-advisor injection (add `Memory` to `bundle`, surface via `contextLines`) and the compact `internal/ui` memory render block.

#### Handoff
- Next role: feature-specialist
- Outstanding questions: (1) Exact ranking tie-break order and whether "dependency-feature match" reads roadmap dependency data already available to plan-advisor or a simpler same-feature/tag heuristic for v1. (2) Tag extraction strategy for deterministic matching — explicit frontmatter tags only, or also keyword-derived from titles. (3) Default values for `recall_max_entries` / `recall_max_bytes`. (4) Whether the `## Decisions` parser captures one entry per bullet (recommended, matches the brief) or one entry per section. (5) Content-hash input — body-only vs. body+source — to ensure stable dedupe without false collisions across features.
