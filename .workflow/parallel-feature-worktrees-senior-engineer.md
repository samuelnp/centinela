### Senior-Engineer Report: parallel-feature-worktrees
**Date:** 2026-05-14

#### Files Touched

| Path | Reason |
|------|--------|
| internal/worktree/path.go | New: derives `.worktrees/<feature>` and detects whether cwd lives inside a worktree (DetectFeatureFromCwd). |
| internal/worktree/provision.go | New: idempotent `git worktree add` wrapper; reuses an existing branch when one already exists. |
| internal/worktree/remove.go | New: idempotent `git worktree remove` + branch cleanup helper. |
| internal/worktree/ignore_sync.go | New: orchestrates ignore-list sync for `.gitignore`, `.eslintignore`, `.prettierignore`, `.dockerignore`, `.rgignore`, plus tsconfig.json. |
| internal/worktree/ignore_append.go | New: line-appender used by every plain-text ignore file. Idempotent. |
| internal/worktree/ignore_tsconfig.go | New: JSON-tolerant patcher for tsconfig.json `"exclude"`. Skips missing files. |
| internal/worktree/start_hook.go | New: `MaybeProvision` — chdirs into the worktree when the flag is on; no-op otherwise. |
| internal/worktree/merger.go | New: hybrid merge orchestrator (dirty-check, git merge, validate, remove). Phase 2. |
| internal/worktree/merger_git.go | New: small git helpers used by the merger (dirty detection, conflicted-paths parser). |
| internal/worktree/steward.go | New: derives Merge Steward evidence paths and the steward-reason tag. |
| internal/worktree/spec_conflicts.go | New: Phase 3 spec-conflict pre-check entry point + formatter. |
| internal/worktree/spec_parser.go | New: minimal Gherkin scanner that pulls Scenario/Given/Then and detects contradictions. |
| internal/config/workflow_config.go | Added `UseWorktrees bool` (`use_worktrees` TOML key) under `[workflow]`. |
| internal/workflow/state.go | Added `WorktreePath` field to `Workflow` so `centinela status` can surface the path. |
| internal/orchestration/policy.go | Added `RoleMergeSteward` constant. Documented that the role runs out-of-band. |
| internal/orchestration/output_rules.go | Split helper functions out and added a `RoleMergeSteward` arm requiring `.workflow/<feature>-merge-steward.md`. |
| internal/orchestration/output_helpers.go | New: receives the helper functions previously in output_rules.go so each file stays ≤100 lines. |
| internal/ui/render_status.go | Surfaces `Worktree <path>` row when the workflow was provisioned with one. |
| cmd/centinela/start.go | Thin wiring: calls `worktree.MaybeProvision`, prints the worktree path, stamps `WorktreePath` on the workflow state, and resumes silently on an existing worktree. |
| cmd/centinela/hook_workflows.go | Hook-side: when cwd is inside `.worktrees/<feature>/`, filter active workflows to that worktree only. |
| cmd/centinela/init_worktree_sync.go | Thin helper: invokes `worktree.SyncIgnores` for the wizard and migrate commands. |
| cmd/centinela/init.go | Calls `syncWorktreeIgnores` after scaffolding so new projects get tooling ignore entries on day 0. |
| cmd/centinela/migrate.go | Calls `syncWorktreeIgnores` on `migrate --apply` only when `workflow.use_worktrees` is true (backward-compatible). |
| cmd/centinela/merge.go | New thin orchestrator for `centinela merge <feature>` — runs the spec pre-check, dispatches to `worktree.Merge`, surfaces Steward escalation. |
| internal/scaffold/assets/centinela.toml | Defaults `use_worktrees = true` for new projects (wizard ON by default). |
| internal/scaffold/assets/docs/architecture/evidence-contract.md | Mirrored from the live doc — scaffold parity test requires this. |
| internal/scaffold/assets/docs/architecture/merge-steward-prompt.md | Mirrored copy of the new role prompt so new projects ship with it. |
| docs/architecture/evidence-contract.md | Added the `merge-steward` entry per the orchestration evidence contract. |
| docs/architecture/merge-steward-prompt.md | New: full role prompt, output format, and escalation contract. |

#### Architecture Compliance

