### Gatekeeper Report: centinela-insights
**Date:** 2026-06-14
**Status:** SAFE

`centinela insights` is a read-only analytics command over the governance-telemetry
log. It introduces a new leaf-consuming aggregator package (`internal/insights`), a
thin `cmd/centinela/insights.go`, and a renderer (`internal/ui/render_insights.go`).
No existing domain entity, port, use case, DTO shape, or workflow state is mutated.
The feature only *reads* `internal/telemetry` (the existing append-only log) and adds
its own package; it changes no shared contract. Every checklist item was verified
mechanically (built `/tmp/cent-insgk` from the worktree; metrics verified empirically
against a synthetic `events.jsonl` in a temp dir — the real repo log was never touched).

#### Analyzed Specs
- `specs/centinela-insights.feature` (the new spec — 36 scenarios)
- All sibling `specs/*.feature` reviewed for shared-contract impact: insights consumes
  only the existing telemetry Event schema (read path) and adds no field, type, port,
  or workflow-state change. No existing scenario depends on insights output.

#### Findings

No blocking or warning findings. Mechanical verification results:

1. **File size (≤100 lines)** — PASS. `find internal cmd -name '*.go' | xargs wc -l | awk '$1>100'`
   returns only the `total` line; every new source AND `_test.go` file is ≤100 (largest
   new files: `cmd/centinela/insights_test.go` 87, `internal/ui/render_insights_test.go` 80,
   `internal/insights/compute_test.go` 73). G1 gate (`validate --full`) reports
   "All files under 100 lines."

2. **Cross-layer imports** — PASS. `internal/insights/*.go` imports ONLY
   `github.com/samuelnp/centinela/internal/telemetry` + stdlib (`sort`). No `cmd/`,
   no `internal/ui` import (grep-confirmed across all six source files). `centinela.toml`
   maps `internal/insights/**` into the `aggregator` layer (`allow = ["domain","leaf"]`),
   and PROJECT.md G2 prose documents the new edge (insights → telemetry leaf, read-only;
   imported only by cmd/; Report type read by internal/ui for rendering). Confirmed
   `internal/insights` does NOT appear in the import_graph "unmapped packages" warn
   (grep count 0 in full-scan output) — the mapping took effect and insights adds no new
   failing edge.

3. **`centinela validate --full` passes** — PASS (exit 0, "All gates passed"). G1,
   G-Build (6 release targets cross-compile), roadmap_drift all ✓; `go test ./...`,
   `go test ./tests/acceptance/...`, `check-coverage.sh`, `check-fmt.sh` all ✓.
   The two ⚠ (`import_graph` unmapped-packages, `spec-traceability` uncovered-scenarios)
   are PRE-EXISTING repo-wide non-failing warns, NOT newly caused by insights:
   - import_graph: `internal/insights` is mapped (not in the unmapped set).
   - spec-traceability: set difference of the 36 feature scenarios against the
     `// Scenario:` comments under `// Acceptance: specs/centinela-insights.feature`
     headers is EMPTY — all 36 covered. No `centinela-insights.feature` scenario is
     listed uncovered.
   NOTE: an earlier `validate --full` run reported acceptance FAILs; these were
   transient Go build-cache corruption (`cannot open .../go-build/...-d`,
   `package ... is not in std`) from a concurrent `go test -cover` thrashing the cache.
   After `go clean -cache`, the re-run passed cleanly and `go test ./...` is green
   (2045 tests). Not a real failure.

4. **No business logic in outer layer** — PASS. `cmd/centinela/insights.go` (56 lines)
   is a thin shell: `telemetry.ReadDefault()` → `insights.Compute(events, top)` →
   `json.MarshalIndent` OR `ui.RenderInsights`. All aggregation/ranking/metric math lives
   in `internal/insights`; the renderer only formats. No reducer logic leaked into cmd/ or ui/.

5. **i18n** — N/A. i18n is disabled for this project (Go CLI, no locale matrix in
   PROJECT.md). User-facing strings are command help/labels, consistent with the rest
   of the binary. Noted, not a violation.

