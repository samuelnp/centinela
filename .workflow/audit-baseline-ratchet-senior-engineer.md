# audit-baseline-ratchet — senior-engineer

## Files Touched

New package `internal/audit/` (each ≤100 lines):
`fingerprint.go` (per-gate identity extractor + `sha256(scheme+gate+key)`),
`baseline.go` (`Baseline`/`GateEntry`, deterministic `Save`, `Load`),
`record.go` (`Record` over `gates.RunWithFilter(cfg,nil)`, Fail-only),
`ratchet.go` (`Diff{New,Baselined,Resolved}`, `Ratchet`, `HasNew`),
`participation.go`, `gate.go` (`Check` → `gates.Result`).

Config: `internal/config/audit_baseline.go` (`AuditBaselineConfig` +
`Normalize`/`validate`, mirrors roadmap_drift); `config.go` (+field),
`defaults.go` (+normalize), `file_size_exceptions.go` (+validate).

Command: `cmd/centinela/audit.go` (ratchet, `--json`), `audit_baseline.go`
(record), `audit_render.go` + `internal/ui/render_audit.go` (render),
`validate_audit.go` (`appendAuditGate` — the cmd-side wiring), `validate.go`
(+call). `centinela.toml` (+`internal/audit/**` in aggregator layer).

## Architecture Compliance

- **G1**: every source file ≤99 lines (`validate.go` kept at 97 by extracting
  `validate_audit.go` rather than baselining the regression).
- **G2 import-graph**: `internal/audit` joins the `aggregator` layer (allows
  domain+leaf). Edges audit→`gates` (domain) + audit→`config` (leaf) add no
  failing edge; `import_graph` gate test passes. **No cycle** — the gate is
  wired from `cmd/centinela/validate.go` (`appendAuditGate`), never from inside
  `gates`, so `internal/gates` has no `audit` import.
- **Scaffold mirror**: no-op by design — `internal/scaffold/assets/centinela.toml`
  carries no `[gates.import_graph]` matrix (verified 0 entries), so there is
  nothing to mirror.

## Type-Safety Notes

Strict Go, no `any`. Deterministic baseline (sorted gates + fingerprints,
versioned `scheme`). `gofmt`/`go vet` clean.

## Trade-Offs

- v1 parses each gate's `Result.Details` via a per-gate identity extractor
  (stable key = path/edge/scenario) + generic fallback; a structured-`Finding`
  refactor of every gate is deferred (out of scope).
- Only `Status == Fail` violations are baselined (excludes import-graph's
  non-failing "unmapped" Warn and skipped gates).

## Two bugs found by independent dogfood + fixed (grind, don't game)

1. `audit baseline` failed in a repo with no `.workflow/` dir — `Save` now
   `os.MkdirAll`s the parent (matching the telemetry writer).
2. A blocking ratchet printed the full cobra usage block — added
   `SilenceUsage`/`SilenceErrors` to both audit commands (sibling convention);
   `main` still prints the returned error once to stderr and exits 1.

Dogfood (throwaway `/tmp` binary, fresh git repo) confirmed: no-baseline →
non-block exit 0; record writes file; 0-new exit 0; oversized file growing
130→170 lines stays baselined (fingerprint stability); new file → exit 1 naming
only the new one; delete → exit 0; re-record byte-identical.

## Handoff

→ qa-senior. Not user-facing (CLI). Tests step: unit (fingerprint stability,
ratchet diff, baseline determinism, participation, config normalize), integration
(record→ratchet over a temp repo), acceptance mapping the 21
`specs/audit-baseline-ratchet.feature` scenarios. Colocated `internal/audit`
tests needed for per-package coverage (each ≤100 lines).
