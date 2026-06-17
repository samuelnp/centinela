# Plan: spec-traceability-gate

Build a diff-aware, config-gated built-in gate that verifies every Gherkin
scenario maps to an acceptance test in the executed suite. Model it on the
existing import-graph and build gates. Total ~200–250 lines across small files
(G1 ≤100 lines each).

## Architecture & layer compliance (G2)

Gate lives in `internal/gates/` (domain). Config lives in `internal/config/`
(leaf). `internal/gates` may import `internal/config` and `internal/gitdiff`
only — both used here. No imports of cmd/, ui, workflow, verify.

## 1. Config (`internal/config/spec_traceability.go`)

```toml
[gates]
spec_traceability = false   # default OFF, like import_graph/build

[gates.spec_traceability]
enabled   = true
spec_dir  = "specs"               # default
test_dir  = "tests/acceptance"    # default
severity  = "fail"                # "fail" | "warn" — strictness knob for adoption
```

- `SpecTraceabilityConfig{ Enabled bool; SpecDir, TestDir, Severity string }`.
- Add `SpecTraceability SpecTraceabilityConfig` to `GatesConfig` (config.go).
- `NormalizeSpecTraceability`: default spec_dir=`specs`, test_dir=`tests/acceptance`,
  severity=`fail`; called from `applyDefaults` (defaults.go).
- `validateSpecTraceability`: severity must be `fail`|`warn`; called from
  `validateConfig` (file_size_exceptions.go). Reject unknown severity.

## 2. Spec parsing (`internal/gates/spec_traceability_parse.go`)

- `parseScenarios(specDir, filter) → []Scenario{ Spec string; Name string }`.
- Walk `specDir/*.feature`; if `filter` is non-nil, include only files in the
  diff set (diff-aware — same contract G1 uses).
- Line regex: `^\s+Scenario(?: Outline)?:\s*(.+?)\s*$`. Normalize the captured
  name (trim, collapse internal whitespace, strip a single trailing period) via
  a shared `normalizeScenario` helper.
- No Gherkin library — stdlib `bufio`/`regexp` only (keeps go.mod thin, matches
  the repo's go-list-only philosophy for the g2 gate).

## 3. Test scanning + match (`internal/gates/spec_traceability_match.go`)

- `coveredScenarios(testDir) → map[specSlug]map[normalizedName]bool`.
- Walk `testDir/*.go`; for each file capture header `// Acceptance: specs/<slug>.feature`
  and every `// Scenario: <name>` comment; key coverage by (slug, normalized name).
- A `.feature` path → slug = base filename without extension; a test header's
  `specs/<slug>.feature` → same slug. Match scenario by (slug, normalized name).
- `uncovered(scenarios, covered) → []Scenario` for reporting.
- **Defensive parsing (measured necessity).** The convention is followed loosely
  in the existing tree: only ~4 of ~32 acceptance files match exactly. The gate
  must therefore: accept a header with trailing annotations
  (`// Acceptance: specs/foo.feature (AC4, AC5)`), tolerate the known `spec/`
  (singular) typo, and normalize scenario names case-insensitively in addition
  to trim/collapse/trailing-period. The gate DEFINES the canonical form; its
  Fail/Warn details restate it so authors can fix mismatches deterministically.

## 4. Gate entry (`internal/gates/spec_traceability.go`)

`checkSpecTraceability(cfg, filter) Result`:
- Skip if no `.feature` files in scope (diff-aware: e.g. CI full-scan vs local
  changed-only). Message mirrors G1/G11 diff-aware wording.
- Parse in-scope scenarios; scan all acceptance tests (coverage is global —
  a scenario may be covered by any acceptance file pointing at its spec).
- If all in-scope scenarios covered → Pass with count.
- Else → `Fail` (or `Warn` if `severity="warn"`); `Details` lists each
  `specs/<slug>.feature: "<scenario>"` uncovered, so the message is actionable.
- Register in `gates.go` `RunWithFilter`: `if cfg.Gates.SpecTraceability.Enabled
  { results = append(results, checkSpecTraceability(cfg, filter)) }`.

## 5. Dogfood wiring (`centinela.toml`) — resolved by big-thinker

Measured reality: **406 scenarios, only 9 covered (397 uncovered)** under exact
`(slug, normalized-name)` matching, and CI is confirmed to force full-scan
(`.github/workflows/validate.yml` runs plain `centinela validate`; GitHub sets
`CI=true` → `ModeFull` → `filter=nil`). A `severity="fail"` gate would therefore
fail Centinela's CI on 397 legacy scenarios on merge. Decision:

```toml
[gates]
spec_traceability = true

[gates.spec_traceability]
enabled = true
severity = "warn"     # CI surfaces the 397-scenario backlog without blocking
```

- **Locally** (diff-aware), only this branch's `specs/spec-traceability-gate.feature`
  is in scope; its 9 scenarios are the bounded dogfood.
- **In CI** (full-scan), the gate reports every uncovered scenario as a WARNING —
  visible, actionable, ratchet-ready — but does not block merge.
- This neither games the gate (no stub tests) nor disables it (the full-scan
  path is real and tested). The dogfood's integrity comes from qa-senior
  genuinely covering the 9 new scenarios so the gate reports them as covered;
  the `severity` ratchets to `fail` in a later step once a backfill /
  `audit-baseline-ratchet` clears the legacy backlog.

> **Why not `fail` + always-diff-aware in CI?** That would diverge from how every
> other gate honors the documented `CI`→full-scan invariant and would mask real
> CI regressions in unchanged specs. Warn keeps the invariant and stays honest.

## Test plan

- Unit (colocated in `internal/gates`, `internal/config` — per-package coverage):
  - `parseScenarios`: Scenario + Scenario Outline, normalization (trailing
    period, extra spaces, casing), diff-filter inclusion/exclusion, malformed
    file tolerated.
  - `coveredScenarios`: header+comment extraction, multiple tests per spec,
    comment without header ignored, slug derivation.
  - matching: covered → Pass; one uncovered → Fail with the name in Details;
    `severity="warn"` → Warn not Fail; no specs in scope → Skip.
  - config: defaults applied; bad severity rejected.
- Integration (`tests/integration`): a temp repo tree (specs + acceptance) run
  through `checkSpecTraceability` end to end for Pass and Fail.
- Acceptance (`tests/acceptance/spec_traceability_gate_test.go`): assert the
  gate is registered and that `centinela.toml` enables it; this file ALSO
  carries the `// Acceptance:` + `// Scenario:` comments that make the feature's
  own spec pass the gate (dogfood closure).

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| CI full-scan fails on 397 legacy uncovered scenarios | High | **Resolved:** ship default-off globally; enable on Centinela with `severity="warn"` so CI reports but never blocks. Ratchet to `fail` after a backfill/`audit-baseline-ratchet`. (See §5.) |
| Comment convention drift (typos, trailing punctuation) causes false "uncovered" | Medium | Normalize aggressively; Details name the exact mismatch so the fix is obvious; document the convention in the gate message. |
| Scenario Outline counted once but has N examples | Low | v1 treats the outline as one scenario (matches docgen's counting); note in edge-cases. |
| New gate pushes a file >100 lines (G1) | Medium | Pre-split into parse/match/entry/config as above. |

## Rollout

1. Config struct + normalize + validate (no behavior yet).
2. Parser + matcher with colocated unit tests.
3. Gate entry + registration; integration test.
4. Enable in `centinela.toml`; write the acceptance test that also closes the
   dogfood; confirm `centinela validate` passes diff-aware.
