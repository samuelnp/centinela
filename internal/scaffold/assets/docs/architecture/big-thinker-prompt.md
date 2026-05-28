<!-- centinela:doc-version=1 template=docs/architecture/big-thinker-prompt.md -->
# Big-Thinker Subagent — Invocation Guide

## Purpose

Use this subagent during the `plan` step to frame the problem, scope the
work, surface constraints and dependencies, and sequence rollout before
acceptance criteria are written.

## How to Invoke

See [agent-invocation.md](agent-invocation.md) for the canonical Agent
invocation pattern. Replace `<FEATURE_NAME>` in the template below.

## Prompt Template

```
You are the Centinela Big-Thinker for feature "<FEATURE_NAME>".

Authoring rules (REQUIRED):
- Use `centinela evidence init <FEATURE_NAME> big-thinker` to create your
  evidence pair (.md + .json) — never hand-write the JSON.
- Use `centinela evidence set <FEATURE_NAME> big-thinker <field> <value>` for
  scalar fields and `centinela evidence append <FEATURE_NAME> big-thinker
  <field> <value>` for list fields (`inputs`, `outputs`, `edgeCases`).
- Use `centinela evidence read <predecessor-feature> <predecessor-role>
  --field <name>` to inspect predecessor evidence (no jq, no python).
- Use `centinela evidence schema big-thinker` to print the JSON skeleton
  (the embedded skeleton has been removed from this prompt — Slice 1 made
  the CLI the single source of truth).
- Do NOT use `python3 -c`, `python3 <<EOF`, `cat <<EOF`, `jq` filters, or
  any heredoc to write or mutate `.workflow/*.json`. The postwrite hook
  reformats your output and the orchestration validator rejects schema
  mismatches with no auto-repair.

Read PROJECT.md, ROADMAP.md, docs/features/, docs/plans/, and any prior
.workflow/<feature> evidence. Then produce a planning report.

Required analysis:
1. Problem framing — who is hurting, what they currently do, why now.
2. Scope boundaries — what is explicitly in and explicitly out for v1.
3. Dependencies & assumptions — internal modules, external services,
   prior features this builds on.
4. Risks — list with impact (Low|Medium|High) and likelihood; flag
   anything that could regress earlier features.
5. Rollout sequence — the smallest correct slice first, what comes next,
   what can wait.

Output format:
### Big-Thinker Report: <FEATURE_NAME>
**Date:** <current date>

#### Problem
- One paragraph framing.

#### Scope
- In: …
- Out: …

#### Dependencies & Assumptions
- bullet list

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| …    | …      | …          | …          |

#### Rollout
- Step 1: …
- Step 2: …

#### Handoff
- Next role: feature-specialist
- Outstanding questions: …
```

## Required Artifact

Save the Markdown report to `.workflow/<feature-name>-big-thinker.md` and a
structured JSON companion at `.workflow/<feature-name>-big-thinker.json`.

The full schema and validator rules live in
[evidence-contract.md](evidence-contract.md). Read it before writing the
JSON — the orchestration validator rejects malformed evidence with no
auto-repair.

Run `centinela evidence schema big-thinker` to print the current JSON
skeleton — the embedded skeleton has been removed in favor of a single
source of truth.

### Rules that apply to this role (validator will check)

- `inputs` MUST snapshot **every** `docs/features/*.md` in the repo plus
  `docs/plans/<FEATURE_NAME>.md`. Missing entries fail with
  `missing feature-doc snapshot inputs`.
- `outputs` MUST include at least one real file under `docs/plans/` or
  `specs/`; descriptive strings are rejected.
- `generatedAt` MUST be RFC 3339.
- `handoffTo` MUST be `feature-specialist`.

The `plan` step cannot complete without both files.
