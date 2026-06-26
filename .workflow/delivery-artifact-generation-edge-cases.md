# Edge Cases: delivery-artifact-generation

Hard paths exercised by the test suite for the read-only delivery-artifact
composer (`internal/delivery`) and its thin cmd orchestrator.

## Covered

| # | Edge case | Expected behavior | Covered by |
|---|-----------|-------------------|------------|
| 1 | Missing evidence source | The dependent PR-body section is OMITTED (never fabricated); other sections still render. | `internal/delivery/sections_test.go` (each section's empty-source branch), `prbody_test.go::TestComposePRBodyOnlyProvenance` |
| 2 | Gate status never faked | When neither a gatekeeper verdict nor a `*verify.VerificationResult` is present, the Gate status section is omitted entirely — no passing gate is asserted. | `sections_test.go::TestGateStatusSection`, spec scenario "gate status line is never faked" |
| 3 | Changelog idempotency | A second `InsertEntry`/`writeChangelog`/`deliver` with the same normalized bullet returns `false` and leaves exactly one copy. | `changelog_insert_test.go::TestInsertEntryFirstThenIdempotent`, `deliver_artifacts_test.go::TestWriteChangelogIdempotent`, integration + acceptance round-trip |
| 4 | Missing CHANGELOG.md | `writeChangelog` is a graceful no-op (`false, nil`) — not an error — when the repo has no `CHANGELOG.md`. | `deliver_artifacts_test.go::TestWriteChangelogMissingFileNoOp` |
| 5 | gh absent (honest failure) | Branch is still pushed, manual instructions print, exit is non-zero, and "Opened pull request" is NEVER printed. | `deliver_pr_more_test.go::TestRunDeliverPRGhAbsent`, acceptance never-false-PR assertions |
| 6 | No origin remote | PR delivery is refused before any push; `writeChangelog` never runs against the network. | `deliver_pr_test.go::TestRunDeliverPRNoOrigin`, `deliver_test.go::TestRunDeliverPRWithoutOrigin` |
| 7 | Category boundary (feat/fix/other) | `feat:`→Added, `fix:`/`bug`→Fixed, `refactor:`/`chore:`/anything else→Changed (default). | `changelog_test.go::TestComposeChangelogCategoryFromStub` |
| 8 | Released sections untouched | Insertion is scoped to the `## [Unreleased]` block; lines below `---` or under a released `## [x.y.z]` heading are never modified. | `changelog_insert_test.go::TestInsertEntryReleasedSectionsUntouched`, integration round-trip |
| 9 | FILL-slot stub ignored | A stub whose first usable line still contains `FILL` is skipped; the seed derives from the brief, or the feature slug when the brief is also empty. | `changelog_test.go::TestComposeChangelogFirstNonFillLine`/`DeriveFromBrief`/`FallbackToSlug` |
| 10 | No `## [Unreleased]` block | `InsertEntry` returns the text unchanged + `false` rather than inventing a block. | `changelog_insert_test.go::TestInsertEntryNoUnreleasedBlock` |
| 11 | Subsection created in canonical order | A new `### <Category>` is inserted respecting Added → Changed → Fixed ordering. | `changelog_insert_test.go::TestInsertEntryCreatesSubsectionInOrder`/`NewFixedAfterAdded` |
| 12 | All sources absent | `gatherEvidence` tolerates a feature with no brief/plan/gatekeeper/spec — returns empty evidence, no error; the PR body is just the provenance footer. | `deliver_gather_test.go::TestGatherEvidenceToleratesMissingSources`, `prbody_test.go::TestComposePRBodyOnlyProvenance` |

## Residual Risks

- **Verification tally not wired on the delivery path (intentional).** The first
  slice leaves `Evidence.Verification == nil`, so the gate-status line is driven
  by the static gatekeeper verdict only. Re-running the suite at delivery is
  slow/fragile; the brief explicitly permits omission. `gateStatusSection`'s
  tally branch is unit-covered with a synthetic `*verify.VerificationResult`.
- **`gh pr create` is exercised only via the seam.** The acceptance test drives a
  local bare origin (no network), where push to a non-existent feature branch
  fails before `gh` is reached — confirming the honest-failure contract but not a
  real PR open. The real `--body-file` wiring is asserted in
  `deliver_pr_changelog_test.go::TestRunDeliverPRPassesBodyFile`.
