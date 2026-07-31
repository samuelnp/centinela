# docstring-gate — qa-senior

**Date:** 2026-07-31

## Test Inventory

Colocated unit/integration coverage already existed from the code step
(`internal/docstring/*_test.go`, `internal/config/docstring_test.go`,
`internal/gates/docstring{,_scope,_degrade}_test.go`,
`cmd/centinela/docs_lint{,_full}_test.go`) — none of it drives the real
binary. This step adds the missing binary-driven acceptance tier, one test
per Gherkin scenario plus one extra (`--json` shape).

| Tier | File | Scenarios |
|------|------|-----------|
| acceptance | `tests/acceptance/docstring_gate_helper_test.go` (51 lines) | shared fixtures — no tests |
| acceptance | `tests/acceptance/docstring_gate_detect_test.go` (60 lines) | Undocumented export fails; Documented passes; Warn doesn't fail; Legacy unchanged never scanned |
| acceptance | `tests/acceptance/docstring_gate_skip_test.go` (61 lines) | Empty scope Skips; No-Go-files scope Skips; Test/generated-only scope Skips; Unresolvable merge base Skips |
| acceptance | `tests/acceptance/docstring_gate_forms_test.go` (62 lines) | Grouped const block; nodoc exemption listed on Pass; struct fields excluded by default; unparseable file reported |
| acceptance | `tests/acceptance/docstring_gate_ratchet_test.go` (24 lines) | CI full-scan mode still ratchets to changed files |
| acceptance | `tests/acceptance/docstring_gate_cli_test.go` (77 lines) | Unknown severity config error; docs lint exit codes; docs lint --full backlog; docs lint --json shape |
| acceptance | `tests/acceptance/docstring_gate_repo_test.go` (47 lines) | This repo ships fail severity; senior-engineer prompt states the duty |

All acceptance files: build the real binary via `buildCentinelaBinary`
(`go build ./cmd/centinela` into `t.TempDir()`), drive it against local git
repos only (`git init` in a temp dir, no network origin), and reuse the
shared helpers already in the package (`runValidate`, `runValidateExpectFail`,
`runCent`, `mustContain`, `mustNotContain`, `commit`, `runGit`, `mustWrite`,
`repoRoot`, `mustHave`) rather than redefining them.

Pre-existing colocated unit tests exercising failure modes the acceptance
tier deliberately does not re-derive: `TestCheckDocstring_ScannerErrorFailsAndNamesTheCause`,
`TestCheckDocstring_MissingScannerSkipsInsteadOfPassing`,
`TestDocstringRatchet_DefaultDelegatesToTheSharedResolver`.

## Coverage Gaps

None. All 18 `Scenario:` lines in `specs/docstring-gate.feature` have an
exact-text-matching `// Scenario:` acceptance test (verified by diffing the
parsed scenario names against the test comments — the spec-traceability
gate's own matching logic, byte for byte). One extra behavior
(`docs lint --json` shape) was added beyond the Gherkin scenarios per the
orchestrator's explicit ask; it carries no `// Scenario:` tag since it does
not correspond 1:1 to a spec line — it extends "docs lint exits 1 on Fail
and 0 on Warn".

During authoring, three `// Scenario:` comments were initially wrapped
across multiple `//` lines (readable, but the spec-traceability matcher
only captures a single line per `^//\s*Scenario:\s*(.+)$`). That produced
real "Scenarios without acceptance coverage" gate warnings on the first
`centinela validate` run against a freshly built worktree binary. Fixed by
keeping the full scenario name on one comment line and moving prose
explanation to unprefixed lines below.

## Acceptance Wiring

`centinela.toml`'s single-run design already covers this: the profiled
`go test ./...` invocation in `validate.commands` is `./...`-scoped, which
includes `tests/acceptance`, so no separate acceptance-only command was
needed or added:

```toml
[validate]
commands = ["go build ./cmd/centinela", "go test ./... -coverprofile=coverage.out", ...]
```

`validate.commands` was not modified (out of scope for this step per the
orchestrator's instruction).

## Deferred Findings

None new. The pre-existing deferred findings from the plan and
senior-engineer steps (`docstring-full-scan-debt-paydown`,
`package-doc-comments`, `docstring-struct-field-docs`) already cover the
residual risk surface this test suite touches; no additional gap was found
during QA.

## Handoff

- Next role: validation-specialist
- Edge-case report: `.workflow/docstring-gate-edge-cases.md` (this step)
