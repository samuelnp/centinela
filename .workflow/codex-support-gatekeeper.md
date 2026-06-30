### Gatekeeper Report: codex-support
**Date:** 2026-06-30
**Status:** SAFE

#### Analyzed Specs
- specs/codex-support.feature
- specs/host-harness-adapters.feature
- specs/adapt-opencode-support.feature
- All other specs in specs/ reviewed for shared-surface conflicts (none found)

#### Findings

**Finding 1 — Registry expansion (orderedAgents + registry map)**
- **Affected spec:** host-harness-adapters.feature
- **Affected scenario:** "Registry returns a typed error for an unknown agent" (AC1) / "both selector composes adapters without a special-case branch" (AC7)
- **Risk:** Adding "codex" to orderedAgents could break `AgentsFor("both")` if the composites map were inadvertently widened; ErrUnknownAgent message now lists four names where the spec named three.
- **Verdict:** No conflict. The composites map is unchanged ("both" → ["claude","opencode"] only). `TestAgentsFor_Both` asserts exactly `[claude opencode]`. The spec says the error "lists the registered harness names claude, opencode, aider" — it does not say exclusively; the additive "codex" in the error message satisfies the assertion. The `adapter_registry_test.go` TestLookup_UnknownAgent_TypedError now checks all four names in the error, which is consistent.

**Finding 2 — Prewrite hook refactor (evalPrewrite → evalPrewriteMulti)**
- **Affected spec:** adapt-opencode-support.feature / host-harness-adapters.feature (blocking scenarios)
- **Affected scenario:** "OpenCode blocks out-of-step writes", "Aider does not wire a prewrite hook"
- **Risk:** Swapping `evalPrewrite` for `evalPrewriteMulti` could change blocking behavior for Claude/OpenCode, which send absolute `file_path`/`filePath`.
- **Verdict:** No regression. `prewriteTargets()` returns a single-element slice for Claude/OpenCode (file_path or filePath present). `EvaluatePrewriteMulti` calls `EvaluatePrewrite` with that one absolute path — identical evaluation path as before. The new `PrewriteDecision.Path` field is additive (zero value empty string); the blocking display uses `d.Path` which `EvaluatePrewriteMulti` sets to the original path on a block. Rendering is identical for single-path callers. `hook_prewrite_block_test.go` and `hook_prewrite_codex_test.go` both cover these paths.

**Finding 3 — coverage_hardening_test.go: TestNoBehaviourChange self-scoping**
- **Affected spec:** specs/coverage-hardening.feature
- **Affected scenario:** "No production behaviour changed"
- **Risk:** Change could weaken the guard if self-scoping logic were incorrect.
- **Verdict:** No weakening. The guard checks `strings.Contains(subj, "coverage-hardening")` against `git log --format=%s main..HEAD`; on any branch other than coverage-hardening (including codex-support), the test skips. The coverage floor assertion (`MIN_COVERAGE:-95.0`) lives in a separate test (`TestCoverageGate_ScriptAndFloor`) which is not modified and runs on every branch.

**Finding 4 — HarnessAdapter / SyncItem / SyncKind interface shapes**
- **Affected spec:** host-harness-adapters.feature (all ACs)
- **Risk:** Shape changes to shared interfaces could silently break existing adapter implementations.
- **Verdict:** No shape changes. HarnessAdapter interface is unchanged. SyncItem and SyncKind are unchanged. The only DTO delta is the additive `Path string` field on `PrewriteDecision` — zero-value safe, no existing code sets or reads it.

**Finding 5 — Golden parity test extended to codex**
- **Affected spec:** host-harness-adapters.feature AC4 / codex-support.feature AC6
- **Risk:** Adding a "codex" case to TestGoldenParityClaudeOpenCode could mask regressions in the claude/opencode cases if the test structure changed.
- **Verdict:** No conflict. The claude and opencode fixture assertions are structurally identical. The codex golden fixture at `testdata/golden/codex/` covers `.codex/config.toml` and `AGENTS.md`. The shared AGENTS.md golden path is the codex-specific one, separate from the opencode golden.

#### File Size Gate (G1)

All changed/new Go source files are within the 100-line limit:

| File | Lines |
|------|-------|
| tests/acceptance/coverage_hardening_test.go | 97 |
| cmd/centinela/hook_postwrite.go | 96 |
| cmd/centinela/hook_prewrite.go | 93 |
| cmd/centinela/init.go | 92 |
| internal/setup/adapter_registry_test.go | 91 |
| internal/setup/codex_config_test.go | 83 |
| internal/setup/sync.go | 82 |
| internal/hookpolicy/prewrite.go | 74 |
| cmd/centinela/init_agent.go | 67 |
| cmd/centinela/hook_prewrite_codex_test.go | 67 |
| internal/hookpolicy/applypatch_multi_test.go | 66 |
| tests/acceptance/codex_support_ac2_test.go | 65 |
| internal/setup/codex_config.go | 61 |
| internal/hookpolicy/applypatch.go | 58 |
| tests/acceptance/codex_support_ac1_test.go | 56 |
| cmd/centinela/hook_prewrite_block_test.go | 51 |
| internal/setup/adapter.go | 50 |
| cmd/centinela/setup_codex_test.go | 50 |
| internal/setup/adapter_codex_test.go | 49 |
| internal/hookpolicy/applypatch_test.go | 48 |
| cmd/centinela/migrate_setup.go | 48 |
| internal/setup/golden_parity_test.go | 47 |
| internal/setup/adapter_codex.go | 24 |

No file exceeds 100 lines. No G1 exceptions required.

#### Cross-Layer Import Violations

- `internal/hookpolicy` imports `internal/config` and `internal/workflow` — both are domain/leaf packages; hookpolicy is infrastructure consumed by cmd/. Allowed.
- `cmd/centinela/hook_prewrite.go` imports `internal/workflow` directly — cmd/ may import internal/*. Allowed per G2 rule.
- No violations detected.

#### i18n

Gate disabled (English-only CLI). Not applicable.

#### Backward Compatibility for Existing Harnesses

- Claude harness: prewrite hook evaluates single absolute path as before. Golden parity fixture unchanged.
- OpenCode harness: same single-path evaluation path. Golden parity fixture unchanged.
- Aider harness: no prewrite hook; not in composites. Unaffected by any change.

#### Deferred Findings

None.

#### Recommendation

SAFE: No conflicts detected. Proceed with implementation.
