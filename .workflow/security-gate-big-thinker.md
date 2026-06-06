### Big-Thinker Report: security-gate
**Date:** 2026-06-06

#### Problem
Centinela's promise is *enforced, not requested* correctness, yet two high-impact
agent failure modes are caught today only by human/subagent review: leaked secrets
(a committed API key/token/private key) and vulnerable dependencies (a dependency
carrying a known CVE). Any team running Centinela on a repo an agent writes to is
exposed — the secret leak is a security incident, the vulnerable dep is latent risk.
This feature adds an opt-in `[gates.security]` gate to `centinela validate` that moves
both from audit-time to validate-time, reusing the existing gate framework
(`gates.Result{Status}`, `RunWithFilter`, `AllPassed`) that `g2-import-graph-gate` and
the build gate already established.

#### Scope
- **In:** `[gates.security]` config (`SecurityGateConfig{Enabled, Secrets, Vuln}` + shape
  validation); secret scan via `gitleaks` (diff-aware locally, full-scan in CI) → hard
  `Fail`; dependency-vuln audit via `govulncheck` + `osv-scanner` (whole-project) → `Warn`;
  `exec.LookPath` presence detection → `Skip` on absent tool; defensive JSON parsing
  (`gitleaks --report-format json`, `govulncheck -json`, `osv-scanner --format json`);
  vuln de-dup by `(package, vuln-id)`; secrets allowlist; two separate `Result` entries
  (`G-Secrets`, `G-Vuln`); off by default (`enabled=false`).
- **Out:** full git-history secret scan + remediation; bundling/installing the scanner
  binaries; license/SBOM/container scanning; auto-fixing vulns or dependency-bump PRs;
  maintaining any vuln DB; **user-configurable scan timeout (deferred — fixed internal
  default only in v1).**

#### Dependencies & Assumptions
- Internal: `internal/config` (gate config + `applyDefaults`/`validateConfig`),
  `internal/gates` (`RunWithFilter`, `Result`/`Status`, filter), `internal/gitdiff`
  (`Set.Contains` for diff scope); `cmd/centinela/validate.go` already renders `[]Result`
  — no change required there.
- Reference shapes: build gate (external-command argv via `strings.Fields`, direct exec,
  no `sh -c`); import-graph gate (whole-project, deliberately ignores the diff filter).
- External (NOT bundled): `gitleaks`, `govulncheck`, `osv-scanner`, invoked via `os/exec`,
  presence via `exec.LookPath`.
- `AllPassed` already fails only on `Fail`, so `Warn` vuln results do not block validate —
  no change to `AllPassed` needed.
- Tool JSON schemas are stable enough to parse defensively; Centinela validates *shape*,
  not the vuln database; tested tool versions pinned in docs.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| External tool not installed → validate breaks | High | High | `exec.LookPath` → `Skip` + warning; never `Fail` (AC#4). |
| Secret-scan false positives → users disable gate | High | Medium | Allowlist (rule ID / path glob); AC#6; document tuning. |
| Vuln noise (transitive/no-fix) blocks work | Medium | High | Vulns are `Warn`, not `Fail` (decision #3). |
| Tool JSON format drift across versions | Medium | Medium | Defensive parse; malformed → `Warn`/`Skip`, never false `Pass`. Pin versions. |
| Slow/hung scan wedges validate | Medium | Low | Fixed-default `context` timeout per exec → `Warn`; configurable value deferred. |
| New gate file > 100 lines | Medium | Medium | 4-file split: `security.go`/`_secrets.go`/`_vuln.go`/`_exec.go`; `_test.go` ≤100. |
| Secret committed before gate enabled | Medium | Medium | v1 scans working tree/diff only; full-history out of scope (follow-up). |
| Diff-scoped local scan misses secret in unchanged file | Medium | Medium | Matches G1/G11 (AC#7); CI full-scan is the backstop. |

#### Rollout
- Step 1: Config + gate skeleton + Skip-on-absent — `SecurityGateConfig` + normalize +
  defaults + `validateConfig` shape check; `security.go` orchestrator + `security_exec.go`
  presence helper; wire into `RunWithFilter` behind `Enabled`. Proves off-by-default emits
  nothing (AC#5), enabled-but-absent → `Skip` not `Fail` (AC#4), never crashes.
- Step 2: Secrets / gitleaks hard-`Fail` — invoke + JSON parse + allowlist + diff-aware
  filter (AC#1, #2, #6; malformed → `Warn`). Highest-impact half, ships standalone.
- Step 3: Vuln / govulncheck + osv-scanner `Warn` — whole-project invoke + parse + `(pkg,id)`
  de-dup (AC#3; no manifest → per-tool `Skip`; same CVE twice → one line). Stacks on top.
- Step 4: Docs / spec / parity — `specs/security-gate.feature` (AC#1–7); config + tuning
  docs (allowlist, pinned versions, CI vs local); parity test + scaffold-asset mirror if a
  config-leaf string set must stay in sync.

#### Handoff
- Next role: feature-specialist
- Outstanding questions:
  - Confirm `checkSecurity(cfg, filter) []Result` signature and `append(..., checkSecurity(...)...)`
    wiring (returns a slice, unlike single-Result gates).
  - gitleaks output transport: `--report-path <tmpfile>` vs stdout capture (pick most
    stable JSON across pinned versions).
  - Exact JSON field paths per tool (gitleaks rule ID/file; govulncheck OSV id/module;
    osv-scanner package/vuln id) against pinned versions.
  - Wording of the combined "both scanners absent" signal (distinct from a clean pass).
  - Fixed internal timeout value (proposed 120s) and that timeout → `Warn`.
  - Allowlist matching semantics: rule-ID equality vs path glob (`filepath.Match`).
