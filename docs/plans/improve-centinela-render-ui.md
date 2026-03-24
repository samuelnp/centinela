---
feature: improve-centinela-render-ui
---

# Plan: Improve Centinela Render UI

## Problem

Users cannot reliably distinguish Centinela system output from LLM-generated
chat text. This causes confusion during hook flows (prewrite/context/postwrite).

## Approach

1. Add reusable branded UI primitives in `internal/ui/styles.go`.
2. Update render functions to include explicit source/channel headers.
3. Keep compact layouts with clear title, key lines, and action hints.
4. Improve hook outputs (`blocked`, `tag`, `context`) for strong distinction.
5. Update tests to assert new explicit branding and visual markers.

## Files

- `internal/ui/styles.go`
- `internal/ui/render*.go`
- `cmd/centinela/hook_*.go` (output integration only)
- `internal/ui/*_test.go` and command tests affected by output text changes

## Validation

- `go test ./...`
- `centinela validate`
