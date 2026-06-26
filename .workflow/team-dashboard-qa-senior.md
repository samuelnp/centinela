# team-dashboard — qa-senior

## Summary

Full test pyramid for `centinela dashboard` (the `internal/teamdashboard` pure
aggregator, the `cmd/centinela` orchestrator + git-owner seam, and the
`internal/ui` renderer). All tiers green; the 95% coverage gate passes at
**95.1% TOTAL** with **internal/teamdashboard at 100.0%** of statements. Every
`_test.go` is ≤100 lines (G1 holds for tests). No production code changed.

## Test Inventory

| Tier | File | Lines | Covers |
|------|------|-------|--------|
| Unit (colocated) | `internal/teamdashboard/features_test.go` | 61 | FeatureRow fields, step index/total, age, profile/archetype/worktree passthrough, owner; nil-skip + order |
| Unit | `internal/teamdashboard/features_age_test.go` | 60 | `ageDays` (zero/future/now/floor), `doneCount` (done/position/unknown), `ownerOf` fallback |
| Unit | `internal/teamdashboard/burndown_test.go` | 62 | nil → Present:false; empty → 0/0; Backlog/Baseline exclusion + schedulable totals |
| Unit | `internal/teamdashboard/burndown_done_test.go` | 36 | done tally from on-disk workflow status (PhaseStatus.Done++ branch) |
| Unit | `internal/teamdashboard/gatehealth_test.go` | 58 | rank + insights.Gates parity, non-gate-failure exclusion, empty, `<none>` bucket |
| Unit | `internal/teamdashboard/compute_test.go` | 58 | three-panel assembly, empty-input empty state, determinism |
| Unit (cmd) | `cmd/centinela/dashboard_test.go` | 85 | `runDashboard` happy panels, empty states, `--json` shape (seam-stubbed, chdir) |
| Unit (cmd) | `cmd/centinela/dashboard_owner_test.go` | 42 | `gitOwner` default "unknown" on bogus branch; `dashboardOwners` map + nil-skip |
| Unit (ui) | `internal/ui/render_dashboard_test.go` | 48 | populated three panels + feature row + burn-down line; ANSI-free |
| Unit (ui) | `internal/ui/render_dashboard_defaults_test.go` | 37 | blank → default/canonical/—; present-empty roadmap "0/0 done" |
| Integration | `tests/integration/team_dashboard_test.go` | 77 | Compute → MarshalIndent round-trip + render stability |
| Acceptance | `tests/acceptance/team_dashboard_test.go` | 70 | built binary: three panels, all-empty states, `--json` keys (local, no network) |

## Coverage Gaps

None blocking. `internal/teamdashboard` is at 100.0% statements via the colocated
unit tests (the only tier the per-package gate counts — no `-coverpkg`). The cmd
seam (`dashboard.go`, `dashboard_owner.go`) and the renderer
(`render_dashboard*.go`) are exercised by their colocated package tests, which is
required since the `tests/` tier does not move per-package coverage for those
packages. The acceptance owner column resolves to "unknown" (non-git temp dir) by
design — real branch-name resolution is covered at the seam/unit level to keep the
suite offline.

## Acceptance Wiring

`centinela.toml` `validate.commands` already includes `go test ./tests/acceptance/...`,
which runs the new `team_dashboard_test.go`. Acceptance is strictly local: a temp
dir seeded with `.workflow/<f>.json` + `roadmap.json`, no git push, no network —
avoiding the known acceptance-hang failure mode. Scenario titles map to
`specs/team-dashboard.feature` (three-panel happy path, all-sources-missing empty
states, `--json` top-level keys).

## Deferred Findings

None.

## Handoff

Handoff to **validation-specialist**. Full suite green (`go test ./...` exit 0),
coverage gate passes (TOTAL 95.1% ≥ 95.0%, internal/teamdashboard 100.0%), all
`_test.go` ≤100 lines, `.workflow/team-dashboard-edge-cases.md` written, evidence
`status=done` and `centinela evidence validate team-dashboard` → "evidence ok".
Validation should run the gatekeeper (G1 file-size on the new test files,
import-graph for the aggregator edges) and `centinela validate` end-to-end.
