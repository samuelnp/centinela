# QA-Senior Report: headless-governance

**Role:** qa-senior · **Step:** tests · **Status:** done · **Handoff:** validation-specialist
**Measured TOTAL coverage:** 95.5% (gate ≥ 95.0% — passing)

## Summary

Wrote colocated unit tests + spec-traceable acceptance tests for the headless
umbrella (`config.IsHeadless`, the two prompt hook deciders) and the verdict
packet (`verdict.AssembleVerdict`, `verdict.EvidenceIndex`, `runVerdict`). All
new code paths are covered; the TOTAL coverage gate is restored to 95.5%.
Every `_test.go` is ≤100 lines (G1).

## Files Created (line counts)

Colocated unit tests:
- `internal/config/headless_test.go` (68) — IsHeadless truth table, nil cfg, envTrue
- `internal/config/headless_load_test.go` (41) — `[headless]` toml round-trip + defaults-off
- `cmd/centinela/review_mode_headless_test.go` (43) — headless beats every_step; back-compat
- `cmd/centinela/plan_advisor_headless_test.go` (35) — advisor silent under headless
- `cmd/centinela/verdict_cmd_test.go` (89) — runVerdict pass→nil+JSON, fail→sentinel+JSON, headless flag
- `internal/verdict/helpers_test.go` (32) — injected fake Deps + fixed Now
- `internal/verdict/assemble_test.go` (89) — pass/gate-fail/verify-fail/warn/provenance/nil-wf
- `internal/verdict/golden_test.go` (90) — byte-stable golden JSON, determinism
- `internal/verdict/mappers_test.go` (66) — gate lowercase, verify uppercase, counts
- `internal/verdict/evidence_index_test.go` (68) — sort by role, empty non-nil slice

Acceptance tests (spec traceability):
- `tests/acceptance/headless_governance_resolve_test.go` (59)
- `tests/acceptance/headless_governance_hooks_test.go` (88)
- `tests/acceptance/headless_governance_verdict_helper_test.go` (34)
- `tests/acceptance/headless_governance_verdict_summary_test.go` (73)
- `tests/acceptance/headless_governance_verdict_detail_test.go` (82)
- `tests/acceptance/headless_governance_command_test.go` (97)

Edge cases: `.workflow/headless-governance-edge-cases.md`

## Scenario → Test Mapping (25/25)

All 25 `Scenario:` lines in `specs/headless-governance.feature` map 1:1 to an
acceptance test via an exact `// Scenario: <title>` marker. Verified empty diff:
`comm -23 <(scenarios) <(// Scenario markers)` → empty.

- Resolution (5): CI auto-detect, CI+detect_ci off, zero-config, env override, empty env.
- Hooks/back-compat (6): env suppresses review, config suppresses review, config
  suppresses advisor, advisor quiet under headless, back-compat review renders,
  back-compat advisor speaks.
- Verdict assembly (10): pass, gate-fail, verify-fail-alone, warnings-don't-fail,
  deterministic JSON, run-info provenance, evidence-index lists, status casing,
  schema field, empty evidence index.
- Verdict command (4): JSON on stdout, sentinel on fail, headless flag, full-scan v1.

## Determinism

`AssembleVerdict` tests inject fake `Gates`/`Verify`/`Evidence` closures and a
fixed `Now` ("2026-06-12T00:00:00Z"), so golden JSON is byte-stable with no real
gates/verify/disk I/O. The golden test asserts two runs are byte-identical and
match an embedded golden document.

## Results

- `go build ./...` — clean
- `gofmt -l internal cmd tests` — empty
- `go test ./...` — 1569 passed
- `./scripts/check-coverage.sh` — coverage gate passed: 95.5% >= 95.0%
