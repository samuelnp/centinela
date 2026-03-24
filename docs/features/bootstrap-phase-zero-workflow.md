---
feature: bootstrap-phase-zero-workflow
type: feat
---

# Feature: Bootstrap Phase 0 Workflow Enforcement

Centinela should enforce a mandatory `Phase 0: Bootstrap` only for greenfield
projects. This prevents teams from skipping foundational setup while avoiding
false requirements for existing projects adopting Centinela midstream.

## Scope

- Add explicit `Project Stage` support (`greenfield | existing`) in `PROJECT.md`.
- For `greenfield`, require roadmap bootstrap features before non-bootstrap work.
- Bootstrap features use `plan -> code -> validate` (no tests step).
- Keep tests mandatory for non-bootstrap features and block placeholder-only test
  directories from satisfying tests artifacts.

## Non-Goals

- Changing gate semantics outside bootstrap-specific workflow behavior.
- Introducing dynamic step counts unrelated to bootstrap classification.
