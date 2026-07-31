# binding-evidence-gates — planner

### Planner Report: binding-evidence-gates

**Date:** 2026-07-31

## Problem

Three gates in the evidence/artifact layer are fail-OPEN: each reports
success on data it documents as rejecting. (1) `handoffTo` is only checked
for non-emptiness, so `handoffTo: banana` completes a step — the documented
role chain (planner → senior-engineer → qa-senior → validation-specialist/
gatekeeper → documentation-specialist → complete) is decoration. (2)
`centinela artifact stamp` splices the `centinela:verification` block's
`commands` array back verbatim with no shape check, so a malformed or
hand-shaped record is only caught later by the reader — or never, if it
happens to satisfy the reader's narrower checks. (3) `validateChangelog`
requires only a non-blank line, so the scaffolded
`- <FILL: type>: <FILL: one-line summary>` stub (written verbatim by
`centinela artifact new <feature> changelog`) passes the docs gate untouched
— hit in practice on `dynamic-model-routing`. For a framework whose thesis is
"no false success," a gate that doesn't bind is worse than no gate: it
produces a green signal nobody earned.

## Scope

**In scope:** the three fixes as specified; the CLI's `evidence init`
pre-fill for `handoffTo` (`handoffForRole`), which must stop seeding a value
the new gate then rejects; a two-line correction to
`docs/architecture/evidence-contract.md` (mirrored to the scaffold asset)
where the documented per-role `handoffTo` target is provably stale relative
to the code (qa-senior, gatekeeper).

**Out of scope:** `revise-to-plan-sheds-no-evidence` and
`spec-conflict-precheck-requires-merging-worktree` (same fail-open class,
different subsystems); redesigning the evidence contract or role chain
itself — this feature enforces what is already documented; retrofitting
existing `.workflow/*.json` files in this repo or downstream; broadening the
`<FILL:>` stub-detection rule beyond the changelog gate (see Deferred
Findings).

## Dependencies & Assumptions

- `workflow.RequiredEvidenceRoles(feature, step)` and `wf.OrderedSteps()`
  already carry every piece of contract/archetype/surface context needed to
  derive the expected `handoffTo` successor — no new policy needs inventing,
  only composing.
- `internal/evidence` already imports `internal/workflow` (one-directional);
  `internal/workflow` importing `internal/evidence` back would cycle, so the
  shared `<FILL:>` marker primitive is placed in `internal/orchestration`
  (which both already depend on), not `internal/evidence`.
- The gatekeeper report's `centinela:verification` block is the only place
  `commands` shape matters; no other artifact carries a `commands` array.
- Assumes the parallel `docstring-gate` session's non-terminal role evidence
  (senior-engineer, qa-senior, etc.) already uses the unambiguous, non-stale
  per-role literals — only its OWN terminal-step evidence (if any) is at risk
  from the derived-vs-literal `complete` distinction.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|---|---|---|---|
| New `handoffTo` chain check invalidates already-completed evidence in this repo's own in-flight workflow state | High | Medium | Only the CURRENT step's roles are re-validated (past-completed steps are never re-checked); full breakage inventory run before coding — 3 confirmed, all test fixtures, not real workflow state |
| Parallel session `docstring-gate` sits at a terminal step with a handoffTo literal copied from the stale doc text instead of the derived `complete` | Medium | Low–Medium | Additive-only check, one-line fix (`centinela evidence set docstring-gate <role> handoffTo complete`); recommend a `centinela status docstring-gate` check before merging slice C |
| Stamp-time commands-schema validation (defect 2) makes an already-malformed on-disk report fail at its NEXT stamp call | Low | Low | Intended behavior, not a regression; every existing repo fixture already writes well-formed `argv`+`exitCode` |
| Changelog stub rejection (defect 3) false-positives on a legitimate line containing `<FILL:` in prose | Low | Very low | Marker match is scoped to the exact rendered string; changelog lines are one-line summaries, not prose about templating |
| Broadening stub-detection to every companion report (tempting — same marker) creates false positives on features that legitimately discuss the marker | Medium if built | N/A | Explicitly deferred, not built (see Deferred Findings) |
| `handoffForRole` prefill left un-updated after only the validator changes, so a freshly-scaffolded stub instantly fails the new gate | High | High if unfixed | Folded into slice C: prefill now calls the same derivation, falling back to the old static default only when no workflow state exists |