- N-tier boundaries respected:
  - `internal/worktree/` only imports `internal/config` (and stdlib). It does not reach into `internal/workflow`, `internal/ui`, or `cmd/`.
  - `cmd/centinela/` imports `internal/worktree`, `internal/config`, `internal/ui`, and `internal/workflow` — all permitted by PROJECT.md → G2.
  - `internal/orchestration/` keeps its self-contained policy + output rules; the merge-steward role is added as a constant and a routing case, no cross-package logic added.
- G1 file size: every new and modified file is ≤100 lines. The `output_rules.go` change required splitting the existing helpers into `output_helpers.go` to stay compliant.
- G7 outer-layer rule: `cmd/centinela/start.go`, `cmd/centinela/merge.go`, `cmd/centinela/init.go`, and `cmd/centinela/migrate.go` remain thin orchestrators. Every decision (idempotency, conflict detection, ignore-list patching, validate dispatch) lives under `internal/worktree/`.
- G2 layer rule for `internal/config`: still imports nothing internal. The new `UseWorktrees` field is a plain TOML-tagged bool.

#### Type-Safety Notes

- No `interface{}`/`any` introduced. The merger uses a typed `ValidateRunner` function alias (`func(repo string) (bool, string)`) instead of an empty interface.
- `MergeOutcome` is a struct with named, typed fields and explicit getters (`StewardHint`, `StewardJSONPath`, `StewardReason`) so callers cannot stringly-dispatch off a payload map.
- `SpecConflict` and `scenarioRecord` are plain structs with named string fields; no map-of-any state machine.
- `gitRunner` is a package-level `var func(...)` to keep tests deterministic without leaking dependency-injection plumbing into callers. Signature is fully typed (`func(repo string, args ...string) ([]byte, error)`).
- The tsconfig patcher uses `map[string]json.RawMessage` deliberately — preserves unknown keys (loader options) without needing a full tsconfig type while still being statically checked at every access site.

#### Trade-Offs

- **Spec conflict detector in `internal/worktree/` instead of a new `internal/specs/` package.** Big-thinker and feature-specialist both flagged this. Kept in `internal/worktree/` for v1 — the only consumer is the merge command, the file count is 2, and splitting now would create a one-import package. Promoting later if a second reader emerges is straightforward.
- **`gitRunner` shell-out vs. a Go git library.** Sticking to `os/exec` keeps `go.mod` untouched, matches the existing `cmd/centinela/git.go` style, and makes the wrapper trivial to stub in tests.
- **`MaybeProvision` chdirs the process** so that `workflow.Save` writes inside the worktree without changing the existing `WorkflowDir = ".workflow"` relative-path contract. The alternative — threading an absolute root through the workflow package — would have touched every caller. Documented in `start_hook.go` so future readers know the chdir is intentional.
- **`use_worktrees = true` in the wizard default but `false` in `applyDefaults`.** The flag opts in only via the scaffolded TOML (or an explicit `centinela migrate` edit) so existing projects on disk see no behavior change until they choose to. Backward compatibility is preserved at the code-level default.
- **`patchTsconfigExclude` tolerates non-JSON content** (returns "no change" instead of failing) because real-world tsconfigs often have JSON comments. The trade-off is silent no-ops for malformed configs; logged as a follow-up if users hit it.
- **Steward prompt + evidence-contract entry shipped now, but no Agent invocation yet.** Phase 2 wiring stops at "the merger surfaces an actionable hint and exits non-zero on Steward-required outcomes." Adding the actual Agent dispatch is left to qa-senior coverage and a follow-up — keeping the senior-engineer change shippable on its own.

#### Handoff

- **Next role:** qa-senior.
- **Key test targets** for qa-senior, in order of risk:
  1. `internal/worktree/path.DetectFeatureFromCwd` — unit tests for cwd inside/outside `.worktrees/<feature>/`, including nested paths and absolute-path normalisation.
  2. `internal/worktree/provision.Create` — integration test against a tmp git repo: fresh worktree, re-run is no-op, branch already exists takes the reuse path.
  3. `internal/worktree/ignore_sync.SyncIgnores` — idempotency unit test across all five ignore filenames; a separate test where tsconfig.json is absent (must be a no-op) and where it is present with existing `exclude`.
  4. `internal/worktree/merger.Merge` — three branches: clean merge → worktree removed; text conflict → outcome carries conflicted paths and `WorktreeKept`; validate fails → `ValidateFail` set and worktree untouched.
  5. `internal/worktree/spec_conflicts.DetectSpecConflicts` — two feature files asserting different Then for the same Given trigger a conflict; same Given + same Then is silent.
  6. `cmd/centinela/start.go` — acceptance scenarios "Start provisions a worktree" and "Restarting a feature with an existing worktree resumes in place" (the latter exercises the idempotent branch).
  7. `cmd/centinela/merge.go` — dirty-tree pre-check fails before `git merge` runs; spec-conflict pre-check blocks the merge and names both files.
  8. Hook integration: `hook_workflows.loadActiveWorkflows` filters by worktree feature when cwd is inside `.worktrees/<feature>/`. Critical for the "two parallel workflows do not collide" scenario.
