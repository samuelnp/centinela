### Feature-Specialist Report: security-gate
**Date:** 2026-06-06

#### Behavior Summary

When `[gates.security] enabled = true` is present in `centinela.toml`, `centinela validate` runs two independent security checks — `G-Secrets` (secret scanning via `gitleaks`) and `G-Vuln` (dependency-vulnerability auditing via `govulncheck` and/or `osv-scanner`) — and appends their `gates.Result` entries to the standard gate result slice via `checkSecurity(cfg, filter) []Result` wired into `RunWithFilter`. A secret finding is a hard `Fail` that sets `AllPassed = false` and blocks validate; a vulnerability finding is a `Warn` that surfaces in output but leaves `AllPassed = true`. If a configured scanner binary is not installed, its check returns `Skip` with a message naming the missing tool — validate never crashes and never silently passes. Malformed or unparseable tool output maps to `Warn` (scan ran, result unusable) or `Skip` (nothing to scan), never a false `Pass`. The gate is off by default (`enabled = false` or absent) so existing projects are unaffected until they opt in. Secret scanning is diff-aware locally (honoring the existing `gitdiff.Set` filter, mirroring G1/G11) and full-scan in CI (nil filter); vulnerability auditing is always whole-project (ignoring the filter, mirroring the build gate).

#### Gherkin Scenarios

All scenarios live in `specs/security-gate.feature`. The table below maps each Acceptance Criterion to its primary scenario(s):

| AC | Scenario name | Expected outcome |
|----|--------------|-----------------|
| AC1 | "A detectable secret is present in a tracked file — secrets gate fails" | `G-Secrets` = `Fail`, `AllPassed` = false, exit 1 |
| AC2 | "No secrets detected — secrets gate passes" | `G-Secrets` = `Pass` |
| AC3 | "A dependency with a known CVE is present — vuln gate warns but does not block" | `G-Vuln` = `Warn`, `AllPassed` = true, exit 0 |
| AC4 | "gitleaks is not installed — secrets gate skips with a clear message" | `G-Secrets` = `Skip`, names "gitleaks", no crash, not `Fail` |
| AC5 | "Security gate is disabled — no security results emitted" | neither `G-Secrets` nor `G-Vuln` in output |
| AC5 | "No `[gates.security]` block present — no security results emitted" | same as above |
| AC6 | "The only secret finding matches an allowlist entry — secrets gate passes" | `G-Secrets` = `Pass` |
| AC7 | "Diff-aware filter is active and the secret is in an unchanged file — locally filtered" | `G-Secrets` = `Pass` locally; CI full-scan catches it |
| EC-both-absent | "Both gitleaks and all vuln tools are absent — two Skips with distinct signal" | `G-Secrets` = `Skip`, `G-Vuln` = `Skip`, "no scanners available" message |
| EC-malformed-secrets | "gitleaks returns malformed JSON — gate reports parse error as Warn, not Pass" | `G-Secrets` = `Warn` |
| EC-malformed-vuln | "A vuln tool returns malformed JSON — gate reports parse error as Warn, not Pass" | `G-Vuln` = `Warn` |
| EC-dedup | "Both govulncheck and osv-scanner report the same CVE — de-duped" | `G-Vuln` = `Warn`, single detail entry per (package, vuln-id) |
| EC-no-manifest | "osv-scanner finds no lockfile or manifest — skips without failing" | `G-Vuln` = `Skip` for that tool, not `Fail` |
| EC-allowlist-noop | "An allowlist entry matches no finding — ignored, not an error" | `G-Secrets` = `Pass`, no extraneous warning |

#### UX States

n/a — this feature has no UI surface. It is a CLI gate running inside `centinela validate`; all output is the existing gate-result rendering via `ui.RenderGateResult`, which already handles `Pass`/`Fail`/`Warn`/`Skip` status display. No new interactive or visual surface is introduced.

#### Out-of-Scope

- Full git-history secret scanning (rewriting / remediating historical commits).
- Bundling or installing the scanner binaries (`gitleaks`, `govulncheck`, `osv-scanner`); presence is the user's responsibility.
- License scanning, SBOM generation, or container/image scanning.
- Auto-fixing vulnerable dependencies or opening dependency-bump PRs.
- Maintaining or mirroring any vulnerability database.
- User-configurable scan timeout (v1 uses a fixed internal default of 120s; configurable value is a follow-up).

#### Handoff

- Next role: senior-engineer
- Open clarifications (resolved for spec purposes, noted for implementation):
  - `checkSecurity(cfg, filter) []Result` signature confirmed; wire as `results = append(results, checkSecurity(cfg, filter)...)` in `RunWithFilter`.
  - gitleaks output transport: `--report-path <tmpfile>` preferred for stable JSON across pinned versions (avoids stdout mixing with banner text).
  - gitleaks JSON fields: `RuleID` (string) + `File` (string) per finding slice element.
  - govulncheck JSON fields: streaming NDJSON; collect `osv`/`finding` records to extract `(module, id)`.
  - osv-scanner JSON fields: `results[].packages[].vulnerabilities[].id` + package name.
  - "Both scanners absent" signal: `G-Secrets` message = "gitleaks not found (install to enable secret scanning)"; `G-Vuln` message = "no vuln tools found (install govulncheck or osv-scanner)"; validate output also prints a summary line "security gate: no scanners available — install tools to enforce this gate".
  - Fixed internal timeout: 120s per exec call; timeout maps to `Warn` with message "scan timed out after 120s".
  - Allowlist matching: rule-ID equality first; if the entry contains a `*` or `/` character, treat as `filepath.Match` path glob.
