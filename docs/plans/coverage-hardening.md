# Plan: coverage-hardening

## Problem

`scripts/check-coverage.sh` enforces a 95.0% total-coverage floor. Three
consecutive features landed at ~95.0%, leaving zero headroom. Validate is
not a required CI check, so a parallel merge that tips `main` below 95%
auto-merges red and unnoticed. Standing policy: never hug a gate — exceed
it by ~2%. This feature raises **total** coverage from the measured
baseline of **95.0%** to **>= 97%** using real, passing tests.

## Scope

### In

- Real, passing unit tests that exercise real logic for the
  **genuinely-testable** gaps: cobra `RunE` error branches and command
  wiring in `cmd/centinela/`, plus pure-logic and error-branch gaps across
  `internal/roadmap`, `internal/gates`, `internal/evidence`,
  `internal/worktree`, `internal/setup`, and the smaller-yield packages
  (`internal/ui`, `internal/docgen`, `internal/analyze`,
  `internal/migration`) as needed to clear >= 97%.
- Colocated `*_test.go` files in the **same package** as the code under
  test (the only thing that moves per-package coverage — see Constraint 1).
- Tiny, behaviour-preserving testability seams **only** where direct
  execution is otherwise impossible (e.g. extracting a pure `argsFor()` from
  an exec wrapper). No production behaviour change.

### Out

- **Lowering, skipping, or otherwise weakening the coverage gate** or its
  threshold. The 95.0% floor stays; we raise actual coverage above it.
- **Gaming** coverage — trivial asserts that execute lines without
  exercising logic, dead test scaffolding, or assertion-free calls.
- **Hard 0% network / server / external-tool paths**: `runMcpServe`,
  `runVulnTool`, `mcpConnectSelf`, and `WriteBytesAtomic` low-level I/O
  error paths. These need real servers, external binaries, or fault
  injection and are explicitly deferred (see Deferred Findings).
- Refactoring production code beyond the tiny testability seams above.
- `tests/`-tier (unit/integration/acceptance) files as a coverage lever —
  they do not move per-package coverage (Constraint 1). Acceptance/spec
  artifacts are still authored in later steps per workflow rules.

## Non-Negotiable Constraints

1. **Coverage is per-package, no `-coverpkg`.** Only colocated `*_test.go`
   files in the **same package** as the target code move that package's
   coverage. Files under `tests/` do **not**. The bulk of the work is
   therefore colocated `*_test.go` in `cmd/centinela` and `internal/*`.
2. **G1 caps every source file — including `_test.go` — at 100 lines**
   (130 only with a justified, configured exception). Split test files into
   `_more_test.go`, `_edge_test.go`, `_error_test.go`, etc.
3. **Tests must be real and pass.** No gaming. Grind, don't game — each test
   must assert on real behaviour/output, not merely invoke a line.
4. **Slice by package group; re-measure after each slice.** A single
   test-writing pass must stay bounded (a long test-writing subagent can die
   mid-write). Sequence highest-yield first and stop once total >= 97% with
   a small safety margin, so we don't over-invest in low-yield packages.

## Dependencies & Assumptions

- Builds on the existing `scripts/check-coverage.sh` gate and the Go test
  toolchain (`go test ./... -coverprofile`, `go tool cover -func`).
- Assumes the measured baseline is total = 95.0% (run at plan time).
- Assumes the highest total-% lever is `cmd/centinela/` (70 sub-100%
  functions — mostly `RunE` error branches and command wiring), followed by
  `internal/roadmap` (23), `internal/gates` (16), `internal/worktree` (13),
  `internal/evidence` (13), `internal/setup` (10), then the ~8–9-function
  packages.
- Assumes cobra commands expose test seams (or accept tiny pure-extraction
  seams) so `RunE` error branches can be driven without real side effects.
- No new production dependencies; no external services in the test path.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Flaky / slow suite from new tests (real filesystem, exec, timing) | High | Medium | Use `t.TempDir()`, in-memory transports, table-driven deterministic cases; no real network/`git push`; keep tests fast and isolated; gate on `go test ./...` staying green and bounded. |
| Test files exceed the 100-line G1 cap | Medium | High | Author each test file lean; split into `_more_test.go` / `_edge_test.go` / `_error_test.go` proactively; gatekeeper + CI full-scan catch any overflow. |
| Gaming coverage (lines hit, logic unexercised) | High | Medium | Every test asserts on real return values/output/state; reviewer + gatekeeper reject assertion-free or trivial tests; grind real branches, not coverage theatre. |
| Touching a hard 0% path and rabbit-holing | Medium | Medium | Explicitly out of scope and deferred to roadmap; stop at the testable boundary. |
| Over-shooting effort past >= 97% into low-yield packages | Low | Medium | Re-measure total after each slice; stop once >= 97% + small margin. |
| A testability seam changes production behaviour | High | Low | Seams limited to pure extractions (e.g. `argsFor()`); existing tests + dogfooding must stay green; no behaviour change allowed. |
| New colocated tests trip layer/import rules | Medium | Low | Tests live in the same package as code under test; use external `_test` package where it avoids import cycles; `import_graph` gate validates. |

## Rollout (sliced, highest-yield first; re-measure after each slice)

Each slice is a bounded test-writing pass that ends with a coverage
re-measure. **Stop as soon as total >= 97%** with a small safety margin —
later slices become optional.

- **Slice 1 — `cmd/centinela/` (the elephant, ~70 fns).** Colocated
  `*_test.go` covering cobra `RunE` error branches and command wiring.
  Biggest total-% lever; expected to do most of the lift. Re-measure.
- **Slice 2 — `internal/roadmap/` (~23 fns).** Pure-logic + error-branch
  gaps. Re-measure.
- **Slice 3 — `internal/gates/` (~16 fns).** Gate evaluation branches and
  error paths. Re-measure.
- **Slice 4 — `internal/evidence/` (~13 fns) and `internal/worktree/`
  (~13 fns).** Schema/marshal edge cases; worktree resolution branches.
  Re-measure.
- **Slice 5 — `internal/setup/` (~10 fns).** Adapter/registry branches.
  Re-measure.
- **Slice 6 (only if still < 97%) — small-yield packages**
  (`internal/ui`, `internal/docgen`, `internal/analyze`,
  `internal/migration`, ~8–9 fns each). Mop up to clear the margin.

After every slice: run `go test ./...` (must stay green) and
`scripts/check-coverage.sh` to re-derive total; record the new total before
starting the next slice.

## Deferred Findings

New out-of-scope discoveries recorded to the validate-exempt Backlog via
`centinela roadmap defer`:

- `unit-test-mcp-server-in-memory-transport` — cover `runMcpServe` /
  `mcpConnectSelf` via an in-memory MCP transport.
- `fault-inject-atomic-write-error-paths` — cover `WriteBytesAtomic` (and
  similar low-level I/O) error branches via fault injection.
- `unit-test-vuln-tool-external-seam` — cover `runVulnTool` by stubbing the
  external vulnerability-scanner binary behind a test seam.

(Exact recorded slugs listed in the big-thinker report.)

## Handoff

- **Next role:** feature-specialist — author the Gherkin `.feature` spec and
  acceptance criteria: total coverage >= 97% via real passing tests, gate
  threshold unchanged, hard 0% paths deferred, all new test files <= 100
  lines.
- **Outstanding questions:** Do any `cmd/centinela/` `RunE` branches require
  a pure-extraction seam (vs. existing seams) to be testable without side
  effects? Resolve during the code step per actual call sites.
