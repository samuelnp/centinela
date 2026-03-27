# Feature Brief: Enforce Plan Snapshot Inputs

## Problem

Plan-step evidence currently proves delegation happened, but not that planners read the
current project feature context before proposing new work.

## Goal

Require `big-thinker` and `feature-specialist` plan evidence to include a full
snapshot of all `docs/features/*.md` files in evidence `inputs`.

## Scope

- Enforce snapshot coverage for plan step evidence validation.
- Include current feature brief in required snapshot set.
- Surface explicit missing-file errors.
- Add tests for positive and negative cases.
- Update architecture docs and scaffold mirror.
