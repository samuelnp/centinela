# Big-Thinker Report — centinela-doctor

## Decision
`centinela doctor`: one read-only health command across 7 checks, with `--fix` for safe idempotent repairs only. Motivated by real incidents THIS session: stale installed binary (0.15.0) silently blocked starts; abandoned worktrees needed hand-cleanup; the ✅-glyph-in-phase-name bug broke all greenfield starts; verify_timeout too low caused spurious claim-verifier failures.

## Architecture
- New `internal/doctor/` domain: `Check`/`Diagnosis{name,status,message,details,repair?}` + ordered registry; `Run` (diagnose) and `Fix` (apply safe repairs then re-diagnose). One check+repair per ≤100-line file with colocated test.
- `cmd/centinela/doctor.go` thin orchestrator (`--fix` flag); `internal/ui/render_doctor.go` report (house style from render_gates.go).
- Reuse, don't reimplement: roadmap drift (gates/roadmap), evidence.Repair (tmp sweep), worktree enumeration, setup hook-wiring, config load/validate. Repo resolution via worktree.DetectFeatureFromCwd (works from worktree or root).

## v1 checks (all 7 in)
1 hook-wiring (safe repair) · 2 roadmap drift + phase-name glyph (safe repair) · 3 abandoned worktrees (report) · 4 stale .workflow (report) · 5 orphaned *.json.tmp (safe repair) · 6 config drift (report) · 7 binary version skew (report).

## Safety / exit codes
Default read-only: exit 0 unless any ERROR (WARN passes), else 1. `--fix`: only Safe&&Idempotent repairs, no-op on rerun, attempts all even if one fails (failure→ERROR), NEVER destructive (worktree/.workflow deletion reported with exact command).

## Layer
internal/doctor = new "aggregator" import_graph layer (allows domain+leaf); PROJECT.md G2 prose mirrors the internal/verify allowance. Coordinates with the deferred roadmap-import-graph-layer-mapping Backlog item.

Handoff → feature-specialist.
