# Edge Cases: docstring-gate

## Covered

- Undocumented exported function in a changed file fails, naming file:line,
  kind and name — `tests/acceptance/docstring_gate_detect_test.go:TestDG_UndocumentedExportFails`
- Fully documented changed file passes, message names the inspected count —
  `tests/acceptance/docstring_gate_detect_test.go:TestDG_DocumentedPasses`
- Warn severity keeps `centinela validate` green while still reporting the
  undocumented identifier (via `docs lint`, since `RenderGateResult` only
  expands Details on Fail) —
  `tests/acceptance/docstring_gate_detect_test.go:TestDG_WarnSeverityDoesNotFail`,
  `tests/acceptance/docstring_gate_cli_test.go:TestDG_DocsLintExitCodes`
- A legacy unchanged file with undocumented exports is never opened when a
  sibling changed file is what's inspected —
  `tests/acceptance/docstring_gate_detect_test.go:TestDG_LegacyUnchangedNeverScanned`
- An empty changed-file scope reports Skip, never a confident Pass —
  `tests/acceptance/docstring_gate_skip_test.go:TestDG_EmptyScopeSkips`
- A changed-file scope with no Go files (e.g. a docs-only branch) reports
  Skip — `tests/acceptance/docstring_gate_skip_test.go:TestDG_NoGoFilesInScopeSkips`
- `_test.go` and `// Code generated ... DO NOT EDIT.` files are excluded;
  when that's the entire scope the gate Skips with empty Details, not a
  false green Pass — `tests/acceptance/docstring_gate_skip_test.go:TestDG_TestAndGeneratedFilesSkip`
- An unresolvable merge base (no baseline branch) reports Skip naming the
  reason, never Pass and never Fail —
  `tests/acceptance/docstring_gate_skip_test.go:TestDG_UnresolvableScopeSkips`
- CI full-scan mode hands every gate a nil filter; the docstring gate
  resolves the same merge-base set itself instead of opening the legacy
  backlog — `tests/acceptance/docstring_gate_ratchet_test.go:TestDG_FullScanStillRatchets`
- A grouped `const ( A; B )` block's single doc comment covers every spec
  inside it — `tests/acceptance/docstring_gate_forms_test.go:TestDG_GroupedConstBlockDocCovers`
- `//centinela:nodoc` exempts an identifier AND the exemption is listed on
  the passing run — the invisibility bug the engineer fixed (`RenderGateResult`
  hides Details on Pass; `docs lint` echoes them via `printDocsLintPassDetails`) —
  `tests/acceptance/docstring_gate_forms_test.go:TestDG_NodocExemptionListedOnPass`
- Undocumented exported struct fields do not fail the gate while
  `check_fields` stays at its default `false` —
  `tests/acceptance/docstring_gate_forms_test.go:TestDG_StructFieldsNotReportedByDefault`
- An unparseable `.go` file is reported by path, never silently dropped from
  the scan — `tests/acceptance/docstring_gate_forms_test.go:TestDG_UnparseableFileReported`
- An unknown `severity` value is rejected at config load, naming the field
  and the accepted values —
  `tests/acceptance/docstring_gate_cli_test.go:TestDG_UnknownSeverityConfigError`
- `centinela docs lint` exits 1 on Fail and 0 on Warn while reporting the
  same identifier both ways —
  `tests/acceptance/docstring_gate_cli_test.go:TestDG_DocsLintExitCodes`
- `centinela docs lint --full` is report-only: always exit 0, prints the
  whole-repo backlog count —
  `tests/acceptance/docstring_gate_cli_test.go:TestDG_DocsLintFullBacklog`
- `centinela docs lint --json` shape (`scope`, `status`, `undocumented`,
  per-violation `details`) —
  `tests/acceptance/docstring_gate_cli_test.go:TestDG_DocsLintJSONShape`
- This repository's own `centinela.toml` ships `[gates.docstring]
  enabled=true severity="fail"` with no scope knob —
  `tests/acceptance/docstring_gate_repo_test.go:TestDG_ShipsFailSeverity`
- `docs/architecture/senior-engineer-prompt.md` states the doc-comment duty
  (byte-identical mirror + 130-line budget already asserted generically by
  `TestScaffoldArchitectureMirrorParity` / `TestPromoteOrchestrationAgents_LineBudget`) —
  `tests/acceptance/docstring_gate_repo_test.go:TestDG_SeniorEngineerPromptStatesDuty`
- Colocated (pre-existing) unit coverage the acceptance tier does not
  re-derive: a scanner that errors mid-scan fails naming the cause, a
  missing scanner registration Skips instead of Pass, and the default
  ratchet delegates to the shared `gitdiff.Resolver` —
  `internal/gates/docstring_degrade_test.go`, `internal/gates/docstring_scope_test.go`

## Residual Risks

- The 171-item whole-repo legacy backlog (`centinela docs lint --full`
  measured at adoption) stays undocumented by design — the ratchet never
  opens those files. Paying it down is the separate deferred feature
  `docstring-full-scan-debt-paydown`; not a gap in this feature's tests.
- `package-doc-comments` (a per-package, not per-file, doc-comment rule) and
  `docstring-struct-field-docs` (the `check_fields` backlog) are both
  pre-existing deferred findings from the plan step, deliberately out of
  v1 scope and unaffected by this test suite.
- Multi-language scanning is a named seam (`docstring.Register`/`For`) with
  only the Go scanner implemented; no test exercises a second language
  because none exists to exercise.
- Hunk-level (line-range) diff scope was rejected at plan time in favor of
  file granularity; there is no test for it because the design explicitly
  does not support it (a `gofmt` reflow or pure rename would defeat it).
- No new coverage gaps were found beyond what the plan and senior-engineer
  steps already tracked in the roadmap Backlog.
