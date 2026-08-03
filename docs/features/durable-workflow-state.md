# Feature: durable-workflow-state

**Phase:** 13 — Lighter Centinela
**Archetype:** canonical
**Depends on:** none

## Problem

`.workflow/<feature>.json` is the file every gate reads to decide what step a
feature is on, which roles its contract requires, and — since
`dynamic-model-routing` — which model each role runs at. It has two
data-integrity defects, both observed in practice rather than theorised.

1. **The write is not atomic.** `workflow.Save` is an unlocked whole-file
   `os.WriteFile`. A process killed between truncate and write leaves a
   truncated or empty file, and the feature's state is gone: `complete` cannot
   load it, and the workflow cannot be advanced or rewound. This session killed
   two long-running `complete` runs by accident (a background job inheriting a
   2-minute timeout) and saw main briefly carry an unparseable
   `.workflow/roadmap.json` from a different write path — the same failure mode,
   one file over. Concurrent writers (a `route set` while a `complete` is
   running) also silently lose one side's update.

2. **There is no schema version.** An older binary unmarshals the file into a
   struct without the newer fields and re-marshals it on the next save, silently
   dropping them. `modelRoutes` is the live example — the installed binary
   routinely lags a worktree build here, so this is a normal condition, not an
   exotic one. Nothing warns; the routing decisions simply vanish.

## Expected outcome

1. **Atomic, durable writes.** `workflow.Save` writes to a temp file in the same
   directory, fsyncs, and renames over the target, so a reader either sees the
   complete previous version or the complete new one — never a partial file. A
   killed process leaves the previous state intact.
2. **Concurrent writes do not silently lose updates.** Either serialize writers
   or detect the conflict and fail loudly. Silently losing a routing decision or
   a step advance is the outcome to eliminate; the plan chooses the mechanism.
3. **A schema version on the file**, written on save and checked on load. A file
   written by a NEWER binary than the one reading it must be refused with an
   actionable error rather than silently round-tripped through a lossy struct.
   Older files without the field load unchanged (absent = version 1).
4. Existing `.workflow/*.json` in this repo and downstream keep loading. No
   migration step, no manual edit.

## Out of scope

- The same durability problem in other `.workflow/` writers
  (`roadmap.json`, evidence JSON). This feature fixes the workflow state file
  and, if the atomic-write helper is reusable, exposes it — but converting the
  other call sites is separate work.
- Locking against a hostile writer. `.workflow/` is agent-writable by design;
  this is about crash and concurrency safety, not tamper resistance.
- Retrofitting a version into files already on disk.

## Constraints

- **Backward compatibility is the hard one.** Every existing workflow file, in
  this repo and in every downstream project, must keep loading with no manual
  step. An absent version field means version 1.
- Forward refusal must be actionable: name the file, the version it carries, the
  version this binary understands, and that upgrading is the fix.
- No gate weakened; no new import edge out of `internal/workflow`.
- 100-line file cap incl. `_test.go` in `cmd/` and `internal/`; per-package
  coverage ≥97% on touched packages.
- Tests must actually exercise the failure modes — a killed write and a
  concurrent write — rather than asserting the happy path twice.
