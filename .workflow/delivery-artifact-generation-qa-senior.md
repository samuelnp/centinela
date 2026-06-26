# delivery-artifact-generation — qa-senior

## Test Inventory

| Tier | File | Lines | What it asserts |
|------|------|-------|-----------------|
| unit (colocated) | `internal/delivery/extract_test.go` | 59 | `ExtractSection` found / case-insensitive / last-section / missing / empty inputs; `headingText` rejects `###` and non-headings; `FirstParagraph` single/multi/blank |
| unit (colocated) | `internal/delivery/sections_test.go` | 74 | Each PR-body section present and OMITTED when its datum is empty; `gateStatusSection` with verdict, with a synthetic `*verify.VerificationResult`, and omitted when both absent; verdict priority; provenance footer always present |
| unit (colocated) | `internal/delivery/prbody_test.go` | 44 | Full-body section ordering golden (Summary→What/Why→Acceptance→Gate→footer); provenance-only body when all sources empty |
| unit (colocated) | `internal/delivery/changelog_test.go` | 42 | Category selection feat→Added/fix→Fixed/refactor+chore→Changed; first non-FILL stub line wins; FILL ignored→derive from brief; fallback to feature slug |
| unit (colocated) | `internal/delivery/changelog_insert_test.go` | 87 | First insert (true) + idempotent re-insert (false, unchanged); subsection created in canonical order; new-Fixed-after-Added fall-through; released sections untouched; no `[Unreleased]`→unchanged+false; EOF append |
| unit (cmd) | `cmd/centinela/deliver_artifacts_test.go` | 79 | `buildPRBody` writes a temp file containing composed sections + provenance; `writeChangelog` true→false idempotent with exactly one line; missing CHANGELOG no-op |
| unit (cmd) | `cmd/centinela/deliver_gather_test.go` | 29 | `gatherEvidence` tolerates all-missing sources (no error, empty fields, feature slug set) |
| unit (cmd) | `cmd/centinela/deliver_pr_changelog_test.go` | 65 | `ghCreatePR` receives a NON-EMPTY `--body-file` path; `commitChangelog` commits CHANGELOG.md only when changed (idempotent re-run adds nothing) |
| integration | `tests/integration/delivery_artifact_changelog_test.go` | 61 | `InsertEntry` round-trip vs a realistic multi-section CHANGELOG: one new bullet inside `[Unreleased]`, released section untouched, second insert is a no-op |
| acceptance | `tests/acceptance/delivery_artifact_generation_test.go` | 49 | Binary `deliver alpha --via pr` against a LOCAL bare origin: CHANGELOG gains exactly one `- feat: alpha` line, idempotent on re-run, never falsely prints "Opened pull request" |

All test files are ≤100 lines (G1). The pre-existing `deliver_pr_more_test.go` /
`deliver_pr_test.go` / `deliver_test.go` and the acceptance `cdp*` helpers are
reused unmodified.

## Coverage Gaps

- **`internal/delivery` package: 98.7% statements** (well above the 95% gate).
  Two tiny defensive branches remain uncovered: `newSubsectionAt`'s path for a
  category not in `canonicalOrder`, and one `headingText` ordering branch.
  Neither is reachable through the public API with the constrained
  `Category` values (`Added`/`Changed`/`Fixed`), so they are dead-defensive.
- **TOTAL repository coverage: 95.1% ≥ 95.0%** — `MIN_COVERAGE=95.0
  ./scripts/check-coverage.sh` exits 0. Coverage is per-package (no `-coverpkg`),
  so the colocated `internal/delivery/*_test.go` and `cmd/centinela/*_test.go`
  files are what move the number; the `tests/` tier files do not.
- **Verification tally on the live delivery path is intentionally unwired** (slice
  one leaves `Evidence.Verification == nil`); the tally branch is covered with a
  synthetic `*verify.VerificationResult` in `sections_test.go`.

## Acceptance Wiring

- `validate.commands` already runs `go test ./tests/acceptance/...`, so the new
  `TestAccDeliveryArtifactChangelog` is picked up by `centinela validate` with no
  config change required.
- The acceptance test uses the existing `cdpRepo(t, true)` LOCAL bare origin
  (offline push — no network URL, per the prior hang incident), `cdpWorkflow`,
  `runDeliverBin`, `commitAll`, and `writeFile`. No helpers were redefined.
- Real-binary behavior confirmed manually: the changelog commit lands BEFORE the
  push, so the line persists even when the subsequent push fails (no `alpha`
  branch on a fresh repo) — which is exactly why `gh` is never reached and
  "Opened pull request" is never printed against a bare origin.

## Deferred Findings

- None new. `centinela-changelog-subcommand` (a standalone changelog command and
  merge-path parity) remains deferred from the big-thinker; not in scope here.

## Handoff

- Next role: **validation-specialist**.
- Gate status from this role: `go test ./...` green (exit 0); coverage gate
  95.1% ≥ 95.0% (exit 0); all new/modified test files ≤100 lines.
- For validate: run the gatekeeper + `centinela validate`. Acceptance execution
  is already in `validate.commands`. The composer is pure (no I/O in
  `internal/delivery`); all disk I/O and git/gh orchestration stays in `cmd/`.
