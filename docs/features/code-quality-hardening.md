# Feature: code-quality-hardening

- surface: internal
- status: planned
- source: Go code-quality review (2026-06-09) — vet/gofmt mechanical pass + senior idiom review

## Problem

A full-repo quality review confirmed four defects that mechanical gates
currently miss:

1. **Evidence key-order drift (real bug).** `internal/hookpolicy/format_evidence_order.go`
   duplicates `internal/evidence/schema.go`'s `jsonKeyOrder` but is missing the
   `"coverage"` key. Its doc comment claims a drift-parity test exists in
   `format_evidence_test.go`; it does not. When the postwrite hook reformats a
   coverage-bearing evidence file, keys are reordered and the byte-stable
   serialization invariant documented in `schema.go` is broken.
2. **Formatting is ungated.** `gofmt -l` currently flags 10+ files, including
   non-test sources (`internal/evidence/schema.go`, `internal/verify/runner.go`,
   `internal/ui/render_gates.go`, `internal/ui/render_status.go`,
   `internal/worktree/merger.go`). Nothing in `[validate] commands` checks
   formatting, so drift accumulates.
3. **Inconsistent config-error policy.** `cmd/centinela/start.go` and
   `cmd/centinela/hook_context.go` swallow `config.Load()` errors
   (`cfg, _ :=`), while `cmd/centinela/complete.go` hard-fails on the same
   condition. A corrupted `centinela.toml` lets a user start a feature with
   empty config, then fails it at complete.
4. **Masked workflow-load errors.** `internal/workflow/state.go` `Load()`
   returns "no workflow found" for *any* read/parse failure, so corruption or
   permission errors are indistinguishable from absence.

## Goal

Fix all four defects and add the missing mechanical enforcement so they
cannot recur: a real key-order parity test and a gofmt check wired into
`centinela validate`.

## Non-goals

- No refactor of the CWD-relative path architecture.
- No test-suite quality overhaul (coverage-padding cleanup is separate work).
- No consolidation of the three shell-exec wrappers.
- No new built-in gate type; the format check rides `[validate] commands`.
