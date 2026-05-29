# Plan: Claim Verification

> Feature brief: [docs/features/claim-verification.md](../features/claim-verification.md)
> Independently re-derive ground truth and reject step completion when an
> agent's evidence claims diverge from reality. Go-first, v1.

## Problem framing

The orchestration validator (`internal/orchestration/`) today checks the
**form** of evidence JSON — field/context match, real-file `outputs`,
snapshot `inputs` — but never whether the **claims are true**. A subagent
can write a perfectly shaped `qa-senior.json` asserting "tests pass" and
"coverage rose" with neither real; the files exist, so the gate passes.
This pushes Centinela's core guarantee onto the human operator as a manual
discipline. Claim verification automates that re-audit: it re-derives
ground truth and hard-blocks `centinela complete` on divergence.

## Scope (locked — from brief)

**In (v1, Go-first):** four claim checks — (1) tests actually pass,
(2) coverage actually moved, (3) outputs aren't empty stubs, (4) edge cases
map to tests. Surfaced as standalone `centinela verify <feature>` AND wired
into the `complete` gate as a HARD block. Worktree-aware.

**Out (v1):** multi-language stub/coverage detection (Go only); auto-fixing
divergence; verifying non-claim prose in the `.md` companion; gating steps
that legitimately have no evidence contract (e.g. `code`).

## Layered design (PROJECT.md G2 / G7)

All decision logic lives in `internal/`; `cmd/` is thin wiring only;
rendering is `internal/ui/`. New domain package `internal/verify/`.

### Domain model — `internal/verify/`

| File | Responsibility | Est. lines |
|------|----------------|-----------|
| `result.go` | `Check{Claim, Role, Status(pass/fail/skip), Detail}` + `VerificationResult{Feature, Checks}` aggregate + `Failed()`/`HasFailures()` helpers. | ~70 |
| `verify.go` | `Verify(feature string, cfg, deps) VerificationResult` — loads evidence for required roles, dispatches each claim check, aggregates. No I/O of its own beyond injected deps. | ~90 |
| `claim_tests.go` | "tests pass" check — runs `cfg.Validate.Commands` (test cmd) via injected runner, maps non-zero exit → fail with command named. Distinguishes missing/misconfigured command (config error) from failed claim. | ~90 |
| `claim_coverage.go` | "coverage moved" check — re-derives per-package coverage (no `-coverpkg`, matching project model), compares claimed vs measured under a documented tolerance. Skips when no coverage claimed. | ~95 |
| `claim_stubs.go` | "outputs aren't empty stubs" — scans each `outputs` file for substantive content; Go test files must contain real assertions, not empty `func Test…(){}` bodies. Conservative; tiny interfaces/helpers exempt. | ~95 |
| `claim_edgecases.go` | "edge cases map to tests" — cross-checks each `edgeCases` entry against test names/assertions in the feature's test files. Reports divergence. | ~90 |
| `runner.go` | `CommandRunner` interface (`Run(cmd) (exit int, out string)`) + thin default impl, so checks are unit-testable without shelling out. | ~40 |

### Coverage / test re-run reuse

To avoid doubling test cost inside `complete`, the test+coverage commands
are run **once** and the result fed to both `claim_tests` and
`claim_coverage`. When `complete` reaches the `validate` step it already
calls `executeValidation()` (which runs `cfg.Validate.Commands`); verify
consumes that run's outcome rather than re-running. `centinela verify`
invoked standalone runs the commands itself (it has no prior run to reuse).
A documented timeout scopes the run; non-deterministic suites are the
operator's responsibility to make deterministic.

### Configuration — `internal/config/`

Reuse existing `cfg.Validate.Commands`, `DiffMode`, `DiffBase`. Add a
small `verify` block only if a tolerance/timeout knob is needed
(`coverage_tolerance`, `verify_timeout`); default values live in
`applyDefaults`. No new business logic in config (leaf layer).

### Worktree — `internal/worktree/`

