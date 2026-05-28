---
feature: roadmap-parallel-readiness
summary: See which roadmap features are safe to start in parallel right now, and be blocked from starting features whose prerequisites aren't done.
audience: end-user
status: done
---

## What it does

The roadmap now declares inter-feature dependencies as a first-class property, and the system shows you which features are safe to start right now — at a glance with 🔓 (ready) and 🔒 (blocked by …) markers. When you run `centinela roadmap ready`, you get just the features whose dependencies are all complete. And if you try to `centinela start` a feature whose prerequisites aren't done yet, the system blocks you with a clear message naming the unmet dependencies.

## When you'd use it

You're driving multiple Claude/Centinela instances in parallel worktrees, fanning out work without manually reasoning about the dependency graph. You want to know instantly which features you can pick up next without colliding on prerequisites — or you just want to spot-check that the feature you're about to start isn't blocked.

## How it behaves

- **Schema & backward compatibility**: Features can declare `dependsOn: ["feature-x", "feature-y"]` in `roadmap.json`; existing roadmaps without this field work exactly as before.
- **Validation**: Unknown dependencies or cycles are caught at load time with a clear error; roadmap validation also ensures declared dependencies refer to existing features.
- **Readiness derivation**: Each feature is classified as done / in-progress / ready (all deps done, planned) / blocked (one or more deps not done); blocked features list their unmet dependencies.
- **Roadmap markers**: `centinela roadmap` shows 🔓 next to ready features and 🔒 next to blocked ones (with the blocker names).
- **Ready-set command**: `centinela roadmap ready` lists just the features safe to start now, one per line; it shows a friendly empty-state message when nothing is ready.
- **Start guard**: `centinela start <f>` refuses to start a feature if any of its dependencies are not done, naming the blockers in the error.
- **Plural rehydration**: When a new session starts, the rehydration output lists all currently-ready features (not just a single "next" feature), so you can see the full parallel frontier and spin up multiple instances at once.

## Examples

```
$ centinela roadmap
Phase 0: Foundation
  ✓ feature-a (done)
Phase 1: Middleware
  🔓 feature-b (ready) — safe to start now
  🔒 feature-c (blocked by feature-b, feature-d)
Phase 2: Integration
  ⋯ feature-d (in-progress)

$ centinela roadmap ready
feature-b
feature-e

$ centinela start feature-c
Error: feature-c has unmet dependencies: feature-b (not done), feature-d (not done)
```
