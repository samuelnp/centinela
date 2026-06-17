# Senior-Engineer Report: spec-traceability-gate

**Date:** 2026-06-11

## Summary
Implemented the config-gated, diff-aware spec-traceability gate that maps every
in-scope Gherkin scenario to a covering acceptance test via the
`// Acceptance: specs/<slug>.feature` + `// Scenario: <name>` convention. Modeled
exactly on the import-graph gate (config leaf + Normalize + validate; gate fn
`check<Name>(cfg, filter)`) and on G1's diff-aware filter predicate
(`filter != nil && !filter.Contains(path)`). Coverage is global; spec scope is
diff-aware. Enabled on Centinela in `severity="warn"` so CI surfaces the legacy
backlog without blocking.

## Files Touched
| File | Reason |
|------|--------|
| `internal/config/spec_traceability.go` | New `SpecTraceabilityConfig` + `NormalizeSpecTraceability` (defaults specs / tests/acceptance / fail) + `validateSpecTraceability` (rejects severity not in {fail,warn}). |
| `internal/config/config.go` | Added `SpecTraceability SpecTraceabilityConfig` field to `GatesConfig`. |
| `internal/config/defaults.go` | Wired `NormalizeSpecTraceability` into `applyDefaults`. |
| `internal/config/file_size_exceptions.go` | Wired `validateSpecTraceability` into `validateConfig`. |
| `internal/gates/spec_traceability_parse.go` | `parseScenarios(specDir, filter)` (diff-aware feature walk, malformed-tolerant) + shared `normalizeScenario`. |
| `internal/gates/spec_traceability_match.go` | `coveredScenarios(testDir)` (header+comment scan, header-annotation/`spec/`-typo tolerant, comment-without-header ignored) + `uncovered`. |
| `internal/gates/spec_traceability.go` | `checkSpecTraceability(cfg, filter)` gate entry: Skip / Pass / Warn / Fail. |
| `internal/gates/gates.go` | Registered gate in `RunWithFilter` (consistent with other gate blocks). |
| `centinela.toml` | Dogfood: `[gates.spec_traceability] enabled=true, severity="warn"`. |

## Architecture Compliance
- **G2 boundaries.** `internal/gates` imports only `internal/config` and
  `internal/gitdiff` (gate entry) plus stdlib (`bufio`, `os`, `regexp`, ...).
  `internal/config` imports only stdlib. No cmd/ui/workflow/verify imports.
  Verified: `go vet ./...` clean; import_graph gate still green.
- **G1 line counts** (all <= 100):
  - `internal/config/spec_traceability.go` — 47
  - `internal/gates/spec_traceability.go` — 58
  - `internal/gates/spec_traceability_parse.go` — 74
  - `internal/gates/spec_traceability_match.go` — 75
- **G7 (no hardcoded user-facing strings).** Gate `Result.Message`/`Details` are
  developer-facing CLI diagnostics, consistent with every existing gate
  (import_graph, G1, security) which emit plain English Results — not end-user
  i18n surfaces. No new i18n keys required (matches house convention).

## Type-Safety Notes
- No `interface{}`/`any`; all signatures concretely typed (`[]Scenario`,
  `map[string]map[string]bool`, `*gitdiff.Set`).
- Error paths wrapped/propagated: `parseScenarios`/`coveredScenarios` return
  `error`; missing dir → `(nil, nil)` (not in scope), unreadable file → skipped.
- Config validation rejects unknown severity at load with a field-naming error.

## Trade-Offs
- **Stdlib bufio/regexp, no Gherkin lib** — keeps go.mod thin and matches the
  repo's go-list-only philosophy; a Scenario Outline counts once (matches docgen).
- **`spec_traceability = true` bare key dropped from `[gates]`.** The
  `[gates.spec_traceability]` table owns that TOML key; adding both collides in
  BurntSushi/toml (`Key 'gates.spec_traceability' has already been defined`).
  Activation is via the table's `enabled = true` — identical to how
  `[gates.import_graph]` works (no `import_graph = true` boolean key exists).
  This is the only deviation from the plan's literal §5 wording and is required
  for the config to load.
- **Severity = warn on Centinela** — per the resolved plan; the 397 legacy
  uncovered scenarios surface as one non-blocking WARN, not a CI failure.

## Verification (from worktree)
- `gofmt -l cmd internal` → empty (clean).
- `go vet ./...` → no issues.
- `go build ./cmd/centinela` → success.
- `go test ./...` → 1232 passed, 0 failed.
- `/tmp/cent-stg validate` → `⚠ spec-traceability-gate` (WARN, not Fail);
  diff-aware scope = `specs/spec-traceability-gate.feature` (10 scenarios);
  validate's only failure is the coverage gate (93.3% < 95%), expected until
  qa-senior adds tests. Full-scan (`--full`, CI behavior) also reports the gate
  as a single non-blocking WARN.

## Handoff → qa-senior
- **Next role:** qa-senior.
- **Outstanding TODO (qa-senior owns ALL tests; I wrote none):**
  1. `tests/acceptance/spec_traceability_gate_test.go` — must honestly cover all
     **10** scenarios in `specs/spec-traceability-gate.feature` with the canonical
     `// Acceptance: specs/spec-traceability-gate.feature` header + one
     `// Scenario: <exact name>` comment per scenario (normalized form: trim,
     collapse ws, strip one trailing period, lowercase) so the diff-aware local
     gate reports them **covered** (closes the dogfood).
  2. Colocated **unit** tests in `internal/gates` and `internal/config`
     (per-package coverage drives the 95% gate): `parseScenarios` (Outline,
     normalization, diff include/exclude, malformed-file tolerance),
     `coveredScenarios` (header+comment, annotation/`spec/`-typo tolerance,
     comment-without-header ignored, slug derivation), `reportTraceability`
     (Pass/Warn/Fail/Skip), and config (defaults + bad-severity rejection).
  3. **Integration** test in `tests/integration`: a temp specs+acceptance tree
     run through `checkSpecTraceability` end-to-end for Pass and Fail.
- Coverage gate is currently red (93.3%) purely because this step added source
  with no tests — qa-senior's unit/integration tests must lift it back over 95%.
