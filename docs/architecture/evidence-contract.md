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
  "role":        "big-thinker | feature-specialist | senior-engineer | ux-ui-specialist | qa-senior | validation-specialist | documentation-specialist",
  "status":      "done",
  "generatedAt": "<RFC 3339 timestamp, e.g. 2026-05-12T14:30:00Z>",
  "inputs":      ["…repo-relative file paths the agent consulted…"],
  "outputs":     ["…repo-relative file paths the agent produced or modified…"],
  "edgeCases":   ["…short statements of cases handled (required for some roles)…"],
  "mobileFirst": true,
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

## Per-role rules

### big-thinker (step: plan)

- `inputs` MUST include **every** `docs/features/*.md` in the repo plus the
  current feature's plan at `docs/plans/<feature>.md`. The validator
  computes the required set via `requiredPlanInputs` and rejects any
  missing entries with `missing feature-doc snapshot inputs`.
- `outputs` MUST include at least one real file under `docs/plans/` or
  `specs/`. Typically: the feature brief at
  `docs/features/<feature>.md` and the plan file at `docs/plans/<feature>.md`.
- `handoffTo` → `feature-specialist`.

### feature-specialist (step: plan)

- Same snapshot-input rule as big-thinker.
- `outputs` MUST include at least one of: `docs/plans/<feature>.md`,
  `specs/<feature>.feature` (typically both, plus
  `docs/features/<feature>.md`).
- `edgeCases` MUST be non-empty.
- `handoffTo` → `senior-engineer`.

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

### validation-specialist (step: validate)

- Only the global rules apply (no role-specific output type).
- `outputs` typically include `.workflow/<feature>-gatekeeper.md` and any
  other gate reports synthesised.
- `handoffTo` → `documentation-specialist`.

### documentation-specialist (step: docs)

- Exempt from the "outputs must be real files" check (the validator skips
  `validateActionableOutputs` for this role).
- All other global rules still apply.
- `handoffTo` → `complete`.

## Worked example — big-thinker

```json
{
  "feature": "demo-feature",
  "step": "plan",
  "role": "big-thinker",
  "status": "done",
  "generatedAt": "2026-05-12T14:30:00Z",
  "inputs": [
    "docs/features/demo-feature.md",
    "docs/features/another-feature.md",
    "docs/plans/demo-feature.md"
  ],
  "outputs": [
    "docs/features/demo-feature.md",
    "docs/plans/demo-feature.md"
  ],
  "edgeCases": [
    "Existing users keep working without migration"
  ],
  "handoffTo": "feature-specialist"
}
```

The `inputs` list must enumerate **every** `docs/features/*.md` in the
repo — abbreviated above for brevity.
