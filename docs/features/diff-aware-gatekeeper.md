# diff-aware-gatekeeper

## Problem

Built-in gates (`G1: File Size`, `G11: i18n`) walk the entire repository
every time `centinela validate` runs. On large repos this is slow, noisy
(historical violations dominate the report), and discourages running
gates during the inner dev loop. Existing violations in untouched files
also block legitimate work even when the current branch did not introduce
them.

## User Stories

- As a developer on a large repo, I want `centinela validate` to flag
  only gate violations introduced by my current branch, so the report
  is actionable.
- As a CI operator, I want CI runs to keep doing a full scan, so the
  pipeline still catches regressions in untouched files.
- As a maintainer configuring Centinela, I want a single TOML setting
  to opt into "diff-aware everywhere" or "always full", so the policy
  is explicit and team-wide.
- As a contributor with untracked new files, I want those files to be
  gated as part of my change, so new violations cannot slip in before
  I `git add`.

## Acceptance Criteria

- `centinela validate` defaults to **diff-aware** mode when run
  locally, and to **full** mode when the `CI` env var is `"true"` (or
  `"1"`).
- The diff base defaults to the merge-base with `main`. It is
  configurable via `[validate] diff_base = "<ref>"` in
  `centinela.toml` (e.g. `"master"`, `"develop"`).
- Diff-aware mode includes **untracked** files (output of
  `git ls-files --others --exclude-standard`) in the change set.
- The mode is selectable explicitly via flags:
  - `centinela validate --changed` forces diff-aware
  - `centinela validate --full` forces full scan
  - flags override the auto/CI defaults
- A new TOML knob `[validate] diff_mode = "auto" | "always" | "off"`
  controls the default:
  - `"auto"` (default): diff-aware locally, full in CI
  - `"always"`: always diff-aware (unless `--full` is passed)
  - `"off"`: always full (unless `--changed` is passed)
- G1 (file size) runs only over files in the change set when in
  diff-aware mode. Files outside the change set are not walked.
- G11 (i18n) runs in full when **any** locale file is in the change
  set; it is **skipped** with a "no locale changes" Pass result when
  no locale file is in the change set. Partial locale-file gating is
  not meaningful for key-completeness comparison.
- User validate commands listed under `[validate] commands` are **not**
  scoped by the diff — they always run in full. (Out of scope for v1.)
- The validate output header indicates the active mode and the diff
  base, e.g. `Built-in Gates (diff-aware: 7 files changed since
  origin/main)` or `Built-in Gates (full scan)`.
- If the project is not a git repo or `git` is unavailable, diff-aware
  silently degrades to full and emits a one-line notice in the report.
- If `git merge-base HEAD <diff_base>` fails (e.g. unrelated history,
  shallow clone), diff-aware degrades to full and emits a notice.
- The `centinela complete <feature>` validate step uses the same
  resolution as `centinela validate` (a passing validate is required).
- Backward-compatible: existing `centinela.toml` files without
  `[validate].diff_mode` or `diff_base` continue to work; default
  remains "auto" with base `main`.

## Edge Cases

- Empty diff set (no files changed since base): G1 passes with
  "no relevant changes"; G11 passes with "no locale changes". No
  violations reported.
- Repo with no `main` branch (only e.g. `master`): user must set
  `diff_base = "master"`; otherwise diff resolution fails and we
  degrade to full with a notice.
- Shallow clone (common in CI containers): `git merge-base` may fail;
  we degrade to full + notice. CI default is full anyway, so this is
  rarely hit.
- File renamed inside the diff: both old and new paths appear in
  `git diff --name-only` only when `--diff-filter=ACMR` is used;
  we filter to existing files on disk, so deleted/old paths are skipped.
- Symlinks pointing outside the repo: walked as in current logic
  (skipped by `isSourceFile`).
- File deleted in the diff: not on disk → naturally skipped.
- New file with G1 violation, not yet `git add`-ed: included in the
  change set via the untracked listing.
- Locale file changed but no new keys: G11 still runs (it must compare
  keys across locales). The change-set check is "any locale file
  touched", not "key was added".

## Risks

- Diff-aware can mask pre-existing violations in untouched files,
  giving a false sense of green. Mitigated by:
  (a) CI default = full scan;
  (b) explicit `diff_mode = "off"` for teams that want strict local
      enforcement;
  (c) the mode banner in every report so the user knows which mode
      ran.
- `git` shell-out introduces a new external dependency. Mitigated by
  graceful degradation to full when git is missing or fails.
- Branch-name assumption (`main`) is wrong for some repos. Mitigated
  by `diff_base` config and the degrade-to-full path with a notice.
- Performance: shelling out to git twice (diff + ls-files-others) on
  every validate adds ~tens of milliseconds. Acceptable given the
  current full-scan cost it replaces.
