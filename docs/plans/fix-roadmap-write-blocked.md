---
feature: fix-roadmap-write-blocked
---

# Plan: Fix Roadmap Write Blocked

## Problem

`docs/features/` is classified as `TypePlan` in `classify.go`. The prewrite hook blocks any
TypePlan write when no feature workflow is active. During the roadmap phase, Claude needs to write
feature briefs to `docs/features/` before any workflow exists.

## Solution

Add `TypeRoadmap` to `classify.go` for files that belong to the roadmap phase:
- `docs/features/`
- `ROADMAP.md`
- `roadmap.json`

`TypeRoadmap` is always allowed by the prewrite hook (returns nil immediately).
Also allowed during plan/code workflow steps via `IsAllowedInStep`.

## Files Changed

- `internal/workflow/classify.go` — add TypeRoadmap constant, `isRoadmapArtifact()` helper,
  update ClassifyFile and IsAllowedInStep
- `cmd/centinela/hook_prewrite.go` — early return nil for TypeRoadmap