## Rollout

Three independently testable, independently revertible slices, smallest and
lowest-risk first:

1. **Slice A — stamp commands schema** (defect 2). Fully self-contained to
   `internal/gatereport/`; zero test breakage.
2. **Slice B — changelog stub rejection** (defect 3). One function change
   plus a shared marker primitive; one test needs a comment/assertion
   update, no functional breakage.
3. **Slice C — handoffTo chain validation** (defect 1). Cross-package
   (orchestration + workflow + evidence prefill), the largest and
   riskiest slice, shipped last once A and B are proven stable. 3 confirmed
   test-fixture breakages to repair in the tests step.

## Behavior Summary

After this feature: an out-of-chain `handoffTo` fails `centinela complete`
naming the expected successor, derived from the workflow's own contract
(never a hardcoded five-role list) — legacy pins, archetype subsets, and the
internal-feature docs-skip all resolve correctly, and the terminal role
always accepts `complete`. `centinela artifact stamp` rejects a malformed
`commands` array at write time with an error naming the offending entry,
using the same shape rule the reader (`gatereport.Assess`) already applies.
`centinela complete <feature>` (docs step) rejects a changelog that still
contains the literal `<FILL: ...>` scaffold marker, naming what to replace it
with; a filled-in one-line summary passes exactly as it does today.

## Acceptance Criteria (Gherkin)

Full scenarios live in `specs/binding-evidence-gates.feature` (11 scenarios,
grouped by defect). Summary:
- Defect 1: a rejected banana handoff; a valid terminal handoff on an
  internal feature (no hardcoded `documentation-specialist` assumption); a
  valid mid-chain handoff on a legacy (pre-adversarial-v1) workflow; a valid
  same-step handoff when UX is required; a rejected and an accepted
  merge-steward handoff.
- Defect 2: a commands entry missing `exitCode` rejected at stamp time; an
  entry with empty `argv` rejected at stamp time; a well-formed entry
  accepted and preserved verbatim.
- Defect 3: the literal scaffolded stub rejected naming the fix; a filled-in
  one-liner accepted.

## UX States

- `centinela complete <feature>` on an invalid `handoffTo`: exits non-zero,
  error names the field, the value found, and the expected successor (e.g.
  `handoffTo must be "qa-senior" for role "senior-engineer" at step "code",
  got "banana"`).
- `centinela artifact stamp <feature>` on a malformed `commands` array: exits
  non-zero, error names the malformed entry index and the missing/wrong key
  (`commands[2]: missing required key "exitCode"`), no partial write (atomic
  rename in `writeReport` is unchanged — a rejected stamp never touches the
  file on disk).
- `centinela complete <feature>` (docs step) on an unfilled changelog: exits
  non-zero, error says to replace `<FILL: ...>` with a real one-line summary
  and names the changelog path.
- All three failures are silent no-ops on success — no new output when the
  gate would have passed anyway.

## Out-of-Scope

- `revise-to-plan-sheds-no-evidence` (rewind path, own state-transition
  semantics).
- `spec-conflict-precheck-requires-merging-worktree` (merge subsystem).
- Reworking the evidence contract or role chain itself.
- Retrofitting existing `.workflow/*.json` files in this repo or downstream
  projects.
- Broadening `<FILL:>` stub-detection beyond the changelog gate (see
  Deferred Findings).

## Deferred Findings

- Broaden `<FILL:>` template-stub detection from changelog-only to every
  companion `.md` report (all use the same marker via `companionSkeleton`),
  IF unfilled-companion-report-passes-gate is ever hit in practice the way
  the changelog case was on `dynamic-model-routing`. Not built now — the
  blast radius (false positives on legitimate prose mentioning the marker)
  isn't worth taking on speculatively. Deferred via `centinela roadmap defer`
  (see Handoff for the exact slug).

## Handoff

Next role: **senior-engineer**. Outstanding questions for the code step:
none blocking — the derivation, file list, and cycle-avoidance path (marker
primitive lives in `internal/orchestration`, not `internal/evidence`) are
fully resolved in `docs/plans/binding-evidence-gates.md`. Implement slices in
order A → B → C; re-run the full suite after slice C specifically to confirm
the 3-item breakage inventory is complete (it was derived by static tracing,
not by running the suite pre-fix).
