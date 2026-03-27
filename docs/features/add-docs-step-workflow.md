# Feature Brief: Add Docs Step to Workflow

## Problem

Documentation generation still depends on user-triggered actions after validate,
which can leave feature docs stale or missing.

## Goal

Extend the workflow to include a required final `docs` step delegated to a
Documentation Specialist subagent.

## Scope

- Change workflow order to `plan -> code -> tests -> validate -> docs`.
- Enforce docs artifacts and strict orchestration evidence at `docs` step.
- Include docs step in bootstrap workflows.
- Update hooks, status views, progress counters, and docs references.
- Add unit/integration/acceptance coverage for the new step.
