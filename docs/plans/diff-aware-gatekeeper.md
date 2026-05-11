# Plan: diff-aware-gatekeeper

## Scope

Add diff-aware execution to the built-in file-scoped gates (G1, G11) so
`centinela validate` can report only violations introduced by the
current branch. Default to diff-aware locally and full in CI. Make the
mode and the diff base configurable through `centinela.toml` and CLI
flags. No changes to user validate commands.

## Work Items

1. **New package `internal/gitdiff/`**
   - `gitdiff/resolver.go`: `func ChangedFiles(base string, includeUntracked bool) (Set, error)`
     - Runs `git merge-base HEAD <base>` to get the merge-base SHA.
     - Runs `git diff --name-only --diff-filter=ACMR <merge-base>` for
       tracked diff.
     - Runs `git ls-files --others --exclude-standard` for untracked.
     - Unions the two lists, normalizes to forward-slash relative paths.
     - Returns a typed `Set` (set-of-paths) plus a `Summary` describing
       resolved base + file count.
   - `gitdiff/set.go`: small `Set` type with `Contains(path string) bool`
     and `Len() int`. Keeps gate code clean of map plumbing.
   - `gitdiff/resolver_test.go` + `gitdiff/set_test.go`: unit tests with
     a tmp git repo fixture (init, commit base, branch, modify, untrack).

2. **Config additions in `internal/config/config.go`**
   - Extend `ValidateConfig` with:
     - `DiffMode string` (toml `diff_mode`) — accepts `"auto"`,
       `"always"`, `"off"`. Normalized in `applyDefaults` to `"auto"`
       when empty or unknown.
     - `DiffBase string` (toml `diff_base`) — defaults to `"main"`.
   - Helper `func (v ValidateConfig) ResolveMode(env Env) Mode` where
     `Mode` is `Full | Changed`. Encapsulates the CI detection
     (`Env.IsCI()` — reads `CI` env var) and the auto/always/off rules.
     Lives in a new `internal/config/validate_mode.go` to keep
     `config.go` thin.
   - `config_test.go`: unit cases for the truth table
     (mode × CI × CLI flag).

3. **Gate runner changes in `internal/gates/`**
   - Add `RunWithFilter(cfg *config.Config, filter *gitdiff.Set) []Result`
     in `gates.go`. `RunAll` stays as a thin wrapper that calls
     `RunWithFilter(cfg, nil)`. `nil` filter = full scan (current
     behavior).
   - `file_size.go`: pass the filter into `findOversizedFiles` and
     skip any path not in the set when the filter is non-nil.
   - `i18n.go` / `i18n_keys.go`: at the top of `checkI18n`, if filter
     is non-nil and no path under `cfg.I18n.Dir` is in the filter,
     return a Pass result with message
     `"No locale changes — gate skipped."`. Otherwise run as today.
   - No behavioral change when filter is nil.

4. **CLI wiring in `cmd/centinela/validate.go`**
   - Add `--changed` and `--full` bool flags. Mutually exclusive; error
     if both set.
   - New helper `resolveValidateMode(cfg, flags)` calls
     `cfg.Validate.ResolveMode` then applies flag overrides.
   - When mode = Changed: call `gitdiff.ChangedFiles`. On error or
     non-git repo, emit a one-line notice and fall back to Full.
   - When mode = Full: pass nil filter.
   - Build the header line: `Built-in Gates (diff-aware: N files
     changed since <base>)` or `Built-in Gates (full scan)`.
   - Refactor `executeValidation` to accept a `mode` argument so
     `runComplete` (validate step) reuses the same resolver. Keep the
     function under 100 lines — split helpers if needed.

5. **Hook surfacing**
   - `cmd/centinela/hook_postwrite.go` already shows step + risk; no
     change required for v1. The validate-step block in
     `cmd/centinela/complete.go` calls `executeValidation` — confirm
     it stays passing.

6. **Acceptance tests**
   `tests/acceptance/diff_aware_gatekeeper_acceptance_test.go`
   - Scenario 1: in a tmp git repo with one oversized file on `main`
     and a clean branch, `centinela validate` in diff-aware mode
     reports Pass.
   - Scenario 2: same repo, branch adds a new oversized file →
     diff-aware mode reports Fail for that file only; full mode
     reports both.
   - Scenario 3: `CI=true` forces full regardless of `diff_mode`.
   - Scenario 4: untracked oversized file is flagged in diff-aware.
   - Scenario 5: `diff_base = "master"` is honored; default `main`
     used otherwise.
   - Scenario 6: non-git directory → diff-aware degrades to full
     with notice line in output.
   - Scenario 7: locale file changed → G11 runs; no locale change →
     G11 short-circuits with Pass and skip message.

7. **Edge cases doc**
   `.workflow/diff-aware-gatekeeper-edge-cases.md` capturing the
   edge cases enumerated in the feature brief plus any discovered
   during code.

8. **Documentation updates** (deferred to docs step)
   - `README.md`: new section under "Gate Checks" — Diff-aware mode.
   - `docs/architecture/gatekeepers.md` (and its scaffold mirror):
     note that G1 and G11 honor the diff filter; G7 (manual) is
     unchanged.
   - `centinela.toml` reference block: new `diff_mode` and
     `diff_base` keys.
   - `internal/scaffold/assets/centinela.toml.template` (if it
     exists) updated to surface the new keys as commented defaults.

## Validation

- `go test ./...` passes including the new acceptance test and
  `internal/gitdiff/` unit tests.
- `centinela validate` passes (G1 file size, full coverage script).
- Coverage script remains ≥ 95%.
- Manual smoke:
  - `centinela validate` on this branch shows diff-aware header.
  - `CI=true centinela validate` shows full-scan header.
  - `centinela validate --full` shows full-scan header.
  - `centinela validate --changed` forces diff-aware even with
    `CI=true` set.
- New files added in this feature stay under 100 lines (G1).

## Compatibility

- Default `diff_mode = "auto"` reproduces today's behavior in CI and
  changes local behavior to diff-aware. Teams that want the old
  behavior set `diff_mode = "off"`.
- No change to gate semantics when filter is nil. `RunAll` remains
  callable and exported.
- No change to user validate commands behavior.
- No change to hook contracts (PreToolUse / PostToolUse JSON shapes).
- Backward compatible with existing `.workflow/<feature>.json` state.

## File-size budget

Target ≤ 100 lines per new file. Planned splits:

- `internal/gitdiff/resolver.go` (~80 lines)
- `internal/gitdiff/set.go` (~25 lines)
- `internal/gitdiff/resolver_test.go` (~90 lines)
- `internal/gitdiff/set_test.go` (~30 lines)
- `internal/config/validate_mode.go` (~50 lines)
- `internal/config/validate_mode_test.go` (~80 lines)
- `cmd/centinela/validate.go` (already 88 lines — add helpers; split
  flag-resolution into `validate_mode.go` if it grows past 100).
- `internal/gates/gates.go` (~60 lines after adding `RunWithFilter`).
- Acceptance test split across two files if needed
  (`*_diff_test.go` + `*_full_test.go`).

## Out of scope

- Diff-aware scoping for user validate commands
  (`[validate] commands`). Tracked as a future feature.
- Diff-aware scoping for G7 (manual code-review gate via gatekeeper
  subagent). The subagent already operates on diff in its prompt;
  no code change needed.
- Performance optimization beyond git shell-out (e.g. caching
  merge-base across invocations).
- Multi-base support (e.g. compare against two refs). Single
  configurable base only.
