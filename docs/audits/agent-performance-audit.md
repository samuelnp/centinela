# Agent Performance Audit

## Scope

This audit reviews Centinela's native OpenCode specialist agents, workflow step coverage, and context-size improvement opportunities.

## Configured Agents

| Agent | Workflow step | Status | Notes |
|---|---|---|---|
| `big-thinker` | plan | Present | Strategy, scope, dependencies, constraints, risks, sequencing |
| `feature-specialist` | plan | Present | Acceptance criteria, specs, UX states, edge cases |
| `senior-engineer` | code | Present | Minimal compliant implementation |
| `ux-ui-specialist` | code | Conditional | Only for user-facing UI features |
| `qa-senior` | tests | Present | Tests, regressions, edge-case report |
| `validation-specialist` | validate | Present | Gatekeeper, validation, readiness checks |
| `documentation-specialist` | docs | Present | Docs updates, docs validation, generated docs |

## Step Coverage

| Step | Required native agent coverage | Result |
|---|---|---|
| plan | `big-thinker`, `feature-specialist` | Covered |
| code | `senior-engineer`, plus `ux-ui-specialist` for UI | Covered |
| tests | `qa-senior` | Covered |
| validate | `validation-specialist` | Covered |
| docs | `documentation-specialist` | Covered |

## Performance Findings

- Agent prompts are already short enough for normal use, but they repeat role names and output reminders.
- Most token cost still comes from repeated project instructions, workflow context, setup panels, and specialist evidence requirements rather than the native OpenCode agent prompts.
- `ux-ui-specialist` is correctly conditional; making it mandatory would waste tokens for backend/internal work.
- `validation-specialist` closes the only missing required-step coverage gap.

## Safe Prompt Reductions

- Replace `You are Centinela <role>` with compact role labels if future tests confirm behavior stays stable.
- Move detailed evidence schemas out of agent prompts and keep them in workflow directives only.
- Keep OpenCode agent descriptions short because they are used for routing.

## Behavior-Changing Ideas

- Add workflow intensity modes: quick, standard, strict.
- Skip `big-thinker` for low-risk docs or one-line fixes.
- Skip docs specialist when no user-facing docs or generated docs inputs changed.
- Use cheaper models for `documentation-specialist` and `qa-senior` when configured by the user.

## Recommendation

Keep the seven-agent set. The validate step previously missed native OpenCode coverage, and `validation-specialist` fixes that gap. Defer prompt trimming to a separate feature because current prompts are not the dominant context cost.
