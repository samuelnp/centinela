---
feature: fix-validate-plan-by-name
---

# Plan: Fix validatePlan to Check by Filename

## Problem

`validatePlan` in `internal/workflow/validate.go` does:
```
for each file in docs/plans/*.md:
    if content contains feature-slug → found
```
A plan file named `project-bootstrap.md` that doesn't repeat the slug in its body fails
this check with: `no plan in docs/plans/ mentions "project-bootstrap"`.

## Fix

Replace the content-search with a direct filename check:
```go
if _, err := os.Stat(fmt.Sprintf("docs/plans/%s.md", feature)); err != nil {
    return fmt.Errorf("plan file not found: docs/plans/%s.md", feature)
}
```

## Files Changed

- `internal/workflow/validate.go` — replace `validatePlan` implementation
