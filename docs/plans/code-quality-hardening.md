# Plan: code-quality-hardening

Four small, independent fixes plus the mechanical enforcement that keeps
them fixed. All changes stay within existing layer boundaries (G2) and the
100-line file limit (G1).

## 1. Evidence key-order drift + parity test

- Add `"coverage"` to `jsonKeyOrder` in
  `internal/hookpolicy/format_evidence_order.go` (position must mirror
  `internal/evidence/schema.go`: between `mobileFirst` and `handoffTo`).
- Replace the false claim in the comment ("Drift is caught by
  format_evidence_test.go") with a reference to the new parity test.
- Add a **behavior-level parity test**: marshal a `RoleEvidence` containing a
  coverage field via `evidence.MarshalJSON`, run the result through the
  hookpolicy formatter, and require byte-identical output. Place the test in
  an external test package (`package hookpolicy_test`) so the
  `internal/evidence` import can never create a cycle if a build-time
  dependency between the packages appears later (today neither imports the
  other). Mirrors the existing `TestBuildMatrixParity` pattern.

## 2. gofmt enforcement

- Run `gofmt -w` on all currently unformatted files (source and test).
- Add `scripts/check-fmt.sh`: runs `gofmt -l` over `cmd internal tests`,
  prints offenders, exits 1 if any (gofmt itself always exits 0, so the
  wrapper script is required; `[validate] commands` run natively without a
  shell, matching the existing `./scripts/check-coverage.sh` precedent).
- Append `./scripts/check-fmt.sh` to `[validate] commands` in
  `centinela.toml`.

## 3. Config-error consistency

- `cmd/centinela/start.go`: replace `cfg, _ := config.Load()` with the same
  hard-fail used by `complete.go` — a corrupted `centinela.toml` must stop
  `start` with an error naming the file.
- `cmd/centinela/hook_context.go`: hooks must never break the host session,
  so instead of hard-failing, surface the failure — inject a one-line
  `config warning: <error>` into the context output and continue with
  defaults. Silent `cfg, _ :=` is removed.
- Audit remaining `config.Load()` call sites for the same silent-discard
  pattern and align them with whichever of the two policies fits their
  surface (command = fail, hook = warn).

## 4. Workflow-load error transparency

- `internal/workflow/state.go` `Load()`: return "no workflow found for %q"
  only when the state file does not exist (`errors.Is(err, fs.ErrNotExist)`);
  wrap *read* failures with `%w` naming the file path. Note: parse failures
  are already wrapped ("invalid workflow file: %w") but lack the file path —
  add it so the spec's corrupted-state scenario can assert on it.
- Check `ActiveWorkflows`/callers so corrupted state files surface as errors
  (or at minimum a warning) rather than being silently skipped.

## Test plan

- Unit: parity test (item 1); `check-fmt.sh` exercised against a temp
  unformatted file; `start` with corrupted TOML errors; hook context with
  corrupted TOML injects warning and exits 0; `Load` distinguishes
  missing vs corrupted state file.
- Integration: `centinela validate` runs the new format check command.
- Acceptance: scenarios in `specs/code-quality-hardening.feature` executed
  from `tests/acceptance/`.

## Risks

- Hard-failing `start` on corrupted TOML is a behavior change — previously it
  silently proceeded. This is the intended fix; release notes must say so.
- New validate command must stay green in CI on a freshly formatted tree.
- G1: touched files must remain ≤100 lines; the hook_context warning path
  may require extracting a helper.

## Rollout

1. Item 1 (bug fix + parity test) — smallest correct slice, lands first.
2. Item 2 (format + script + toml wiring).
3. Items 3 and 4 (error-policy alignment) together, sharing test fixtures.
