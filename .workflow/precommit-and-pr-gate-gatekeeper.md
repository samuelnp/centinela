### Gatekeeper Report: precommit-and-pr-gate

**Date:** 2026-06-17
**Status:** SAFE

## Analyzed Specs

- `specs/precommit-and-pr-gate.feature` (17 scenarios: precommit block/pass/skip-build/warn, installer idempotency + preserve/uninstall, pr-gate markdown verdict, fail_on_warning, custom+audit composition, determinism)
- Sibling specs skimmed for conflict: `specs/custom-gate-sdk.feature`, `specs/audit-baseline-ratchet.feature`, `specs/g2-import-graph-gate.feature`, `specs/cross-platform-build-gate.feature`, `specs/enforce-coverage-in-validate.feature`. No scenario in any sibling spec references `precommit`/`pr-gate`/`pr_gate` — no spec collision.

## Findings

**1. config.go → gates_config.go extraction is behavior-preserving (PASS).**
`GatesConfig` moved verbatim into `internal/config/gates_config.go`: identical 12 fields and identical `toml` tags (`file_size`, `file_size_exceptions`, `i18n`, `production_readiness`, `build`, `import_graph`, `security`, `spec_traceability`, `roadmap_drift`, `audit_baseline`, `custom`). config.go diff is purely the removal of that block plus two additive `Config` fields. `go build ./...` succeeds — every consumer of `GatesConfig` resolves. config.go is now 87 lines (was the source of a prior 102-line CI failure; now well under 100).

**2. New `[precommit]`/`[pr_gate]` sections are additive and safe (PASS).**
New `Config` fields `Precommit PrecommitConfig` / `PrGate PrGateConfig`. Both normalize to safe defaults when absent: precommit `Enabled=false` (advisory only), `SkipBuild` defaults true via `*bool RawSkipBuild` (omitted→nil→true, distinguishable from explicit false); pr_gate `Enabled=false`, `FailOnWarning=false`. `validatePrecommit`/`validatePrGate` are no-ops wired into `validateConfig`. Absent sections → zero behavior change for existing configs.

**3. centinela.toml import-graph: only ADDS `internal/githooks/**` to the leaf layer (PASS).**
`internal/githooks/{install,splice}.go` import only stdlib (`os`, `path/filepath`, `strings`); zero project packages (grep-confirmed). Leaf `allow=[]` is satisfied — no new failing edge. `centinela validate` import_graph result is a single ⚠ "packages match no configured layer" (the pre-existing non-failing warning for unmapped ui/roadmap/verify), not a failure.

**4. internal/ui/render_markdown.go adds no new layer edge (PASS).**
`RenderGatesMarkdown` imports `internal/gates`; `ui` already imports `gates` (render_gates.go). `ui` is an unmapped package → non-failing warning, unchanged.

**5. .github/workflows/validate.yml change is correctly guarded (PASS).**
Existing `validate` job is untouched and still runs on both `push` and `pull_request`. New `pr-gate` job is gated `if: github.event_name == 'pull_request'` with scoped `permissions: pull-requests: write`. Verdict rendered with `|| true` so the comment always posts; a final "Enforce gate exit code" step re-runs `pr-gate` to fail the check. Comment uses `--edit-last` with fallback (single-comment update). No interference with the validate job.

**6. Gate-keepers checklist:**
- File size ≤100: PASS. Full-repo scan of internal/ + cmd/ finds zero .go files >100. All 27 new/changed internal/cmd files (incl. every `_test.go`) ≤100; max is 96 (precommit_cmd_test.go). No G1 exception needed.
- Cross-layer imports / cycles: PASS. githooks is a stdlib-only leaf; gates never imports ui/audit; no cycle introduced.
- `centinela validate`: PASS (exit 0; G1, G-Build all-6-targets, spec-traceability 15/15, roadmap_drift, go test, coverage, fmt all green).
- Business logic in outer layer: N/A — cmd/ commands delegate to gates/gitdiff/githooks/ui; no domain logic embedded.
- i18n: N/A — Centinela's own operator-facing Go CLI has no i18n layer (PROJECT.md locales govern validated *projects*, not Centinela's CLI strings); consistent with all sibling features.
- Production readiness gate: not enabled.

**7. Test suite:** `go test ./...` → 2317 passed across 31 packages.

## Deferred Findings

none

## Recommendation

Ship. The shared-config refactor is provably behavior-preserving (verbatim struct/tags + clean compile), all new config is additive with safe defaults, the import-graph leaf addition introduces no failing edge, and the CI workflow change is correctly scoped and guarded. Every file ≤100 lines, full validate passes, 2317 tests green. Hand off to validation-specialist.
