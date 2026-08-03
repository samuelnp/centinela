# Plan: binding-evidence-gates

**Phase:** 13 — Lighter Centinela · **Archetype:** canonical

## 1. Problem framing

Three evidence/artifact gates report success on input they document as
rejecting. Each is fail-OPEN. Fixing them means: (1) `handoffTo` is checked
against the workflow's own derived chain, not just non-empty; (2) `artifact
stamp` validates the `commands` array shape it splices, not just the reader;
(3) `validateChangelog` rejects the literal `<FILL: ...>` stub the CLI itself
scaffolds. None of the three redesign the evidence contract — they enforce
what `docs/architecture/evidence-contract.md` and the CLI's own templates
already promise.

## 2. Scope boundaries (v1)

**In scope:** the three fixes exactly as specified in the feature brief;
the CLI evidence-init pre-fill for `handoffTo` (must stop seeding a value the
new gate then rejects — see 3d below); a doc correction to
`evidence-contract.md` for the two per-role lines that are provably stale
(qa-senior → validation-specialist, gatekeeper → documentation-specialist),
mirrored to the scaffold asset copy.

**Out of scope (per brief):** `revise-to-plan-sheds-no-evidence`,
`spec-conflict-precheck-requires-merging-worktree`, redesigning the evidence
contract/role chain, retrofitting existing `.workflow/*.json` files in this
repo or downstream. Also out of scope for v1: broadening the `<FILL:>` stub
check beyond the changelog (see 4d).

## 3. Dependencies & assumptions, and the four resolved open points

**a) Expected-successor rule for `handoffTo`.** There is no existing "next
role" computation anywhere in the codebase — `docs/architecture/
evidence-contract.md`'s per-role lines are static prose, and two of them are
already stale relative to the code: `qa-senior → validation-specialist` is
wrong for every `adversarial-v1` workflow (the real successor is `gatekeeper`,
per `RoleGatekeeper` replacing `RoleValidationSpec` at the validate step), and
`gatekeeper → documentation-specialist` is wrong for internal (non-user-facing)
features under the canonical archetype, where `RequiredRolesForFeature`
returns **no** role for the docs step (project rule: internal features ship
only a changelog, no documentation-specialist evidence). A hardcoded chain
would get both of these wrong and break exactly the legacy and
archetype/surface-subset cases the brief calls out.

The derivation implemented instead is entirely built from functions that
already exist and already carry the contract/archetype/surface knowledge —
`workflow.RequiredEvidenceRoles(feature, step)` (contract-pin aware) and
`wf.OrderedSteps()` (archetype-aware):

1. Resolve `roles := RequiredEvidenceRoles(feature, step)` — the same slice
   the existing structural check already validates against.
2. Find `role`'s index `i` in `roles`. If `i+1 < len(roles)`, the expected
   value is `roles[i+1]` (a same-step hop — e.g. `senior-engineer →
   ux-ui-specialist` when UX is required, `big-thinker → feature-specialist`
   for the legacy plan pair).
3. Otherwise (role is last/only in its step), walk `wf.OrderedSteps()`
   forward from `step`, skipping any step whose `RequiredEvidenceRoles`
   resolves to **empty** (archetype omission OR the docs-step internal-feature
   skip — both collapse to the same "nothing required here" case, so no
   special-casing is needed for either), and take the first role of the first
   non-empty step found.
4. If no such step remains, the expected value is the literal `"complete"`.

This one algorithm reproduces every documented case (planner→senior-engineer,
big-thinker→feature-specialist, feature-specialist→senior-engineer,
senior-engineer→qa-senior or →ux-ui-specialist, ux-ui-specialist→qa-senior,
documentation-specialist→complete) **and** correctly derives the two stale
cases dynamically instead of by literal doc text, **and** naturally produces
`"complete"` for every terminal role under every archetype (spike's
`senior-engineer` at `code`, hotfix/refactor's `gatekeeper`/
`validation-specialist` at `validate`, canonical's `documentation-specialist`
at `docs`, and validate-step roles on an internal feature where docs requires
no evidence). `merge-steward` is out-of-band (step `"merge"`, never in
`OrderedSteps()`) and is handled as a separate, self-contained literal check:
`handoffTo` MUST be `"complete"` or `"user"` — this needs no workflow context,
so it lives directly in `orchestration.ValidateEvidence`, not the derived
chain.

**Doc correction:** `docs/architecture/evidence-contract.md`'s qa-senior and
gatekeeper/validation-specialist `handoffTo` lines are corrected to describe
the derivation (next required-evidence step's first role, or `complete`)
instead of a static literal, mirrored byte-for-byte to
`internal/scaffold/assets/docs/architecture/evidence-contract.md`. This is a
documentation-accuracy fix, not a contract redesign — the enforced behavior
was already implied by the code (`RequiredEvidenceRoles`'s own doc comment
calls out the gatekeeper/validation-specialist swap explicitly), the prose
just hadn't caught up.

**b) Independent slices, independent revert.** Yes — three slices, no shared
edit in any single file, ordered smallest/lowest-risk first (see §5).

**c) Where the commands-schema validator for defect 2 lives.** A single new
exported function, `gatereport.ValidateCommandsSchema(raw json.RawMessage)
error`, in a new file `internal/gatereport/commands_schema.go`. It is the only
place that knows the shape (`[]{argv: []string (non-empty), exitCode: int
(key MUST be present, not just defaulted), durationMs: int (optional)}`), and
is called from both sides that currently look at `commands`:
`gatereport.ParseVerification` (read path — `Assess` and everything built on
it) and `gatereport.Stamped` (write path — currently splices the raw bytes
verbatim with **no** shape check at all). Requiring the `exitCode` key to be
explicitly present (not just absent-and-zero) is a small independent
strengthening this unification buys for free: today a record missing
`exitCode` entirely silently defaults to `0` (= "passed") in the typed
`Command` struct, which is itself a smaller instance of the same "accepts
what it should reject" class the brief is about. `hasPassingValidate`'s
semantic "did a real `centinela validate` pass" check stays reader-only per
the brief ("the gate that reads it keeps its own checks") — it is an
admissibility judgment, not a shape check, and stamp time has no basis to
assert it (the commands array pre-dates the stamp; stamp never invents
claims).

**d) Stub-detection: changelog-specific or shared?** Changelog-specific for
v1. `internal/verify/claim_stubs.go`'s `outputs-not-stubs` check is a
**different, unrelated** mechanism — it inspects `.go` output files for empty
test bodies / boilerplate-only content; it has no notion of the `<FILL: ...>`
template marker and doesn't touch Markdown. It is not the right integration
point. The right shared primitive is the marker itself:
`evidence.FillMarker`/`FillSlot` (`internal/evidence/fill.go`) already
produces the exact string `validateChangelog` must detect, since
`changelogBody()` (`internal/evidence/artifact_changelog.go`) renders the stub
with `FillSlot(...)`. A new `evidence.HasFillMarker(s string) bool` sits next
to that constant as the single source of truth for "is this still a
template," and `internal/workflow/validate_docs.go` is the **only** caller
for v1. Broadening it to every companion `.md` report (which ALL use the same
marker via `companionSkeleton`) is explicitly deferred, not built now: a
legitimately-written report that discusses the fill-marker mechanism in prose
(plausible for any future feature touching evidence templating itself) would
be a false positive under a blanket rule, and the brief's Expected Outcome
only asks for the changelog. Flagged as a deferred finding (§ Deferred
Findings in the evidence report) rather than silently scope-creeped in.

## 4. Risks

| Risk | Impact | Likelihood | Mitigation |
|---|---|---|---|
| New `handoffTo` chain check invalidates already-completed evidence in THIS repo's own in-flight `.workflow/*.json` (this feature is built inside Centinela governing itself) | High (blocks `centinela complete` for this or another feature) | Medium | Confirmed via static grep: only the CURRENT step's roles are re-validated by `validateOrchestration` (past-completed steps are never re-checked); ran the full breakage inventory (§6) before writing code; the 3 confirmed breakages are test fixtures, not real workflow state |
| Parallel session `docstring-gate` (a different worktree) has its own `.workflow/docstring-gate.json` and role evidence; if it is currently sitting at a TERMINAL step with a handoffTo literal copied from the (stale) doc text (`documentation-specialist`/`validation-specialist`) instead of the derived `complete`, its next `centinela complete` would newly fail | Medium (blocks another feature's workflow, not this one's code) | Low–Medium (only bites if that session is exactly at its own terminal step AND used the stale literal) | Additive-only check, easy to spot-fix (`centinela evidence set docstring-gate <role> handoffTo complete`); non-terminal hops (senior-engineer→qa-senior etc.) are unambiguous today and unaffected; recommend a `centinela status docstring-gate` sanity check before merging this slice |
| `artifact stamp` schema validation (defect 2) makes a currently-malformed on-disk gatekeeper report (this repo's or docstring-gate's) newly fail at the NEXT stamp call | Low (stamp is re-run by the verifier at will, and a malformed record was already going to be non-grounded eventually) | Low (every existing test/fixture in this repo already writes well-formed `argv`+`exitCode`) | This is the intended fix, not a regression — a malformed record should fail loudly at write time; no repo fixture depends on tolerating one |
| Changelog stub rejection (defect 3) rejects a real, human-authored changelog line that happens to legitimately contain the substring `<FILL:` (e.g. describing this very feature) | Low | Very low | Scoped to literal marker match against `FillMarker`'s exact rendered form; changelog lines are one-line prose summaries, not documentation about templating, so collision is implausible; if it ever happens, `centinela evidence set <feature> changelog` (or hand-edit) still works, this isn't a hard block on legitimate text elsewhere |
| Broadening stub-detection to all companion reports (tempting, since `FillMarker` is shared) creates false positives on future features whose evidence legitimately discusses the marker | Medium if built | N/A (not building it) | Explicitly deferred (§3d) rather than shipped speculatively |
| `handoffForRole` prefill (evidence-init) still hardcodes the stale defaults after only the validator is fixed, so a freshly-scaffolded stub is invalid the moment it's written | High (foot-gun: the CLI's own prefill fails its own new gate) | High (certain, if left unfixed) | Folded into slice 1: `handoffForRole` now calls the same derivation (`workflow.ExpectedHandoff`) feature/contract-aware, falling back to the old static per-role default only when no workflow state is found (preserves existing no-workflow-context unit tests byte-for-byte) |

## 5. Rollout sequence (smallest correct slice first)

**Slice A — stamp commands schema (defect 2).** Zero test breakage, fully
self-contained to `internal/gatereport/`. Files:
- NEW `internal/gatereport/commands_schema.go` — `ValidateCommandsSchema(raw
  json.RawMessage) error`, `validateCommandEntry` (checks `argv` present +
  `[]string` + non-empty, `exitCode` present + int; extra keys ignored).
- `internal/gatereport/block.go` (`ParseVerification`) — after computing
  `body, ok := blockBody(report)`, before/alongside unmarshalling into
  `Verification`, extract the raw `commands` sub-value via `map[string]
  json.RawMessage` and call `ValidateCommandsSchema`; wrap a failure as
  `malformed centinela:verification block: %w`.
- `internal/gatereport/stamp.go` (`Stamped`) — validate the `commands` raw
  string (existing-block branch AND the `emptyCommands` fallback, for
  consistency/no-special-case) via `ValidateCommandsSchema` before splicing;
  return the error wrapped with the report path context (caller
  `workflow.StampVerification` already wraps with `%s: %w`).
- `internal/gatereport/check.go` (`assessRecord`) — no functional change
  required (its own `len(c.Argv)==0` loop still runs on the typed slice), but
  simplify the comment to note the shape is now ALSO guaranteed by
  `ParseVerification` upstream; keep the loop as defense-in-depth since
  `Command`'s zero-value Argv can still be empty if `ValidateCommandsSchema`
  somehow isn't invoked from some future caller — cheap belt-and-suspenders,
  not a duplicate rule (the SHAPE rule lives in one place; this remaining
  check is the pre-existing ADMISSIBILITY rule, semantically different).

**Slice B — changelog stub rejection (defect 3).** One file changed, one
constant reused, minimal breakage (comment-only, see §6). Files:
- `internal/evidence/fill.go` — add `HasFillMarker(s string) bool` (checks
  `strings.Contains(s, "<FILL:")`).
- `internal/workflow/validate_docs.go` (`validateChangelog`) — after the
  existing non-blank-line scan finds a candidate line, additionally reject it
  with a new error (`changelog entry is still a template placeholder for %q:
  %s (replace <FILL: ...> with a real one-line summary)`) when
  `evidence.HasFillMarker(line)` is true. Import `internal/evidence`
  (workflow already imports evidence-adjacent packages elsewhere in this
  tree, no cycle: evidence imports workflow, so this direction — workflow
  importing evidence — must be checked; see verification note below).

  **Cycle check:** `internal/evidence` currently imports `internal/workflow`
  (e.g. `roles_retired.go`). If `internal/workflow` imports `internal/
  evidence` back, that's a cycle. Resolution: define `HasFillMarker` (and the
  marker constant it wraps) somewhere BOTH sides can already reach without a
  new edge — since `internal/evidence` already depends on `internal/
  orchestration`, and `internal/workflow` already depends on `internal/
  orchestration` too, the marker helper is placed in `internal/orchestration`
  instead (a new tiny file, e.g. `internal/orchestration/fill_marker.go`,
  exporting `FillMarker` constant + `HasFillMarker`), and `internal/evidence/
  fill.go`'s existing `FillMarker`/`FillSlot` become thin wrappers/aliases so
  every existing caller of `evidence.FillMarker`/`FillSlot` is unaffected.
  `internal/workflow/validate_docs.go` then calls
  `orchestration.HasFillMarker` directly — no new cross-package edge, no
  cycle.

**Slice C — handoffTo chain validation (defect 1).** Largest, highest-risk,
shipped last so A and B are proven independently first. Files:
- `internal/orchestration/evidence.go` (`ValidateEvidence`) — add the
  self-contained merge-steward literal check: when `role ==
  RoleMergeSteward`, `e.HandoffTo` MUST be `"complete"` or `"user"`.
- NEW `internal/orchestration/handoff_read.go` — `ReadHandoffTo(path string)
  (string, error)`, a minimal JSON read (mirrors the existing read pattern in
  `ValidateEvidence`) so `internal/workflow` can read a role's handoffTo
  without duplicating JSON parsing.
- NEW `internal/workflow/handoff.go` — `ExpectedHandoff(feature, step
  string, role orchestration.Role) (expected string, ok bool)` (the
  derivation from §3a); `validateHandoffChain(feature, step string, roles
  []orchestration.Role) error` (loops `roles`, reads each via
  `orchestration.ReadHandoffTo`, compares against `ExpectedHandoff`, skips
  silently on read errors — those are already reported by the existing
  structural `ValidateRoles` call so this never double-reports); private
  `nextChainRole` helper (walks `wf.OrderedSteps()`).
- `internal/workflow/validate_orchestration.go` (`validateOrchestration`) —
  after `orchestration.ValidateRoles` succeeds, additionally call
  `validateHandoffChain(feature, step, RequiredEvidenceRoles(feature,
  step))` and fold any error through the same `annotatePlanContract` wrap.
- `internal/evidence/schema_init.go` (`handoffForRole` → `Skeleton`) —
  change signature to `handoffForRole(feature string, role Role) string`;
  first try `workflow.ExpectedHandoff(feature, stepForRole(role), role)`
  (guard merge-steward separately, unconditionally `"complete"`); fall back
  to the existing static switch (renamed `legacyHandoffForRole`) when
  `ExpectedHandoff` reports `ok == false` (no workflow state found —
  preserves the no-chdir unit tests in `schema_init_planner_test.go`
  byte-for-byte).
- `docs/architecture/evidence-contract.md` + its mirror in
  `internal/scaffold/assets/docs/architecture/evidence-contract.md` — correct
  the qa-senior and gatekeeper/validation-specialist `handoffTo` lines (§3a).

## 6. Test-breakage inventory (starting list; tests step re-runs full suite
to confirm no others)

Confirmed by tracing which call paths actually invoke
`workflow.ValidateArtifacts`/`validateOrchestration` (the only place the new
chain check runs) versus the many call sites that go through
`orchestration.ValidateEvidence` directly, `internal/verify.Verify` (claim
checking, unaffected), or `internal/verdict`/`internal/hookpolicy` (read-only
formatting, unaffected):

1. `cmd/centinela/complete_cmd_test.go::TestRunCompleteDoneAndValidatePath`
   (the `f2` sub-case) — legacy `validation-specialist` evidence at
   `step:"validate"` with `"handoffTo":"orchestrator"`. Feature `f2` has no
   `docs/features/f2.md` (internal). Expected under the derived rule:
   `"complete"` (validate is the last step with any required evidence for an
   internal feature under the canonical/default order). **Fix:** update the
   fixture's `handoffTo` to `"complete"`.
2. `cmd/centinela/complete_branches_test.go::TestRunCompleteValidateErrorAndWarningBranches`
   — same shape, feature `f`, same fix (`"complete"`).
3. `internal/workflow/validate_orchestration_docs_test.go::TestValidateArtifactsDocsStrictOrchestration`
   — `documentation-specialist` evidence at `step:"docs"` for a **user-facing**
   feature `f`, `"handoffTo":"orchestrator"`. `docs` is the last step in the
   canonical order, so expected is `"complete"`. **Fix:** update the fixture.

Not breakage (verified, not just assumed): every other `handoffTo` value
found across the test suite (`feature-specialist`, `validation-specialist` as
a value, `senior-engineer`, `qa-senior`, `documentation-specialist`,
`complete`, and the merge-steward `complete`/`user` fixtures) either already
matches the derived rule (`internal/workflow/validate_orchestration_plan_gate_test.go`,
`plan_evidence_helper_test.go` callers) or is consumed by a path that never
calls `validateOrchestration`
(`internal/verify` claim checks in `cmd/centinela/verify_gate_test.go`,
`complete_validate_gates_test.go`'s `runValidateGates` — which explicitly
defers report-admissibility/orchestration checks to a different call site per
its own doc comment —, `evidence_validate_uipaths_test.go`'s
`runEvidenceValidate` which calls `orchestration.ValidateEvidence` directly,
`internal/verdict`, `internal/hookpolicy` formatting tests, and
`mcp_verdict_helpers_test.go`'s MCP verdict path).

Also needs a **non-breaking update** (comment/assertion, not a fix): `internal/
evidence/artifact_changelog_test.go::TestRenderTemplateChangelogEmitsNonBlankOneLiner`
asserts the freshly-scaffolded changelog stub's first line is non-blank "so
the docs gate passes" — after slice B this specific stub (still carrying
`<FILL:>`) will no longer pass the docs gate (by design). The test's own
assertion (non-blank) still holds, so it won't fail, but its comment becomes
misleading and should be corrected in the tests step, alongside a NEW test
asserting `validateChangelog` rejects the literal scaffolded stub and passes
once filled in.

Defect 2 (stamp schema): **zero** breakage — every existing
`centinela:verification` fixture in the repo already has well-formed
`argv`/`exitCode` per command entry (verified across all 7 files carrying the
fence marker).

## 7. File-by-file summary

| File | Change | Slice |
|---|---|---|
| `internal/gatereport/commands_schema.go` (new) | `ValidateCommandsSchema` + entry validator | A |
| `internal/gatereport/block.go` | `ParseVerification` validates raw `commands` | A |
| `internal/gatereport/stamp.go` | `Stamped` validates before splicing | A |
| `internal/orchestration/fill_marker.go` (new) | `FillMarker` const + `HasFillMarker` | B |
| `internal/evidence/fill.go` | `FillMarker`/`FillSlot` become wrappers over orchestration's | B |
| `internal/workflow/validate_docs.go` | `validateChangelog` rejects unfilled stub | B |
| `internal/orchestration/evidence.go` | merge-steward literal handoff check | C |
| `internal/orchestration/handoff_read.go` (new) | `ReadHandoffTo` | C |
| `internal/workflow/handoff.go` (new) | `ExpectedHandoff`, `validateHandoffChain`, `nextChainRole` | C |
| `internal/workflow/validate_orchestration.go` | wire in `validateHandoffChain` | C |
| `internal/evidence/schema_init.go` | `handoffForRole` feature/contract-aware, legacy fallback | C |
| `docs/architecture/evidence-contract.md` + scaffold mirror | correct 2 stale per-role lines | C |
| `cmd/centinela/complete_cmd_test.go` | fixture fix (breakage #1) | C |
| `cmd/centinela/complete_branches_test.go` | fixture fix (breakage #2) | C |
| `internal/workflow/validate_orchestration_docs_test.go` | fixture fix (breakage #3) | C |
| `internal/evidence/artifact_changelog_test.go` | comment fix + new negative test | B |

All new/changed files stay well under the 100-line cap (largest new file,
`internal/workflow/handoff.go`, estimated ~45 lines).
