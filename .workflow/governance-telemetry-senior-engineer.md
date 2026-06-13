# Senior-Engineer Report: governance-telemetry (code step)

## Summary

Implemented the governance event log as specified in the approved plan: a
non-blocking, append-only JSONL recorder in the new leaf package
`internal/telemetry` (config + stdlib only), a `*bool` opt-out config mirroring
`[memory]`, and emission wired into the three `cmd/` chokepoints. Domain
producers stay pure — all emission happens from `cmd/`. `go build`, `go vet`,
and `gofmt` are clean; full suite (1569 tests, 26 packages) is green; every
new/changed source file is ≤100 lines.

## Files Touched

Created:

| File | Lines | Role |
|------|-------|------|
| `internal/config/telemetry.go` | 10 | `TelemetryConfig{Enabled *bool}` + `IsEnabled()` (opt-out, default ON) |
| `internal/telemetry/event.go` | 44 | Contract: `Schema = "centinela.telemetry/v1"`, 5 type consts, flat `Event` + owned `CheckRef` |
| `internal/telemetry/record.go` | 51 | Non-blocking `Record(cfg, e)` + `appendEvent` (`O_APPEND|O_CREATE|O_WRONLY`); overridable `now` |
| `internal/telemetry/constructors.go` | 37 | 5 typed one-liner helpers (Block/GateFailure/VerifyRejection/CompleteRejected/StepAdvanced) |
| `internal/telemetry/read.go` | 46 | Lenient `Read(dir)` (skip unparseable lines, missing file → nil,nil) + `ReadDefault()` |
| `cmd/centinela/telemetry_emit.go` | 40 | `emitGateFailures`, `emitVerifyRejection`, `toCheckRefs` (verify.Check → telemetry.CheckRef) |
| `cmd/centinela/complete_verify.go` | 26 | Extracted `runClaimVerification` (keeps complete.go ≤100; carries verify-rejection emit) |

Changed:

| File | Lines | Change |
|------|-------|--------|
| `internal/config/config.go` | — | Added `Telemetry TelemetryConfig` after `Memory` (no defaults — nothing to normalize) |
| `cmd/centinela/hook_prewrite.go` | 76 | `RecordBlock` before each `exitPrewrite(2)`: need-init (no feature/step) + out-of-step (full context) |
| `cmd/centinela/validate.go` | 97 | `emitGateFailures(cfg, results)` after `gates.RunWithFilter` |
| `cmd/centinela/complete.go` | 88 | `complete-rejected{gates}` / `{verify}` on the two aborts; `step-advanced` beside `memory.Capture` |
| `.gitignore` | — | `.workflow/telemetry/` (local-only in THIS repo; feature contract elsewhere is git-tracked) |

Slice → code: **S1** contract/storage/config = telemetry `{event,record,read}.go`
+ `config/telemetry.go` + `config.go` + `.gitignore`; **S2** gate-failure +
complete-rejected + step-advanced = `validate.go` + `complete.go`; **S3**
verify-rejection = `complete_verify.go`; **S4** block = `hook_prewrite.go`
(emitted only on genuine blocks, before `os.Exit`; `d.Allow` early-return keeps
the allow path zero-cost).

## Architecture Compliance

- **`internal/telemetry` is a config-only leaf**: imports only `internal/config`
  + stdlib. It owns its own `CheckRef` copy (no `internal/verify` import) and
  hardcodes the storage path (`telemetryDir`/`eventsFile`, no `internal/workflow`
  import).
- **Emission is `cmd/`-only**: domain types it reads (`hookpolicy.PrewriteDecision`,
  `gates.Result`, `verify.Check`) are unmodified; no domain→telemetry edge.
- **Unmapped-not-leaf decision**: `centinela.toml` `import_graph` was NOT edited.
  The `leaf` layer has `allow=[]`, so a `telemetry → config` edge (both leaf)
  would be a leaf→leaf violation. `internal/memory` already imports `config` and
  is likewise unmapped → a non-failing `⚠ import_graph` warning. Telemetry mirrors
  memory exactly. `validate` confirms import_graph stays `⚠` (warn), not `✖`.
- **All source files ≤100 lines** (largest validate.go 97, complete.go 88).

## Type-Safety Notes

Flat, fully-typed `Event` struct with `omitempty` tags — no `map[string]any`,
so JSON key order is deterministic (declaration order) and field names are
`go vet`-checkable. `gofmt`/`go vet` clean. `Enabled *bool` distinguishes
absent (default ON) from explicit `false`, matching `MemoryConfig`.

## Trade-Offs

- **Non-blocking contract**: `Record` returns nothing; nil/disabled cfg no-op;
  all I/O errors warn to stderr and are swallowed. No exit code / block decision
  / advance outcome changes — identical to `memory.Capture`. Block events are
  written *before* `exitPrewrite(2)` so an I/O failure still lets the block
  proceed.
- **Extracted `runClaimVerification` to `complete_verify.go`**: the
  verify-rejection emit pushed complete.go to 105 lines (>G1). Moved the
  self-contained helper to a sibling file (complete.go → 88) rather than grow it.
- **gate-failure carries no feature** (validate is not feature-scoped); the
  feature/step context lives in the paired `complete-rejected{gates}` event.

## Test, build & dogfood results

- `go build ./...` clean; `go vet ./...` clean; `gofmt -l internal cmd` empty.
- `go test ./...` → 1569 passed, 26 packages; no existing tests broken
  (telemetry unit/acceptance tests are qa-senior's job).
- Dogfood (`/tmp/cent-gt` from `./cmd/centinela`):
  - Out-of-step block (`code` file during `plan`) → exit 2 AND appended
    `{"schema":"centinela.telemetry/v1","type":"block",…,"feature":"dogfeat","step":"plan","reason":"out-of-step","fileType":"code","targetPath":"…/src/foo.go"}`.
  - Need-init block → `reason":"need-init"`, no feature/step (omitempty verified).
  - `[telemetry] enabled = false` → block still exits 2 but NO file/dir written
    (clean no-op).
- `centinela validate`: import_graph ⚠ (telemetry unmapped, expected); only ✖ are
  coverage 94.6%<95% and spec-traceability — both because telemetry has no tests
  yet (qa-senior's step).

## Handoff

**Next role:** qa-senior. The 17 scenarios in
`specs/governance-telemetry.feature` map 1:1 to Go tests. Testable seams:
overridable `now` (deterministic timestamps), `Read(dir)` accepting any dir,
disabled/nil-cfg no-op paths, and the lenient reader skipping garbage lines.
The 95% coverage gate currently fails only because telemetry is untested —
close it with real tests, not by lowering the gate.
