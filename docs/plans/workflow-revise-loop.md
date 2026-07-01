# Plan: workflow-revise-loop

Add `centinela revise <feature> --to <step> --reason "<why>"`: a controlled
backward transition. Rewinding re-opens every downstream step to `pending` and
deletes only those steps' *certification evidence* — never source/test code —
so the next `centinela complete` re-runs their gates on the corrected tree.
The "all gates pass on the final tree" guarantee falls out of reusing
`Complete()`; no new gate logic is written. Mirrors how `complete.go` wires
`workflow` + `telemetry` + `memory`: the cmd composes pure pieces.

## Layer compliance (G2 / G7)

- **`RewindTo` (pure state rewind)** lives in `internal/workflow` (domain) —
  step state and ordering are a workflow concern. It mirrors `Complete()` in
  `steps.go` and reuses `OrderedSteps()` / `stepIndexIn()`. It MUST NOT import
  `internal/evidence` (forbidden by G2; `workflow` is below `evidence`).
- **`Revision` type + `Revisions []Revision` field** live on the `Workflow`
  struct in `internal/workflow/state.go`, persisted by the existing
  `Save`/`Load` (JSON), exactly like the `Archetype` field precedent.
- **`Invalidate(feature, role)` primitive** lives in `internal/evidence`
  (`io.go` neighbourhood) — it owns the `.workflow/<feature>-<role>.{json,md}`
  path convention via `pathFor`/`companionPath`. Deleting an evidence pair is
  an evidence concern, not a workflow one.
- **Role→step mapping** is read from `internal/orchestration`
  (`RequiredRolesForFeature(feature, step)`) — already a leaf importable by
  both `workflow` and `cmd`. The set of roles to invalidate per re-opened step
  comes from there plus the two non-step roles in `evidence.AllRoles()`
  (`gatekeeper`, `production-readiness`) and the `-edge-cases.md` artifact.
- **`cmd/centinela/revise.go`** is the outer layer (G7): a thin orchestrator
  that loads the workflow, calls `RewindTo`, loops the invalidation, saves,
  and records telemetry. No business logic — every decision sits in `internal/`.
  Mirrors `complete.go` lines 29-98.
- Untouched: `complete.go` ship gate, `validate*.go` validators,
  `orchestration/policy.go`, `internal/verify`, `internal/gates`. Re-gating is
  pure reuse of the existing forward path.

## Why this needs almost no new gate wiring

- **Re-gating** — re-opened steps are `pending`/`in-progress`; the next
  `complete` calls `ValidateArtifacts` + (for `validate`) `executeValidation` +
  `runClaimVerification` exactly as on the first pass. **No edit to
  complete.go.** ✔
- **Step-gating** (`IsAllowedInStep` / prewrite hook) — after a rewind to
  `code`, `CurrentStep == "code"`, so the prewrite hook allows TypeCode writes
  again with zero new logic. The bypass disappears because the legitimate path
  re-opens. ✔
- **Evidence re-creation** — deleting the role `.json`+`.md` means the
  orchestration validator (`validateOrchestration`) again reports the evidence
  as missing, forcing the agent to re-run the subagent before re-advancing. ✔

## Target validation (the "right target" property)

`RewindTo(target)` rejects unless ALL hold (errors name the offending value):
- `target` is in `wf.OrderedSteps()` (a real step for THIS feature/archetype).
- `target` is strictly before `wf.CurrentStep` (`stepIndexIn(target) <
  stepIndexIn(current)`). Equal or forward → error (`revise` is backward-only).
- `wf.CurrentStep != "done"` (revising a shipped workflow is out of scope —
  reopen via a new feature).

Common cases: validate-found-bug → `--to code`; validate-found-spec-gap →
`--to plan`.

## Implementation

### 1. State: Revision type + audit field (`internal/workflow/state.go`)
- `type Revision struct { From, To, Reason string; At time.Time }` with JSON
  tags. Append-only.
- Add `Revisions []Revision \`json:"revisions,omitempty"\`` to `Workflow`.
  Empty/absent on pre-existing workflows (back-compat, like `Archetype`).

### 2. Pure rewind (`internal/workflow/rewind.go`, NEW — keep steps.go ≤100)
- `func (wf *Workflow) RewindTo(target, reason string) ([]string, error)`:
  - validate target per "Target validation" above; `reason` must be non-empty
    (defence-in-depth; the cmd also requires the flag).
  - capture `from := wf.CurrentStep`.
  - for every step strictly AFTER `target` in `OrderedSteps()`: set
    `Status="pending"`, `CompletedAt=nil`. Set `target` to `in-progress`,
    `CompletedAt=nil`. (Steps before `target` keep their `done` state.)
  - `wf.CurrentStep = target`.
  - append `Revision{From:from, To:target, Reason:reason, At:now}`.
  - return the list of **re-opened step names** (every step after target) so
    the caller knows which steps' evidence to invalidate.
- A tiny `reopenedSteps(order, target) []string` helper keeps it small.

### 3. Evidence invalidation primitive (`internal/evidence/invalidate.go`, NEW)
- `func Invalidate(feature string, role Role) (bool, error)`: remove
  `pathFor(feature, role)` and `companionPath(feature, role)`; missing files
  are not an error (idempotent); return whether anything was removed.
- `func InvalidateArtifact(feature, suffix string) (bool, error)` (or reuse a
  path helper) for the non-role `-edge-cases.md` artifact.
- These NEVER touch `docs/`, `tests/`, or source — only `.workflow/<feature>-*`.

