### Gatekeeper Report: audit-baseline-ratchet

**Date:** 2026-06-16
**Status:** SAFE

## Analyzed Specs

- `specs/audit-baseline-ratchet.feature` (this feature — 21 scenarios: record, ratchet, prune/reintroduce, fingerprint stability, missing-baseline default, newly-enabled gate, full-scan enforcement, determinism, config severity/path).
- Sibling specs touching gates/validate (read for conflict surface):
  - `specs/g2-import-graph-gate.feature` — aggregator-layer rule the new `internal/audit` edges must satisfy.
  - `specs/security-gate.feature` — `G-Secrets: Secret Scan` is a participating gate.
  - `specs/spec-traceability-gate.feature` — `spec-traceability-gate` is a participating gate.
  - `specs/enforcement-profiles.feature` — profile defaults vs the new `[gates.audit_baseline]` block.

## Implementation Reviewed

- `internal/audit/{fingerprint,baseline,record,ratchet,participation,gate}.go`
- `internal/config/audit_baseline.go`, `config.go`, `defaults.go`, `file_size_exceptions.go`
- `cmd/centinela/{audit,audit_baseline,audit_render}.go`, `validate.go`, `validate_audit.go`
- `internal/ui/render_audit.go`, `centinela.toml` (import_graph aggregator layer)

## Conflict Analysis (verified by reading code + running gates)

1. **Config parsing / enforcement-profile defaults — NO CONFLICT.**
   `GatesConfig` gains one additive field `AuditBaseline AuditBaselineConfig` (`toml:"audit_baseline"`). Absent in an old config it is zero-valued then defaulted by `applyDefaults` → `NormalizeAuditBaseline` (severity `warn`, path `.workflow/audit-baseline.json`, `Enabled` false). `validateAuditBaseline` is a no-op when disabled and only rejects an unknown severity when enabled. No existing config key is touched; enforcement profiles are unaffected because the gate defaults OFF.

2. **`appendAuditGate` is a strict NO-OP when disabled — VERIFIED.**
   `cmd/centinela/validate_audit.go`: `if cfg.Gates.AuditBaseline.Enabled { results = append(...) }`. Default `Enabled=false`, so `centinela validate` produces an identical result set to before for every existing scenario. Wired in `validate.go` as `appendAuditGate(cfg, gates.RunWithFilter(...))` — appends only, never mutates other results.

3. **`centinela.toml` aggregator-layer change is purely additive — VERIFIED.**
   Diff adds only `"internal/audit/**"` to the existing `aggregator` layer (`allow = ["domain","leaf"]`); no other layer's `paths`/`allow` changed. `internal/audit` imports `internal/gates` (domain) + `internal/config` (leaf) — both allowed. No other layer's verdict is altered.

4. **No import cycle — VERIFIED.**
   `grep -rn internal/audit internal/gates/` is empty; `gates` never imports `audit`. The audit→gates edge is one-directional. The gate is wired from `cmd/` (the correct seam) precisely to avoid gates→audit. `go build ./...` and `go vet` clean.

5. **Full-scan bypass of diff_mode — VERIFIED.**
   `Record` and `Ratchet` both call `gates.RunWithFilter(cfg, nil)` (nil filter = full repo), independent of `[validate] diff_mode`. Satisfies the "audit scans the full repo even when diff-aware" scenario. Only the validate-embedded gate path uses the diff filter; the standalone `audit`/`audit baseline` commands always full-scan.

6. **Severity vs blocking semantics — consistent.**
   Standalone `centinela audit` blocks on any new violation via `Diff.HasNew()` regardless of severity (severity only maps the validate-embedded gate's Status via `newStatus`). This matches the spec's warn/disabled scenarios, which target the validate gate, and the acceptance test comments confirm the same model.

## Gate-Keepers Checklist (by inspection)

- [x] **File size ≤100 (incl. `_test.go` in `internal/`+`cmd/`):** all `internal/audit/*.go` (max 94), all `cmd/centinela/audit*.go` + `validate_audit*.go` (max 86), `internal/config/audit_baseline*.go` (max 69), `internal/ui/render_audit.go` (40) are ≤100. The three `tests/{acceptance,integration,unit}/*` tier files (440/116/115) exceed 100 but live under `tests/` which is exempt from G1 (only `internal/`+`cmd/` `_test.go` are gated). G1 gate ran ✓ "All files under 100 lines."
- [x] **No cross-layer import violation / no cycle:** import_graph gate ran ✓ failing-status clear; `internal/audit` correctly mapped to aggregator; not present in the unmapped-packages warning.
- [x] **`centinela validate` passes:** context-confirmed (G1 ✓, cross-compile ✓, spec-traceability all 21 ✓, roadmap_drift ✓). Independently: `go build ./...` ✓, `go vet` ✓, full suite **1179 tests pass**, `validate --full` gate run exits without "validation failed."
- [x] **No business logic in outer layer:** `cmd/centinela/audit*.go` are thin orchestrators — load config, call `audit.Load/Record/Ratchet/Save`, render, set exit code. All decisions (fingerprinting, partitioning, severity mapping, participation) live in `internal/audit`.
- [x] **i18n:** N/A — PROJECT.md declares English-only, `gates.i18n` disabled; siblings hardcode English CLI strings by convention.
- [x] **Gatekeeper report:** SAFE.
- [ ] **Production readiness:** not evaluated here — gate not indicated as enabled for this feature; defer to validation-specialist if `gates.production_readiness=true`.

## Findings

None of WARNING or BLOCKING severity.

Informational (non-blocking):
- The spec narrative (Background line 20; scenario line 238) refers to the path key loosely as "baseline"; the implemented + tested toml key is `baseline_path` (acceptance tests write `baseline_path = "..."`). No scenario step asserts a literal key name, so spec-traceability is unaffected. Cosmetic only.

## Deferred Findings

None.

## Recommendation

SAFE to advance to the validate step. The feature is fully additive: the gate defaults OFF (safe-adoption), `appendAuditGate` is a verified no-op when disabled, the import-graph change only adds `internal/audit` to the aggregator layer with no cycle, and the standalone audit commands force a full-repo scan. All 1179 tests pass; build and vet are clean. Handoff to validation-specialist.
