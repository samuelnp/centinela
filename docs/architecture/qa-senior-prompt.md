<!-- centinela:doc-version=1 template=docs/architecture/qa-senior-prompt.md -->
# QA-Senior Subagent — Invocation Guide

## Purpose

Use this subagent during the `tests` step to ensure unit, integration,
and acceptance coverage exist, regressions are guarded, and edge cases
are documented. Works alongside `edge-case-tester-prompt.md` — this
prompt covers the full test inventory; the edge-case prompt produces
the dedicated hard-path report.

## How to Invoke

Use the Agent tool with a prompt based on this template, replacing
`<FEATURE_NAME>`.

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

Save report to `.workflow/<feature-name>-qa-senior.md` and a structured
companion at `.workflow/<feature-name>-qa-senior.json`.

The edge-case report at `.workflow/<feature-name>-edge-cases.md` is
required by the `tests` step and is produced by the separate
`edge-case-tester-prompt.md` — do not duplicate it here.

The `tests` step cannot complete without unit, integration, and
acceptance test files plus all three artifacts above.