### 4. Role/artifact set per re-opened step (`cmd/centinela/revise.go` helper)
- For each re-opened step name, gather roles via
  `orchestration.RequiredRolesForFeature(feature, step)`. Additionally, when
  `validate` is re-opened, include `gatekeeper` + `production-readiness`
  (the two non-step roles in `evidence.AllRoles()`); when `tests` is
  re-opened, include the `-edge-cases.md` artifact. Dedupe.
- This mapping table is the one piece of new "policy"; keep it as a small,
  data-driven function so it is trivially testable.

### 5. Command wiring (`cmd/centinela/revise.go`, NEW — mirror complete.go)
- `revise <feature> --to <step> --reason <why>` cobra command, self-registered
  via `init()` → `rootCmd.AddCommand`.
- `--to` and `--reason` both REQUIRED (cobra `MarkFlagRequired`); empty
  `--reason` after trim → error.
- Flow: `config.Load` → `workflow.Load` → `wf.RewindTo(to, reason)` →
  for each re-opened step, resolve roles/artifacts and call
  `evidence.Invalidate` / artifact removal → `saveWorkflow(wf)` →
  `telemetry.RecordRevised(cfg, feature, from, to, model)` →
  success render listing the new current step + invalidated evidence count.
- Resolve `model` via `resolveEmitModel(wf, cfg)` like complete.go.

### 6. Telemetry (`internal/telemetry/constructors.go` + event type)
- Add `TypeStepRevised` event type and
  `RecordRevised(cfg, feature, from, to, model string)` sibling of
  `RecordStepAdvanced`, carrying both endpoints of the jump.

### 7. Status surface (`internal/ui/render_status.go` + `internal/workflow`)
- Render a `Revisions  N` line (and, when non-empty, the last reason) below the
  step list. Compute the display string in `internal/workflow`
  (`RevisionsSummary(wf) string`) so `internal/ui` stays logic-free — mirrors
  `DisplayArchetype`/`ProfileProvenance`.

## Test plan

- **Unit (colocated, per-package, each ≤100 lines):**
  - `RewindTo`: validate→code re-opens `[tests,validate,docs]` as pending,
    sets code in-progress, leaves plan/code-as-done? (code becomes the target →
    in-progress); appends one Revision with From/To/reason.
  - Target rejection: forward target, equal target, unknown step, and
    `CurrentStep=="done"` each error; empty reason errors.
  - `reopenedSteps` returns exactly the steps after target for canonical AND a
    non-canonical order (hotfix `[code,tests,validate]`) — pins
    archetype-awareness.
  - State round-trip: `Revisions` persists in/out of `.workflow` JSON.
  - `evidence.Invalidate`: removes both `.json`+`.md`; idempotent on missing;
    **safety test** — a sibling source/test file in the dir is untouched.
  - Per-step role/artifact mapping: re-opening `validate` includes
    `gatekeeper`+`production-readiness`; re-opening `tests` includes
    `-edge-cases.md`; user-facing vs internal `code` differs (ux-ui).
  - `RecordRevised` writes a `TypeStepRevised` event with from/to/model.
  - `RevisionsSummary` / status render shows the count + reason.
- **Integration (`tests/integration`):** drive `RewindTo` + `Invalidate`
  end-to-end on a temp `.workflow`: after a rewind, downstream evidence files
  are gone, `docs/`+`tests/` fixtures remain, state shows the revision.
- **Acceptance (`tests/acceptance/workflow_revise_loop_test.go`):** compiled
  binary — `start` → fake-advance to `validate` → `revise --to code --reason
  "bug"` → assert exit 0, current step `code`, downstream evidence absent,
  source intact; and `revise` without `--reason` exits non-zero. Carries the
  `// Acceptance:` + `// Scenario:` comments closing this feature's own spec.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|---|---|---|---|
| Thrashing / abuse (endless rewind loops) | High | Med | `--reason` REQUIRED is the friction; every rewind appended to `Revisions` and surfaced in `status` (visible count) — a high count is a smell a human/gate can see. Telemetry `RecordRevised` makes it analyzable. |
| Accidental deletion of source or test CODE | High | Low | `Invalidate` operates ONLY on `.workflow/<feature>-<role>.{json,md}` + `-edge-cases.md` via `pathFor`/`companionPath`; it never touches `docs/`/`tests/`/`internal/`. A dedicated safety unit test asserts a sibling source file survives. Code is work product, not certification. |
| Re-gating not actually forced (evidence "looks" present) | High | Low | Deleting the role `.json` makes `validateOrchestration` report it missing, so `complete` blocks until re-run. Re-using `Complete()` means zero special-case code can drift from the forward path. Acceptance pins downstream-evidence-absent. |
| `RewindTo` wrong for non-canonical archetypes | Med | Low | Operates on `wf.OrderedSteps()`, never `DefaultStepOrder`; a hotfix-order unit test pins it. |
| Revising a `done` workflow | Med | Low | Explicitly rejected in `RewindTo`; out of scope for v1 (reopen via a new feature). |
| G1 >100 lines per file | Low | Low | Split across `rewind.go`, `invalidate.go`, `revise.go`; the role/artifact map is data. |

## Rollout

Smallest correct slice first, each independently testable:

1. **State** — `Revision` type + `Revisions` field + JSON round-trip test
   (pure data, no behaviour).
2. **`RewindTo`** pure domain rewind (`rewind.go`) + target-validation +
   archetype-order unit tests. No I/O, no evidence.
3. **`evidence.Invalidate`** primitive + safety test (source survives). No
   workflow coupling.
4. **`revise.go`** wiring composing 2+3 + the per-step role/artifact map +
   `RecordRevised` telemetry.
5. **Status render** of the revision count/reason
   (`RevisionsSummary` + `render_status.go`).
6. **Acceptance** test closing the dogfood (binary-driven: validate→code
   rewind, missing-`--reason` rejection).
