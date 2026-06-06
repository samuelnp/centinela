### Senior-Engineer Report: security-gate
**Date:** 2026-06-06

Implemented the opt-in `[gates.security]` gate per the plan's four rollout
slices: config leaf + Skip-on-absent skeleton, gitleaks secrets (hard `Fail`),
govulncheck + osv-scanner vuln audit (`Warn`), and the orchestrator wiring.
This is the `code` step — implementation only; tests are qa-senior's next step.
`go build ./...`, `go vet ./internal/... ./cmd/...`, `gofmt`, and the existing
config+gates suites (128 tests) all pass.

#### Files Touched
| Path | Reason | Lines |
|------|--------|-------|
| internal/config/security_gate.go (new) | `SecurityGateConfig`/`SecretsConfig`/`VulnConfig`, `NormalizeSecurityGate`, `validateSecurityGate` | 76 |
| internal/config/config.go (edit) | add `Security SecurityGateConfig` to `GatesConfig` | +1 |
| internal/config/defaults.go (edit) | `cfg.Gates.Security = NormalizeSecurityGate(...)` | +1 |
| internal/config/file_size_exceptions.go (edit) | wire `validateSecurityGate` into `validateConfig` | +3 |
| internal/gates/security.go (new) | `checkSecurity(cfg, filter) []Result` orchestrator + "no scanners" signal | 22 |
| internal/gates/security_exec.go (new) | `toolPresent` (LookPath), `runScanner` (120s ctx timeout), `exitCode` | 54 |
| internal/gates/security_secrets.go (new) | gitleaks invoke + report-file + Result mapping | 82 |
| internal/gates/security_secrets_parse.go (new) | gitleaks JSON parse + allowlist + diff filter | 74 |
| internal/gates/security_vuln.go (new) | vuln orchestration, dedup, fold to one `G-Vuln` Result | 98 |
| internal/gates/security_vuln_parse.go (new) | govulncheck NDJSON + osv-scanner JSON defensive parsers | 92 |
| internal/gates/gates.go (edit) | `if cfg.Gates.Security.Enabled { append(..., checkSecurity(cfg, filter)...) }` | +4 |

#### Locked Decisions — how each is implemented
- **(1) secrets + vuln both** — `checkSecurity` returns `[]Result{G-Secrets, G-Vuln}`.
- **(2) gitleaks / govulncheck / osv-scanner** — `Secrets.Tool` (gitleaks) and
  `Vuln.Tools` ([govulncheck, osv-scanner]); `vulnArgs` builds each tool's JSON argv.
- **(3) secrets Fail / vuln Warn** — `classifySecrets` returns `Fail` on retained
  findings; `foldVuln` returns `Warn` on any finding. `AllPassed` (Fail-only) is unchanged.
- **(4) missing tool = Skip, never Fail/crash** — `toolPresent` (exec.LookPath) gates
  every scan; absent secrets -> `Skip` naming the tool; all vuln tools absent -> `Skip`.
- **(5) diff-aware secrets / whole-project vuln** — `retainFindings` drops findings where
  `filter != nil && !filter.Contains(file)`; `checkVuln` ignores the filter entirely.
- **(6) off by default** — `Enabled` zero value false; `RunWithFilter` only appends when
  enabled; `NormalizeSecurityGate`/`validateSecurityGate` are no-ops when disabled.

#### Architecture Compliance
- internal/config stays a leaf: security_gate.go imports only `fmt`/`strings`.
- internal/gates imports only `internal/config` + `internal/gitdiff` (matches build/import_graph).
- cmd/ untouched — it already renders `[]Result`.
- G1: every source file ≤100 lines (max 98). Secrets and vuln each split into
  invoke + parse siblings to stay under the cap.
- G7: no business logic added to the outer layer.

#### Type-Safety Notes
- No `any`/`interface{}`; tool JSON decoded into named structs with explicit `json` tags.
- `Status` is the existing typed enum; `vulnKey struct{Pkg,ID string}` is the dedup key.
- Defensive parsing: empty output -> no findings; non-empty unparseable -> `Warn` (never false `Pass`); launch failure (exit -1) -> `Warn`.

#### Trade-Offs
- gitleaks output via `--report-path <tmpfile>` (not stdout) for a stable JSON array;
  empty/missing file = clean. Rejected stdout capture (banner/log interleave risk).
- Diff scoping done by filtering gitleaks findings post-scan via `filter.Contains` rather
  than feeding a file list to gitleaks — simpler, and the Set exposes no path slice.
- Fixed 120s internal timeout (not yet configurable) per Resolved Q5; a `timeout` field is a clean follow-up.

#### Handoff
- Next role: qa-senior — write unit/integration/acceptance tests for the 14 scenarios.
- Outstanding: tests must keep each `_test.go` ≤100 lines; cover off-by-default (0 results),
  Skip-on-absent, allowlist glob vs rule-ID, vuln dedup, malformed-JSON -> Warn.
