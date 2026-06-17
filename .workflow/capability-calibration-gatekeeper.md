# Gatekeeper Report: capability-calibration

**Date:** 2026-06-15
**Status:** SAFE

## Scope

Two-part feature: (1) telemetry per-event model stamping; (2) `internal/calibration`
per-model governance-friction analyzer + `centinela calibrate` command. Every Gate
Keepers Checklist item was verified mechanically against a freshly built worktree
binary (`go build ./cmd/centinela`) and an isolated synthetic telemetry log — never
trusted from prose.

## Checklist Verification

### 1. File size (<=100 lines, src + colocated `_test.go` in internal/ + cmd/)
- `find internal cmd -name '*.go' | xargs wc -l | awk '$1>100'` → **empty**. No file
  exceeds 100 lines. G1 gate in `validate --full`: PASS ("All files under 100 lines").

### 2. Cross-layer imports / leaf purity
- `go list -deps ./internal/telemetry | grep samuelnp` → only `internal/config` +
  `internal/telemetry`. **Telemetry remains a config-only leaf** — Part 1 added no new
  import (value resolved at cmd/ emit sites via `resolveEmitModel`/`resolveEmitModelFrom`
  in `cmd/centinela/telemetry_model.go`, model passed in as a trailing `string`).
- `go list -deps ./internal/calibration | grep samuelnp` → only `internal/config` +
  `internal/telemetry` (+ stdlib). **No import of cmd/ or internal/ui.**
- `internal/calibration/**` is mapped into the `aggregator` layer in `centinela.toml`
  (`paths = ["internal/doctor/**", "internal/insights/**", "internal/calibration/**"]`,
  `allow = ["domain","leaf"]`). PROJECT.md G2 prose updated to document calibration as
  an aggregator over the telemetry+config leaves, imported only by cmd/ (Report type by
  internal/ui for rendering).
- `validate --full` `import_graph`: **WARN** (pre-existing, unchanged) — see note below.

### 3. `centinela validate --full` passes
- Exit 0. G1 PASS, G-Build (6 cross-compile targets) PASS, roadmap_drift PASS,
  `go test ./...` PASS, `go test ./tests/acceptance/...` PASS.
- `import_graph` ⚠ and `spec-traceability-gate` ⚠ are **pre-existing repo-wide warnings**.
  Confirmed by building the `main` binary and running `validate --full` on `main`: it
  emits the identical two warnings. This feature introduces **no new** unmapped package
  (`internal/calibration` is mapped) and **no new** uncovered scenario (see item 8).

### 4. No business logic in the outer (cmd/) layer
- `cmd/centinela/calibrate.go` is thin: `config.Load` → `telemetry.ReadDefault` →
  `calibration.Calibrate` → `ui.RenderCalibration` (or `json.MarshalIndent`). No analysis.
- `cmd/centinela/telemetry_model.go` is thin model-resolution only (precedence:
  `wf.DriverModel` → `config.DriverModelFrom`). All friction tallying, classification,
  thresholds, and sort live in `internal/calibration`.

### 5. i18n — N/A (disabled for this project).

### 6. Correctness (empirical, isolated synthetic log)
Built a temp repo (`centinela.toml` mapping `model-*` ids to classes; a 31-event
`events.jsonl` + one malformed line) and ran the real binary from that CWD. JSON +
render results, all exit 0:
- **Undergoverned**: `model-frontier` (frontier→outcome, 3 advances + 3 rework,
  Rate=1.0) → **Undergoverned / Tighten → guided**, cites `advances=3 rework=3 rate=1.00`.
- **Overgoverned**: `model-limited` (limited→strict, 4 advances + 1 rework, Rate=0.25)
  → **Overgoverned / Loosen → guided**, cites the counts.
- **Boundary inclusivity**: Rate exactly 1.0 → Undergoverned; exactly 0.25 → Overgoverned
  (`>=`/`<=` confirmed in `classify.go` and via the synthetic log).
- **Maxed → Keep**: `model-strict-hi` (strict + Rate=1.0, no tighter) → WellCalibrated/Keep;
  `model-frontier-lo` (outcome + Rate=0.25, no looser) → WellCalibrated/Keep.
