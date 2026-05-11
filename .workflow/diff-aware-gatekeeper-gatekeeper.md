### Gatekeeper Report: diff-aware-gatekeeper
**Date:** 2026-05-11
**Status:** SAFE

#### Analyzed Specs

- `specs/diff-aware-gatekeeper.feature` (new)
- All other `specs/*.feature` reviewed for domain / config / gate /
  CLI conflicts.

#### Findings

No conflicts detected. The feature is additive at the domain boundary
and backward-compatible at the user-facing boundary.

**Layer dependencies (G2 / n-tier).** All new imports respect the
stack:

- `cmd/centinela/` imports `internal/config`, `internal/gates`,
  `internal/gitdiff`, `internal/ui`. Outer layer only.
- `internal/gates/` adds `internal/gitdiff` and `internal/config`
  imports. Domain → supporting domain → leaf config. OK.
- `internal/gitdiff/` imports stdlib only (`os/exec`, `strings`,
  `errors`, `fmt`, `path/filepath`). Leaf-clean.
- `internal/config/` unchanged at the import level — `validate_mode.go`
  imports `strings` only. No new internal imports.

**Outer-layer purity (G7).** `cmd/centinela/validate.go` is a thin
orchestrator: parses flags, loads config, asks
`cfg.Validate.ResolveMode` for the mode, asks
`resolveDiffFilter` for the filter, hands both to
`gates.RunWithFilter`, renders results. No business decisions live
in `cmd/`. Mode-resolution truth table lives in
`internal/config/validate_mode.go`. Filter rendering helper lives in
`cmd/centinela/validate_mode.go` but is a presentation concern (the
human-readable header string) — no domain logic.

**File size (G1).** All new and edited files ≤ 100 lines (max 98 at
`internal/gitdiff/resolver.go`). `cmd/centinela/validate.go` was at
111 lines after the first refactor; helper extracted into
`validate_runner.go` brings it to 96.

**Spec first (G5) / Plan first (G6).** Both artifacts exist and
predate the code step in this branch's history:
`docs/plans/diff-aware-gatekeeper.md`,
`specs/diff-aware-gatekeeper.feature`,
`docs/features/diff-aware-gatekeeper.md`.

**Single responsibility (G8).** `Set` (membership), `Resolver`
(git shell-out), `Mode`/`ResolveMode` (policy), `i18n_filter`
(short-circuit), `validate_mode.go` (CI detection + filter resolve),
`validate_runner.go` (shell-out helper) each export exactly one
concern.

**Type safety.** No `interface{}` introduced (existing
`map[string]interface{}` in `i18n_keys.go` predates this feature).
All exported types in `internal/gitdiff/` are concrete structs;
`Run` field is a typed function value to allow test injection.

**DTO / state-machine impact.** None.
- `.workflow/<feature>.json` schema unchanged.
- Orchestration evidence JSON shapes unchanged (this feature
  produces big-thinker, feature-specialist, senior-engineer,
  qa-senior, plus pending validation-specialist + documentation-
  specialist artifacts in the existing format).
- `centinela.toml` adds two optional keys under `[validate]`;
  default behavior unchanged for unset values.

**Backward compatibility.**
- `gates.RunAll(cfg)` preserved as a wrapper over `RunWithFilter`
  with nil filter. Any external caller stays green.
- `executeValidation()` keeps its no-arg signature for `complete.go`
  and the pre-existing test suite.
- Configs without `diff_mode` / `diff_base` normalize to
  `auto` / `main` — behavior matches today's full scan in CI; only
  the local default flips to diff-aware. Documented as a user-visible
  change in the feature brief; not a contract break.

**CI risk.** `auto` mode reads `CI=true` / `CI=1` (universal
convention covering GitHub Actions, GitLab CI, CircleCI, Travis,
Buildkite, Drone). Other CI systems that don't set `CI` would need
either `diff_mode = "off"` or an explicit `--full` flag in their
pipeline. Not a blocker — projects can opt out — but worth flagging
in the docs step.

#### Recommendation

SAFE: No conflicts detected. Proceed with `centinela validate`.
