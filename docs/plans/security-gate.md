# Plan: security-gate

> Implements the brief in [docs/features/security-gate.md](../features/security-gate.md).
> A new opt-in `[gates.security]` gate inside `centinela validate` that mechanically
> enforces secret-scanning (hard `Fail`) and dependency-vuln audit (`Warn`),
> extending the existing gate framework (`g2-import-graph-gate`, build gate).

## Locked Decisions (honored, not relitigated)

1. **Scope = both** secret-scanning AND dependency-vuln audit in v1.
2. **Tooling = `gitleaks` (secrets) + `govulncheck` + `osv-scanner` (vulns).**
3. **Severity split:** secrets = hard `Fail` (blocks validate); vulns = `Warn` (surface only).
4. **Missing/uninstalled scanner = `Skip` + warning, NEVER `Fail`** (must not break validate).
5. **Diff-aware secret scan locally** (honor the existing `gitdiff.Set` filter, like
   G1/G11); **vuln audit is whole-project** (ignores the filter, like the build gate).
6. **Gate OFF by default** (`enabled = false`); zero-config-safe.

## Problem Framing

Centinela's promise is *enforced, not requested* correctness. Two high-impact
security failure modes are today caught only by human/subagent review: leaked
secrets (a committed API key/token/private key) and vulnerable dependencies (a
dependency carrying a known CVE). Both are exactly the agent-failure classes the
roadmap targets. The existing gate framework — `gates.Result{Status: Pass|Fail|Warn|Skip}`,
`RunWithFilter`, `AllPassed` — already provides the mechanism; `g2-import-graph-gate`
(whole-project, filter-ignoring) and the build gate (external-command, whole-project)
are the reference shapes we extend. This gate moves both failure modes from
audit-time to `validate`-time.

## Scope (v1)

### In
- New `[gates.security]` config block: `SecurityGateConfig{ Enabled; Secrets; Vuln }`
  with shape validation, mirroring `BuildGateConfig`/`ImportGraphConfig`.
- Secret scan via `gitleaks`, diff-aware locally / full-scan in CI → hard `Fail` on finding.
- Vuln audit via `govulncheck` (Go) + `osv-scanner` (multi-ecosystem manifests),
  whole-project → `Warn` on finding.
- `exec.LookPath`-driven presence detection → `Skip` + clear message on absent tool.
- Defensive JSON parsing (`gitleaks --report-format json`, `govulncheck -json`,
  `osv-scanner --format json`); malformed/empty output → `Warn`/`Skip`, never false `Pass`.
- De-dup of vuln findings by `(package, vuln-id)` across the two vuln tools.
- Secrets allowlist (gitleaks rule IDs / path globs) for known false positives.
- Two **separate** `Result` entries per enabled gate (`G-Secrets`, `G-Vuln`).
- Off by default; absent/`enabled=false` → no security results emitted.

