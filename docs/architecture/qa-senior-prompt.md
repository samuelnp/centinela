<!-- centinela:doc-version=1 template=docs/architecture/qa-senior-prompt.md -->
# QA-Senior Subagent — Invocation Guide

## Purpose

Use this subagent during the `tests` step to ensure unit, integration,
and acceptance coverage exist, regressions are guarded, and edge cases
are documented. Works alongside `edge-case-tester-prompt.md` — this
prompt covers the full test inventory; the edge-case prompt produces
the dedicated hard-path report.

## How to Invoke

See [agent-invocation.md](agent-invocation.md) for the canonical Agent
invocation pattern. Replace `<FEATURE_NAME>` in the template below.

## Prompt Template

```
You are the Centinela QA-Senior for feature "<FEATURE_NAME>".

Read docs/plans/<FEATURE_NAME>.md, specs/<FEATURE_NAME>.feature, the
senior-engineer report at .workflow/<FEATURE_NAME>-senior-engineer.md,
and the edge-case report at .workflow/<FEATURE_NAME>-edge-cases.md.

Required analysis:
1. Test inventory by tier — list every unit, integration, and acceptance
   test added or modified for this feature.
2. Coverage gaps — scenarios from the .feature spec that are not yet
   covered by an executable assertion.
3. Acceptance test execution wiring — confirm validate.commands runs
   the acceptance tests (not just unit/integration).
4. Regression guards — tests added specifically to prevent prior bugs
   from recurring.

Output format:
### QA-Senior Report: <FEATURE_NAME>
**Date:** <current date>

#### Test Inventory
| Tier        | File | Scenarios |
|-------------|------|-----------|
| unit        | …    | …         |
| integration | …    | …         |
| acceptance  | …    | …         |

#### Coverage Gaps
- list scenarios from the .feature spec not yet asserted

#### Acceptance Wiring
- centinela.toml validate.commands snippet showing acceptance run

#### Handoff
- Next role: validation-specialist
- Edge-case report: produced separately by edge-case-tester subagent
```

## Required Artifact

Save the Markdown report to `.workflow/<feature-name>-qa-senior.md` and a
structured JSON companion at `.workflow/<feature-name>-qa-senior.json`.

The full schema and validator rules live in
[evidence-contract.md](evidence-contract.md). Read it before writing the
JSON — the orchestration validator rejects malformed evidence with no
auto-repair.

### qa-senior JSON skeleton

```json
{
  "feature": "<FEATURE_NAME>",
  "step": "tests",
  "role": "qa-senior",
  "status": "done",
  "generatedAt": "<RFC 3339 timestamp>",
  "inputs": [
    "docs/plans/<FEATURE_NAME>.md",
    "specs/<FEATURE_NAME>.feature",
    ".workflow/<FEATURE_NAME>-senior-engineer.md",
    ".workflow/<FEATURE_NAME>-edge-cases.md"
  ],
  "outputs": [
    "tests/unit/<file>_test.go",
    "tests/integration/<file>_test.go",
    "tests/acceptance/<file>_test.go",
    ".workflow/<FEATURE_NAME>-edge-cases.md"
  ],
  "edgeCases": [
    "Short, specific cases the test suite now covers (REQUIRED — non-empty)"
  ],
  "handoffTo": "validation-specialist"
}
```

### Rules that apply to this role (validator will check)

- `outputs` MUST include **at least one real path under `tests/`** AND
  exactly `.workflow/<FEATURE_NAME>-edge-cases.md`. Missing either fails
  with `qa-senior outputs must include at least one real test file and …`.
- `edgeCases` MUST be non-empty.
- All output paths MUST exist on disk when `centinela complete` runs.
- `generatedAt` MUST be RFC 3339.
- `handoffTo` MUST be `validation-specialist`.

The edge-case report at `.workflow/<feature-name>-edge-cases.md` is
required by the `tests` step and is produced by the separate
`edge-case-tester-prompt.md` — do not duplicate it here.

The `tests` step cannot complete without unit, integration, and
acceptance test files plus all three artifacts above.
