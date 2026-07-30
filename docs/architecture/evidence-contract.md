<!-- centinela:doc-version=1 template=docs/architecture/evidence-contract.md -->
# Orchestration Evidence Contract

Every step subagent writes a `.workflow/<feature>-<role>.json` file that the
orchestration validator parses before `centinela complete <feature>` will
advance the step. The validator lives in `internal/orchestration/` —
`evidence.go`, `output_rules.go`, `plan_snapshot.go`, `evidence_ux.go`.

Following this contract exactly avoids the round-trip where evidence is
rejected and rewritten.

## JSON schema (all roles)

```json
{
  "feature":     "<feature-slug>",
  "step":        "plan | code | tests | validate | docs",
  "role":        "planner | senior-engineer | ux-ui-specialist | qa-senior | gatekeeper | documentation-specialist (legacy: big-thinker, feature-specialist, validation-specialist)",
  "status":      "done",
  "generatedAt": "<RFC 3339 timestamp, e.g. 2026-05-12T14:30:00Z>",
  "inputs":      ["…repo-relative file paths the agent consulted…"],
  "outputs":     ["…repo-relative file paths the agent produced or modified…"],
  "edgeCases":   ["…short statements of cases handled (required for some roles)…"],
  "mobileFirst": true,
  "coverage":    85.0,
  "handoffTo":   "<next role or 'complete'>"
}
```

## Global rules

1. `feature`, `step`, `role` MUST match the invocation context — a mismatch
   fails validation with `mismatched evidence fields`.
2. `status` MUST be the literal string `"done"`.
3. `generatedAt` MUST parse as RFC 3339 (e.g. `2026-05-12T14:30:00Z`).
4. `inputs`, `outputs`, and `handoffTo` MUST be non-empty.
5. `outputs` entries MUST be real file paths that exist on disk **when
   `centinela complete` runs** — every role except documentation-specialist.
   Descriptive strings like `"Updated workflow"` will be rejected as
   `actionable outputs must be real files`.
6. `mobileFirst` is omitted unless the role is `ux-ui-specialist`.
7. `coverage` is an **optional** number — the per-package coverage percentage
   the role claims (e.g. `85.0` for 85%). Omit it when the role makes no
   coverage claim. When present, `centinela verify` / the `complete` gate
   re-derives measured coverage and **hard-fails** if the claim exceeds
   measured coverage beyond `verify.coverage_tolerance` (default 0.1%). A
   nil/absent value skips the coverage check rather than failing it. Set it
   with `centinela evidence set <feature> <role> coverage <value>` — it is
   typed (a bare number, optionally suffixed `%`) to forbid free-form prose
   claims.

## Per-role rules

### planner (step: plan)

- `inputs` MUST include the feature's own brief `docs/features/<feature>.md`
  and its plan `docs/plans/<feature>.md`. Additional inputs are allowed — the
  rule is *include*, not *only* — so evidence written before this rule shrank
  still validates. The validator computes the required set via
  `RequiredPlanInputs` (construction-derived, no filesystem glob) and rejects
  any missing entries with `missing feature-doc snapshot inputs`.
- `outputs` MUST include at least one real file under `docs/plans/` or
  `specs/`. Typically both the plan file at `docs/plans/<feature>.md` and
  the Gherkin spec at `specs/<feature>.feature`.
- `edgeCases` MUST be non-empty — the planner carries the spec lens.
- `handoffTo` → `senior-engineer`.

#### Legacy (pre `planner-v1`) workflows

A workflow whose `.workflow/<feature>.json` has no `planContract` predates the
unified planner and still requires the COMPLETE retired pair — `big-thinker`
(same snapshot-input rule, `handoffTo` → `feature-specialist`) and
`feature-specialist` (same snapshot rule, non-empty `edgeCases`, `handoffTo` →
`senior-engineer`). A partial set fails, and a workflow pinned to `planner-v1`
cannot satisfy its gate with legacy-named files.

### senior-engineer (step: code)

- `outputs` MUST include at least one **real implementation file** outside
  these prefixes: `.workflow/`, `tests/`, `docs/features/`, `docs/plans/`,
  `specs/`, `docs/project-docs/`. Pointing only at evidence or doc files
  fails with `senior-engineer outputs must include a real non-evidence
  implementation file`.
