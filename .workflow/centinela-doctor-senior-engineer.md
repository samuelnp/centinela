# Senior-Engineer Report — centinela-doctor

Implemented `centinela doctor` + `--fix` as a new `internal/doctor` aggregator package.

## Files
- `internal/doctor/`: doctor.go (Diagnosis types), registry.go (ordered Run/Fix), context.go (repo-root resolution + chdir + config load), git.go (shared git/worktree helpers), and one file per check: check_hooks, check_roadmap, check_roadmap_glyph (+roadmap_helpers), check_worktrees, check_workflow_state, check_evidence, check_config (+config_keys), check_version. All ≤100 lines.
- `cmd/centinela/doctor.go` (thin, --fix flag); `internal/ui/render_doctor.go` (report, house style).
- `centinela.toml`: new `aggregator` import_graph layer (internal/doctor → domain+leaf); cmd allows aggregator. `PROJECT.md` G2 prose for internal/doctor (mirrors internal/verify).

## Reuse (no reimplementation)
roadmap drift via roadmap.RenderMarkdown; evidence.Repair for *.json.tmp; worktree enumeration; setup sync for hook wiring; config.Load/validate.

## Safety / behavior
Read-only default; exit 1 only on ERROR (WARN passes). --fix: only Safe&&Idempotent repairs, attempts all even if one fails, NEVER destructive (worktree/.workflow deletion reported with exact command). Resolves canonical repo root when run from a worktree. Non-TTY plain output, deterministic ordering. Runs with NO active workflow.

## Verification (independently re-checked by orchestrator)
- All new files ≤100 lines; gofmt/vet/build clean.
- Dogfood on clean repo: 7 OK, exit 0. Isolated temp repo: glyph injected → ✗ roadmap ERROR exit 1; --fix strips glyph + regenerates ROADMAP.md → clean heading; second --fix byte-identical (idempotent).

## Deviation
workflowDone: a missing workflow file is NOT "done" (under worktree flow the state lives on-branch); abandonment requires explicit done or merged/absent branch.

Handoff → qa-senior.