- **Outstanding TODOs** (not blocking qa-senior):
  - Wire the actual Merge Steward Agent invocation behind the "Merge Steward required" exit in `cmd/centinela/merge.go`. Today the command exits non-zero with the evidence-path hint; the Agent dispatch is a Phase 2.1 task.
  - Confirmed: `merge-steward` lives outside the 5-step workflow and is NOT part of `RequiredRoles(step)` — validator only fires on the role when evidence is written.

---

### Patch addendum (2026-05-14): edge-case fixes

Two surgical patches in response to the edge-case tester report at `.workflow/parallel-feature-worktrees-edge-cases.md` (Top Gaps: feature-slug validation, symlink normalization).

| Path | Reason |
|------|--------|
| internal/worktree/slug.go | New: `ValidateFeatureSlug(name string) error`. Enforces kebab-case ASCII `^[a-z0-9]+(-[a-z0-9]+)*$` before any feature name reaches `git worktree add` or the filesystem. Rejects shell-unsafe and path-traversal inputs (`alpha/../beta`, `alpha;rm`, names with `/`, spaces, uppercase). Placed in its own file so `path.go` stays focused; the project had no prior slug validator to reuse (confirmed by grep across `internal/workflow/` and `internal/roadmap/`). |
| internal/worktree/provision.go | `Create` now calls `ValidateFeatureSlug(feature)` first thing. Replaces the previous bare empty-string guard, since the slug regex already rejects empty input. Belt-and-braces: validation also fires from the cmd layer, but the package-level guard ensures any future caller is safe by default. |
| internal/worktree/path.go | `DetectFeatureFromCwd` now resolves symlinks via `filepath.EvalSymlinks(abs)` before scanning for the `.worktrees/<feature>` segment. Fixes macOS `/tmp` -> `/private/tmp` and any other symlinked parent path; without normalization the un-resolved cwd could fail to match the on-disk worktree root. EvalSymlinks failure is treated as "not inside a worktree" (falls through to the unresolved path), matching the existing no-feature-no-error contract. |
| cmd/centinela/start.go | Defensive call to `worktree.ValidateFeatureSlug(feature)` at the top of `runStart` — feature slug is user-supplied CLI input. Fails fast before any filesystem or workflow state touches the disk. |
| cmd/centinela/merge.go | Same defensive call at the top of `runMerge`. Prevents a stray `centinela merge "alpha;rm -rf"` from reaching `worktree.DetectSpecConflicts` or `worktree.Merge`. |

#### Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./...` — all packages pass (cached + cmd/centinela re-run at 1.027s)
- File sizes: `slug.go` 26L, `path.go` 51L, `provision.go` 51L, `start.go` 81L, `merge.go` 48L — all well under the G1 100-line ceiling.

#### Scope decisions

- Single regex slug (`^[a-z0-9]+(-[a-z0-9]+)*$`) — kebab-case only, no underscores, no leading/trailing hyphen. Matches the wording in `internal/ui/render_roadmap.go` ("Feature names must be valid centinela slugs (lowercase, hyphens)"). Tightening rather than broadening avoids future edge cases.
- `worktree.Create` short-circuits on the slug error instead of falling through to `Exists()`. A bad slug must never produce a path or touch git, even by accident.
- Validation in `cmd/centinela/start.go` runs **before** the `PROJECT.md` existence check so the error message is about the slug, not a missing project file, when both are invalid.
- No tests added — qa-senior owns coverage (test gaps for `TestCreate_InvalidFeatureSlug` and `TestDetectFeatureFromCwd_WithSymlinks` are already in the edge-case report).