- `handoffTo` → `qa-senior` (or `ux-ui-specialist` when the feature is
  user-facing and that role is required).

### ux-ui-specialist (step: code, user-facing features only)

- `mobileFirst` MUST be present and set to `true`.
- `edgeCases` MUST contain all eight required UX tags (case- and
  separator-insensitive — `loading state`, `loading-state`, `loading_state`
  all match):
  - `mobile-first`
  - `visual-hierarchy`
  - `typography-hierarchy`
  - `responsive-layout`
  - `loading-state`
  - `empty-state`
  - `error-state`
  - `motion-and-reduced-motion`
- `outputs` MUST include real UI/asset paths declared for the feature
  (validator checks against `uiPaths` for the feature surface).
- `handoffTo` → `qa-senior`.

### qa-senior (step: tests)

- `outputs` MUST include at least one path under `tests/` **AND**
  `.workflow/<feature>-edge-cases.md`. Missing either fails with
  `qa-senior outputs must include at least one real test file and …`.
- `edgeCases` MUST be non-empty.
- `handoffTo` → `validation-specialist`.

### gatekeeper (step: validate)

The validate step's adversarial verifier. It is the only role
`RequiredRolesForFeature(feature, "validate")` returns.

- `outputs` MUST include `.workflow/<feature>-gatekeeper.md`.
- That report carries the machine-readable grounding record the complete
  gate enforces: a fenced ```` ```json centinela:verification ```` block
  with the `revision` and `treeDigest` that were verified and a non-empty
  `commands` array proving `centinela validate` was actually run. See
  [gatekeeper-prompt.md](gatekeeper-prompt.md). The block lives in the
  Markdown report, NOT in this JSON companion — the evidence schema is
  unchanged.
- `handoffTo` → `documentation-specialist`.

### validation-specialist (step: validate) — LEGACY

Retained for workflows started before the `adversarial-v1` validate
contract. No step requires this role any more; new features write
gatekeeper evidence instead.

- Only the global rules apply (no role-specific output type).
- `outputs` typically include `.workflow/<feature>-gatekeeper.md` and any
  other gate reports synthesised.
- `handoffTo` → `documentation-specialist`.

### merge-steward (out-of-band on `centinela merge`)

The Merge Steward runs outside the 5-step workflow. `centinela merge`
invokes it when `git merge` produces a text conflict OR when
`centinela validate` fails on the merged tree, so it is NOT included in
`RequiredRoles(step)`. When the steward writes evidence, the validator
applies these rules:

- `feature` MUST be the feature being merged.
- `step` MAY be the literal string `"merge"` — there is no workflow
  step gate keyed off this role.
- `role` MUST be `"merge-steward"`.
- `outputs` MUST include `.workflow/<feature>-merge-steward.md` (the
  human-readable report). Additional entries (proposed-diff files,
  test reproductions) are allowed.
- `handoffTo` MUST be the literal string `"complete"` on a successful
  resolution, or `"user"` when the steward escalates due to low
  confidence.
- `edgeCases` SHOULD enumerate every conflict class detected
  (text-conflict, post-merge-validate-failed, spec-contract).

### documentation-specialist (step: docs)

- Exempt from the "outputs must be real files" check (the validator skips
  `validateActionableOutputs` for this role).
- All other global rules still apply.
- `handoffTo` → `complete`.

## Worked example — planner

```json
{
  "feature": "demo-feature",
  "step": "plan",
  "role": "planner",
  "status": "done",
  "generatedAt": "2026-05-12T14:30:00Z",
  "inputs": [
    "docs/features/demo-feature.md",
    "docs/plans/demo-feature.md"
  ],
  "outputs": [
    "docs/plans/demo-feature.md",
    "specs/demo-feature.feature"
  ],
  "edgeCases": [
    "Existing users keep working without migration"
  ],
  "handoffTo": "senior-engineer"
}
```

The `inputs` list must include the feature's own brief
`docs/features/<feature>.md` and its plan `docs/plans/<feature>.md`.
Additional inputs are allowed — the rule is *include*, not *only* — so
evidence written before this rule shrank still validates.
