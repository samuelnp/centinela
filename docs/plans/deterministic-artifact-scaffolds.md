# Plan: deterministic-artifact-scaffolds

### Big-Thinker Report: deterministic-artifact-scaffolds

**Date:** 2026-06-12

#### Problem

`evidence-cli` removed JSON-by-heredoc but not *shape toil*. Two concrete pains:

1. `centinela evidence init <feature> <role>` pre-fills only scalars; `inputs`
   and `outputs` are empty. The validator `validatePlanSnapshotInputs` then
   requires big-thinker / feature-specialist `inputs` to snapshot **every**
   `docs/features/*.md` + `docs/plans/<feature>.md` — today ~80 paths the agent
   hand-loops on every run. That set is already computed mechanically inside the
   validator (`requiredPlanInputs`), so the agent re-derives what the framework
   already knows.
2. `artifact new` bodies and companion markdown ship italic-prose placeholders.
   Under the `strict` profile a `limited`-class model mis-shapes the artifact and
   retries on *structure* rather than substance.

The fix mirrors the docs CLI-fallback pattern (`internal/docgen.Generate`):
stamp everything mechanically derivable, leave only substance for the LLM.

#### Scope (In / Out)

**In:**
- Export `orchestration.RequiredPlanInputs(feature) []string` (promote the
  existing unexported `requiredPlanInputs`); the validator calls the exported
  name so there is ONE source of truth.
- `evidence init` pre-fills `inputs` with `RequiredPlanInputs(feature)` for
  big-thinker + feature-specialist only; other roles unchanged (empty).
- One canonical `<FILL: …>` marker (constant + helper) in `internal/evidence`.
- Per-role companion markdown skeletons seeded with `<FILL: …>` slots.
- `artifact new` markdown: italic prose → `<FILL: …>` slots + mechanical
  pre-fill of derivable lists (gatekeeper "Analyzed Specs" ← `specs/*.feature`).

**Out (v1):**
- **No `outputs` pre-fill in evidence JSON** and **no `PredictedOutputs` API** —
  see Divergence. Outputs stay a fill slot.
- No `<FILL: …>` in any JSON list field.
- No `--minimal` flag. No profile-gating of pre-fill (unconditional).

#### Dependencies & Assumptions

- **Code dep:** `evidence-cli` (shipped). `enforcement-profiles` /
  `model-capability-profiles` are the motivating *why*, not code deps.
- **Assumption (verified):** `internal/evidence` already imports
  `internal/orchestration` (roles.go, orchestration_bridge.go, artifact_docs.go),
  so promoting `requiredPlanInputs` to public adds **zero** new import-graph
  edge. `internal/orchestration` stays a leaf (imports no internal pkg).
