### QA-Senior Report: security-gate
**Date:** 2026-06-06

The security gate ships two `gates.Result` entries — `G-Secrets: Secret Scan`
(gitleaks, diff-aware locally, hard Fail) and `G-Vuln: Dependency Audit`
(govulncheck + osv-scanner, whole-project, Warn-only) — wired into
`gates.RunWithFilter` behind `cfg.Gates.Security.Enabled`. The tests verify the
spec semantics through both the colocated unit tier and a new acceptance tier
that drives the real gate via the public `RunWithFilter` surface with fake
scanner binaries on PATH.

#### Test Inventory

| Tier        | File | Scenarios |
|-------------|------|-----------|
| unit        | internal/gates/security_classify_test.go | secrets Fail/Pass/Warn folding; launch-failure (incl. timeout sentinel) never Pass |
| unit        | internal/gates/security_fake_bin_test.go | fake gitleaks finding→Fail, clean→Pass, malformed→Warn; vuln arg shape |
| unit        | internal/gates/security_secrets_parse_test.go | gitleaks report parse (empty/missing/malformed/happy); allowlist by rule-id and path glob; diff-filter in/out |
| unit        | internal/gates/security_vuln_parse_test.go | govulncheck + osv-scanner parse (empty/happy/malformed); dispatch; sorted details |
| unit        | internal/gates/security_fold_test.go | vuln Warn/Pass folding; dedup by (pkg,id); AllPassed Warn-vs-Fail |
| unit        | internal/gates/security_retain_test.go | retain/allowlist/dedup/detail-format; unmatched allowlist ignored |
| unit        | internal/gates/security_skip_test.go | absent-tool Skip; both-absent two-Skips+no-scanners note; disabled emits nothing |
| unit        | internal/gates/security_exec_test.go | toolPresent; runScanner stream capture; exitCode |
| unit        | internal/config/security_gate_test.go | normalize defaults/trim; validate unknown tool names |
| integration | (covered by colocated `RunWithFilter` assertions in security_skip_test.go) | enabled→appends two results; disabled→none |
| acceptance  | tests/acceptance/security_gate_test.go | AC1 secret→Fail+AllPassed false; AC2 clean→Pass; AC4 absent→Skip; AC5 disabled→no results |
| acceptance  | tests/acceptance/security_gate_more_test.go | AC3 vuln→Warn+AllPassed true; AC6 allowlisted→Pass; AC7 diff-aware out/in; dup-CVE dedup |
| acceptance  | tests/acceptance/security_gate_helpers_test.go | shared PATH/fake-bin/config/result helpers |

Bug fixed this step: the fake gitleaks scripts wrote their JSON report to `$5`,
but the real invocation makes the report-path the 6th positional arg
(`detect --no-banner --report-format json --report-path <path>`). Corrected to
`$6`; the two previously-failing fake-bin tests now pass.

#### Coverage Gaps

- No `.feature` scenario lacks an executable assertion. AC1–AC7 and every
  documented edge case map to a test (see `.workflow/security-gate-edge-cases.md`).
- The live 120s `context.WithTimeout` deadline inside `runScanner` is not driven
  by a unit test (would require a long-running fake binary); the timeout→Warn
  *mapping* is asserted via the `errScanTimeout` sentinel. Documented as a
  residual risk, not a spec gap.
- Acceptance fakes emit fixed JSON, so they validate parsing/severity wiring, not
  real tool output-format fidelity (mitigated by defensive parsers + pinned
  versions in the plan).

#### Acceptance Wiring

`centinela.toml` `[validate].commands` already executes the acceptance tier:

```toml
[validate]
commands = [
  "go test ./...",
  "go test ./tests/acceptance/...",
  "./scripts/check-coverage.sh"
]
```

`go test ./tests/acceptance/...` runs the new `security_gate*_test.go` files, so
the acceptance assertions run on every `centinela validate`.

#### Handoff

- Next role: validation-specialist
- Edge-case report: `.workflow/security-gate-edge-cases.md` (each plan edge case
  mapped to its asserting test; residual risks noted)
- All unit + acceptance tests green; coverage gate passes (95.1% ≥ 95.0%); every
  security `_test.go` file is ≤100 lines.
