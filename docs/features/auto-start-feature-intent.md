# Feature Brief: Auto-start Workflow from New Feature Intent

## Problem
After a feature is completed, agents can continue coding without starting a new workflow when users discuss adding or extending functionality.

## Goal
Ensure Centinela automatically starts a new workflow when prompt intent indicates a new feature and no active workflow exists.

## Scope
- Prevent `done` workflows from allowing unrestricted writes.
- Add prompt hook to detect new-feature intent and auto-run `centinela start`.
- Show explicit directive when no active workflow exists.

## Acceptance Criteria
- If all workflows are `done`, non-roadmap writes are blocked until a new feature starts.
- Prompt intent like "add", "extend", or "new feature" auto-starts a workflow.
- Existing active workflows are never auto-replaced.
- Hook output clearly reports auto-started feature name.
