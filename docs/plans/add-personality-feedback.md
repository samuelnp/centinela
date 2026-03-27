---
feature: add-personality-feedback
---

# Plan: Add personality to Centinela feedback

## Problem

Current UI uses strong structure and color but language is neutral and repetitive.
Users want outputs to feel more intentional and memorable while remaining concise.

## Approach

1. Add a small persona helper in `internal/ui/` that maps tone to expression.
2. Route shared render primitives through this helper so all messages adopt it.
3. Update CLI success/status wording to avoid duplicated rigid tokens like `OK`.
4. Keep Lipgloss-based tone colors as-is to retain ANSI behavior when supported.
5. Extend UI tests to assert persona text appears across core render paths.

## Files

- `internal/ui/panel.go`
- `internal/ui/render_status.go`
- `internal/ui/persona.go` (new)
- `internal/ui/*_test.go` affected by output changes

## Validation

- `go test ./internal/ui/...`
- `go test ./...`
