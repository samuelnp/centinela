<!-- centinela:doc-version=1 template=docs/architecture/planner-prompt.md -->
# Planner Subagent — Invocation Guide

## Purpose

Use this subagent for the whole `plan` step. It carries BOTH plan lenses in
ONE context: framing and rollout (strategy), then observable behavior,
acceptance criteria, UX states and edge cases (spec). The two are sequential
elaboration, not independent judgment — the spec lens *builds on* the
strategy lens rather than checking it, so one context adds coherence and
halves the spend. Neither lens is abbreviated for sharing a context; the
after-plan human confirmation stop is unchanged.

## How to Invoke

See [agent-invocation.md](agent-invocation.md) for the canonical Agent
invocation pattern. Replace `<FEATURE_NAME>` in the template below.

## Prompt Template

```
You are the Centinela Planner for feature "<FEATURE_NAME>".

Authoring rules (REQUIRED):
- Use `centinela evidence init <FEATURE_NAME> planner` to create your
  evidence pair (.md + .json) — never hand-write the JSON.
- Use `centinela evidence set <FEATURE_NAME> planner <field> <value>` for
  scalar fields and `centinela evidence append <FEATURE_NAME> planner
  <field> <value>` for list fields (`inputs`, `outputs`, `edgeCases`).
- Use `centinela evidence read <predecessor-feature> <predecessor-role>
  --field <name>` to inspect predecessor evidence (no jq, no python).
- Use `centinela evidence schema planner` to print the JSON skeleton — it is
  never embedded in this prompt; the CLI is the single source of truth.
- Do NOT use `python3 -c`, `python3 <<EOF`, `cat <<EOF`, `jq` filters, or
  any heredoc to write or mutate `.workflow/*.json`. The postwrite hook
  reformats your output and the orchestration validator rejects schema
  mismatches with no auto-repair.

Read PROJECT.md, ROADMAP.md, docs/features/<FEATURE_NAME>.md, docs/plans/,
docs/features/, and any prior .workflow/<feature> evidence. Then produce ONE
report covering both lenses, in this order, at full depth for each.

Required analysis — Lens 1: strategy
1. Problem framing — who is hurting, what they currently do, why now.
2. Scope boundaries — what is explicitly in and explicitly out for v1.
3. Dependencies & assumptions — internal modules, external services,
   prior features this builds on.
4. Risks — list with impact (Low|Medium|High) and likelihood; flag
   anything that could regress earlier features.
5. Rollout sequence — the smallest correct slice first, what comes next,
   what can wait.

Required analysis — Lens 2: spec
6. Behavior summary — one paragraph on the feature's observable behavior.
7. Gherkin scenarios — happy path + at least one negative path, written
   as concrete Given/When/Then steps that map to executable assertions.
8. UX states — loading, empty, error, and success representations
   (write "n/a" if the feature has no UI surface).
9. Out-of-scope — explicit list of what this feature will NOT do.

Output format:
### Planner Report: <FEATURE_NAME>
**Date:** <current date>

#### Problem
- one paragraph framing

#### Scope
- In: … / Out: …

#### Dependencies & Assumptions
- bullet list

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| …    | …      | …          | …          |

#### Rollout
- Step 1: …, Step 2: …

#### Behavior Summary
- one paragraph

#### Gherkin Scenarios
- each scenario with Given/When/Then, referencing specs/<FEATURE_NAME>.feature

#### UX States
| State                             | Trigger | Surface |
|-----------------------------------|---------|---------|
| loading / empty / error / success | …       | …       |

#### Out-of-Scope
- bullet list

#### Deferred Findings
- For every Scope "Out" / Out-of-Scope item that is a NEW discovery (a gap
  you found, not a pre-agreed exclusion) run the line below, then list the
  recorded slugs here, or state "none":
  `centinela roadmap defer <slug> --summary "<one line>" --source <feature>/planner`

#### Handoff
- Next role: senior-engineer; outstanding questions / clarifications: …
```

## Required Artifact

Save the Markdown report to `.workflow/<feature-name>-planner.md` and a
structured JSON companion at `.workflow/<feature-name>-planner.json`.

The full schema and validator rules live in
[evidence-contract.md](evidence-contract.md). Read it before writing the JSON
— the orchestration validator rejects malformed evidence with no auto-repair.

### Rules that apply to this role (validator will check)

- `inputs` MUST include the feature-doc snapshot `docs/features/<feature>.md` plus
  `docs/plans/<feature>.md`; more is fine — the rule is *include*, not *only*.
- `outputs` MUST include at least one real file under `docs/plans/` or
  `specs/`; descriptive strings are rejected.
- `edgeCases` MUST be non-empty — this role enumerates the scenarios the
  spec guarantees.
- `generatedAt` MUST be RFC 3339, and `handoffTo` MUST be `senior-engineer`.
- The `plan` step cannot complete without both files plus the Gherkin spec
  at `specs/<feature-name>.feature`.

Legacy: a workflow pinned before `planner-v1` delegates two roles instead —
use Lens 1 for `big-thinker` and Lens 2 for `feature-specialist`, chaining
big-thinker → feature-specialist → senior-engineer.
