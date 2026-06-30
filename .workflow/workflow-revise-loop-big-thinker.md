### Big-Thinker Report: workflow-revise-loop
**Date:** 2026-06-30

#### Problem

Centinela's workflow is forward-only — `plan → code → tests → validate → docs`
moves exclusively through `centinela complete`. When the `validate` step finds
a defect that needs a `code` change (or a spec gap needing `plan`), there is no
enforcer-sanctioned way back. The driving LLM bypasses the enforcer entirely,
editing source via raw `bash` outside its gated step, which silently defeats
the architecture rules, the per-step validators, and the certification evidence
the framework exists to guarantee. `revise` replaces that bypass with an
auditable, gate-preserving backward transition. The design is already decided;
this report frames, de-risks, and sequences it.

#### Scope

- **In (v1):** `centinela revise <feature> --to <step> --reason "<why>"`;
  pure-domain `RewindTo` on `*Workflow` (`internal/workflow/rewind.go`);
  `evidence.Invalidate(feature, role)` delete-pair primitive
  (`internal/evidence`); `Revision` type + `Revisions []Revision` audit field
  on the state; `telemetry.RecordRevised`; a revision count/reason line in
  `centinela status`.
- **Out (v1) — deliberate decisions, not omissions:**
  - **Revising a completed (`done`) workflow:** OUT. A shipped feature reopens
    via a new follow-up feature, not a rewind. `RewindTo` errors when
    `CurrentStep == "done"`.
  - **Marking evidence `.stale` vs deleting it:** v1 **DELETEs** the evidence
    pair. Git history + the `Revisions` audit log already preserve the full
    record, so a parallel `.stale` lifecycle is redundant complexity.
  - Forward-skip / arbitrary jumps (backward-only); auto-detecting the target
    step (the agent supplies `--to`).

#### Dependencies & Assumptions

- `internal/workflow`: `Complete()`/`stepIndexIn()` (`steps.go`) is the forward
  mover `RewindTo` mirrors; `OrderedSteps()` (`order.go`) gives the per-feature
  step set (archetype-aware — never `DefaultStepOrder`); `Workflow`/`StepState`
  + `Save`/`Load` (`state.go`) carry the new `Revisions` field, exactly like the
  existing `Archetype` field precedent.
- **G2 layering:** `internal/workflow` (domain) may import `config`,
  `gitdiff`, `orchestration` but **NOT** `internal/evidence`. So the rewind is
  pure-domain; the evidence deletion lives in `internal/evidence` (owns
  `pathFor`/`companionPath`); `cmd/centinela/revise.go` (outer layer, may import
  everything) composes both — mirroring how `complete.go` wires
  workflow+telemetry+memory.
- Role→step mapping: `orchestration.RequiredRolesForFeature(feature, step)`
  (leaf, importable by cmd) + the two non-step roles in `evidence.AllRoles()`
  (`gatekeeper`, `production-readiness`) + the `-edge-cases.md` artifact.
- Re-gating needs NO new code: re-opened (`pending`) steps force the next
  `complete` to re-run `ValidateArtifacts` + (for validate) `executeValidation`
  + `runClaimVerification`. The "all gates pass on the FINAL tree" guarantee
  falls out of reusing `Complete()`.
- Assumes the prewrite hook keys allowed file types on `wf.CurrentStep`; after a
  rewind to `code`, code writes are legitimate again with zero new logic.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Thrashing / abuse (endless rewinds) | High | Med | `--reason` REQUIRED is the friction; every rewind appended to `Revisions` and surfaced as a visible count in `status`; `RecordRevised` telemetry makes it analyzable. |
| Accidental deletion of source/test CODE | High | Low | `Invalidate` touches ONLY `.workflow/<feature>-<role>.{json,md}` + `-edge-cases.md` via `pathFor`/`companionPath`; never `docs/`/`tests/`/`internal/`. Dedicated safety unit test asserts a sibling source file survives. Code = work product, not certification. |
| Re-gating not actually forced | High | Low | Deleting the role `.json` makes `validateOrchestration` report it missing; `complete` blocks until re-run. Reusing `Complete()` means no special path can drift. Acceptance pins downstream-evidence-absent. |
| `RewindTo` wrong for non-canonical archetypes | Med | Low | Operates on `wf.OrderedSteps()`, never `DefaultStepOrder`; hotfix-order unit test pins it. |
| Revising a `done` workflow | Med | Low | Explicitly rejected in `RewindTo`; v1 out-of-scope. |
| G1 >100 lines per file | Low | Low | Split across `rewind.go`, `invalidate.go`, `revise.go`; role/artifact map is data. |

#### Rollout

- Step 1: **State** — `Revision` type + `Revisions` field + JSON round-trip
  test (pure data, no behaviour).
- Step 2: **`RewindTo`** pure-domain rewind (`rewind.go`) + target-validation
  (backward-only, real-step, not-done, non-empty reason) + archetype-order unit
  tests. No I/O, no evidence coupling.
- Step 3: **`evidence.Invalidate`** delete-pair primitive + idempotency +
  safety test (source survives).
- Step 4: **`revise.go`** wiring composing 2+3 + the per-step role/artifact
  invalidation map + `RecordRevised` telemetry.
- Step 5: **Status render** of the revision count/reason (`RevisionsSummary`
  in `internal/workflow` + `render_status.go`).
- Step 6: **Acceptance** test closing the dogfood (validate→code rewind;
  missing-`--reason` rejection).

#### Deferred Findings

- none. The two boundary calls (no-rewind-when-done; delete-not-stale) are
  deliberate v1 decisions recorded in Scope/Out, not new out-of-scope
  discoveries warranting a roadmap defer.

#### Handoff

- Next role: feature-specialist
- Outstanding questions: (1) confirm the exact per-step role set to invalidate
  when re-opening `code` for a user-facing feature (senior-engineer +
  ux-ui-specialist) vs internal (senior-engineer only) — the feature-specialist
  should pin this in the Gherkin scenarios. (2) Decide whether `status` shows
  only the latest reason or the full revision list (recommend: count + latest
  reason inline, full list in a `--verbose` future slice).