### Out (explicitly deferred)
- Full git-history secret scanning + rotation/remediation guidance.
- Bundling/installing scanner binaries (presence is the user's responsibility).
- License scanning, SBOM, container/image scanning.
- Auto-fixing vulns or opening dependency-bump PRs.
- Maintaining/mirroring any vulnerability database.
- **Configurable scan timeout: OUT for v1** (see Resolved Design Question 5).

## Dependencies & Assumptions

- **Internal modules:** `internal/config` (gate config + `applyDefaults`/`validateConfig`
  wiring), `internal/gates` (`RunWithFilter`, `Result`/`Status`, `gitdiff.Set` filter),
  `internal/gitdiff` (`Set.Contains` for diff-scoping), `cmd/centinela/validate.go`
  (already renders `[]Result` via `ui.RenderGateResult` — no change needed).
- **External binaries** (NOT bundled): `gitleaks`, `govulncheck`, `osv-scanner`.
  Invoked via `os/exec`; absence detected with `exec.LookPath`.
- **Build gate is the external-command reference** (argv via `strings.Fields`,
  direct exec, no `sh -c`); **import-graph is the whole-project / filter-ignored reference.**
- **Assumption:** JSON output schemas of the three tools are stable enough to parse
  defensively; we validate *shape*, not the vuln DB. Tested tool versions pinned in docs.
- **Assumption:** `AllPassed` already fails only on `Fail`, so `Warn` vuln results
  naturally do not block validate — no change to `AllPassed`.

## Resolved Design Questions

1. **Result naming:** `G-Secrets` and `G-Vuln` (matching the `G-Build` convention).
2. **Skip detection:** each scanner runner calls `exec.LookPath(tool)` first via the
   shared `security_exec.go` helper; on `ErrNotFound` it returns a `Skip` Result naming
   the missing tool — before any scan is attempted. Both families absent → two `Skip`
   results + a clear "no scanners available" signal (so it's not mistaken for clean).
3. **JSON parse strategy (per tool, all defensive):**
   - `gitleaks detect --no-banner --report-format json --report-path <tmp>` (or stdout):
     unmarshal into a slice of findings `{RuleID, File, ...}`. Non-zero exit with valid
     JSON findings → `Fail`. Non-zero exit with unparseable output → `Warn` (tool/parse
     error), never `Pass`.
   - `govulncheck -json ./...`: streaming NDJSON; collect `osv`/`finding` records into
     `(package, vuln-id)` pairs. Parse error → `Warn`.
   - `osv-scanner --format json -r .`: single JSON document `{results:[{packages:[{vulnerabilities:[...]}]}]}`;
     extract `(package, vuln-id)`. No manifest/lockfile → tool reports nothing scannable
     → `Skip` for that tool (not `Fail`).
   - Rule: **malformed/empty JSON never yields a false `Pass`** — it maps to `Warn`
     (scan ran, output unusable) or `Skip` (nothing to scan / tool absent).
4. **Vuln de-dup:** aggregate findings from both tools into a `map[struct{pkg,id}]bool`,
   emit sorted `Details` lines; identical CVE reported by both tools appears once.
5. **Configurable scan timeout: OUT for v1.** Rationale: keeps the first slice small and
   the config surface minimal; the brief lists it as a "design detail," and a wedged scan
   is a low-likelihood risk. We still wrap each `exec` in `context.WithTimeout` with a
   **fixed internal default** (e.g. 120s) so a hung tool degrades to `Warn` rather than
   wedging validate — but the value is not user-configurable in v1. A `timeout` config
   field is a clean follow-up.

## Touchpoints (concrete)

| File | Change | Notes |
|------|--------|-------|
| `internal/config/security_gate.go` (NEW, ≤100) | `SecurityGateConfig`, `SecretsConfig`, `VulnConfig` + `NormalizeSecurityGate` | mirror `build_gate.go`/`import_graph.go`; trim + validate known tool names |
| `internal/config/config.go` | add `Security SecurityGateConfig \`toml:"security"\`` to `GatesConfig` | one field |
| `internal/config/defaults.go` | `cfg.Gates.Security = NormalizeSecurityGate(cfg.Gates.Security)` | mirror Build/ImportGraph lines |
| `internal/config/file_size_exceptions.go` (`validateConfig`) | add shape validation call for security tool names | reject unknown `tool`/`tools` values with a clear error |
| `internal/gates/gates.go` | add `if cfg.Gates.Security.Enabled { results = append(results, checkSecurity(cfg, filter)...) }` to `RunWithFilter` | note: `checkSecurity` returns `[]Result` (multiple entries), so append with `...` |
| `internal/gates/security.go` (NEW, ≤100) | `checkSecurity(cfg, filter) []Result` orchestrator → calls secrets + vuln, returns 1–2 Results | |
| `internal/gates/security_secrets.go` (NEW, ≤100) | gitleaks invoke + JSON parse + allowlist + diff-aware filter → one `G-Secrets` Result | honors `filter.Contains` for local diff scope |
| `internal/gates/security_vuln.go` (NEW, ≤100) | govulncheck + osv-scanner invoke + parse + `(pkg,id)` de-dup → one `G-Vuln` Result | whole-project; ignores `filter` like import-graph |
| `internal/gates/security_exec.go` (NEW, ≤100) | `lookPath`/presence helper + `runJSON(ctx, tool, args...)` exec wrapper (capture stdout/stderr, timeout) | shared by secrets + vuln |
| `specs/security-gate.feature` (NEW) | Gherkin acceptance scenarios mapping AC#1–7 | feature-specialist authors |

**File-size discipline:** every new source file (incl. `_test.go` in `internal/`) ≤100
lines. The orchestrator/secrets/vuln/exec split exists specifically to stay under the cap;
if any file approaches the limit, split the parse helper into a `*_parse.go` sibling.

## Risk Table

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| External tool not installed → validate breaks | High | High | `exec.LookPath` → `Skip` + warning; never `Fail` on absence (AC#4). Resolved Q2. |
| Secret-scan false positives → users disable the gate | High | Medium | Allowlist (rule ID / path glob); AC#6; document tuning in docs step. |
| Vuln noise (transitive / no-fix) blocks work | Medium | High | Vulns are `Warn`, not `Fail` (decision #3). `AllPassed` unaffected. |
| Tool JSON format drift across versions | Medium | Medium | Defensive parse; malformed → `Warn`/`Skip`, never false `Pass` (Resolved Q3). Pin tested versions in docs. |
| Slow / hung scan wedges validate | Medium | Low | Fixed-default `context` timeout per exec → `Warn` on timeout; configurable value deferred (Resolved Q5). |
| Any new gate file > 100 lines | Medium | Medium | 4-file split (`security.go`/`_secrets.go`/`_vuln.go`/`_exec.go`); `_test.go` ≤100 too. |
| Secret committed before gate was enabled | Medium | Medium | v1 scans working tree / diff only; full-history scan is out of scope (follow-up). |
| Diff-scoped local scan misses a secret in an unchanged file | Medium | Medium | Documented as matching G1/G11 behavior (AC#7); CI full-scan is the backstop. |
| `checkSecurity` returning `[]Result` mis-wired into `RunWithFilter` | Low | Low | Append with `...`; covered by gates wiring unit test (off→0, on→correct count). |

## Rollout Sequence (smallest correct slice first)

- **Slice 1 — Config + skeleton + Skip-on-absent (zero behavior risk).**
  `SecurityGateConfig`/`SecretsConfig`/`VulnConfig` + `NormalizeSecurityGate` + defaults +
  `validateConfig` shape check; `security.go` orchestrator + `security_exec.go` presence
  helper; wire into `RunWithFilter` behind `cfg.Gates.Security.Enabled`. With no tools
  configured-but-present, both checks return `Skip`. Proves: off-by-default emits nothing
  (AC#5); enabled-but-absent emits `Skip` not `Fail` (AC#4); validate never crashes.

- **Slice 2 — Secrets / gitleaks hard-`Fail` (highest impact).**
  `security_secrets.go`: invoke gitleaks with JSON output, parse findings, apply allowlist,
  honor diff-aware `filter` locally. Finding → `Fail` (AC#1); clean → `Pass` (AC#2);
  allowlisted-only → `Pass` (AC#6); malformed output → `Warn` (edge case). Delivers the
  blocking half on its own.

- **Slice 3 — Vuln / govulncheck + osv-scanner `Warn` (surface only).**
  `security_vuln.go`: invoke both tools whole-project, parse, de-dup by `(pkg,id)`.
  Finding → `Warn` with `AllPassed` still true (AC#3); no manifest → per-tool `Skip`;
  both report same CVE → one detail line. Stacks on top without touching Slice 2.

- **Slice 4 — Docs / spec / parity.**
  `specs/security-gate.feature` (AC#1–7), config-reference + tuning docs (allowlist,
  pinned tool versions, CI vs local behavior). If a config-leaf string set (known tool
  names) must stay in sync with the domain, add a parity test — and mirror any edited
  architecture doc into `internal/scaffold/assets` per project memory.

## Handoff to feature-specialist

- Confirm the `checkSecurity(...) []Result` signature and the `append(results, checkSecurity(cfg, filter)...)`
  wiring in `RunWithFilter` (it returns a slice, unlike the single-Result gates).
- Decide gitleaks output transport: `--report-path <tmpfile>` (read+cleanup) vs stdout
  capture — pick the one with the most stable JSON across pinned versions.
- Confirm exact field paths for each tool's JSON (rule ID / file for gitleaks; OSV id +
  module for govulncheck; package + vuln id for osv-scanner) against the pinned versions.
- Confirm the "both scanners absent" combined signal wording (a distinct message so it is
  not read as a clean pass).
- Confirm the fixed internal timeout value (proposed 120s) and that timeout maps to `Warn`.
- Confirm allowlist matching semantics: rule-ID equality vs path-glob (`filepath.Match`).
