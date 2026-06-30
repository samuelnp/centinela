### Feature-Specialist Report: coverage-hardening
**Date:** 2026-06-30

#### Behavior Summary

The coverage-hardening feature raises Centinela's total Go statement coverage
from its current baseline of 95.0% to >= 97.0% by writing real, passing,
colocated `*_test.go` unit tests that exercise genuinely-testable logic gaps —
cobra `RunE` error branches in `cmd/centinela/`, pure-logic and error-branch
paths across `internal/roadmap`, `internal/gates`, `internal/evidence`,
`internal/worktree`, `internal/setup`, and smaller-yield packages as needed to
clear the target. The 95.0% gate threshold in `scripts/check-coverage.sh`
remains unchanged; the feature exceeds rather than replaces the floor. Tests
are delivered in slices, highest-yield package first, with a coverage
re-measure after each slice so effort stops as soon as total >= 97% plus a
small safety margin. Network/server/external-tool paths that would require real
infrastructure or fault injection to cover are explicitly deferred to the
roadmap rather than covered with hollow or assertion-free tests.

#### Gherkin Scenarios

Spec file: `specs/coverage-hardening.feature`

1. **Total coverage meets the hardened target**
   - Given: full Go test suite including all new colocated `*_test.go` files
   - When: `scripts/check-coverage.sh` runs with the default threshold
   - Then: reported total statement coverage >= 97.0%; script exits 0
   - Verified by: running `scripts/check-coverage.sh` after all slices are
     merged; confirming exit 0 and reading the printed total percentage.

2. **Coverage gate still passes at the configured floor**
   - Given: gate threshold remains 95.0% in `scripts/check-coverage.sh`
   - When: script runs after coverage-hardening tests are merged
   - Then: gate passes because actual >= 97.0% exceeds the floor; threshold
     value in the script is still exactly 95.0
   - Verified by: `grep 95` on `scripts/check-coverage.sh` + a passing run.

3. **New tests are colocated and within size limits**
   - Given: the set of `*_test.go` files added by this feature
   - When: each file is inspected for package declaration and line count
   - Then: every new test file declares the same package as the file under
     test; no new `*_test.go` exceeds 100 lines; no new file (any kind)
     exceeds 130 lines (G1 cap with justified exception ceiling)
   - Verified by: `wc -l` over new test files; `head -1` package check; the
     gatekeeper full-scan in the validate step.

4. **No production behaviour changed**
   - Given: full test suite including all new tests
   - When: `go test ./...` completes
   - Then: exit 0, no existing acceptance or behavioural spec regresses; any
     added testability seams are pure extractions with no observable side
     effects
   - Verified by: green `go test ./...` run + `centinela validate` in the
     validate step; no changes to production function signatures beyond
     optional pure-extraction seams.

5. **Hard-to-unit-test paths are explicitly deferred, not faked**
   - Given: `runMcpServe`, `mcpConnectSelf`, `runVulnTool`, and
     `WriteBytesAtomic` I/O error branches remain at 0% coverage
   - When: the feature is complete
   - Then: none are covered by hollow or assertion-free tests; all three
     deferred slugs appear in `centinela roadmap list --status deferred`
   - Verified by: reviewing roadmap deferred list for the three recorded slugs;
     per-function coverage report confirms those functions remain at 0% with no
     assertion-free wrappers in the test suite.

#### UX States

n/a — this feature has no UI surface; it is a quality/meta improvement
delivered entirely via test files and an updated coverage metric.

#### Out-of-Scope

- Lowering, skipping, or otherwise weakening the 95.0% coverage gate threshold.
- Gaming coverage with trivial or assertion-free tests that execute lines
  without exercising logic.
- Covering `runMcpServe`, `mcpConnectSelf`, `runVulnTool`, or
  `WriteBytesAtomic` I/O error branches in this feature (require real
  infrastructure or fault injection — deferred to roadmap).
- Refactoring production code beyond tiny, behaviour-preserving testability
  seams (e.g. pure `argsFor()` extraction).
- Adding coverage via `tests/`-tier files (unit/integration/acceptance under
  `tests/`) — those don't move per-package coverage due to the no-`-coverpkg`
  constraint.
- Raising coverage above the 97% target into low-yield territory once the
  margin is already satisfied; effort stops at >= 97% + small safety margin.

#### Edge Cases

- **Colocated-only constraint**: coverage must come from `*_test.go` files in
  the same package as the code under test. Tests in `tests/unit/`,
  `tests/integration/`, or `tests/acceptance/` do not move per-package
  coverage and must not be relied on as a coverage lever.
- **<=100-line cap on every test file**: each added `_test.go` must stay <=100
  lines; split into `_more_test.go`, `_edge_test.go`, `_error_test.go` etc. if
  needed. The gatekeeper full-scan in validate catches any overflow that
  diff-aware mode misses.
- **Re-measure after each slice**: a single large test-writing subagent can die
  mid-write (socket error, context expiry); slicing by package group with a
  re-measure after each provides bounded, resumable progress and a clear stop
  criterion.
- **Defer hard paths, don't fake them**: the three 0%-coverage network/external
  functions must remain at 0% coverage in this feature; covering them with
  hollow wrappers would satisfy the metric while violating the "grind, don't
  game" policy.
- **Gate threshold must remain 95.0%**: the feature raises actual coverage, not
  the floor. Any accidental edit to `scripts/check-coverage.sh` must be caught
  before the validate gate.
- **No real network or git push in tests**: tests must use `t.TempDir()`,
  in-memory transports, and local bare repos as origins; real network calls or
  git pushes cause the suite to hang (see acceptance-test-network-push-hangs
  memory entry).

#### Deferred Findings

The following deferred roadmap items were recorded by the big-thinker and
remain deferred through this step (no new discoveries):

- `unit-test-mcp-server-in-memory-transport` — cover `runMcpServe` /
  `mcpConnectSelf` via an in-memory MCP transport.
- `fault-inject-atomic-write-error-paths` — cover `WriteBytesAtomic` (and
  similar low-level I/O) error branches via fault injection.
- `unit-test-vuln-tool-external-seam` — cover `runVulnTool` by stubbing the
  external vulnerability-scanner binary behind a test seam.

No new deferred findings discovered at the feature-specialist stage.

#### Handoff

- **Next role:** senior-engineer
- **Open clarifications:** Do any `cmd/centinela/` `RunE` branches require a
  new pure-extraction seam (vs. existing seams) to be testable without side
  effects? Resolve during the code step against the real call sites.