- **Assumption (verified):** `Skeleton` is reused by `SchemaSkeleton` (repair,
  feature=`<feature-slug>`) and `docsSpecialistPair`. Pre-fill must therefore
  live in the **`evidence init` command path**, applied to the skeleton *after*
  construction — NOT inside `Skeleton` — or repair/docs templates would carry
  bogus glob results. (Divergence from proposed #2.)
- **Assumption (verified):** `evidence init` is non-overwriting today (existence
  guarded); re-running with the new pre-fill is idempotent because
  `RequiredPlanInputs` is deterministic and sorted.

#### Divergence from proposed design (with reasons)

1. **DROP `outputs` pre-fill and `PredictedOutputs` (proposal #1b, #2).**
   `validateActionableOutputs` → `missingOutputFiles` rejects any `outputs` entry
   that is not a real file on disk *at validate-time*. At `init` time the
   predicted outputs (`docs/plans/<feature>.md`, `specs/<feature>.feature`,
   `.workflow/<feature>-edge-cases.md`, the role `.md`) do **not** exist. Two
   failure modes: (a) if the agent runs `evidence validate` early it fails on a
   real file it hasn't written; (b) pre-seeding `outputs` commits the agent to
   producing exactly those files — but the *real* substance outputs (impl file,
   test file path) are unknowable, so a partial pre-fill is misleading. Net: it
   trades a benign "forgot to list it" miss for a hard "listed a phantom file"
   failure and silently narrows what the agent thinks it must produce.
   **Decision:** outputs stay empty (a genuine fill slot). The 80-path `inputs`
   hand-loop is the real pain; that is what we kill. `inputs` is safe to pre-fill
   because `RequiredPlanInputs` returns *existing* `docs/features/*.md` files
   (plus the feature's own brief, which the plan step creates first) — and the
   validator only checks the snapshot is a *superset*, never that the listed
   files exist on disk.

2. **Pre-fill in the command path, not in `Skeleton` (refines proposal #2).**
   `Skeleton` is shared with repair + docs templates; mutating it would poison
   those. New helper `evidence.PlanInputs(feature, role) []string` returns the
   pre-fill (delegating to `orchestration.RequiredPlanInputs` for the two plan
   roles, else `nil`); `runEvidenceInit` applies it to the skeleton.

3. **FILL marker confirmed `<FILL: …>` (resolves open question).** Survives the
   no-HTML-escape marshal (`SetEscapeHTML(false)`), greppable, and is invalid in
   finished prose so a leftover is trivially caught. Used in **markdown only**.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|------------|
| Pre-filled `inputs` fail `validatePlanSnapshotInputs` (drift) | High — blocks plan step | Low | Same source fn (`RequiredPlanInputs`) for pre-fill AND validation; add a test asserting an init'd plan-role evidence validates with zero appends. |
| Pre-fill leaks into repair/docs templates via `Skeleton` | Medium — bogus globs in templates | Low (mitigated) | Pre-fill lives in command path, never in `Skeleton`; test asserts `SchemaSkeleton`/`docsSpecialistPair` inputs stay empty. |
| `<FILL: …>` marker survives into finished/validated content | Medium — ugly artifacts | Low | Marker only in markdown bodies; existence-only check on companions means no JSON-validator exposure; doc the grep `grep -r '<FILL:' .workflow/` as a self-check. |
| Re-running `evidence init` clobbers agent-edited inputs | Medium | Low | Honor existing existence-check / `--force`; pre-fill is idempotent (sorted, deterministic) so a `--force` re-run is safe. |
| New import-graph FAIL from promoting fn to public | High | None | Verified: edge already exists; promotion is rename-only, no new package boundary crossed. |
| File-size G1 (>100 lines) on touched files | Low | Low | Touched files are 37–73 lines; add small new files (`fill.go`, `companion_skeletons.go`) rather than grow existing ones. |
| Companion skeleton drift vs prompt section headers | Low | Medium | Keep skeleton headers minimal and aligned to existing prompt docs; not validator-enforced, so drift is cosmetic. |

#### Rollout (slices)

- **Slice 1 — `inputs` pre-fill (the value blocker).** Promote
  `RequiredPlanInputs`; add `evidence.PlanInputs`; wire into `runEvidenceInit`.
  Kills the 80-path hand-loop. Shippable alone.
- **Slice 2 — FILL marker + companion skeletons.** `evidence.FillSlot` + per-role
  companion markdown skeletons replacing the one-liner.
- **Slice 3 — `artifact new` body upgrade.** Italic prose → `<FILL: …>` slots +
  mechanical pre-fill (gatekeeper Analyzed Specs ← `specs/*.feature`).

Slices are independent; ship 1 first. If sequencing pressure appears, 2 and 3
can become follow-up features.

#### Handoff

**Next role:** feature-specialist.

Outstanding questions for the feature-specialist:
- Confirm exact companion section headers per role against the live
  `*-prompt.md` docs (align skeleton to what each agent already writes).
- Decide whether gatekeeper "Analyzed Specs" pre-fill should glob `specs/*.feature`
  (all specs) or just `specs/<feature>.feature` (lean to all — gatekeeper reviews
  cross-spec conflicts).
- Confirm `--force` is the only overwrite path (no new flag needed).

---

## Implementation Plan

### Changed files

| File | Pkg | Change | Budget |
|------|-----|--------|--------|
| `internal/orchestration/plan_snapshot.go` | orchestration | Rename `requiredPlanInputs` → exported `RequiredPlanInputs`; update caller `validatePlanSnapshotInputs`. Pure rename. | stays ≤62 |
| `cmd/centinela/evidence_init.go` | main | After `Skeleton`, apply `skel.Inputs = evidence.PlanInputs(feature, role)` when non-nil. ~3 lines. | stays ≤68 |

### New files

| File | Pkg | Purpose | Budget |
|------|-----|---------|--------|
| `internal/evidence/plan_inputs.go` | evidence | `PlanInputs(feature, role)` delegating to `orchestration.RequiredPlanInputs` for the two plan roles, else `nil`. | ~20 |
| `internal/evidence/fill.go` | evidence | `FillMarker` const + `FillSlot(desc) string`. | ~15 |
| `internal/evidence/companion_skeletons.go` | evidence | Per-role markdown skeletons; `DefaultCompanionTemplate` switches on role. | ~70 |

### New / changed signatures

```go
// internal/orchestration/plan_snapshot.go  (PROMOTE — same body)
func RequiredPlanInputs(feature string) []string

// internal/evidence/plan_inputs.go  (NEW)
// PlanInputs returns the mechanical inputs pre-fill for a role, or nil when the
// role has no derivable inputs. Only big-thinker + feature-specialist derive.
func PlanInputs(feature string, role Role) []string {
    switch role {
    case orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial:
        return orchestration.RequiredPlanInputs(feature)
    default:
        return nil
    }
}

// internal/evidence/fill.go  (NEW)
const FillMarker = "<FILL: %s>"
// FillSlot renders a substance slot, e.g. FillSlot("the impl file path")
// -> "<FILL: the impl file path>". Markdown bodies only; never JSON lists.
func FillSlot(desc string) string { return fmt.Sprintf(FillMarker, desc) }

// internal/evidence/companion.go  (CHANGED — delegate to skeleton)
func DefaultCompanionTemplate(feature string, role Role) string // role-aware
```

### `evidence init` change (command path)

In `runEvidenceInit`, after `skel := evidence.Skeleton(...)`:

```go
if pre := evidence.PlanInputs(feature, role); pre != nil {
    skel.Inputs = pre
}
```

`Skeleton` itself is **unchanged** (keeps repair/docs templates clean).
Companion write already calls `DefaultCompanionTemplate(feature, role)`, which
now returns the role-aware skeleton — no wiring change there.

### Companion skeleton matrix (markdown only, `<FILL: …>` slots)

| Role | Sections seeded with FILL slots |
|------|--------------------------------|
| big-thinker | Problem / Scope / Risks / Rollout / Handoff |
| feature-specialist | Acceptance Criteria / Edge Cases / Data Model / Integration |
| senior-engineer | Files Changed / Design Notes / Tradeoffs |
| ux-ui-specialist | UI Paths / Mobile-First Notes / Components |
| qa-senior | Covered / Residual Risks / Acceptance Mapping |
| validation-specialist | Gate Results / Verify Results / Verdict |
| documentation-specialist | KB Pages / project-docs Entries / Outcome |
| merge-steward | Conflicts / Resolution / Ship Decision |
| (default) | one-line placeholder fallback (current behavior) |

Headers stay short; the feature-specialist aligns exact wording to the live
`*-prompt.md` docs (handoff question above).

### `artifact new` body upgrade (slice 3)

Per-kind body fns swap `_italic prose_` for `FillSlot("…")` and pre-fill
derivable lists:

- **gatekeeper** (`artifact_gatekeeper.go`): "Analyzed Specs" pre-filled by
  globbing `specs/*.feature` (deterministic; empty list if none). Findings /
  Recommendation become FILL slots. Needs a tiny filesystem helper (glob) kept
  in the body fn or a shared `artifact_derive.go` (~15 lines) to stay ≤100.
- **edge-cases / production-readiness / changelog / docs**: italic → FILL slots;
  no new derivation (nothing else is mechanically derivable). `**Status:**` and
  `**Date:**` lines are UNCHANGED (parsed by `centinela validate`).

### Pre-fill matrix (what each role's evidence JSON gets at `init`)

| Role | `inputs` pre-fill | `outputs` pre-fill | `edgeCases` |
|------|-------------------|--------------------|-------------|
| big-thinker | `RequiredPlanInputs(feature)` (every `docs/features/*.md` + plan) | empty (fill slot) | empty |
| feature-specialist | `RequiredPlanInputs(feature)` | empty | empty |
| senior-engineer | empty | empty | empty |
| ux-ui-specialist | empty | empty | empty |
| qa-senior | empty | empty | empty |
| validation-specialist | empty | empty | empty |
| documentation-specialist | empty | empty | empty |
| merge-steward | empty | empty | empty |
| gatekeeper / production-readiness | empty | empty | empty |

Scalars (Feature, Step, Role, Status, GeneratedAt, HandoffTo, MobileFirst)
unchanged from current `Skeleton`.

### Back-compat note

- No schema change: `inputs`/`outputs` stay `[]string`. Pre-existing `.workflow/`
  JSON validates unchanged.
- Pre-filled `inputs` pass `validatePlanSnapshotInputs` **by construction** — the
  pre-fill and the validator call the same `RequiredPlanInputs`. The validator
  checks the snapshot is a superset of required, never that files exist on disk,
  so a pre-filled init validates with zero appends.
- `outputs` stays empty → unchanged validator behavior (agent still lists its
  real outputs; `validateActionableOutputs` still enforces real files).
- `RequiredPlanInputs` promotion is a rename; the import-graph edge
  `internal/evidence → internal/orchestration` already exists, so no G2 change.
- No `<FILL: …>` ever enters a validated JSON list; companions are existence-only.

### Test plan (for qa-senior, not authored here)

- `evidence init <f> big-thinker` then `evidence validate <f>` passes for the
  big-thinker role with no manual `append` (slice 1 acceptance).
- `PlanInputs` returns `nil` for non-plan roles; `SchemaSkeleton` /
  `docsSpecialistPair` inputs stay empty (no `Skeleton` poisoning).
- `FillSlot("x")` == `"<FILL: x>"`; companion skeletons contain `<FILL:` per role.
- gatekeeper artifact "Analyzed Specs" lists existing `specs/*.feature`.
- Pre-existing minimal JSON (no pre-fill) still validates (back-compat).
