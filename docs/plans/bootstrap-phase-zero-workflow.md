---
feature: bootstrap-phase-zero-workflow
---

# Plan: Bootstrap Phase 0 Workflow

## Problem

Roadmaps can miss foundational setup work. In greenfield projects, this allows
teams to start product features before infrastructure and quality scaffolding is
ready. Conversely, existing projects may not need a bootstrap phase at all.

## Design

1. Add a parser for `Project Stage` in `PROJECT.md`:
   - Values: `greenfield`, `existing`
   - Missing/unknown defaults to `greenfield`
2. Add roadmap helpers to identify `Phase 0: Bootstrap` features and completion.
3. Gate `centinela start`:
   - Greenfield: only bootstrap features can start until bootstrap is done.
   - Existing: skip bootstrap gating.
4. Make workflow steps dynamic per feature:
   - Bootstrap: `plan -> code -> validate`
   - Non-bootstrap: `plan -> code -> tests -> validate`
5. Harden tests artifact detection to ignore `.gitkeep`/dotfiles.
6. Update renderers/docs/templates for stage + dynamic step counts.

## Validation

- `go test ./...`
- `go run ./cmd/centinela validate`
