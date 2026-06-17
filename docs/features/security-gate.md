# Feature: security-gate

> Phase 3 (Close the Mechanical-Verification Gap). Sibling to the shipped
> `g2-import-graph-gate` and the build/file-size/i18n gates. Converts "did an
> agent leak a secret or pull a vulnerable dependency?" from a *requested*
> subagent-review concern into a *mechanically enforced* gate inside
> `centinela validate`.

## Problem

Centinela's flagship promise is **enforced, not requested** correctness. Today
two high-impact security failure modes are caught only by human/subagent review
(if at all):

1. **Leaked secrets** — an agent commits an API key, token, private key, or
   `.env` value. The gatekeeper subagent may miss it; nothing mechanical stops it.
2. **Vulnerable dependencies** — an agent adds or bumps a dependency carrying a
   known CVE. No gate audits the dependency set.

Both are exactly the "treat every agent failure as an engineering problem to fix
permanently" cases the roadmap targets. The existing gate framework
(`internal/gates/`, `gates.Result{Status}`, `RunWithFilter`, `AllPassed`) already
gives us the mechanism — `g2-import-graph-gate` is the reference implementation.

**Who is hurting:** any team running Centinela on a repo an agent writes to —
the secret leak is a security incident; the vulnerable dep is latent risk. The
gate makes both fail loudly at `validate` time instead of at audit time.

## Outcome

A new opt-in `[gates.security]` gate, wired into `centinela validate` alongside
the existing gates, that runs:

- **Secret scanning** (`gitleaks`) over the working tree (diff-aware in local
  validate, full-scan in CI) → a finding is a **hard `Fail`** (blocks validate).
- **Dependency vulnerability audit** (`govulncheck` for Go + `osv-scanner` for
  multi-ecosystem manifests) over the project's dependency set → a finding is a
  **`Warn`** (surfaces, does not block — fixes are often unavailable, transitive,
  or need human judgement).

Each underlying scanner is an **external binary**. If a configured scanner is not
installed, its check returns **`Skip`** with a clear message (never crashes
validate, never silently passes). Availability/parsing of tool output is the
gate's job; Centinela does not bundle the binaries.

```toml
[gates.security]
enabled = true

# Secret scanning (hard-fail). Optional allowlist for known false positives.
[gates.security.secrets]
tool = "gitleaks"            # v1: gitleaks
allowlist = []               # gitleaks rule IDs or path globs to ignore

# Dependency vulnerability audit (warn-only in v1).
[gates.security.vuln]
tools = ["govulncheck", "osv-scanner"]   # run those present; Skip+warn if absent
```

## Design Decisions (locked with the user)

1. **Scope = both** secret-scanning AND dependency-vuln audit in v1.
2. **Tooling = `gitleaks` + `govulncheck` + `osv-scanner`.** govulncheck covers
   Go-native vulns; osv-scanner adds multi-ecosystem manifest coverage (npm, pip,
   etc.) for the non-Go projects Centinela governs.
3. **Severity split: secrets `Fail`, vulns `Warn`.** A leaked secret blocks
   validate; a vulnerable dependency warns (a fix may not exist / may be
   transitive). Implemented via the existing `gates.Status` — `AllPassed` already
   fails only on `Fail`.
4. **Missing tool = `Skip` + warning, never `Fail`.** A gate must not break a
   validate run just because a scanner isn't installed; it reports the gap.
5. **Diff-aware where it makes sense.** Secret scanning honors the existing
   diff-aware `filter` in local validate (scan changed files) and full-scans in
   CI — mirroring G1/G11. Vuln audit is inherently whole-project (the dependency
   set), so it ignores the filter — mirroring the build gate.
6. **Opaque tool output, parsed to findings.** Each scanner runs with JSON output
   (`gitleaks detect --report-format json`, `govulncheck -json`,
   `osv-scanner --format json`); the gate parses findings into `Result.Details`.
   Centinela validates the *shape* of output, not the vuln database.

### Severity → gate Status mapping

| Check | Tool(s) | Finding | No finding | Tool absent |
|-------|---------|---------|------------|-------------|
| secrets | gitleaks | `Fail` (blocks) | `Pass` | `Skip` + warn |
| vuln | govulncheck, osv-scanner | `Warn` (surfaces) | `Pass` | `Skip` + warn |

Secrets and vuln are surfaced as **separate `gates.Result` entries** (e.g.
`G-Secrets`, `G-Vuln`) because they carry different severities — `RunWithFilter`
already returns a `[]Result`, so one enabled gate may append more than one result.

## User Stories

- As a team lead, I want a committed secret to **fail `centinela validate`** so it
  can never reach a merge, the same way a file-size violation does.
- As a developer, I want a newly introduced vulnerable dependency **surfaced as a
  warning** at validate time so I can triage it, without blocking unrelated work
  when no fix is available yet.
- As an operator on a machine without `osv-scanner` installed, I want validate to
  **skip that scanner with a clear message**, not crash or falsely pass.
- As a maintainer, I want an **allowlist** for known-false-positive secret matches
  so the gate stays trustworthy and isn't disabled out of frustration.
- As a zero-config user, I want the gate **off by default** (`enabled = false`)
  so existing projects are unaffected until they opt in.

## Acceptance Criteria (→ Gherkin in `specs/security-gate.feature`)

1. Given `[gates.security] enabled = true` and a file containing a detectable
   secret, when `centinela validate` runs, then the security/secrets gate result
   is `Fail` and `AllPassed` is false (validate blocks).
2. Given the secrets scanner finds nothing, when validate runs, then the secrets
   gate result is `Pass`.
