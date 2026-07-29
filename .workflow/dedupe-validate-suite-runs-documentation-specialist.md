# dedupe-validate-suite-runs — documentation-specialist

## KB Pages

None — the feature brief declares `surface: internal`, so per the
documentation-generator prompt no `docs/project-docs/kb/<feature>.md` guide is
written. `docs/project-docs/kb/index.html` was refreshed as a side effect of
the portal regeneration.

## project-docs Entries

- `docs/project-docs/index.html` regenerated via
  `centinela docs generate --out docs/project-docs/index.html` (exit 0, after
  `centinela docs validate` passed). The portal now lists
  `dedupe-validate-suite-runs` in the workflow status matrix and spec index.
- `.workflow/dedupe-validate-suite-runs-changelog.md` — concise changelog
  entry: single-run validate design (one profiled suite run serves tests +
  coverage), complete-gate `PriorTestRun` reuse behind a post-run
  `VerificationFresh` re-check, gatekeeper/qa-senior prompt mandates updated
  in both scaffold-mirrored copies, CI bare-suite step dropped; cycle cost
  ~7 → 2 full-suite runs.
- `ROADMAP.md` regenerated via `centinela roadmap generate` — clears the
  `roadmap_drift` warning the gatekeeper flagged by publishing the two
  Backlog deferrals recorded during this workflow:
  `cross-process-suite-result-reuse` (planner) and
  `validate-flake-diagnosability` (gatekeeper).

## Outcome

Docs step complete for an internal-surface feature. No user-facing guide
(README / docs/guides/) required edits: grep for the old two-command flow
(`check-coverage`, `go test ./...`, `COVERAGE_PROFILE`, `coverprofile`)
shows the guides only carry generic scaffold-style examples, which this
feature deliberately leaves unchanged (declared non-goal: scaffolded
projects keep the self-contained `check-coverage.sh` behavior and the
gatekeeper's bare-suite obligation). Gatekeeper verdict was WARNING with
three non-blocking findings: the transient suite flake is tracked as the
`validate-flake-diagnosability` deferral (now in ROADMAP.md), the
roadmap drift is fixed here, and the coverage-claim caveat (a gatekeeper
`coverage` claim re-triggers a suite run at complete time that
`PriorTestRun` does not suppress) is recorded as an edge case in the
evidence JSON for the follow-up.
