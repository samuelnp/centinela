# brownfield-roadmap-generation — qa-senior

## Test Inventory

| Tier | File | Covers (scenarios / seams) |
|------|------|----------------------------|
| colocated unit | `internal/brownmap/generate_test.go` | Generate: Baseline+gaps from TODOs+goal; empty/doc-only → empty Baseline, 0 gaps; deterministic |
| colocated unit | `internal/brownmap/baseline_test.go` | baselinePhase one-feature-per-target in order; empty→non-nil; roleOrModule/baselineDescription |
| colocated unit | `internal/brownmap/gaps_test.go` | gapPhases: TODO→`-confirm` + goal features; no-work→nil; goal-only |
| colocated unit | `internal/brownmap/write_test.go` | WriteDraft writes draft path; refuses canonical; deterministic; mkdir failure |
| colocated unit | `internal/brownmap/write_more_test.go` | atomicWrite `dir==""` branch; rename-over-dir failure + temp cleanup |
| colocated unit | `internal/roadmap/baseline_test.go` | isBaselinePhaseName / IsBaselinePhaseName / isNonSchedulablePhase; Summary / NonBacklogFeatureSet / DeriveReadiness exclude Baseline; Backlog regression guard |
| colocated unit | `internal/reconstruct/todotargets_test.go` | TodoTargets returns TODO-bearing in order; zero-able (nil) |
| colocated unit | `internal/ui/render_brownfield_test.go` | RenderBrownfieldSummary counts+path; no-gaps hint branch |
| colocated cmd | `cmd/centinela/roadmap_brownfield_test.go` | happy: draft written, canonical untouched; `--json` |
| colocated cmd | `cmd/centinela/roadmap_brownfield_errors_test.go` | missing inventory→error+no draft; malformed; refuse canonical `--out` |
| tests/unit | `tests/unit/brownfield_baseline_unit_test.go` | Scenario 5 (Baseline exempt from status/coverage) |
| tests/integration | `tests/integration/brownfield_pipeline_test.go` | real analyze→Save→Load→Generate→WriteDraft; existing roadmap.json byte-unchanged |
| tests/acceptance | `tests/acceptance/brownfield_helper_test.go` | binary runner + fixtures (shared real binary) |
| tests/acceptance | `tests/acceptance/brownfield_happy_test.go` | Scenarios **1, 2, 3, 4, 10** |
| tests/acceptance | `tests/acceptance/brownfield_edge_test.go` | Scenarios **5, 6, 7, 8, 9** |

**Scenario traceability:** all 10 spec scenarios carry an exact-match `// Scenario:` comment across the two acceptance files (5 in happy, 5 in edge); each acceptance file carries `// Acceptance: specs/brownfield-roadmap-generation.feature`.

## Coverage Gaps

Coverage gate **passed: 95.1% >= 95.0%**. New `internal/brownmap` statements are exercised by colocated unit tests (acceptance shells out to the built binary and contributes no `internal/brownmap` coverage). Two residual defensive branches remain unit-uncovered and are documented in the edge-cases file:

- `atomicWrite` mid-write fault returns (`tmp.Write` / `tmp.Close` errors) — require OS fault injection.
- `json.MarshalIndent` error inside `WriteDraft` — a Roadmap of plain structs cannot fail to marshal.

The realistic crash-safety edges (mkdir failure, rename-over-directory + temp cleanup) **are** covered. No coverage gap was deferred to the roadmap.

## Acceptance Wiring

`centinela.toml` `[validate].commands` already includes the acceptance run and the coverage gate — wiring exists, only the test files were added:

```toml
[validate]
commands = [
  "go test ./...",
  "go test ./tests/acceptance/...",
  "./scripts/check-coverage.sh",
  "./scripts/check-fmt.sh"
]
```

`centinela.toml` was **not** edited.

## Verification (run in the worktree)

- `go build ./...` → `Success`.
- `go test ./...` → `2536 passed in 37 packages` (all pass).
- `go test ./tests/acceptance/...` → `598 passed in 1 packages`.
- `gofmt -l` over all new test files → empty (clean).
- Line-count check on all 15 new `_test.go` files → max 83 lines; nothing > 100.
- `./scripts/check-coverage.sh` → `coverage gate passed: 95.1% >= 95.0%`.
- `centinela evidence validate brownfield-roadmap-generation` → `evidence ok`.

## Deferred Findings

None.

## Handoff

Next role: **validation-specialist**. Suite is green, coverage gate is green at 95.1%, all 10 acceptance scenarios are traced, the edge-cases artifact is filled, and evidence validates. The validation step should run `centinela validate` (lint + type check + full suite + acceptance + coverage + fmt + spec-traceability) and author the gatekeeper report. The only intentionally-uncovered code is two defensive fault branches in `internal/brownmap/write.go` (documented; gate still ≥ 95%).