3. Given a project dependency with a known vulnerability, when validate runs, then
   the vuln gate result is `Warn` (surfaced in output) and `AllPassed` stays true
   (validate does not block on it).
4. Given `gitleaks` is not installed, when validate runs with the gate enabled,
   then the secrets gate result is `Skip` with a message naming the missing tool
   (validate does not crash and does not `Fail` on the absence).
5. Given `[gates.security] enabled = false` (or absent), when validate runs, then
   no security gate result is produced (zero-config-safe, unchanged behavior).
6. Given a secret match whose rule ID / path is in `secrets.allowlist`, when
   validate runs, then that match is excluded and (if it was the only finding) the
   secrets gate result is `Pass`.
7. Given the diff-aware filter is active (local validate) and the only secret is
   in an unchanged file, then v1 behavior matches the existing file-scoped gates
   (G1/G11): the secrets scan is filter-scoped locally; CI full-scan still catches it.

## Edge Cases

- Scanner present but returns malformed/empty JSON → gate reports an internal
  error as a `Warn`/`Skip` with the parse failure, never a false `Pass`.
- No dependency manifest for a configured vuln tool (e.g. osv-scanner with no
  lockfile) → `Skip` for that tool, not `Fail`.
- Both vuln tools present and both report the same CVE → de-duplicate by
  (package, vuln-id) in `Details`.
- Secret allowlist entry that matches nothing → ignored (not an error).
- Gate enabled but BOTH scanner families absent → two `Skip` results + warnings;
  validate still passes (nothing to enforce), with a clear "no scanners available"
  signal so it isn't mistaken for a clean scan.
- Large repo / slow scan → respect a configurable timeout; on timeout, `Warn`
  (don't wedge validate). (Timeout value: design detail for plan.)
- Secret scanning over the diff vs full history: v1 scans the working tree /
  changed files, not full git history (history rewriting is out of scope).

## Data Model

No persisted runtime entities. Configuration + per-run results only:

- **`internal/config/` (leaf):**
  - `GatesConfig.Security SecurityGateConfig` (`toml:"security"`).
  - `SecurityGateConfig{ Enabled bool; Secrets SecretsConfig; Vuln VulnConfig }`.
  - `SecretsConfig{ Tool string; Allowlist []string }`,
    `VulnConfig{ Tools []string }`. Shape validation: known tool names, etc.
- **`internal/gates/` (domain):**
  - `checkSecurity(cfg, filter) []Result` (or `checkSecrets` + `checkVuln`),
    wired into `RunWithFilter` behind `cfg.Gates.Security.Enabled`.
  - Per-scanner runner files kept ≤100 lines:
    `security.go` (orchestrator), `security_secrets.go` (gitleaks invoke+parse),
    `security_vuln.go` (govulncheck + osv-scanner invoke+parse), plus a small
    `security_exec.go` for the external-command + tool-presence helper.
  - Reuse `gates.Result`/`Status`; no new status type needed.
- **`cmd/centinela/validate*.go`** stays thin — it already renders `[]Result`.

## Integration Points

- **`internal/gates/gates.go`** — add the `Security` branch to `RunWithFilter`.
- **`internal/config/`** — `SecurityGateConfig` + shape validation; mirror the
  `BuildGateConfig`/`ImportGraphConfig` pattern; parity test if a config-leaf
  string set (tool names) must stay in sync with the domain.
- **External binaries** — `gitleaks`, `govulncheck`, `osv-scanner`. Invoked via
  `exec`; presence detection (`exec.LookPath`) drives the `Skip` path. JSON output
  parsed into findings. **No network/database work in Centinela.**
- **CI vs local** — full-scan in CI (existing `centinela validate` CI job),
  diff-aware locally (existing `gitdiff.Set` filter).
- **`custom-gate-sdk` (Phase 3, depends on g2)** — security-gate is another
  reference built-in; its per-scanner shape should stay generalizable so the SDK
  can later express equivalent gates.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| External tool not installed → validate breaks | High | High | `exec.LookPath` → `Skip` + warning; never `Fail` on absence. AC#4. |
| Secret-scan false positives → users disable the gate | High | Medium | Allowlist (rule ID / path glob); AC#6; document tuning. |
| Vuln noise (transitive, no-fix) blocks work | Medium | High | Vulns are `Warn`, not `Fail` (decision #3). |
| Tool output format drift across versions | Medium | Medium | Parse JSON defensively; malformed output → `Warn`/`Skip`, never false `Pass`. Pin tested tool versions in docs. |
| Slow scans wedge validate | Medium | Low | Configurable timeout → `Warn` on timeout. |
| Each gate file > 100 lines | Medium | Medium | Split per scanner (`security.go`/`_secrets.go`/`_vuln.go`/`_exec.go`); _test.go ≤100 too. |
| Secret committed before gate enabled | Medium | Medium | v1 scans working tree/diff; full-history scan + remediation is out of scope (note for follow-up). |

## Decomposition

Sized for one feature. Natural split if big-thinker finds it too large:

- **`security-gate-secrets`** — `[gates.security.secrets]` + gitleaks invoke/parse
  + hard-`Fail` wiring + Skip-on-absent. Delivers the highest-impact half.
- **`security-gate-vuln`** — `[gates.security.vuln]` + govulncheck + osv-scanner
  + `Warn` wiring on top.

Explicitly **out of scope** for v1:

- Full git-history secret scanning + rotation/remediation guidance.
- Bundling/installing the scanner binaries (presence is the user's responsibility).
- License scanning, SBOM generation, container/image scanning.
- Auto-fixing vulnerabilities or opening dependency-bump PRs.
- Maintaining or mirroring any vulnerability database.
