# Feature Brief: Enforce Step Subagent Orchestration

## Problem
Agents can skip delegation discipline and perform step work directly, which weakens Centinela's process guarantees and makes subagent outputs inconsistent.

## Goal
Enforce strict step-to-subagent orchestration where the main agent coordinates and required specialist subagents produce auditable evidence.

## Scope
- Define strict role requirements per step.
- Require paired markdown and JSON evidence for each required role.
- Block step completion when evidence is missing or invalid.
- Apply enforcement only to workflows started after this feature.

## Acceptance Criteria
- `plan` requires `big-thinker` and `feature-specialist` evidence pairs.
- `code` requires `senior-engineer` evidence pair.
- `tests` requires `qa-senior` evidence pair.
- `centinela complete` fails with explicit missing/invalid evidence details.
- Existing in-progress workflows are not retroactively blocked.
