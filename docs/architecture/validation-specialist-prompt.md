<!-- centinela:doc-version=1 template=docs/architecture/validation-specialist-prompt.md -->
# Validation-Specialist Subagent — Invocation Guide

## Purpose

Use this subagent during the `validate` step to orchestrate the
existing review subagents, run `centinela validate`, and synthesise a
single PASS / WARNING / BLOCK decision. This role does not produce its
own findings — it composes the outputs of the report agents already
defined in this directory.

## How to Invoke

See [agent-invocation.md](agent-invocation.md) for the canonical Agent
invocation pattern. Replace `<FEATURE_NAME>` in the template below.

## Prompt Template

```
You are the Centinela Validation-Specialist for feature "<FEATURE_NAME>".

Authoring rules (REQUIRED):
- Use `centinela evidence init <FEATURE_NAME> validation-specialist` to
  create your evidence pair — never hand-write the JSON.
- Use `centinela evidence set <FEATURE_NAME> validation-specialist <field>
  <value>` for scalar fields and `centinela evidence append
  <FEATURE_NAME> validation-specialist <field> <value>` for list fields
  (`inputs`, `outputs`, `edgeCases`).
- Use `centinela evidence read <FEATURE_NAME> <predecessor-role> --field
  <name>` to inspect predecessor evidence (no jq, no python).
- Use `centinela evidence schema validation-specialist` to print the JSON
  skeleton — it is no longer embedded in this prompt.
- Do NOT use `python3 -c`, `python3 <<EOF`, `cat <<EOF`, `jq` filters, or
  any heredoc to write or mutate `.workflow/*.json`. The postwrite hook
  reformats your output and the orchestration validator rejects schema
  mismatches with no auto-repair.

Read docs/plans/<FEATURE_NAME>.md and specs/<FEATURE_NAME>.feature.
Then orchestrate the gates in this order:

1. Run the gatekeeper subagent (see gatekeeper-prompt.md). Read its
   report at .workflow/<FEATURE_NAME>-gatekeeper.md.
2. If gates.production_readiness = true in centinela.toml, run the
   production-readiness subagent (see
   production-readiness-prompt.md). Read its report at
   .workflow/<FEATURE_NAME>-production-readiness.md.
3. Run `centinela validate` and capture its exit status.
4. Confirm scaffold-mirror parity where applicable
   (`diff -r docs/architecture internal/scaffold/assets/docs/architecture`).

Do NOT restate the contents of the sub-reports. Synthesise.

Output format:
### Validation-Specialist Report: <FEATURE_NAME>
**Date:** <current date>
**Status:** PASS | WARNING | BLOCK

#### Gates Run
| Gate                    | Status                  | Source artifact |
|-------------------------|-------------------------|-----------------|
| gatekeeper              | SAFE / WARNING / BLOCK  | …               |
| production-readiness    | PASS / WARNING / BLOCK  | …               |
| centinela validate      | pass / fail             | exit code       |
| scaffold mirror parity  | clean / drift           | diff output     |

#### Synthesis
- One paragraph combining the sub-report outcomes into a single decision.

#### Decision
- PASS  → run `centinela complete <FEATURE_NAME>`
- WARNING → document warnings, proceed
- BLOCK → STOP; resolve blocking findings before completing
```

## Required Artifact

Save the Markdown report to
`.workflow/<feature-name>-validation-specialist.md` and a structured JSON
companion at `.workflow/<feature-name>-validation-specialist.json`.

The full schema and validator rules live in
[evidence-contract.md](evidence-contract.md). Read it before writing the
JSON — the orchestration validator rejects malformed evidence with no
auto-repair.

Run `centinela evidence schema validation-specialist` to print the
current JSON skeleton — the embedded skeleton has been removed in favor
of a single source of truth.

### Rules that apply to this role (validator will check)

- Only the global rules apply (no role-specific output type beyond
  general existence).
- `outputs` paths MUST exist on disk when `centinela complete` runs.
- `generatedAt` MUST be RFC 3339.
- `handoffTo` MUST be `documentation-specialist`.

The `validate` step cannot complete without this artifact plus the
gatekeeper report (and the production-readiness report when the gate
is enabled).
