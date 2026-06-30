### Big-Thinker Report: coverage-hardening
**Date:** 2026-06-30

#### Problem
Centinela's `scripts/check-coverage.sh` gate enforces a 95.0% total
statement-coverage floor, but the last three features each landed *exactly*
on ~95.0%, leaving zero headroom. Because `centinela validate` is not a
required CI check, a red coverage result on `main` does not block a merge —
so when two PRs land in parallel the second merge can tip the trunk below
95% and auto-merge red, unnoticed (a near-miss happened this week). Standing
policy is to never sit on a gate: this feature raises **total** coverage from
the measured baseline of 95.0% to **>= 97%** using real, passing tests.

#### Scope
- **In:** Real, passing colocated `*_test.go` tests (same package as the code
  under test) exercising the genuinely-testable gaps — cobra `RunE` error
  branches + command wiring in `cmd/centinela/`, and pure-logic/error-branch
  gaps in `internal/roadmap`, `internal/gates`, `internal/evidence`,
  `internal/worktree`, `internal/setup`, and the smaller packages as needed.
  Tiny behaviour-preserving testability seams only where direct execution is
  otherwise impossible.
- **Out:** Lowering/skipping/weakening the gate or its threshold; gaming
  coverage (trivial/assertion-free tests); the hard 0% network/server/
  external-tool paths (`runMcpServe`, `runVulnTool`, `mcpConnectSelf`,
  `WriteBytesAtomic` I/O error branches); refactoring production code beyond
  tiny seams; `tests/`-tier files as a coverage lever (they don't move
  per-package coverage).

#### Dependencies & Assumptions
- Builds on `scripts/check-coverage.sh` + the Go test toolchain
  (`go test ./... -coverprofile`, `go tool cover -func`).
- Measured baseline total = 95.0% (run at plan time).
- Highest total-% lever is `cmd/centinela/` (70 sub-100% fns), then
  `internal/roadmap` (23), `internal/gates` (16), `internal/worktree` (13),
  `internal/evidence` (13), `internal/setup` (10), then ~8–9-fn packages.
- Coverage is per-package with no `-coverpkg`: only colocated `*_test.go` in
  the same package moves that package's coverage.
- G1 caps every file, including `_test.go`, at 100 lines.
- No new production deps; no external services / real network in the test
  path.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Flaky/slow suite from new tests | High | Medium | `t.TempDir()`, in-memory transports, deterministic table tests; no real network/git push; keep fast + isolated. |
| Test files exceed 100-line G1 cap | Medium | High | Author lean; split into `_more_/_edge_/_error_test.go`; gatekeeper + CI full-scan catch overflow. |
| Gaming coverage (lines hit, logic not exercised) | High | Medium | Assert on real return values/output/state; reject assertion-free tests; grind real branches. |
| Rabbit-holing on hard 0% paths | Medium | Medium | Explicitly out of scope + deferred; stop at testable boundary. |
| Over-shooting past >= 97% into low-yield packages | Low | Medium | Re-measure after each slice; stop at >= 97% + small margin. |
| Testability seam changes production behaviour | High | Low | Pure extractions only; existing tests + dogfooding stay green. |

#### Rollout
- **Step 1:** `cmd/centinela/` (~70 fns) — RunE error branches + command
  wiring. Biggest lever. Re-measure.
- **Step 2:** `internal/roadmap/` (~23 fns) — pure-logic + error branches.
  Re-measure.
- **Step 3:** `internal/gates/` (~16 fns) — gate evaluation + error paths.
  Re-measure.
- **Step 4:** `internal/evidence/` (~13) + `internal/worktree/` (~13) —
  schema/marshal edges + worktree resolution branches. Re-measure.
- **Step 5:** `internal/setup/` (~10 fns) — adapter/registry branches.
  Re-measure.
- **Step 6 (only if still < 97%):** small-yield packages (`internal/ui`,
  `internal/docgen`, `internal/analyze`, `internal/migration`). Mop up.
- After every slice: `go test ./...` stays green + re-run
  `scripts/check-coverage.sh`. **Stop once total >= 97% + small margin.**

#### Deferred Findings
Recorded to the validate-exempt Backlog via `centinela roadmap defer`:
- `unit-test-mcp-server-in-memory-transport` — cover `runMcpServe` /
  `mcpConnectSelf` via an in-memory MCP transport.
- `fault-inject-atomic-write-error-paths` — cover `WriteBytesAtomic` (and
  similar low-level I/O) error branches via fault injection.
- `unit-test-vuln-tool-external-seam` — cover `runVulnTool` by stubbing the
  external vulnerability-scanner binary behind a test seam.

#### Handoff
- **Next role:** feature-specialist
- **Outstanding questions:** Do any `cmd/centinela/` `RunE` branches require
  a pure-extraction seam (vs. existing seams) to be testable without side
  effects? Resolve during the code step against the real call sites.
