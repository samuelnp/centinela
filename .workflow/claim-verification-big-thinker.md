### Big-Thinker Report: claim-verification
**Date:** 2026-05-28

#### Problem
Centinela's promise is *trustworthy* enforcement of plan → code → tests →
validate → docs. Today the orchestration validator (`internal/orchestration/`)
checks the **form** of evidence JSON — field/context match, real-file
`outputs`, snapshot `inputs` — but never whether the **claims are true**. A
subagent can write a perfectly shaped `qa-senior.json` asserting "tests pass"
and "coverage rose" with neither real; the files exist, so the gate passes.
The most important check — independent claim verification — is left to the
human operator as a manual discipline. The user hurting is the developer who
delegates steps to subagents and relies on `centinela complete` to mean the
work is genuinely done, not merely well-described. Why now: parallel worktrees
and richer orchestration mean more delegated, harder-to-audit handoffs.

#### Scope
- In (v1, Go-first): four claim checks — (1) tests actually pass, (2) coverage
  actually moved, (3) outputs aren't empty stubs, (4) edge cases map to tests.
  Standalone `centinela verify <feature>` AND wired into the `complete` gate as
  a HARD block (no warn-only bypass). Worktree-aware.
- Out (v1): multi-language stub/coverage detection (Go only); auto-fixing
  divergence; verifying prose in the `.md` companion; gating steps with no
  evidence contract (e.g. `code`).

#### Dependencies & Assumptions
- New domain package `internal/verify/` (VerificationResult model + per-claim
  checks). G2/G7: no business logic in `cmd/`; rendering in `internal/ui/`.
- `internal/orchestration/` — verification plugs into the existing `complete`
  gate beside structural evidence validation; reads evidence via the existing
  `Evidence` struct / role paths.
- `internal/config/` — reuses `Validate.Commands`, `DiffMode`, `DiffBase`;
  optional `coverage_tolerance`/`verify_timeout` knobs with defaults.
- `internal/worktree/` — `DetectFeatureFromCwd` resolves the active tree so
  `.workflow/`, test commands, and coverage profile resolve inside
  `.worktrees/<feature>/` when `use_worktrees` is on.
- `cmd/centinela/` — thin `verify` command + ~15-line wiring in `complete.go`.
- Cost reuse: the `validate` step already runs `cfg.Validate.Commands` via
  `executeValidation()`; verify consumes that run rather than re-running.
  Coverage uses the project's per-package, no-`-coverpkg` model.
- `internal/ui/` — `RenderVerification` renders PASS/FAIL/SKIP; pure.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Regression to `complete` gate (false block on honest evidence) | High | Medium | Verify only required-role evidence that exists; absent claim → skip not fail; acceptance test asserts a clean workflow verifies green. |
| Stub heuristic false positives (tiny interfaces/helpers flagged) | High | Medium | Conservative: only flag empty-body `func Test…`, zero-assertion test files, boilerplate-only files; non-test files below threshold exempt; unit-tested. |
| Coverage re-derivation drift (exact match flaky) | Medium | High | Documented tolerance; claim *above* measured fails; per-package no-`-coverpkg`. |
| Doubling test cost inside `complete` | Medium | High | Reuse validate-step run; standalone verify runs once with a scoped timeout. |
| Worktree mis-resolution (root tree vs feature tree) | Medium | Medium | Resolve via `DetectFeatureFromCwd`; acceptance test runs verify inside a worktree. |
| Misconfigured/missing test command read as failed claim | Low | Medium | Surface config error distinctly from claim-fail. |

#### Rollout
- Step 1: Core slice — `internal/verify/` (result, verify, runner,
  claim_tests, claim_stubs) + `ui/render_verify.go` + `cmd/centinela/verify.go`
  + HARD-block wiring in `complete.go`. Tests-pass + outputs-stub only.
- Step 2: Coverage slice — `claim_coverage.go` + tolerance config, reusing the
  validate-step coverage run.
- Step 3: Edge-cases slice — `claim_edgecases.go` mapping edgeCases → tests.
- Step 4: Tests — unit per check (injected runner), integration for complete
  wiring + worktree resolution, acceptance for each fabrication scenario plus
  the honest-green path.
- Step 5: Docs — README + workflow-enforcement.md + evidence-contract hard-block
  + centinela.toml knobs.

#### Handoff
- Next role: feature-specialist
- Outstanding questions: edge-case check hard-fail vs warn-only (recommend
  report-and-warn in v1; tests/coverage/stubs hard-fail); exact coverage
  tolerance + verify timeout defaults; source of the *claimed* coverage figure
  (dedicated evidence field vs parsed prose — may need an evidence-schema add).
