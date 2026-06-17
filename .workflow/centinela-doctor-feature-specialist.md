# Feature-Specialist Report — centinela-doctor

## Spec
`specs/centinela-doctor.feature` — 38 scenarios, 1:1 acceptance-traceable, grouped: happy path, per-check diagnose+repair (7 checks), exit-code/output contract, --fix behavioral contract, multi-problem, robustness/env edges.

## Key guarantees asserted
- Per-check: hook-wire (repair+idempotent), roadmap drift + glyph (flag+strip+idempotent), abandoned worktree (report, NOT removed by --fix), stale .workflow (report, NOT deleted), orphaned *.json.tmp (remove+idempotent), config drift (WARN, no mutation), version skew (WARN, no reinstall).
- Exit 0 on OK/WARN, 1 on any ERROR; deterministic ordering; non-TTY plain output.
- --fix attempts all safe repairs even if one fails; never destructive; multi-problem single pass.
- Robustness: no .claude / not a git repo / no worktrees / binary off PATH → graceful WARN, no panic; doctor needs NO active workflow.

## Edge cases added to brief
Multiple simultaneous problems; binary-not-on-PATH; doctor requires no active workflow; check dependency missing at runtime (unparseable toml → that check ERROR, others still run); multiple fixable problems in one --fix; --fix partial-success summary.

Handoff → senior-engineer.
