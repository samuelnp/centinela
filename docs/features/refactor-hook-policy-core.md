# Feature Brief: Refactor Hook Policy Core

## Problem
Hook policy decisions (what to block, what to allow, and how to message) are embedded inside command handlers.
This makes agent integrations harder to keep aligned and harder to test in isolation.

## Users
- Maintainers evolving Centinela hook integrations.
- Teams using Claude and OpenCode in parallel.

## Goals
- Move hook decision logic into a shared internal package.
- Keep `cmd/centinela/hook_*` as thin adapters.
- Preserve current behavior and messaging.

## Non-Goals
- Changing workflow step semantics.
- Changing gate or artifact requirements.

## Acceptance Criteria
- Prewrite policy can be evaluated through shared internal functions.
- Existing hook command behavior remains unchanged for current tests.
- New tests validate policy behavior without CLI wiring.

## Risks
- Behavior drift during extraction.
- Over-engineering abstractions that add little value.

## Decomposition
- Create `internal/hookpolicy` package for prewrite decision flow.
- Rewire `hook_prewrite` to call shared logic.
- Add targeted unit tests for policy scenarios.
