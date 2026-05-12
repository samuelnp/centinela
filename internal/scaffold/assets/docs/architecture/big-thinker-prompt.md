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

### big-thinker JSON skeleton

```json
{
  "feature": "<FEATURE_NAME>",
  "step": "plan",
  "role": "big-thinker",
  "status": "done",
  "generatedAt": "<RFC 3339 timestamp>",
  "inputs": [
    "docs/features/<FEATURE_NAME>.md",
    "docs/plans/<FEATURE_NAME>.md",
    "…every other docs/features/*.md in the repo (full snapshot)…"
  ],
  "outputs": [
    "docs/features/<FEATURE_NAME>.md",
    "docs/plans/<FEATURE_NAME>.md"
  ],
  "edgeCases": [
    "Optional but recommended — risks or invariants you flagged"
  ],
  "handoffTo": "feature-specialist"
}
```

### Rules that apply to this role (validator will check)

- `inputs` MUST snapshot **every** `docs/features/*.md` in the repo plus
  `docs/plans/<FEATURE_NAME>.md`. Missing entries fail with
  `missing feature-doc snapshot inputs`.
- `outputs` MUST include at least one real file under `docs/plans/` or
  `specs/`; descriptive strings are rejected.
- `generatedAt` MUST be RFC 3339.
- `handoffTo` MUST be `feature-specialist`.

The `plan` step cannot complete without both files.
