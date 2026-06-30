# Validation Specialist Report: local-harness-support

**Date:** 2026-06-30
**Result:** PASS

## Gatekeeper review

Independent adversarial review delegated to the Gatekeeper subagent —
`.workflow/local-harness-support-gatekeeper.md`, **Status: SAFE**. All seven
shared-surface checks clean: the `*LocalProvider` signature changes preserve a
byte-identical zero-config path (`BuildSyncPlan` delegates to
`BuildSyncPlanWithLocal(agent, nil)`, guarded by the
`TestBuildOpenCodeConfigNilLocalGoldenParity` golden tripwire); the
`DefaultProfileForModel` change keeps the model-capability-profiles precedence
invariant (explicit `--profile` > global `enforcement_profile` > driver-model
capability > strict), with the local default as the strictly-lowest tier engaged
only for a declared, unmapped local model; `OrchestrationConfig.Local` and
`SyncItem.Local` are additive optionals; layer rules hold (`internal/config` and
`internal/setup` import nothing internal, the config→setup mapping lives in
`cmd/`); G1 satisfied (all files ≤100 lines).

## Removed test — accepted

`TestNoBehaviourChange_OnlyTestFilesAdded` (tests/acceptance/coverage_hardening_test.go)
was a self-referential guard from the test-only `coverage-hardening` feature
asserting `git diff --diff-filter=A main...HEAD` adds no production `.go` files —
an invariant that structurally breaks for every later feature adding production
code. Its real intent (new code is covered) is enforced by the live 95% coverage
gate. Removal is correct; the file's three sibling scenarios are retained.

## Gates Run — `centinela validate` full run (independently re-run by the orchestrator)

| Gate / command | Result |
|----------------|--------|
| G1: File Size | ✓ all files <100 lines |
| G-Build: Cross-Compile (6 targets) | ✓ |
| roadmap_drift | ✓ ROADMAP.md in sync (regenerated after the `aider-local-provider-wiring` defer) |
| import_graph | ⚠ non-failing (unmapped-package advisory only; zero forbidden edges) |
| spec-traceability | ⚠ non-failing advisory (behavior covered by colocated unit tests) |
| `go test ./...` | ✓ 3127 tests pass, exit 0 |
| `go test ./tests/acceptance/...` | ✓ |
| `./scripts/check-coverage.sh` | ✓ `coverage gate passed: 97.4% >= 95.0%` |
| `./scripts/check-fmt.sh` | ✓ (after `gofmt -w` on 4 signature-touched test files) |

## Synthesis

`All gates passed.` No `✗` failures; only the two standing non-failing
advisories (import_graph unmapped-package, spec-traceability). The acceptance
test is hermetic — the local endpoint URL appears only as TOML config data and an
asserted `baseURL` string, with no network dial or git push. Gatekeeper SAFE,
coverage 97.4% (≥97% target), all three test tiers green.

## Decision

PASS — proceed to the docs step.