6. **Metric correctness** — PASS (verified empirically end-to-end via the binary against a
   hand-constructed synthetic log with known counts):
   - **Gates (most-failed)** ranks gate-failure by Gate, count desc then key asc.
     coverage×3, then tied-at-1 `<none>`,fmt,g1 in key-asc order. Empty Gate → `<none>`. ✓
   - **Blocks (most-triggered)** ranks `<reason> · <fileType>`, missing field → `<none>`,
     count desc then key asc. need-init·go (2) before `<none> · <none>` (1) before
     out-of-step·md (1). ✓
   - **Rework** = per-feature (gate-failure + verify-rejection + complete-rejected);
     step-advanced NOT counted; empty-Feature excluded. alpha=4 (2 gf + 1 vr + 1 cr),
     beta=4 (3 gf + 1 cr); empty-feature gate-failure correctly dropped. ✓
   - **Mean steps-to-green = (complete-rejected + step-advanced) / step-advanced.**
     1 rejection + 2 advances → 1.50 (verified). 1 advance + 1 rejection → 2.00.
     1 advance, 0 rejections → 1.00. ✓
   - **Zero advances → n/a**, no panic, exit 0 (denominator guarded by `Advances>0`;
     Mean stays 0, HasValue=false, renderer prints "n/a"). ✓
   - **Empty file AND missing file** → "no telemetry yet" empty-state, exit 0. ✓
   - **Malformed JSONL lines skipped, valid still counted** (telemetry.Read is lenient;
     verified malformed line dropped while surrounding valid events aggregated). ✓
   - **`--top N` truncation**: top 1 → each section ≤1; top 0 → "(no events)"; top 99
     (N>buckets) → all buckets returned. ✓
   - **`--json`**: valid JSON, BYTE-STABLE across two consecutive runs (cmp identical). ✓
   - **Determinism**: human output byte-identical across runs; ties broken count-desc
     then key-asc in both human and JSON. ✓
   - **Non-TTY**: piped output contains ZERO ANSI escape bytes (lipgloss auto-strips;
     `grep -c $'\x1b'` = 0). ✓

7. **Coverage** — PASS, genuine. `./scripts/check-coverage.sh`: 95.2% ≥ 95.0% gate.
   Per-package (raw `go tool cover -func`):
   - `internal/insights`: **100.0%** of statements (every reducer + rankTop + span fully
     exercised). Not a rounding pass.
   - `internal/ui`: 97.2% total; `RenderInsights`/`spanLabel`/`rankSection`/`stepsSection`
     100%; `dateOnly` 66.7% (the short-timestamp `len<10` fallback branch only).
   - `cmd/centinela`: 93.8% total; `runInsights` 83.3% (uncovered branch is the defensive
     `json.MarshalIndent` error path — effectively unreachable for a value-typed Report).
   NOTE: the repo-wide gate margin is thin (95.2% vs 95.0% = 0.2%); insights itself is at
   100% and improves the pool. Flagged for visibility, not a violation.

8. **Tests real & executed** — PASS. `go test ./...` = 2045 tests green across 28 packages.
   All 36 feature scenarios carry a matching `// Scenario:` acceptance comment under a
   `// Acceptance: specs/centinela-insights.feature` header (8 acceptance files; the 9th,
   `_helper_test.go`, is a shared helper with no scenarios — correct). Spot-checked
   `centinela_insights_steps_test.go` (asserts exact means 1.50/2.00/1.00 scoped to the
   Steps-to-Green section body, plus exit codes) and `centinela_insights_determinism_test.go`
   (asserts actual a<m<z tie-break ordering; asserts valid events survive AND the malformed
   line is dropped). Assertions check real metric values and exit codes — no trivially-true
   test theater.

#### Deferred Findings
- none

#### Recommendation
**SAFE.** No conflicts with existing specs or shared contracts. All eight checklist items
pass on mechanical verification: every file ≤100 lines, no cross-layer import violation
(`internal/insights` → telemetry leaf only, correctly mapped), `validate --full` green,
cmd layer thin, metrics empirically correct (including the steps-to-green formula, the
zero-advance n/a guard, malformed-line skipping, `--top` truncation, JSON byte-stability,
and non-TTY ANSI-stripping), coverage genuine (insights 100%, repo 95.2%), and tests real
and executed (2045 green, all 36 scenarios covered). The two ⚠ gate warnings are
pre-existing repo-wide non-failing warns to which insights adds nothing. The only watch-item
is the thin 0.2% repo coverage margin — informational, not blocking. Proceed.
