# Orchestration Evidence: senior-engineer

- Feature: `diff-aware-gatekeeper`
- Step: `code`
- Outcome: Implemented the v1 diff-aware gatekeeper end-to-end while
  keeping every new and edited source file under the 100-line G1
  limit. Followed the plan in `docs/plans/diff-aware-gatekeeper.md`:

  - New package `internal/gitdiff/` (`set.go`, `resolver.go`):
    `Set` exposes `Contains` / `Len` / `HasPrefix` (the prefix helper
    is what G11 uses to test for "any locale file touched"). `Resolver`
    is injectable via a `Run` field so tests can stub the git
    shell-out. `ChangedFiles` unions `git diff --name-only
    --diff-filter=ACMR <merge-base>` with
    `git ls-files --others --exclude-standard`, and degrades to
    full scan by returning `(nil, Summary{Degrade: ...}, nil)`
    on any git failure.

  - Config (`internal/config/validate_mode.go`): adds `DiffMode` /
    `DiffBase` keys with `NormalizeDiffMode` / `NormalizeDiffBase`
    defaults, plus `ResolveMode(env, flag)` implementing the
    precedence "CLI flag > config mode > CI env".
    `applyDefaults` calls the normalizers so legacy configs keep
    working with no edits.

  - Gates (`internal/gates/gates.go`, `file_size.go`, new
    `i18n_filter.go`): `RunAll` is preserved as a thin shim over
    new `RunWithFilter(cfg, filter)`. `checkFileSize` skips any
    walked file not in the filter when filter is non-nil; the empty
    set emits a "No relevant changes — gate skipped" Pass.
    `checkI18nFiltered` short-circuits with the same Pass message
    when no path under `cfg.I18n.Dir` is in the filter; otherwise
    delegates to the existing `checkI18n` for the full comparison.

  - CLI (`cmd/centinela/validate.go`, new `validate_mode.go`, new
    `validate_runner.go`): adds `--changed` / `--full` cobra flags
    (mutually exclusive). `executeValidationWithFlag(flag)` is the
    new pipeline; `executeValidation()` stays callable with no args
    so `complete.go` and existing tests are untouched. The header
    line always shows the mode and base, e.g. `Built-in Gates
    (diff-aware: 7 files changed since main)` or
    `Built-in Gates (full scan)`. Diff resolution failures emit a
    one-line `notice:` and degrade to full.

  Updated four unit test call sites to pass `nil` for the new
  filter argument (preserves the legacy full-scan semantics under
  test). No production behavior change for the nil-filter path.

  Verified: `go build ./...` succeeds, `go vet ./...` clean,
  `go test ./...` green on `cmd/centinela`, `internal/gates`,
  `internal/config`. `internal/gitdiff` ships without tests at the
  code step (qa-senior to add unit + acceptance coverage at tests).

- Inputs: `docs/plans/diff-aware-gatekeeper.md`,
  `docs/features/diff-aware-gatekeeper.md`,
  `specs/diff-aware-gatekeeper.feature`,
  `.workflow/diff-aware-gatekeeper-big-thinker.md`,
  `.workflow/diff-aware-gatekeeper-feature-specialist.md`,
  the prior `cmd/centinela/validate.go`, `internal/gates/*.go`,
  `internal/config/config.go`.
- Outputs:
  `internal/gitdiff/set.go`,
  `internal/gitdiff/resolver.go`,
  `internal/config/validate_mode.go`,
  `internal/config/config.go` (edited),
  `internal/gates/gates.go` (edited),
  `internal/gates/file_size.go` (edited),
  `internal/gates/i18n_filter.go`,
  `cmd/centinela/validate.go` (edited),
  `cmd/centinela/validate_mode.go`,
  `cmd/centinela/validate_runner.go`.
- Handoff: `qa-senior`
