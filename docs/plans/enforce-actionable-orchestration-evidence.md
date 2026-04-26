# Plan: Enforce Actionable Orchestration Evidence

1. Extend orchestration evidence validation to reject output entries that are not real repo paths.
2. Add role-specific output rules for `big-thinker`, `feature-specialist`, `senior-engineer`, and `qa-senior`.
3. Keep validation logic in `internal/orchestration/` and preserve thin command wiring.
4. Add focused unit tests for evidence validation error branches and workflow-step enforcement.
5. Add integration and acceptance coverage proving `centinela complete` blocks insight-only evidence.
6. Update docs or templates only if required by the docs step after runtime behavior is in place.
