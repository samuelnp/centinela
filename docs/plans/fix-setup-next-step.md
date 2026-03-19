---
feature: fix-setup-next-step
---

# Plan: Fix Setup Next Step

## Problem

`RenderSetupNeeded()` in `internal/ui/render_setup.go` step 4 says:
"Suggest: centinela start <first-feature>"

This is wrong — the user should define the roadmap before starting any feature.

## Change

Replace step 4 in `RenderSetupNeeded()`:

```go
// before:
StyleMuted.Render("4. Suggest: centinela start <first-feature>"),

// after:
StyleMuted.Render("4. Tell the user: \"PROJECT.md is ready — next, let's define your roadmap.\""),
StyleMuted.Render("   Then immediately start the roadmap conversation (phases, features, briefs)."),
```

## Files Changed

- `internal/ui/render_setup.go` — 1 line replaced, 1 line added
