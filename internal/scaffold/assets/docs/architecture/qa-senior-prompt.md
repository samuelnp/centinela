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

Authoring rules (REQUIRED):
- Use `centinela evidence init <FEATURE_NAME> qa-senior` to create your
  evidence pair — never hand-write the JSON.
- Use `centinela evidence set <FEATURE_NAME> qa-senior <field> <value>`
  for scalar fields and `centinela evidence append <FEATURE_NAME>
  qa-senior <field> <value>` for list fields (`inputs`, `outputs`,
  `edgeCases`).
- Use `centinela evidence read <FEATURE_NAME> senior-engineer --field
  <name>` to inspect predecessor evidence (no jq, no python).
- Use `centinela evidence schema qa-senior` to print the JSON skeleton —
  it is no longer embedded in this prompt.
- For the mandatory edge-cases companion artifact, run
  `centinela artifact new <FEATURE_NAME> edge-cases` first. This drops a
  templated `.workflow/<FEATURE_NAME>-edge-cases.md` stub you then fill.
- Do NOT use `python3 -c`, `python3 <<EOF`, `cat <<EOF`, `jq` filters, or
  any heredoc to write or mutate `.workflow/*.json`. The postwrite hook
  reformats your output and the orchestration validator rejects schema
  mismatches with no auto-repair.

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

Run `centinela evidence schema qa-senior` to print the current JSON
skeleton — the embedded skeleton has been removed in favor of a single
source of truth.

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
