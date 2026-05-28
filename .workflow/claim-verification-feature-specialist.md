### Feature-Specialist Report: claim-verification
**Date:** 2026-05-28

#### Behavior Summary

`centinela verify <feature>` independently re-derives ground truth for every claim made in a feature's evidence JSON files and prints a per-claim PASS/FAIL/SKIP/WARN/TIMEOUT report, exiting non-zero whenever any claim cannot be confirmed. Three of the four checks — tests-pass, coverage-moved, and outputs-not-stubs — are **hard failures**: a divergence immediately blocks `centinela complete` with no bypass path. The fourth check — edge-cases-to-tests mapping — is a **warning** in v1 because the mapping is heuristic (string matching between edgeCases prose and test function names); it is surfaced in the complete output but does not prevent advancement. Claim types absent from evidence (e.g. no coverage field) are skipped, not failed; steps with no evidence at all produce "no claims to verify" and do not block. Configuration problems (missing `validate.commands`, binary not on PATH, suite timeout) surface as distinct CONFIG ERROR or TIMEOUT results, never as false claim failures. Verification is worktree-aware: paths, evidence files, and test commands are resolved inside `.worktrees/<feature>/` when `use_worktrees` is on. Inside `centinela complete`, verification reuses the test-suite run already performed by `executeValidation()` to avoid doubling cost.

#### Gherkin Scenarios

Full spec at `specs/claim-verification.feature` (20 scenarios).

**Happy path:**
- Honest evidence with passing tests, coverage within tolerance, substantive outputs, and matched edge cases → all PASS, exit 0, complete advances.

**Negative paths — HARD FAIL:**
1. Fabricated tests-pass (commands exit non-zero) → FAIL for tests-pass check; `complete` hard-blocked.
2. Inflated coverage (claimed 92%, measured 78%, delta > 0.1% tolerance) → FAIL for coverage check; `complete` hard-blocked.
3. Coverage within tolerance (claimed 85.0%, measured 84.95%, delta 0.05% < 0.1%) → PASS.
4. Empty-stub output file (empty `func TestFoo(t *testing.T) {}` body) → FAIL for outputs-stub check; `complete` hard-blocked.
5. Legitimately tiny non-test file (small interface) → PASS (exempt from stub check).

**Negative paths — WARN (not hard-fail):**
6. Edge case with no matching test name → WARN; `complete` advances but surfaces the warning.
7. All edge cases matched → PASS.

**Skip scenarios:**
8. No evidence files → SKIP all, exit 0, complete not blocked.
9. Evidence omitting coverage field → SKIP coverage check only.
10. Evidence with empty edgeCases → SKIP edge-cases check.

**Config error scenarios:**
11. `validate.commands` missing/empty → CONFIG ERROR (not FAIL), instructs user to configure.
12. Test command binary not on PATH → CONFIG ERROR naming the missing binary.

**Worktree scenarios:**
13. CWD inside worktree → evidence and test commands resolved relative to `.worktrees/<feature>/`.
14. CWD at repo root → verify resolves the correct worktree by feature name.

**Timeout scenario:**
15. Suite exceeds `verify_timeout` (30s configured) → TIMEOUT result; `complete` hard-blocked.

**CLI surface scenarios:**
16. Mixed-result report → one labeled line per check + summary line.
17. `complete` gate → full verify report shown inline before gate error line; claim-fail message distinct from structural-evidence failure.

#### UX States

| State   | Trigger                                             | Surface                                      |
|---------|-----------------------------------------------------|----------------------------------------------|
| loading | `centinela verify` invoked; test commands running   | CLI: "Running verification for <feature>..."  |
| empty   | No evidence files present for the feature           | CLI: per-check SKIP lines + "no claims to verify" |
| error   | CONFIG ERROR (missing command / binary not on PATH); TIMEOUT | CLI: CONFIG ERROR or TIMEOUT line with detail; exit non-zero |
| success | All present checks pass                             | CLI: PASS lines + summary; exit 0            |

#### Hard-Fail vs Warn Policy (pinned)

| Check                    | Policy in v1    | Rationale                                              |
|--------------------------|-----------------|--------------------------------------------------------|
| tests-pass               | HARD FAIL       | Binary, deterministic — exit code cannot be ambiguous |
| coverage-moved           | HARD FAIL       | Numeric, objectively measurable (within tolerance)    |
| outputs-not-stubs        | HARD FAIL       | Structural, conservative heuristic, unit-tested       |
| edge-cases-to-tests      | WARN ONLY       | Heuristic string matching; false positives expected in v1 |

TIMEOUT and CONFIG ERROR are also hard-blockers for `complete` (the claim cannot be confirmed, which is as bad as confirmed failed).

#### Out-of-Scope (v1)

- Multi-language stub/coverage detection (Go only in v1).
- Auto-fixing divergence — verify reports, does not repair.
- Verifying narrative prose in the `.md` companion files.
- Gating steps with no evidence contract (e.g. the `code` step).
- Warn-only bypass flag on any of the three hard-fail checks.
- Evidence-schema addition for the claimed coverage figure field (open item — see below).

#### Handoff

- Next role: senior-engineer
- Open clarifications:
  1. **Claimed coverage figure source**: The Gherkin asserts "evidence claiming X% coverage" but the current evidence schema has no dedicated `coverage` field. The senior-engineer must decide: (a) add a `coverage` field to the evidence JSON schema (requires updating `internal/orchestration/` and the `evidence-contract.md`), or (b) parse the figure from `edgeCases` prose using a convention (e.g. `"coverage: 85%"`). Option (a) is cleaner and recommended; option (b) is fragile. This is a schema decision that affects the code design.
  2. **verify_timeout default**: The feature-specialist has set the default to **60 seconds** based on typical Go suite run times. The Gherkin scenario uses 30s as a configurable override. Senior-engineer should confirm this default in `applyDefaults()` or adjust based on known suite durations.
  3. **coverage_tolerance default**: Pinned at **0.1%** (0.001 fractional). Senior-engineer should wire this into `applyDefaults()` and ensure the comparison uses the same rounding as `go test -cover` output.
  4. **Stub check threshold for non-test files**: The Gherkin says "under content threshold" — senior-engineer must define the exact line/byte threshold in `claim_stubs.go` and ensure the heuristic is unit-tested for interface-only files.
