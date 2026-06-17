# QA-Senior Report: right-size-docs-step

**Date:** 2026-06-12
**Handoff →** validation-specialist

Surface-aware docs step. Tests are colocated per-package (no `-coverpkg`), every
file ≤100 lines, and assert observable behavior (role lists, error names,
changelog presence/blank distinction, the merge-regen seam, surface-aware hook
nags, default-internal surface) — no `!= ""` padding.

## Test Inventory

### Unit (colocated, same package as code)
| File | Lines | Asserts |
|------|------|---------|
| internal/orchestration/policy_docs_surface_test.go | 47 | docs role gating: user-facing → includes documentation-specialist; internal → empty; code-step ux-ui unchanged |
| internal/orchestration/feature_surface_default_test.go | 27 | `IsUserFacingFeature` default=false for no-surface brief and for missing brief |
| internal/workflow/validate_docs_internal_test.go | 59 | `validateDocsInternal`: missing names file; blank names "empty" distinctly; one-liner passes with NO KB |
| internal/ui/render_changelog_test.go | 22 | `RenderChangelogNeeded` panel names feature + changelog + artifact command |
| internal/evidence/artifact_changelog_test.go | 34 | `KindChangelog` in KindsAllowed/ParseKind; `RenderTemplate` emits one `<f>-changelog.md` with non-blank first line |
| cmd/centinela/hook_statusline_docs_internal_test.go | 35 | internal docs nag = `MISSING_DOCS_OUTPUT`/`write-changelog`; clears when changelog present |
| cmd/centinela/hook_context_docs_internal_test.go | 56 | internal banner = `RenderChangelogNeeded` (not Documentation); silent when changelog present |
| cmd/centinela/merge_regen_helper_test.go | 48 | seeds a real git repo + committed worktree for a clean merge |
| cmd/centinela/merge_regen_test.go | 44 | **MERGE-REGEN SEAM:** seam invoked once on clean merge (8); regen error → merge still succeeds + `notice:` emitted (9); original restored in defer |

Existing preserved: `validate_docs_test.go` (user-facing KB+portal), `policy_user_facing_test.go` (code ux gating), `evidence/artifact_test.go` loop over KindsAllowed (covers changelog generically), `hook_statusline_rules_test.go` / `hook_context_docs_test.go` (user-facing paths).

### Integration (tests/integration)
| File | Lines | Asserts |
|------|------|---------|
| right_size_docs_step_integration_test.go | 52 | internal feature reaches docs-valid with only a changelog (no roles, no KB); user-facing still needs full bundle (role present, portal demanded) |

### Acceptance (tests/acceptance — `// Acceptance:` + one `// Scenario:` per real func)
| File | Lines | Scenarios |
|------|------|-----------|
| right_size_docs_step_test.go | 49 | 1, 2, 10 |
| right_size_docs_step_userfacing_test.go | 39 | 3, 4 |
| right_size_docs_step_internal_test.go | 46 | 5, 6, 7 |
| right_size_docs_step_merge_test.go | 48 | 8, 9 |

## Coverage Gaps — 10-scenario → test mapping (NONE uncovered)

| # | Scenario | Acceptance test | Behavioral backing |
|---|----------|-----------------|--------------------|
| 1 | user-facing requires documentation-specialist role | TestRDSUserFacingRequiresDocsSpecialist | policy_docs_surface_test |
| 2 | internal does not require documentation-specialist role | TestRDSInternalDropsDocsSpecialist | policy_docs_surface_test |
| 3 | user-facing docs still requires KB guide | TestRDSUserFacingPassesWithKnowledgeBase | validate_docs_test (existing) |
| 4 | user-facing docs fails without KB guide | TestRDSUserFacingFailsWithoutKnowledgeBase | validate_docs_test (existing) |
| 5 | internal passes with only a one-line changelog | TestRDSInternalPassesWithChangelog | validate_docs_internal_test |
| 6 | internal fails without a changelog entry | TestRDSInternalFailsWithoutChangelog | validate_docs_internal_test |
| 7 | internal fails when changelog entry is blank | TestRDSInternalFailsWhenChangelogBlank | validate_docs_internal_test |
| 8 | clean merge regenerates the documentation portal | TestRDSCleanMergeRegeneratesPortal | merge_regen_test (seam invoked once) |
| 9 | portal regen failure does not fail a clean merge | TestRDSPortalRegenFailureDoesNotFailMerge | merge_regen_test (notice + success) |
| 10 | default surface is internal when none declared | TestRDSDefaultSurfaceIsInternal | feature_surface_default_test |

Spec-traceability gate: **All 10 scenarios have acceptance coverage** (dogfood proof below).

## Acceptance Wiring

- Each acceptance file carries `// Acceptance: specs/right-size-docs-step.feature`.
- Each `// Scenario:` line copies the exact spec text and sits directly above a
  real test func (no orphans); the gate normalizer (trim/collapse/strip-period/
  lowercase) matches them.
- Scenarios 1–7, 10 drive real production functions (`RequiredRolesForFeature`,
  `IsUserFacingFeature`, `workflow.ValidateArtifacts`→`validateDocsOutput`).
  Scenarios 8/9 assert the merge.go wiring contract at the acceptance tier (the
  `docsPortalRegen` seam is unexported package `main`); the deep behavioral test
  runs a real git merge in `cmd/centinela/merge_regen_test.go`.

## Verification

- `gofmt -l cmd internal tests` → clean
- `go vet ./...` → no issues
- `go test ./...` → all packages ok (acceptance + integration + 5 touched units)
- `./scripts/check-coverage.sh` → **95.4% ≥ 95.0%**; per-package: orchestration
  97.3, workflow 97.7, evidence 95.7, ui 97.3, cmd 93.6 (cmd pre-existing <95,
  improved from 93.0 baseline). New code lines all 100% (validateDocsInternal,
  RenderChangelogNeeded, runHookContext internal branch).
- prompt-doc ↔ scaffold-mirror diff → empty (parity held).
- Dogfood: `cent validate` spec-traceability → all 10 scenarios COVERED.
- All 14 new test files ≤100 lines (max 59).

## Handoff → validation-specialist

Run the gatekeeper + full `centinela validate`. No blocking gaps. cmd/centinela
sits at 93.6% (pre-existing, raised by this work, not regressed); the global
coverage gate passes at 95.4%. Merge-regen behavior is verified end-to-end via a
real git merge with the seam swapped.