Verification must run against the active worktree tree, not root. Reuse
`worktree.DetectFeatureFromCwd(cwd)` so paths (`.workflow/`, test commands,
coverage profile) resolve inside `.worktrees/<feature>/` when
`use_worktrees` is on. No new worktree code expected; verify takes the
resolved root as a dependency.

### Presentation — `internal/ui/`

| File | Responsibility | Est. lines |
|------|----------------|-----------|
| `render_verify.go` | `RenderVerification(VerificationResult) string` — per-claim PASS/FAIL/SKIP lines + summary. Pure rendering; no decisions. Reuses `styles.go`. | ~70 |

### CLI wiring — `cmd/centinela/`

| File | Responsibility | Est. lines |
|------|----------------|-----------|
| `verify.go` | `centinela verify <feature>` — loads config, resolves worktree root, calls `verify.Verify(...)`, prints `ui.RenderVerification(...)`, exits non-zero on any fail. Thin. | ~55 |
| `complete.go` (edit) | After/with `executeValidation()` in the `validate` step, call `verify.Verify` reusing the just-run results and return an error on any failure (HARD block, no warn bypass). | +~15 |

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Regression to existing `complete` gate (false block on honest evidence) | High | Medium | Verify only required-role evidence that exists; absent claim → skip, not fail. Acceptance test: a clean truthful workflow verifies green and completes unchanged. |
| Stub heuristic false positives (tiny interfaces/helpers flagged) | High | Medium | Conservative rules: only flag empty-bodied `func Test…`, zero-assertion test files, and whitespace/boilerplate-only files. Exempt files below a content threshold that are non-test. Unit tests for each heuristic. |
| Coverage re-derivation drift (exact match flaky) | Medium | High | Documented tolerance (claimed must hold within tolerance; claim *above* measured fails). Per-package, no `-coverpkg`, matching project model. |
| Doubling test cost inside `complete` | Medium | High | Reuse the `validate`-step run; standalone `verify` runs once with a scoped timeout. |
| Worktree mis-resolution (verifies root tree instead of feature tree) | Medium | Medium | Resolve root via `worktree.DetectFeatureFromCwd`; acceptance test runs verify from inside a worktree. |
| Misconfigured/missing test command read as a failed claim | Low | Medium | Distinct config-error path surfaced separately from claim-fail. |

## Rollout sequence (smallest correct slice first)

1. **Slice 1 — core (`claim-verification-core`).** `internal/verify/`
   `result.go` + `verify.go` + `runner.go` + `claim_tests.go` +
   `claim_stubs.go`; `internal/ui/render_verify.go`; `cmd/centinela/verify.go`;
   wire the HARD block into `complete.go` (validate step). Tests-pass and
   outputs-stub checks only. This is shippable on its own and delivers the
   primary guarantee.
2. **Slice 2 — coverage (`claim-verification-coverage`).** Add
   `claim_coverage.go` + tolerance config + reuse of the validate-step
   coverage run. Plugs into the existing dispatch in `verify.go`.
3. **Slice 3 — edge cases (`claim-verification-edgecases`).** Add
   `claim_edgecases.go` mapping `edgeCases` → test names/assertions.
4. **Tests.** Unit per claim check (table-driven, injected runner);
   integration for `complete`-gate wiring and worktree resolution;
   acceptance scenarios: fabricated tests-pass blocked, stub output
   blocked, coverage overclaim blocked, edge-case-without-test reported,
   honest workflow verifies green.
5. **Docs.** Add a `verify` section to README + workflow-enforcement.md;
   document the hard-block in evidence-contract.md and the tolerance/timeout
   knobs in the `centinela.toml` reference.

## Open questions for feature-specialist

- Edge-case check: hard-fail or warn-only? Brief leaves the policy to plan;
  recommend **report-and-warn** in v1 (mapping is heuristic) while
  tests/coverage/stubs hard-fail — confirm in the Gherkin.
- Exact coverage tolerance value (e.g. 0.1%) and verify timeout default.
- Where does verify read the *claimed* coverage figure from — a dedicated
  evidence field, or parsed from `outputs`/edgeCases prose? May need an
  evidence-schema addition (coordinate with evidence-contract).
