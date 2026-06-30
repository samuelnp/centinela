# coverage-hardening — senior-engineer
**Date:** 2026-06-30

Raised total statement coverage **95.0% → 97.4%** with real, branch-exercising colocated unit tests. Test-only change: no production behavior modified, gate config untouched. 55 new `*_test.go` files, each colocated in the package under test and ≤100 lines (G1).

## Files Touched

Test-only additions, by package (statement coverage before → after):

| Package | Before | After |
|---------|--------|-------|
| cmd/centinela | 91.8% | 96.8% |
| internal/roadmap | 93.7% | 97.4% |
| internal/gates | 94.5% | 97.9% |
| internal/evidence | 95.9% | 96.6% |
| internal/worktree | 92.1% | 98.7% |
| internal/setup | 95.7% | 98.3% |
| internal/ui | 95.6% | 99.6% |
| internal/migration | 92.7% | 98.8% |
| internal/analyze | 95.3% | 99.2% |

## Architecture Compliance

- All new files are `*_test.go` colocated in the same package as the code under test (per-package coverage rule — `tests/` tier files do not move the gate).
- G1: every new test file ≤100 lines (split into `_more_`/`_edge_`/`cov2_*` files as needed).
- No production source modified; no new cross-layer imports. `go vet ./...` exits 0 across the whole module (verified after the parallel rounds — no duplicate declarations).

## Type-Safety Notes

- Tests drive genuine branches with real fixtures (temp dirs, temp git repos, fake binaries on PATH, oversize scanner inputs); no reflection hacks or `interface{}` shortcuts. Errors/stdout/written files are asserted — no hollow asserts.

## Trade-Offs

- Authored slice-by-slice across independent packages in parallel for speed, then verified the whole module together (`go vet ./...` + full coverage run) to catch any cross-round collisions.
- Deferred genuinely un-unit-testable paths rather than faking them: MCP server loop (`runMcpServe`, `mcpConnectSelf`), atomic-write syscall fault-injection, external vuln-tool timeout seam, and a few unreachable `json.Marshal`/embedded-FS error returns. Recorded as roadmap items.

## Handoff

- Next role: qa-senior — add the `tests/` tier artifact + `.workflow/coverage-hardening-edge-cases.md` and re-verify the ≥97% gate (currently 97.4%).
