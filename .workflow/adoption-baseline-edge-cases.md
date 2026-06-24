# Edge Cases: adoption-baseline

## Covered

- **Skip-if-exists leaves the baseline byte-unchanged (text mode)** — a second `centinela adopt`
  without `--force` exits non-zero with "baseline already exists … use --force" and never rewrites
  the file. Tests: `cmd/centinela/adopt_test.go:TestRunAdoptSkip`,
  `tests/acceptance/adoption_baseline_edge_test.go:TestAccAdoptSkipIfExists`,
  `internal/audit/adopt_test.go:TestAdoptSkipsWhenExists`.
- **Skip-if-exists in `--json` mode** — emits `{adopted:false, skipped:true, per_gate:{}}`, exits
  non-zero, file byte-unchanged. Tests: `cmd/centinela/adopt_json_test.go:TestRunAdoptJSONSkip`,
  `tests/acceptance/adoption_baseline_edge_test.go:TestAccAdoptJSONSkip`.
- **`--force` overwrites and widens** — re-records the current violations including a newly added
  one. Tests: `internal/audit/adopt_test.go:TestAdoptForceOverwrites`,
  `tests/acceptance/adoption_baseline_edge_test.go:TestAccAdoptForce`.
- **Clean repo → zero-finding baseline** — adopt writes a baseline with 0 fingerprints and the
  report says "0 accepted findings — nothing to ratchet." Tests:
  `internal/audit/adopt_more_test.go:TestAdoptCleanRepoZeroFindings`,
  `internal/ui/render_adopt_test.go:TestRenderAdoptionZeroFindings`,
  `tests/acceptance/adoption_baseline_edge_test.go:TestAccAdoptCleanRepo`.
- **Byte-identical to `audit.Record` + `audit.Save`** — adopt adds semantics, not different data.
  Tests: `internal/audit/adopt_more_test.go:TestAdoptByteIdenticalToRecordSave` (literal Record+Save
  identity) and `tests/acceptance/adoption_baseline_test.go:TestAccAdoptDeterministic` (binary-level
  determinism over identical repos).
- **Post-adoption ratchet is clean** — a fresh `centinela audit` over the unchanged repo reports
  "0 new", so day-one validate is not drowned. Tests:
  `tests/integration/adoption_baseline_integration_test.go:TestAdoptThenRatchetClean`,
  `tests/acceptance/adoption_baseline_test.go:TestAccAdoptThenAuditClean`.
- **`--json` adopt verdict shape** — `{adopted:true, skipped:false, path, total, per_gate}`, and the
  human report prose is suppressed. Tests: `cmd/centinela/adopt_json_test.go:TestRunAdoptJSONAdopted`,
  `tests/acceptance/adoption_baseline_test.go:TestAccAdoptJSON`.
- **Load error propagation** — a baseline path that is a directory makes `Load` fail (not a
  missing-file); `Adopt` surfaces the error without writing. Test:
  `internal/audit/adopt_more_test.go:TestAdoptLoadErrorPropagates`.
- **Config-load failure** — surfaced as a non-zero exit. Test:
  `cmd/centinela/adopt_test.go:TestRunAdoptConfigError`.

## Residual Risks

- **Adoption records a point-in-time snapshot.** Violations introduced *after* adoption are correctly
  flagged as new by the ratchet (`centinela audit`); adopt does not and should not retro-accept them.
  This is by design — the ongoing ratchet (`centinela audit baseline`) owns re-baselining.
- **`adopt` accepts the participating-gate set as configured.** If `[gates.audit_baseline]` is
  misconfigured before adoption (e.g. a gate disabled), the baseline reflects that config. Mitigation:
  adoption is a deliberate, visible act with a per-gate report, so the adopter sees exactly which
  gates contributed and can re-run with `--force` after fixing config.
