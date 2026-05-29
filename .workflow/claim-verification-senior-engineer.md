### Senior-Engineer Report: claim-verification

**Date:** 2026-05-29

#### Files Touched
| Path | Reason |
|------|--------|
| internal/verify/result.go | `Check` + `VerificationResult` model, status helpers (`HasFailures`/`HasWarnings`/`Failed`/`Tally`). |
| internal/verify/verify.go | `Verify(feature, step, cfg, Deps)` dispatcher over required-role evidence; injected `EvidenceLoader`/`CommandRunner` for testability. |
| internal/verify/runner.go | `CommandRunner` interface + `execRunner` default impl with timeout. |
| internal/verify/claim_tests.go | tests-pass check; distinguishes config error (missing/misconfigured cmd) from failed claim. |
| internal/verify/claim_coverage.go | coverage-moved check; per-package mean (no `-coverpkg`), claim-above-measured beyond tolerance fails; absent claim skips. |
| internal/verify/claim_stubs.go | outputs-not-stubs check; flags empty-bodied tests / zero-assertion files; conservative. |
| internal/verify/claim_edgecases.go | edge-case→test mapping; WARN only (heuristic). |
| internal/ui/render_verify.go | pure `RenderVerification` PASS/FAIL/SKIP/WARN report. |
| cmd/centinela/verify.go | thin `centinela verify <feature>`; resolves worktree root via `worktree.DetectFeatureFromCwd`. |
| cmd/centinela/complete.go | wires `runClaimVerification` into the validate-step gate as a HARD block (blocks on `HasFailures()` only). |
| internal/config/verify_config.go + config.go | `VerifyConfig` (`verify_timeout`=60, `coverage_tolerance`=0.001) + `applyDefaults`. |
| internal/evidence/schema.go, setter.go, setter_parse.go, schema_marshal.go | typed optional `Coverage *float64` field + `evidence set ... coverage` support + marshalling. |
| docs/architecture/evidence-contract.md | documented the new optional `coverage` field + global rule 7. |

#### Architecture Compliance
- Boundary checks: `internal/verify` imports `config` (leaf), `evidence`, `orchestration` only — no `cmd/` or `ui/` imports. `cmd/` wiring stays thin (decisions live in `internal/verify`). `internal/ui/render_verify.go` renders only.
- G1 file size: every source file ≤ 100 lines (`config.go` 93, `setter.go` 82 after extracting `setter_parse.go`/`verify_config.go`; `complete.go` 93). Full-tree scan clean.
- G7 outer-layer: no business logic in `cmd/`; the gate calls into `verify.Verify` and only inspects `HasFailures()`.
- **For gatekeeper:** `internal/verify` is a new package not named in PROJECT.md's G2 rule. Its dependency set (config/evidence/orchestration) is a reasonable domain-service shape; PROJECT.md → G2 rule should be updated to ratify it.

#### Type-Safety Notes
- `Coverage` is `*float64` (nil = no claim) — no `any`/`interface{}`; prose coverage claims are impossible by construction.
- `Status` is a typed enum with a `blocking()` method; warn vs fail is type-driven, not string-compared at call sites.
- `CommandRunner`/`EvidenceLoader` are interfaces/func types injected via `Deps`, so checks are unit-testable without shelling out or touching disk.

#### Trade-Offs
- The validate-step gate currently re-runs the test/coverage command rather than reusing `executeValidation()`'s output. `Deps.PriorTestRun` exists to plumb that reuse later; deferred to avoid refactoring `executeValidation`'s signature in this slice. **TODO(qa/perf):** wire `PriorTestRun` to halve cost at the gate.
- edge-case→test mapping is intentionally warn-only (string-heuristic; hard-blocking would be too false-positive-prone in v1).

#### Handoff
- Next role: qa-senior
- Outstanding TODOs: unit tests per claim check (table-driven, injected runner); integration test for the complete-gate hard block + worktree resolution; acceptance scenarios from specs/claim-verification.feature (fabricated pass blocked, stub blocked, coverage overclaim blocked, edge-case-without-test warns, honest workflow green); wire `Deps.PriorTestRun`; add `verify` knobs to the centinela.toml reference doc.
