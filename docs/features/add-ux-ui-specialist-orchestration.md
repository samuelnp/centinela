---
surface: internal
---

# Feature Brief: Add UX/UI Specialist Orchestration

## Problem
Strict orchestration can enforce planning, implementation, QA, and docs evidence, but it has no
specialist focused on the quality of the end-user experience for user-facing features.

## Goal
Add a conditional `ux-ui-specialist` role that is required only for user-facing features and that
enforces actionable UI/UX review outputs during the `code` step.

## Scope
- Add feature-aware orchestration role selection.
- Detect `user-facing` scope from feature brief metadata.
- Add configurable UI path prefixes with safe defaults.
- Require `ux-ui-specialist` alongside `senior-engineer` during `code` for user-facing features.
- Enforce actionable UI-layer outputs and edge cases for UX evidence.

## Acceptance Criteria
- Backend or internal features keep the current required role set.
- User-facing features require `ux-ui-specialist` evidence during `code`.
- UX evidence fails if outputs are summaries or do not touch UI paths.
- UX evidence passes when it points to real UI files and records edge cases.
- Orchestration directive output clearly includes the UX role when required.
