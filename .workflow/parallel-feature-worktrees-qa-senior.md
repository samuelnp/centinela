### QA-Senior Report: parallel-feature-worktrees
**Date:** 2026-05-14

#### Test Inventory

| Tier        | File                                                          | Scenarios |
|-------------|---------------------------------------------------------------|-----------|
| unit        | tests/unit/worktree_slug_test.go                              | Accepts kebab-case; rejects empty, path-escape (`alpha/../beta`), slashes, shell metas (`;`, spaces), and uppercase |
| unit        | tests/unit/worktree_path_test.go                              | `Path` joins; `DetectFeatureFromCwd` inside, outside, after-dir-only, **symlinked-cwd (macOS /tmp)**; `IsInsideWorktree`; `Exists` dir vs missing vs file |
| unit        | tests/unit/worktree_ignore_sync_test.go                       | Appends `.worktrees/` to every ignore file; idempotency; tsconfig missing is no-op; tsconfig present patched; malformed tsconfig JSON is no-op |
| unit        | tests/unit/worktree_ignore_tsconfig_test.go                   | tsconfig missing `exclude` key gets one added; `exclude` as string is repaired to array |
| unit        | tests/unit/worktree_spec_conflicts_test.go                    | Same Given / different Then flags; same Given+Then no flag; missing specs/ dir not an error; same owner is not a conflict; `FormatSpecConflicts` empty case |
| integration | tests/integration/worktree_provision_test.go                  | Fresh worktree + branch; `Create` idempotent; reuses existing branch; rejects invalid slug before any disk write |
| integration | tests/integration/worktree_merger_test.go                    | Clean merge removes worktree; text conflict keeps worktree + enumerates paths + `StewardReason=git-text-conflict`; validate failure keeps worktree + `StewardReason=post-merge-validate-failed`; dirty main fails fast and does not invoke validator |
| integration | tests/integration/worktree_hook_resolution_test.go            | Workflow inside worktree returns only that feature; cwd outside worktree returns empty |
| acceptance  | tests/acceptance/parallel_feature_worktrees_test.go           | `MaybeProvision` honors flag (on provisions, off no-ops); `SyncIgnores` idempotency from a project without `.worktrees/` entries; spec-conflict pre-check flags contradictory Gherkin across an in-flight worktree |

Total: 21 unit, 6 integration (some functions cover multiple branches), 4 acceptance. All pass with `go test ./...`.

#### Coverage Gaps

The following Gherkin scenarios from `specs/parallel-feature-worktrees.feature` are exercised only at unit or integration level and not via a full CLI exec:

- **"Wizard syncs tool ignore lists for new projects"** — covered by `SyncIgnores` unit tests + `internal/scaffold/assets/centinela.toml` having `use_worktrees = true`; a full `centinela init` end-to-end acceptance run is deferred.
- **"Merge Steward escalates uncertain resolutions to the user"** — the `MergeOutcome.StewardReason()` and `StewardHint()` contracts are covered, but the actual Agent-dispatch is acknowledged as a v1.1 follow-up in the senior-engineer report; no test stub for the dispatch yet.
- **"Restarting a feature with an existing worktree resumes in place"** — covered by `TestCreate_Idempotent_SecondCallNoOp` (lower-tier), not by a full `centinela start` re-run acceptance scenario.

These are explicit gaps for the validation-specialist to weigh against the production-readiness gate.

#### Acceptance Wiring

`centinela.toml` `[validate].commands` already runs the full Go test tree, which includes acceptance tests:

```toml
[validate]
commands = [
  "go test ./...",
  "./scripts/check-coverage.sh"
]
```

No change required — `go test ./...` recurses into `tests/acceptance/` and the new `parallel_feature_worktrees_test.go` runs as part of every validate.

#### Edge Cases Covered (selected)

- Shell-injection / path-escape slugs (`alpha/../beta`, `;rm`, spaces) rejected by `ValidateFeatureSlug` before any disk side-effect.
- macOS symlink-masked cwd resolved via `filepath.EvalSymlinks` in `DetectFeatureFromCwd`.
- Idempotent `SyncIgnores` re-runs across all five ignore files + tsconfig.
- Malformed tsconfig JSON treated as no-op (does not crash); missing `exclude` key gets one added; string `exclude` repaired to array.
- Merger pre-flight blocks on dirty main without invoking the validator.
- Spec-conflict detection ignores same-owner self-comparisons (no false positives within a single feature's spec).

#### Handoff

- Next role: validation-specialist
- Edge-case report: `.workflow/parallel-feature-worktrees-edge-cases.md` (51 risks, 44 proposals — 41 implemented as tests or covered by behavior, 3 documented as gaps above)
- Open clarifications: Phase 2 Steward Agent dispatch is stubbed; the validation-specialist should confirm with the user whether shipping without auto-dispatch is acceptable or whether it blocks the gatekeeper.
