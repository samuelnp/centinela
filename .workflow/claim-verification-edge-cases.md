# Edge Cases: claim-verification

## Covered

Each edge case below maps to a test name (unit/integration/acceptance) so the
heuristic edge-case→test mapping itself recognises them.

- **Absent coverage field skips the coverage check** — `ev.Coverage == nil`
  yields `StatusSkip`, not a fail. Covered: `TestCheckCoverage/absent-skip`,
  `TestVerifyDispatchesFourChecks`.
- **Coverage tolerance boundary** — claim within tolerance passes, claim above
  measured beyond tolerance fails. Covered:
  `TestCheckCoverage/within-tolerance-pass`, `/overclaim-fail`,
  `TestCheckCoverageWithinTolerance`, `TestAcceptance_CoverageOverclaimBlocked`.
- **Misconfigured test command vs a failed claim** — empty `validate.commands`
  and a missing binary surface as `CONFIG-ERROR`, distinct from a non-zero exit
  (`FAIL`). Covered: `TestCheckTestsPassNoCommands`,
  `TestCheckTestsPass/missing-binary`, `/fail`.
- **Timeout is not a claim failure** — a timed-out run surfaces as `TIMEOUT`
  (still blocking) rather than `FAIL`. Covered: `TestCheckTestsPass/timeout`,
  `TestCheckCoverage/timeout`, `TestExecRunnerTimeout`.
- **No-evidence skip does not block** — a started feature with no evidence JSON
  produces a single `SKIP` placeholder and never blocks. Covered:
  `TestVerifyNoEvidenceSkips`, `TestVerifyDefaultLoader`,
  `TestRunVerifyNoEvidenceClean`, `TestAcceptance_NoEvidenceSkips`.
- **Stub false-positive avoidance** — empty-bodied test funcs and zero-assertion
  test files fail; tiny interfaces/helpers and non-Go/unreadable outputs pass.
  Covered: `TestCheckStubs` (all rows), `TestCheckStubsEmptyAndNonGo`,
  `TestAcceptance_StubOutputBlocked`.
- **Edge-case mapping is WARN-only** — an unmatched edge case warns but does not
  hard-block; matched entries pass. Covered: `TestCheckEdgeCases`,
  `TestEdgeCaseMatches`, `TestAcceptance_UnmappedEdgeWarnsNotBlocks`.
- **Worktree vs root resolution** — verification reads evidence and test files
  from the resolved root, not the process CWD. Covered:
  `TestVerifyResolvesWorktreeRoot`, `TestVerifyRoot`.
- **Complete-gate hard block** — the gate blocks on `HasFailures()` only;
  honest evidence advances, fabricated claims are blocked with divergence named.
  Covered: `TestRunClaimVerificationHardBlocksOnFailedTests`,
  `TestRunClaimVerificationPassesOnHonestEvidence`,
  `TestCompleteGateBlocksFabricatedClaim`, `TestAcceptance_FabricatedTestsBlocked`.

## Residual Risks

- **Non-Go stub detection** out of scope in v1 (Go-only heuristics). Mitigation:
  non-Go outputs are conservatively treated as substantive (never false-flagged).
- **Coverage re-derivation drift** — per-package mean (no `-coverpkg`), so the
  measured figure is an average across packages and can differ from a single
  aggregate number. Mitigation: documented tolerance; claim-above-measured is
  the only failing direction.
- **Edge-case heuristic** is significant-word string matching (≥4 chars) and may
  miss paraphrased mappings. Mitigation: WARN-only, never blocks completion.
- **PriorTestRun reuse** is plumbed but the gate still re-runs the suite (TODO
  from senior-engineer). No correctness risk; a performance follow-up.
