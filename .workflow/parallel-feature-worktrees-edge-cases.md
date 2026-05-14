# Edge-Case Report: parallel-feature-worktrees

**Date:** 2026-05-14

## Risk Matrix

| Case | Impact | Likelihood | Why |
|------|--------|------------|-----|
| Feature name contains shell-unsafe characters (`/`, spaces, `..`); `git worktree add` fails silently or creates wrong path | High | Medium | `branchName()` identity transform, no validation of feature slug against `[a-z0-9_-]`; caller responsibility unclear. `Create()` will pass malformed name to git, which may fail or misbehave. |
| `.worktrees/<feature>` already exists as a tracked file (not directory) — collision with existing artifact | High | Low | `Exists()` checks `IsDir()`, so it would detect a file, but `MkdirAll()` on the parent would succeed, then `git worktree add` would fail with a cryptic error. No pre-flight check. |
| `worktree.DetectFeatureFromCwd()` misclassifies path on case-insensitive filesystems (macOS) — returns wrong feature | High | Medium | Path splitting uses `filepath.ToSlash()` (forward slash) then splits on `/`, but `filepath.FromSlash()` on reconstruction. Symlinks in parent path to `.worktrees/` (e.g. `link -> real/.worktrees`) break the logic. No normalization via `filepath.EvalSymlinks()`. |
| Hook cwd is a subdirectory inside a worktree (e.g. `.worktrees/alpha/src`) — feature resolution returns `("alpha", ".worktrees/alpha")` correctly, but `.workflow/` writes happen in `.worktrees/alpha/.workflow/` instead of repo root | Medium | Medium | `DetectFeatureFromCwd()` walks parents and stops at the first `.worktrees/<feature>` segment. If cwd is `.worktrees/alpha/src/subdir`, it correctly identifies `alpha`, but `MaybeProvision()` chdirs to `.worktrees/alpha`, making relative `.workflow/` writes repo-relative to the worktree, not the main checkout. **This is by design**, but callers outside hooks (e.g., `centinela validate`) must run from repo root or worktree root. Ambiguous in cwd-relative validate commands. |
| `centinela validate` runs from worktree with commands like `npm test` that assume repo root; test runner looks for `package.json` in worktree instead of main — tests fail or run wrong test suite | Medium | High | The spec says "commands are run from cwd of the worktree", so `npm test` in `.worktrees/alpha/` would look for `package.json` there, not at repo root. If the validate command is repo-root-relative and not worktree-self-contained, this breaks. No validation that commands are portable across cwd contexts. |
| `git worktree add` fails but `Create()` catches the error and returns the target path anyway if `Exists()` returns true after the failure — could return a partial/broken state | Medium | Low | `Create()` first checks `Exists()` (fast path), then calls `git worktree add`. If git fails mid-way (e.g., out of disk) and `os.MkdirAll()` left the directory in place, a second call to `Create()` would find the directory and return it as success even though it is not a valid worktree. |
| Merging with feature branch deleted out-of-band (user manually `git branch -D <feature>`) — `Merge()` attempts `git merge --no-ff <branch>` which fails, but the worktree still exists; unclear cleanup | Medium | Medium | `Merge()` will fail with "no such branch", caught as text-conflict path, worktree kept. User must manually clean up or re-delete the branch. No message guiding cleanup. |
| Main branch is detached HEAD (e.g., after a botched rebase) — `isDirty()` and merge both assume a branch-based main; behavior undefined | Medium | Low | `isDirty()` uses `git status --porcelain`, which works in detached state. `git merge --no-ff <branch>` fails with a specific error in detached state. No pre-flight check. Error message may confuse the user. |
| Main has staged but unpushed commits; user runs `centinela merge` — dirty check passes (no unstaged changes), merge proceeds, pollutes main with local commits | Medium | Medium | `isDirty()` checks `git status --porcelain`, which reports both staged and unstaged changes. So this is **not** a risk — the dirty check will catch staged commits. But the error message "commit or stash" may be unclear that staged is included. |
| `use_worktrees` config flag is missing from `centinela.toml` — defaults to false; existing project tries to use worktrees, feature silently goes to main checkout | Medium | High | The `WorkflowConfig` struct has `UseWorktrees bool` with default false. Config loading via `config.Load()` + TOML unmarshal will leave it false if the key is absent. No warning. Behavior is backward-compatible but confusing if the user expects per-feature isolation. |
| Feature flag flips from `true -> false` with active worktrees running — `centinela start` on a new feature will use the main checkout even though old worktrees exist | Medium | Medium | `MaybeProvision()` checks `cfg.Workflow.UseWorktrees` at call time. If the flag flips between features, old worktrees are orphaned. Spec says "warning not error" but no warning is implemented in the code. |
| Ignore-list sync idempotency: `.gitignore` already contains `.worktrees/` with trailing whitespace or comments; `containsLine()` trims and exact-matches, so it would NOT detect it as present and would append again | Medium | Medium | `containsLine()` does `strings.TrimSpace()` on both sides, so `  .worktrees/  ` matches `.worktrees/`. But an inline comment `*.log # ignore logs\n.worktrees/` would not match due to the comment. `.gitignore` doesn't support inline comments, so this is **not** a real risk, but the code assumes no trailing comments. |
| `tsconfig.json` with `"exclude"` as a non-array (malformed, e.g., `"exclude": "value"`) — `patchTsconfigExclude()` decodes, gets nil, appends, re-encodes as array, corrupting config | Low | Low | `decodeExcludes()` tries to unmarshal into `[]string`. If the JSON value is a string, it fails and returns nil. Then `append(nil, ".worktrees")` produces `[".worktrees"]`. This **repairs** the malformed config by converting to array, which is probably good behavior, but it is silent. Logged as a potential follow-up but not a blocker. |
| `tsconfig.json` missing the `"exclude"` key — `patchTsconfigExclude()` adds it as a new array | Low | Low | Code handles this: `decodeExcludes()` checks `if len(raw) == 0`, returns nil, then appends and encodes. Correct. |
| `tsconfig.json` is a symlink to another file; write permission denied or symlink points to read-only location — `os.WriteFile()` fails | Low | Medium | No check for symlinks before write. Error bubbles up correctly. Acceptable — mirrors existing file-handling behavior. |
| `.gitignore` is a symlink to `/dev/null` or read-only file; `appendIgnoreLine()` fails | Low | Medium | `os.WriteFile()` will fail with permission denied. Error propagates. Acceptable. |
| Ignore file ends without newline (e.g., `file.txt.gitignore` has no trailing newline) — `appendIgnoreLine()` checks suffix and adds newline if missing, then appends; idempotent on re-run | Low | Low | Code explicitly handles this: `if !bytes.HasSuffix(data, []byte("\n"))` then add one before appending. Idempotent. Correct. |
| Spec conflict pre-check: two features have identical `Given` clause but different wording for `Then` — correctly detected as conflict | Low | Low | `scenariosConflicts()` groups by exact `Given` string and compares `Then` strings. If wording differs, conflict is raised. Correct. |
| Spec conflict pre-check: two features have the same `Given` and same `Then` but different `Scenario` name — not flagged (should not be) | Low | Low | `scenariosConflicts()` only compares `Then` values when `Given` is the same. Identical `Given` + `Then` with different scenario names is **not** a conflict. Correct. |
| Feature has no `.feature` file at all (delete-only change) — spec pre-check silently skips it | Low | Low | `readSpecsFrom()` returns empty slice if directory does not exist or has no `.feature` files. `collectScenarios()` appends empty slice. No error. Feature is not included in conflict detection. Acceptable for v1 — deletion is not a spec conflict. |
| Spec file has multiple `Scenario:` blocks with the same `Given` but different first `Then` — parser only captures the first `Then` per scenario | Low | Medium | `parseScenarios()` uses `&& cur.Then == ""` guard, so only the first `Then` is captured per scenario. If a scenario has multiple `Then` steps (compound), only the first is used for conflict detection. Spec violations (multiple Then per scenario) are not caught by the parser, but this is a feature, not a bug — the parser is deliberately minimal. Acceptable if specs are well-formed Gherkin. |
| Main branch is modified between merge attempt and Steward verdict (e.g., another agent merges a feature while Steward is thinking) — Steward writes to a stale main tree | High | Low (by design) | The merge command does not lock main. Steward operates on the merged tree and may be out of date by the time its verdict applies. No distributed lock or re-validation. This is **out of scope for v1** per the big-thinker report, but it is a real concurrent-merge risk. Mitigation: operators should not run multiple merges in parallel; a v2 note should recommend advisory locking. |
| Steward times out during evaluation (e.g., agent invocation hangs) — merge command waits indefinitely or times out after a long delay, leaving the worktree in limbo | Medium | Low | Phase 2 wiring is not yet complete (per senior-engineer notes), so the actual Agent dispatch is a follow-up. No timeout handling is defined. Acceptable for v1 — the infrastructure for invoking Steward is in place, but the actual invocation is stubbed. |
| Steward returns "success" but no JSON evidence file is written (e.g., agent crash mid-write) — merge command returns success even though the evidence is incomplete | Medium | Low | The evidence-contract validator checks for `.workflow/<feature>-merge-steward.json` only when the `merge-steward` role is explicitly declared in CLAUDE.md. If the Steward role is invoked but fails to write evidence, the validator will catch it at the next `centinela validate` run. Acceptable if roles are enforced. |
| Merging when feature branch is identical to main (no-op) — `git merge --no-ff` succeeds with a fast-forward, creating a merge commit with no new code | Low | Low | `--no-ff` always creates a merge commit, even on fast-forward. This is correct — it preserves the feature integration boundary. No risk. |
| Merging when feature branch was force-pushed (diverged then reset) — git merge may produce unexpected results if the shared commit history is unknown | Low | Low | The merge command does not validate commit history. Users are responsible for ensuring branches are not force-pushed. Acceptable; this is a user workflow responsibility, not a Centinela guarantee. |
| `centinela validate` from a worktree writes output to `.workflow/` — which `.workflow/` directory? Main checkout or worktree? | Medium | High | `MaybeProvision()` chdirs into the worktree. `workflow.WorkflowDir = ".workflow"` is a relative path. So writes are to `.worktrees/alpha/.workflow/`. This is **correct** — per-worktree isolation. But if a validate command or post-step hook tries to read `centinela.toml` or `.workflow/` state, it must look in the worktree, not the main checkout. No clear documentation in the code. |
| `centinela validate` reads `centinela.toml` from main checkout instead of worktree `.workflow/` — gates and validate commands disagree on config | Medium | Medium | `centinela.toml` is at repo root. All worktrees share the same config. This is **correct** — config is global, not per-feature. But if a user expects per-worktree config overrides (e.g., different test commands per feature), this will surprise them. Out of scope for v1, but should be documented. |
| Hook cwd is exactly `.worktrees/` (no feature subdirectory) — `DetectFeatureFromCwd()` splits on `/`, looks for `.worktrees` segment, finds it, then tries to read `parts[i+1]`, which would be empty if cwd is `.worktrees` with no trailing content | Medium | Low | `cwd = ".worktrees"` would be split into `["", "worktrees"]` or similar depending on `filepath.ToSlash()` behavior on relative vs. absolute paths. The loop `for i := 0; i+1 < len(parts)` would not find `.worktrees/` because there is no next part after it. So it would return `("", "")`. Correct behavior — not inside a worktree. |
| `.worktrees/` exists as a tracked git file (e.g., someone committed it as a file, not directory) — `os.MkdirAll()` fails | Low | Low | `os.MkdirAll()` returns an error if the path exists but is not a directory. `Create()` would return a wrapped error. Acceptable. |
| Git binary is missing or `git` command fails in a way that is not caught — e.g., `gitRunner` panics or hangs | Medium | Low | `gitRunner` is a `var func()` for testability. The default uses `exec.Command("git", ...)`. If git is missing, `Command()` succeeds but `CombinedOutput()` fails with "exec format error" or similar. Errors are caught and wrapped. Acceptable. |
| Project root is not a git repository — `isGitRepo()` returns false, `MaybeProvision()` returns `("", nil)` silently, feature goes to main checkout | Medium | Medium | Correct behavior for non-git projects. But a user who misconfigures `use_worktrees = true` on a non-git project will see silent no-op, not a clear error. No warning is raised. Acceptable if the assumption is "use_worktrees only makes sense in git repos", but documentation should clarify. |
| Worktree path contains spaces or special shell chars — `git worktree add "rel/.worktrees/my feature"` — shell expansion might occur | Low | Low | `gitRunner()` uses `exec.Command()` with args as a slice, so no shell parsing happens. Path is safely passed. Correct. |
| A new feature is created while an old feature's worktree is being merged; both try to read/write `.workflow/` simultaneously — race condition | Low | Low | Each feature has its own `.workflow/<feature>.json` file. Simultaneous writes to different files are safe at the OS level. No risk. |
| `centinela merge` is run twice concurrently on the same feature — `git merge` is attempted twice on main | Medium | Low | The merge command does not use advisory locks. If run twice, both would attempt the merge. The second would fail (branch already merged or conflict). This is a user error (running concurrent merges), not a Centinela bug. Acceptable; document the serial merge requirement. |
| Worktree feature name is derived from branch name, which is derived from feature slug — if branch name is not exactly feature slug (e.g., `feature/<slug>` prefix), branch detection breaks | Low | Low | `branchName()` returns feature as-is (identity). So if the workflow feature is `"alpha"`, the branch is `"alpha"`, and the worktree is `.worktrees/alpha/`. Consistent. Correct. |
| `DetectFeatureFromCwd()` on Windows with mixed `\` and `/` separators — path splitting may be inconsistent | Low | Low | `filepath.ToSlash()` normalizes to `/`, split happens on `/`, then `filepath.FromSlash()` converts back. On Windows, backslashes are handled correctly by the stdlib. Acceptable. |
| Worktree `.workflow/<feature>.json` is written by `worktree.MaybeProvision()` chdir, then the process crashes before returning — worktree is left in an inconsistent state (exists but `.workflow/` is incomplete) | Medium | Low | `MaybeProvision()` only chdirs; it does not write `.workflow/`. The caller (`centinela start`) writes the workflow state after `MaybeProvision()` returns. If the process crashes between chdir and write, the worktree exists but `.workflow/` is missing. This is a general crash-safety issue, not specific to worktrees. Acceptable; document that `centinela start` should be re-run if it crashes. |
| Ignore-list sync is called but `.gitignore` is not writable (permissions or filesystem issue) — `appendIgnoreLine()` fails, but the feature start continues anyway (if sync is not critical-path) | Medium | Low | `syncWorktreeIgnores()` is called from `migrate.go` and `init.go`. Errors are returned and propagate, failing the command. Correct — ignore-list sync is critical for tooling to work. |
| Multiple concurrent `centinela start` calls for the same feature — race on `git worktree add` or `Exists()` check | Low | Low | `Exists()` and `Create()` are not atomic. If two processes call `Create(".", "alpha")` concurrently, both might find `Exists()` false and both call `git worktree add`. The second call would fail with "worktree already exists". The first succeeds, the second returns an error. Users must serialize `centinela start` calls; this is acceptable. |
| Steward escalates a merge with "low confidence" but the evidence JSON does not include the proposed diff — user cannot understand what the Steward proposed | High | Medium | The `merge-steward` prompt and evidence-contract are in place, but the Steward role wiring is not complete (phase 2 follow-up). Once wired, the evidence must include `proposed_diff` and `confidence` fields per the contract. Acceptable if the contract is enforced during role validation. |
| Steward escalates but the `.workflow/<feature>-merge-steward.json` file is never written — the evidence-contract validator does not catch the missing file because the role was not declared in CLAUDE.md | Low | Low | The `merge-steward` role is declared in the output_rules validator. If a Steward agent is invoked but the role is not declared, the validator will fail on the next `centinela validate`. Acceptable — the evidence contract is enforced. |
|`.worktrees/<feature>` directory is deleted manually while the feature workflow is in progress — hooks assume the worktree still exists and may crash or misbehave | Medium | Low | `DetectFeatureFromCwd()` would still find the feature from cwd if cwd is still inside the deleted directory (cwd is a process-local state, not a filesystem check). But `Exists()` would return false. Some functions may fail gracefully (e.g., `Remove()` is idempotent), others may fail cryptically. Acceptable if we document that worktree deletions must go through `centinela merge` or manual cleanup. |
| Two in-flight features edit the same `.feature` file, each adding their own Scenario — spec conflict detector sees both scenarios from both files but conflicting scenarios are in different files; detection works correctly, no false negative | Low | Low | If `specs/shared.feature` exists in both `.worktrees/alpha` and `.worktrees/beta` with different scenarios, the conflict detector reads both files and compares. If scenarios have the same Given but different Then, they are flagged. Correct. |
| Feature has a Gherkin feature file but no scenarios — `parseScenarios()` would return empty list; no conflict is detected | Low | Low | Correct — no scenarios = no conflicts. Acceptable. |
| Spec file is binary or has invalid UTF-8 — `os.ReadFile()` reads it, `parseScenarios()` receives garbled text, scan fails or produces nonsense | Low | Low | The scanner stops on invalid UTF-8 (bufio.Scanner skips malformed sequences). Worst case: the spec file is ignored for conflict detection. Acceptable; document that `.feature` files must be valid UTF-8. |

## Missing or Weak Scenarios

1. **Worktree path resolution under symlink traversal** — No test for cwd inside a symlinked parent directory (e.g., `/Users/link-to-project/.worktrees/alpha` where `link-to-project` is a symlink). `DetectFeatureFromCwd()` does not normalize via `filepath.EvalSymlinks()`.
   - **Test gap:** `tests/unit/worktree_path_test.go::TestDetectFeatureFromCwd_WithSymlinks`

2. **Shell-unsafe feature names** — No validation of feature slug. Names like `"alpha/../beta"`, `"alpha /beta"`, `"alpha;rm -rf"` are not rejected, passed directly to `git worktree add`.
   - **Test gap:** `tests/unit/worktree_provision_test.go::TestCreate_InvalidFeatureSlug`
   - **Mitigation needed:** Add `ValidateFeatureSlug()` function that enforces `[a-z0-9_-]+`.

3. **Ignore-list idempotency with whitespace variations** — `containsLine()` trims, so `.worktrees/` and `  .worktrees/  ` are treated as identical. But if the file has Windows CRLF line endings and the append logic uses LF, the idempotency might fail.
   - **Test gap:** `tests/unit/worktree_ignore_append_test.go::TestAppendIgnoreLine_WithCRLF`

4. **`centinela validate` command portability across cwd contexts** — The spec says "commands are run from the worktree cwd", but validate commands often assume repo-root context (e.g., `go test ./...`). No validation that commands are relative or cwd-portable.
   - **Test gap:** `tests/integration/worktree_validate_cwd_test.go::TestValidateFromWorktree_WithRepoRootCommands`

5. **Merge with detached main branch** — No pre-flight check for detached HEAD on main. `git merge --no-ff` fails with a specific error, but the error message is not user-friendly.
   - **Test gap:** `tests/integration/worktree_merger_test.go::TestMerge_WithDetachedMain`

6. **Steward escalation without evidence JSON** — Wiring is incomplete (phase 2), but the contract assumes `.workflow/<feature>-merge-steward.json` exists. No test for the case where Steward is invoked but the JSON is not written.
   - **Test gap:** `tests/acceptance/worktree_merge_steward_escalation_test.go::TestStewardEscalation_NoEvidence`

7. **Concurrent `centinela merge` on the same feature** — No locking. If two processes merge the same feature, both attempt `git merge` on main, the second fails. Error message is cryptic.
   - **Test gap:** `tests/integration/worktree_merger_concurrent_test.go::TestMergeConcurrency_SameFeature`

8. **Worktree collision with tracked `.worktrees` file** — Rare, but if the repo has a file called `.worktrees`, `os.MkdirAll()` fails. Error is not caught explicitly.
   - **Test gap:** `tests/integration/worktree_provision_test.go::TestCreate_WithTrackedWorktreeFile`

9. **Spec conflict with identical Given and Then across multiple features** — Should not flag as conflict. Test to ensure no false positive.
   - **Test gap:** `tests/unit/worktree_spec_conflicts_test.go::TestDetectSpecConflicts_SameGivenSameThen`

10. **Ignore-list sync with read-only files or symlinks** — No pre-flight permission check. Write failures are silent if sync is non-critical.
    - **Test gap:** `tests/integration/worktree_ignore_sync_test.go::TestSyncIgnores_ReadOnlyFile`

11. **`use_worktrees` flag missing from config** — Silent default to false. No warning that worktrees are not enabled.
    - **Test gap:** `tests/unit/worktree_config_test.go::TestMaybeProvision_NoFlagInConfig`

12. **Feature flag flip (true → false) with active worktrees** — Old worktrees are orphaned. No cleanup or warning.
    - **Test gap:** `tests/integration/worktree_config_migration_test.go::TestConfigFlip_TrueToFalse`

## Proposed/Added Tests

### Unit Tests

1. **tests/unit/worktree_path_test.go**
   - `TestPath_Basic` — `Path(".", "alpha")` returns `.worktrees/alpha`
   - `TestDetectFeatureFromCwd_InsideWorktree` — cwd inside `.worktrees/alpha/src` returns `("alpha", ".worktrees/alpha")`
   - `TestDetectFeatureFromCwd_OutsideWorktree` — cwd in main checkout returns `("", "")`
   - `TestDetectFeatureFromCwd_WithSymlinks` — cwd is symlinked; detection should work (or fail gracefully with documentation)
   - `TestDetectFeatureFromCwd_CaseInsensitive` — on macOS, `.WORKTREES/alpha` vs `.worktrees/alpha` (edge case; document behavior)
   - `TestIsInsideWorktree_True` — cwd inside `.worktrees/alpha` reports true
   - `TestIsInsideWorktree_False` — cwd in main returns false
   - `TestExists_True` — directory exists, reports true
   - `TestExists_False` — directory missing, reports false
   - `TestExists_File` — `.worktrees` is a file, not directory; reports false

2. **tests/unit/worktree_provision_test.go**
   - `TestCreate_Fresh` — creates new worktree and branch
   - `TestCreate_Idempotent` — re-run returns same path, no error
   - `TestCreate_BranchExistsLocally` — branch already exists locally, reuses it
   - `TestCreate_InvalidFeatureSlug` — rejects `"alpha/../beta"`, `"alpha;rm"`, names with spaces
   - `TestCreate_GitBinaryMissing` — graceful error when git is not available (mock gitRunner)
   - `TestCreate_WithTrackedWorktreeFile` — collision with tracked `.worktrees` file
   - `TestCreate_EmptyFeatureSlug` — rejects empty feature name

3. **tests/unit/worktree_ignore_append_test.go**
   - `TestAppendIgnoreLine_Fresh` — creates file and appends line
   - `TestAppendIgnoreLine_Idempotent` — re-run appends nothing
   - `TestAppendIgnoreLine_WithWhitespace` — trims and matches despite whitespace
   - `TestAppendIgnoreLine_WithoutTrailingNewline` — adds newline before appending
   - `TestAppendIgnoreLine_WithCRLF` — handles Windows line endings correctly
   - `TestContainsLine_ExactMatch` — finds exact match
   - `TestContainsLine_TrimmedMatch` — finds match after trimming
   - `TestContainsLine_CaseInsensitive` — does NOT match case-insensitively (asserts case-sensitive)

4. **tests/unit/worktree_ignore_tsconfig_test.go**
   - `TestPatchTsconfigExclude_Fresh` — adds `.worktrees` to empty or missing `exclude`
   - `TestPatchTsconfigExclude_Idempotent` — re-run makes no change
   - `TestPatchTsconfigExclude_MissingExclude` — creates `exclude` array if missing
   - `TestPatchTsconfigExclude_ExcludeNotArray` — silently repairs malformed config (convert string to array)
   - `TestPatchTsconfigExclude_MissingFile` — tolerates missing tsconfig.json
   - `TestPatchTsconfigExclude_InvalidJSON` — tolerates comments and malformed JSON (returns no-op, no error)
   - `TestPatchTsconfigExclude_PreservesOtherKeys` — does not lose other tsconfig fields

5. **tests/unit/worktree_ignore_sync_test.go**
   - `TestSyncIgnores_AllFiles` — patches all five ignore files
   - `TestSyncIgnores_Idempotent` — re-run touches no files
   - `TestSyncIgnores_PartialFiles` — when some ignore files are missing, creates them

6. **tests/unit/worktree_merger_git_test.go**
   - `TestIsDirty_Clean` — clean tree reports false
   - `TestIsDirty_Staged` — staged changes report true
   - `TestIsDirty_Untracked` — untracked files report false (workspace dirty, not tree dirty)
   - `TestIsDirty_Unstaged` — unstaged changes report true
   - `TestParseConflictedPaths_NoConflict` — returns empty slice on clean merge
   - `TestParseConflictedPaths_WithConflict` — parses conflicted file paths

7. **tests/unit/worktree_merger_test.go**
   - `TestMerge_CleanSuccess` — clean merge, validate passes, worktree removed
   - `TestMerge_TextConflict` — git merge fails, worktree kept, outcome flagged
   - `TestMerge_ValidateFail` — clean merge but validate fails, worktree kept
   - `TestMerge_DirtyMain` — pre-check fails before git merge
   - `TestMerge_ParseConflictedPaths` — outcome includes conflicted file list
   - `TestMerge_BranchDoesNotExist` — handles missing branch gracefully

8. **tests/unit/worktree_spec_parser_test.go**
   - `TestParseScenarios_SingleScenario` — parses one Scenario/Given/Then block
   - `TestParseScenarios_MultipleScenarios` — parses multiple blocks
   - `TestParseScenarios_OnlyFirstThen` — captures only first `Then` per scenario (as designed)
   - `TestParseScenarios_NoGivenOrThen` — skips incomplete scenarios
   - `TestParseScenarios_WithComments` — ignores comment lines

9. **tests/unit/worktree_spec_conflicts_test.go**
   - `TestDetectSpecConflicts_NoConflict` — same Given, same Then, no conflict
   - `TestDetectSpecConflicts_Conflict` — same Given, different Then, flags conflict
   - `TestDetectSpecConflicts_SameFeature` — ignores conflicts within the same feature branch
   - `TestDetectSpecConflicts_MultipleConflicts` — returns all conflicts
   - `TestDetectSpecConflicts_EmptySpecs` — handles missing spec directories gracefully
   - `TestFormatSpecConflicts_Readable` — formats conflict list in human-readable form

10. **tests/unit/worktree_start_hook_test.go**
    - `TestMaybeProvision_FlagOn` — creates worktree and chdirs
    - `TestMaybeProvision_FlagOff` — no-op when flag is false
    - `TestMaybeProvision_NotGitRepo` — no-op when project is not a git repo
    - `TestMaybeProvision_ChdirFails` — handles chdir failure (permissions)

### Integration Tests

1. **tests/integration/worktree_provision_e2e_test.go**
   - `TestProvision_Fresh` — full flow: create worktree, branch, verify on disk
   - `TestProvision_Resume` — re-run on existing worktree is safe
   - `TestProvision_BranchUpstreamOnly` — branch exists on origin but not locally; `git worktree add` handles it

2. **tests/integration/worktree_ignore_sync_e2e_test.go**
   - `TestSyncIgnores_RealFiles` — creates real `.gitignore`, `.prettierignore`, etc. and verifies content
   - `TestSyncIgnores_RealTsconfig` — patches a real `tsconfig.json` and verifies structure
   - `TestSyncIgnores_Permission` — attempts to patch read-only files and reports error

3. **tests/integration/worktree_merger_e2e_test.go**
   - `TestMerge_CleanMerge_Full` — real git repo, feature branch with changes, clean merge into main, validate succeeds, worktree removed
   - `TestMerge_TextConflict_Full` — overlapping edits, git merge fails, worktree kept, error returned
   - `TestMerge_ValidateFail_Full` — clean merge but validate command fails (mock validate), worktree kept
   - `TestMerge_DirtyMain_Full` — main has staged changes, merge refused

4. **tests/integration/worktree_hook_integration_test.go**
   - `TestHookWorkflows_InsideWorktree` — `loadActiveWorkflows()` filters to current worktree only
   - `TestHookWorkflows_OutsideWorktree` — `loadActiveWorkflows()` includes all workflows

5. **tests/integration/worktree_validate_cwd_test.go**
   - `TestValidateFromWorktree_WithRepoRootCommands` — runs `npm test` from `.worktrees/alpha/`; documents expected behavior

6. **tests/integration/worktree_merger_concurrent_test.go**
   - `TestMergeConcurrency_SameFeature` — two merges on same feature; documents error behavior

### Acceptance Tests

1. **tests/acceptance/worktree_parallel_start_test.go**
   - Scenario: "Start provisions multiple features in parallel"
     - Two features are started concurrently; both get their own worktrees and branches
     - No collision; both workflows are independent

2. **tests/acceptance/worktree_merge_clean_test.go**
   - Scenario: "Clean merge when git applies cleanly and validate passes"
     - Feature branch has no conflicts with main, validate passes
     - Worktree is removed, success message is printed

3. **tests/acceptance/worktree_merge_text_conflict_test.go**
   - Scenario: "Text conflict invokes the Merge Steward"
     - Overlapping edits cause git merge failure
     - Steward is invoked (stub for phase 2)
     - Worktree is kept for inspection

4. **tests/acceptance/worktree_merge_semantic_conflict_test.go**
   - Scenario: "Semantic conflict after a clean text merge invokes the Steward"
     - Clean git merge, but validate fails
     - Steward is invoked, worktree kept

5. **tests/acceptance/worktree_spec_conflict_test.go**
   - Scenario: "Spec conflict across in-flight worktrees is detected before merging"
     - Two features have conflicting Gherkin specs
     - Merge pre-check fails, no commits added to main

6. **tests/acceptance/worktree_dirty_main_test.go**
   - Scenario: "Merge fails fast when the main working tree is dirty"
     - Main has staged changes, merge refused

7. **tests/acceptance/worktree_resume_test.go**
   - Scenario: "Restarting a feature with an existing worktree resumes in place"
     - Worktree is not recreated, workflow state is preserved

## Residual Risks

### Hard to Test, Require Mitigation

1. **Symlink-based cwd traversal** — Testing symlinked parent directories is platform-specific and difficult to mock. Mitigation: Add a note in documentation that symlinks in the worktree path are unsupported or add explicit symlink normalization (via `filepath.EvalSymlinks()`) and test on CI.

2. **Concurrent merge attempts** — Full concurrency testing requires process-level synchronization and is expensive. Mitigation: Document that `centinela merge` is not concurrent-safe; users must serialize merges. Consider adding advisory locking in v2.

3. **Main branch modification during Steward evaluation** — Testing requires simulating a long-running Steward and concurrent git operations. Mitigation: Out of scope for v1; add a v2 note about distributed merge locking.

4. **Validate command portability across cwd** — Testing requires a diverse set of real test commands (npm, go, pytest, etc.) and verifying they work from both main checkout and worktree. Mitigation: Document that validate commands must be cwd-portable (use relative paths, avoid hardcoded repo-root assumptions). Add smoke tests for common commands (npm test, go test ./..., pytest).

5. **Steward evidence-contract enforcement** — Testing the full cycle (invoke Steward, write JSON, validate contract) requires the complete agent wiring, which is phase 2. Mitigation: Once Steward is wired, acceptance test the full escalation and evidence validation cycle.

6. **Case-insensitive filesystem behavior** — macOS HFS+ is case-insensitive but case-preserving. Testing requires a case-insensitive filesystem. Mitigation: Document that Centinela assumes case-sensitive filesystems (Linux, case-sensitive APFS on macOS). Test manually or add a CI step for case-insensitivity edge cases.

7. **File permission and symlink edge cases** — Full coverage would require testing on multiple filesystems (NTFS, HFS+, ext4) and with various permission configurations. Mitigation: Accept platform-specific behavior as out of scope. Document expectations (Unix file permissions, no hardlinks in ignore files).

8. **Shell-unsafe feature names** — Currently, feature slugs are passed as-is to `git worktree add`. No validation. Mitigation: **CRITICAL** — Add `ValidateFeatureSlug()` function that enforces `[a-z0-9_-]+` and rejects dangerous names. Document the naming constraint in PROJECT.md.

9. **Git binary failure modes** — `gitRunner()` may fail in ways not covered (e.g., git hangs on network issues, git segfaults). Mitigation: Add a timeout to `exec.Command()` (e.g., 30s) for git operations. Document timeout behavior.

10. **Worktree removal failure cascade** — If `git worktree remove` fails mid-way (e.g., dirty worktree), the branch cleanup is not attempted. Manual cleanup required. Mitigation: Add `--force` flag handling and clear error messaging. Document manual cleanup steps.

### Design Constraints

1. **Per-worktree config isolation not supported** — All worktrees share `centinela.toml` and `PROJECT.md`. Per-feature test commands or validation rules are impossible in v1. Mitigation: Document as a v2 feature request. Users can fork validate commands in PROJECT.md (e.g., `test_suffixes` includes both unit and integration tests by default).

2. **Manual main-branch sync** — Worktrees do not auto-pull from main. Long-running features may diverge. Mitigation: Document manual `git rebase origin/main` from the worktree. Add a `centinela worktree sync` command in v2 if auto-rebase becomes necessary.

3. **No shared artifact isolation** — Two features that edit shared docs (e.g., `docs/architecture/`) can have non-overlapping text but contradictory intent. The spec-conflict detector does not catch this (it is Gherkin-spec-only). Mitigation: Document as a v2 "shared-asset allow-list" feature. For v1, recommend code review for shared files.

4. **Steward escalation depends on prompt quality** — The Merge Steward is only as good as its prompt and the role definition. If the prompt is incomplete or the evidence contract is unclear, escalations may be inappropriate. Mitigation: **CRITICAL** — The merge-steward-prompt.md must be carefully authored with explicit escalation criteria. Acceptance tests must validate that escalations include full reasoning and proposed diffs.

---

### Summary of Top Gaps

- **Feature slug validation** (HIGH) — No check for shell-unsafe names; must add `ValidateFeatureSlug()`.
- **Symlink normalization** (MEDIUM) — `DetectFeatureFromCwd()` does not handle symlinked parent paths; consider `filepath.EvalSymlinks()`.
- **Validate command portability** (MEDIUM) — No enforcement that validate commands work from a worktree cwd; document and test key patterns.
- **Steward evidence contract** (HIGH) — Phase 2 wiring is incomplete; once added, must validate that evidence JSON is written and includes confidence + diff.
- **Concurrent merge safety** (MEDIUM) — No locking for multi-process merges; document serial-only or add advisory locks in v2.
- **Case-insensitive filesystem** (LOW) — macOS HFS+ behavior is untested; document assumptions and test manually.

