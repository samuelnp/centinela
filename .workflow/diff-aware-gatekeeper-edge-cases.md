# Edge Cases — diff-aware-gatekeeper

Captured during the plan + tests steps. Every case has a deterministic
behavior and (where it can be exercised) a test that proves it.

## Diff resolution

- **Empty diff set.** Branch identical to base. `checkFileSize` returns
  Pass with `"No relevant changes — gate skipped."` Covered by
  `TestCheckFileSize_FilterEmptySetSkipsAllFiles`.
- **Non-git directory.** `git merge-base` errors with "not a git
  repository"; resolver returns nil set + degrade summary. Validate
  prints `"notice: diff-aware degraded to full scan — not a git
  repository"` and continues with a full scan. Covered by
  `TestChangedFiles_DegradesOnNonGitRepo` and acceptance subtest
  `non-git directory degrades to full`.
- **Missing diff base.** Branch named differently from configured base
  (or unrelated history / shallow clone). `git merge-base` exits with
  "Not a valid object name". Degrade reason becomes
  `'diff base "main" not found'`. Covered by
  `TestChangedFiles_DegradesOnMissingBase` and acceptance subtest
  `missing diff base degrades to full with notice`.
- **Generic git failure.** Any other git error surfaces as
  `"git merge-base failed: ..."`. Covered by
  `TestChangedFiles_GenericMergeBaseFailureSurfacesMessage`.
- **`git diff` failure after a valid merge-base.** Surfaces as
  `"git diff failed: ..."`. Covered by
  `TestChangedFiles_DegradesOnDiffFailure`.
- **`git ls-files --others` failure.** Same degrade path, message
  `"git ls-files failed: ..."`. Covered by
  `TestChangedFiles_DegradesOnLsFilesFailure`.

## Change-set membership

- **Renamed file.** `--diff-filter=ACMR` lists the new path; the old
  path does not appear on disk, so `WalkDir` never visits it. No
  spurious violation.
- **Deleted file.** Not on disk → naturally skipped by `WalkDir`.
- **Untracked file with G1 violation.** Picked up via
  `git ls-files --others --exclude-standard` and flagged. Covered by
  acceptance subtest `untracked oversized file is included`.
- **Untracked file ignored by .gitignore.** Not listed by
  `--exclude-standard`. No false positive.
- **File outside the configured source roots.** Not walked. No false
  positive even if in the diff.

## i18n short-circuit

- **No locale file in change set.** `checkI18nFiltered` returns Pass
  with `"No locale changes — gate skipped."` without parsing locale
  files. Covered by `TestCheckI18nFiltered_SkipsWhenNoLocaleChanged`.
- **One locale file changed.** Full key comparison runs (must
  whole-repo for cross-locale parity). Covered by
  `TestCheckI18nFiltered_RunsWhenLocaleChanged`.
- **Locale touched but no keys added.** Still runs the full
  comparison. The change-set predicate is "any locale file touched",
  not "any key added".
- **`I18n.Dir` empty.** `localeFileInFilter` returns true so the
  gate runs as today; protects misconfigured projects from silently
  passing.
- **Nil filter (full-scan mode).** `checkI18nFiltered` falls through
  to `checkI18n`. Covered by
  `TestCheckI18nFiltered_NilFilterFallsThroughToFullCheck`.

## Mode resolution

- **Mutually exclusive CLI flags.** `--changed --full` exits non-zero
  with `"--changed and --full are mutually exclusive"`. Covered by
  acceptance subtest `mutually exclusive flags rejected`.
- **CLI flag beats config.** `[validate] diff_mode = "off"` +
  `--changed` runs diff-aware anyway. Covered by
  `TestResolveModeFlagOverrides`.
- **Config beats CI env.** `diff_mode = "always"` keeps diff-aware in
  CI; `"off"` keeps full locally. Covered by
  `TestResolveModeAlwaysAndOff`.
- **Auto mode in CI.** `CI=true` (or `CI=1`) forces full scan. Other
  values (including empty) treated as not-CI. Covered by
  `TestResolveModeAutoFlipsOnCI` and acceptance subtest
  `CI default is full scan`.
- **Empty `DiffMode`.** Normalized to `"auto"` so legacy configs keep
  working with no edits. Covered by `TestNormalizeDiffMode`.
- **Empty `DiffBase`.** Normalized to `"main"`. Covered by
  `TestNormalizeDiffBase`.

## Output contract

- **Header always shows mode.** Either `Built-in Gates (full scan)`
  or `Built-in Gates (diff-aware: N files changed since BASE)`. A
  green report cannot be misread as "fully scanned" by accident.
  Covered by acceptance subtests `local default is diff-aware` and
  `full mode reports historical violations`.
- **User `[validate] commands` are not scoped.** They always execute
  in full, regardless of mode. Covered by acceptance subtest
  `user validate commands always run in full`.
