# Feature: lean-evidence-footprint

## Problem

Every completed feature commits ~15–25 files into `.workflow/`, and the
directory is tracked in git (only `telemetry/` and `analysis.json` are
ignored). Across the project's history this has grown to **1,419 tracked
files**, of which:

- **212 are zero-byte `.lock` files** — advisory locks that are never
  deleted (`internal/evidence/lock.go`'s release closure releases the OS
  lock and closes the handle but never `os.Remove`s the file). A test
  comment states outright: *"Lock files accumulate silently."*
- **535 are per-feature `.json` evidence** — the machine contract the
  orchestration validator reads. Nothing reads these after a feature
  merges; they are inert receipts in `main` and noise in every PR diff.

This is a token burner on two axes: PR diff size (review + reading) and
repo bloat (anything that greps or fans out over the tree).

## Goal

Stop committing the machine-only evidence plumbing while **keeping the
human-readable `-<role>.md` narratives** — those are valuable to reviewers
and as an LLM knowledge base. The workflow and CI must keep functioning
unchanged.

## Scope

1. **Ignore future machine plumbing.** Add to `.gitignore`:
   - `.workflow/*.json` (per-feature evidence + per-feature root state)
   - `!.workflow/roadmap.json` (the one required-committed json — the
     SessionStart rehydration hook reads it; `hook_setup.go` enforces it)
   - `.workflow/*.lock`
2. **Retroactively untrack** the 535 `.json` (excluding `roadmap.json`)
   and 212 `.lock` files already in the index via `git rm --cached`.
   Local copies are preserved; only the git index is cleared.
3. **Keep `-<role>.md` committed** — unchanged, this is the KB.

## Non-goals / explicitly out of scope

- **Not** modifying the locking semantics. Adding `os.Remove` to the lock
  release introduces a classic unlink-after-unlock race (a fresh `open`
  on the same path can lock a different inode), which is *why* the code
  leaves the files. Gitignore solves the committed-noise problem without
  that risk. Local 0-byte locks remain but are harmless and untracked.
- **Not** collapsing the `.json`/`.md` pair into one file — a larger
  change to the evidence contract + validator + role prompts, deferred.

## Why it's safe

- `centinela validate` (what CI runs, `cmd/centinela/validate.go:30`) runs
  **only gates** (lint/type/test/audit + `validate.commands`). It never
  calls `ValidateStep`/`ValidateEvidence`.
- Evidence `.json` is parsed **only** by `centinela complete`, locally,
  while the feature is active — where the files exist on disk regardless
  of gitignore.
- The rehydration hook reads only `roadmap.json`, which we keep.
