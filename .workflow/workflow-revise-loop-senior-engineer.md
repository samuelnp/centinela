# workflow-revise-loop — senior-engineer

`centinela revise <feature> --to <step> --reason "<why>"` — a controlled,
auditable backward transition. Rewinding re-opens every downstream step to
`pending`, deletes ONLY those steps' certification evidence
(`.workflow/<feature>-*`), and appends a `Revision` to the workflow state. The
"all gates pass on the final tree" guarantee falls out of reusing the existing
forward `Complete()` path — no new gate logic was written.

## Files Touched

| File | Lines | Change |
|------|------:|--------|
| `internal/workflow/state.go` | 92 | Added `Revisions []Revision` audit field on `Workflow` (omitempty, back-compat like `Archetype`). |
| `internal/workflow/rewind.go` | 92 | NEW. `Revision` type, `RewindTo`, `RevisionsSummary`, `reopenedSteps` helper. |
| `internal/evidence/invalidate.go` | 42 | NEW. `Invalidate(feature, role)` + `InvalidateArtifact(feature, suffix)` + `removeBoth` (idempotent). |
| `internal/evidence/invalidation_targets.go` | 22 | NEW. `InvalidationTargets(feature, step)` — the per-step role/artifact policy. |
| `internal/telemetry/event.go` | 50 | Added `TypeStepRevised` + `From` event field. |
| `internal/telemetry/constructors.go` | 56 | Added `RecordRevised(cfg, feature, from, to, model)`. |
| `internal/ui/render_status.go` | 86 | Render `Revisions N (last: "…")` row when non-empty. |
| `cmd/centinela/revise.go` | 74 | NEW. Cobra `revise` command (thin orchestrator). |
| `cmd/centinela/revise_invalidate.go` | 44 | NEW. `invalidateDownstream` — loop/dedup/count composition. |

## Architecture Compliance

**G1 (≤100 lines).** Every file above is ≤92 lines. The plan's `rewind.go` /
`invalidate.go` / `revise.go` split was followed; `state.go` was kept at 92 by
moving the `Revision` type definition into `rewind.go` (alongside the logic that
appends it) while the `Revisions` field stays on the `Workflow` struct.

**G2 layering (domain `internal/workflow` must NOT import `internal/evidence`).**
- `RewindTo` is **pure state** in `internal/workflow` — imports only `fmt`,
  `strings`, `time`. It returns the re-opened step names so the caller (not the
  domain) drives evidence deletion. No evidence import; no import cycle.
- `Invalidate` / `InvalidationTargets` live in `internal/evidence` (which already
  imports `internal/orchestration`), so the role→step policy and the
  `pathFor`/`companionPath` deletion sit where the path convention is owned.
- `cmd/centinela/revise.go` composes the two pure pieces — it imports everything
  (config, workflow, evidence, telemetry, ui), the outer layer is allowed to.

**G7 (no business logic in cmd/).** Both decisions live in `internal/`: the
rewind/validation rules in `workflow.RewindTo`, the per-step role/artifact set in
`evidence.InvalidationTargets`. `revise.go` only wires Load → RewindTo →
invalidate → Save → telemetry → render; `invalidateDownstream` only loops,
dedupes, and counts.

**Re-gating is pure reuse.** Re-opened steps become `pending`/`in-progress` and
their evidence `.json` is gone, so the existing `validateOrchestration` reports
it missing and the unchanged `Complete()` blocks until the subagent re-runs.
`complete.go`, the validators, and `orchestration/policy.go` are untouched.

## Type-Safety Notes

- `RewindTo` returns `([]string, error)` and rejects every bad input (unknown /
  equal / forward target, `done` workflow, empty reason) BEFORE mutating any
  state, each error naming the offending value.
- `evidence.Role` is the existing `orchestration.Role` alias — no stringly-typed
  drift; `InvalidationTargets` reuses `RequiredRolesForFeature` so feature-aware
  gating (ux-ui only for user-facing code) is inherited, not re-implemented.
- `Invalidate`/`InvalidateArtifact` return `(bool, error)`; absence is success
  (`os.IsNotExist`), real I/O errors surface with the path. No panics, no
  silent swallow of unexpected errors.
- `Revision.At` is `time.Time` stamped `time.Now().UTC()`; JSON tags mirror the
  `Archetype` back-compat precedent (`omitempty`).

## Trade-Offs

- **`Revision` type in `rewind.go`, not `state.go`** — purely to keep both files
  ≤100 under G1. The field stays on `Workflow`; a comment cross-references.
- **`InvalidationTargets` per-step (caller dedupes)** rather than taking the
  whole step list — keeps the policy trivially unit-testable and matches the
  plan; dedup is a 2-map loop in the cmd.
- **Invalidation count is per role-pair/artifact, not per file** — a role's
  `.json`+`.md` count as one "evidence artifact invalidated" in the success
  line, which reads naturally for the user.
- **Auto-commit deferred** — unlike `complete.go`, `revise.go` does not
  auto-commit; a rewind is an interactive correction and the surrounding
  workflow tooling owns the commit. No `cfg.Workflow.DisableAutoCommit` branch
  was added (keeps the orchestrator thin).

## Deferred Findings

- No tests were written (qa-senior's step). Dogfood (scratch binary) confirmed:
  happy path validate→code (code in-progress; tests/validate/docs pending; plan
  stays done; touched-step `CompletedAt` cleared; 4 evidence artifacts removed;
  `senior-engineer` evidence and `internal/demo/handler.go` survive; revision
  appended), idempotent second rewind to plan (2 revisions accumulate), and
  rejection of missing/whitespace `--reason`, forward/equal/unknown targets.
- `RewindTo` archetype-awareness (hotfix `[code,tests,validate]`) is implemented
  via `OrderedSteps()` but only canonical order was exercised by hand — pin it
  with the hotfix-order unit test the plan calls for.

## Handoff → qa-senior

Build, `go vet`, and `gofmt` are clean; `centinela evidence validate
workflow-revise-loop` passes. Write the colocated unit tests (each ≤100 lines)
per the plan's Test plan: `RewindTo` validation + re-open matrix +
`reopenedSteps` canonical/hotfix; `Revisions` JSON round-trip; `evidence.Invalidate`
removal + idempotency + **sibling-source-survives safety test**;
`InvalidationTargets` (validate→gatekeeper+production-readiness; tests→edge-cases;
internal vs user-facing code); `RecordRevised` event; `RevisionsSummary`/status
render. Then integration (temp `.workflow`) and the binary-driven acceptance test
`tests/acceptance/workflow_revise_loop_test.go` carrying the `// Acceptance:` +
`// Scenario:` comments, and add its execution to `validate.commands`.