- **<3 advances**: `model-capable-few` (2 advances) → WellCalibrated/Keep.
- **Zero advances**: `model-zero` (blocks only) → HasRate=false, `rate=n/a`, Keep, no
  division/panic, exit 0.
- **Unclassified + unattributed**: `model-unknown` (no class) → Unclassified/None;
  the empty-model events bucket as `unattributed` → Unclassified, **rendered LAST**.
- **Empty/missing log**: missing `events.jsonl` and empty file both → "no telemetry yet"
  line, exit 0; `--json` → `{"ModelCount":0,...,"Models":[]}`.
- **Malformed line skipped**: the non-JSON line is silently dropped (counts stay clean).
- **`--json` valid + byte-stable**: parses under `json.load`; two consecutive runs are
  byte-identical (`diff` clean). Deterministic order (model id asc, unattributed last).
- **Non-TTY no ANSI**: piped render contains no `\x1b[` escapes.
- **Stamping**: a constructor call with a model marshals `"model":"<id>"`; with no model,
  omitempty drops the field entirely; a legacy line without `model` reads back with
  empty Model and `Calibrate` buckets it as `unattributed`. All emit sites
  (`complete.go`, `complete_verify.go`/`telemetry_emit.go`, `hook_prewrite.go`,
  `validate.go`) resolve and thread the model through (verified in the diff).

### 7. Coverage
- `./scripts/check-coverage.sh` → **95.3% >= 95.0%** (genuine; gate uses the 1e-9-margin
  float compare, not a rounding pass — actual value clears the bar).
- Per-package on the changed code: `internal/calibration` **100.0%**;
  `internal/ui/render_calibration.go` all four funcs **100.0%**;
  `cmd/centinela/telemetry_model.go` both funcs **100.0%**; `calibrate.go` `runCalibrate`
  80% (only the `config.Load`/`ReadDefault` error returns uncovered — acceptable).
  The new feature carries its own weight; the 95.3% aggregate floor is set by
  pre-existing packages, not by thin new-code coverage.

### 8. Tests real & executed
- `go test ./...` → **2104 passed across 29 packages**, exit 0.
- All **27** capability-calibration scenarios have acceptance coverage: normalized
  (lowercase / collapsed-whitespace / trailing-period-stripped, mirroring the gate's
  `normalizeScenario`) diff of spec `Scenario:` names vs acceptance `// Scenario:`
  comments → **0 uncovered**. The repo-wide spec-traceability ⚠ is pre-existing (present
  on `main`), not caused by this feature.
- Spot-checked tests are real, not theater:
  - `capability_calibration_boundary_test.go` builds exact advance/rework counts, runs
    the real binary, and asserts verdict + recommendation + raw counts + exit code.
  - `capability_calibration_stamping_test.go` does a constructor→ReadDefault round-trip
    asserting the stamped model, the empty-model case, and legacy→unattributed bucketing.
  - `tests/integration/calibration_roundtrip_test.go` stamps via `Record*`, reads back,
    Calibrates, and asserts each model's classification end to end.

## Findings

- **Affected spec:** (none — repo-wide gate warnings)
  - **Risk:** `import_graph` and `spec-traceability-gate` show ⚠ in `validate --full`.
  - **Assessment:** Both are **pre-existing**, reproduced identically by the `main`
    binary. `internal/calibration` is mapped to the aggregator layer (no new unmapped
    edge — its only internal edges are to the telemetry + config leaves), and all 27 new
    scenarios are covered. **No new failure or warning is attributable to this feature.**
  - **Suggestion:** No action required for this feature.

## Deferred Findings
- none.

## Recommendation

**SAFE.** No conflicts with existing specs. Telemetry leaf purity is preserved
(config-only; model value resolved at the cmd/ boundary). `internal/calibration` is a
clean read-only aggregator over the telemetry + config leaves, correctly mapped and
documented. Classification is correct and boundary-inclusive across all required cases,
output is deterministic and byte-stable, empty/malformed/zero-advance inputs are handled
without panic, new code is ~100% covered, and the full 2104-test suite is green. The two
`validate` warnings are pre-existing and not introduced here. Proceed.

**Verdict: SAFE**
