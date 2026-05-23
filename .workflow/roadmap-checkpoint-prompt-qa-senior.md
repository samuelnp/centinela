### QA-Senior Report: roadmap-checkpoint-prompt
**Date:** 2026-05-23

#### Test Inventory
| Tier        | File | Scenarios |
|-------------|------|-----------|
| unit (in-package) | `internal/roadmapcheckpoint/checkpoint_decide_test.go` | `Decide` all 8 outcome branches (no-marker emit, equal-at suppress, after suppress, ROADMAP-newer stale, malformed-JSON stale, unparseable-at stale, workflow-file suppress, no-artifacts suppress); `Decide` no-first suppress; `LatestMtime` max + empty; `parseMarkerAt` 4 cases; `RequiredArtifacts` shape; `FirstIncompleteBootstrap` 6 cases (nil, no-bootstrap, first-non-done, skip-done, in-progress, all-done); `WriteMarker`/`ReadMarker` round-trip + missing + malformed + unparseable; `NewOSFS` Stat/ReadFile/Exists present + missing. |
| unit (consumer-facing) | `tests/unit/roadmap_checkpoint_prompt_unit_test.go` | Public-contract `Decide` 9 outcomes via fake FS (table); no-first suppress; `FirstIncompleteBootstrap` public behavior (nil, first, skip-done, all-done); `WriteMarker`/`ReadMarker` round-trip. |
| integration (cmd/centinela, package main) | `cmd/centinela/roadmap_checkpoint_prompt_test.go` | Drives real `runHookSetup` end-to-end against on-disk artifacts: scenarios 1-12 plus the anti-spam regression that calls the real `runRoadmapIterate` then asserts silence (13 tests). |
| acceptance (built binary exec) | `tests/acceptance/roadmap_checkpoint_prompt_test.go` | Builds and execs `centinela hook setup` (and `roadmap iterate`) against temp fixtures: all 12 spec scenarios + anti-spam regression (13 tests). |

#### Scenario → assertion map (all 12 spec scenarios covered at integration AND acceptance)
1. Happy-path emit (no marker) — `TestCheckpoint_Emit_NoMarker` / `TestAccept_Checkpoint_Emit`
2. Suppress fresh marker — `TestCheckpoint_Suppressed_FreshMarker` / `TestAccept_Checkpoint_SuppressFreshMarker`
3. Stale (ROADMAP.md newer) — `TestCheckpoint_Stale_RoadmapNewer` / `TestAccept_Checkpoint_StaleRoadmap`
4. Stale (analysis artifact newer) — `TestCheckpoint_Stale_AnalysisArtifactNewer` / `TestAccept_Checkpoint_StaleAnalysisArtifact`
5. Suppress bootstrap complete — `TestCheckpoint_Suppressed_BootstrapComplete` / `TestAccept_Checkpoint_SuppressBootstrapComplete`
6. Suppress no Phase 0 — `TestCheckpoint_Suppressed_NoPhaseZero` / `TestAccept_Checkpoint_SuppressNoPhaseZero`
7. Suppress workflow file exists — `TestCheckpoint_Suppressed_WorkflowFileExists` / `TestAccept_Checkpoint_SuppressWorkflowFileExists`
8. Precedence: missing ROADMAP → roadmap-required — `TestCheckpoint_Precedence_MissingRoadmap` / `TestAccept_Checkpoint_PrecedenceMissingRoadmap`
9. Precedence: invalid roadmap.json → roadmap-json — `TestCheckpoint_Precedence_InvalidRoadmapJSON` / `TestAccept_Checkpoint_PrecedenceInvalidRoadmapJSON`
10. Multi-feature picks second — `TestCheckpoint_MultiFeature_PicksSecond` / `TestAccept_Checkpoint_MultiFeaturePicksSecond`
11. Malformed marker re-emits, no crash — `TestCheckpoint_MalformedMarker_ReEmits` / `TestAccept_Checkpoint_MalformedMarkerReEmits`
12. Unparseable `at` → stale re-emit — `TestCheckpoint_UnparseableAt_ReEmits` / `TestAccept_Checkpoint_UnparseableAtReEmits`
Regression (anti-spam): `TestCheckpoint_AntiSpam_IterateThenSilent` / `TestAccept_Checkpoint_AntiSpamIterateThenSilent`

#### Coverage Gaps
- None of the 12 spec scenarios is unasserted; every one has an executable assertion at BOTH integration and acceptance tiers.
- Untriggerable error paths only: `WriteMarker`'s `os.MkdirAll`/`os.WriteFile` failure branches and `ReadMarker`'s non-NotExist `os.ReadFile` error branch are not deterministically reproducible without filesystem-permission injection; left uncovered (package still clears the 95% gate).
- mtime-granularity boundary documented as a deliberate trade-off (see edge-case report §4), not a gap — second-precision marker `at` vs sub-second artifact mtimes. No production change made; surfaced to validation-specialist.

#### Acceptance Wiring
`centinela.toml` `validate.commands` already runs the acceptance tier — no edit needed:
```toml
[validate]
commands = [
  "go test ./...",          # includes tests/acceptance (and tests/unit, tests/integration)
  "./scripts/check-coverage.sh"
]
```
`go test ./...` compiles and runs `github.com/samuelnp/centinela/tests/acceptance`, so the built-binary acceptance suite executes inside `centinela validate`.

#### Regression Guards
- Anti-spam idempotency: after the real `centinela roadmap iterate` writes a fresh marker, a second `hook setup` over unchanged disk stays silent — guards the core "one prompt, then quiet" UX promise at both integration and acceptance tiers.
- Precedence guards (scenarios 8 & 9) assert the checkpoint directive is ABSENT when an earlier setup directive fires — catches any future reordering of the hook chain that would let the checkpoint leak through.
- Coverage gate: the new `internal/roadmapcheckpoint` package shipped in the `code` step with no in-package tests, dropping the aggregate to 92.4%. The in-package unit suite restores the package to ~95-100% per function and the aggregate to 95.1% (gate ≥ 95.0%). Gate closed with real tests, not by lowering the threshold.

#### Handoff
- Next role: validation-specialist
- Edge-case report: `.workflow/roadmap-checkpoint-prompt-edge-cases.md` (produced alongside this report).
- No production code was modified. The mtime-granularity trade-off (§4 of the edge-case report) is the only behavioral nuance flagged for validation review.
