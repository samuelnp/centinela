# Senior-Engineer Report — headless-governance (code step)

## Summary

Implemented the two approved deliverables across two slices with zero behavior
change when headless is off (default zero-config path resolves to `false`).

## Slice 1 — Headless umbrella (non-interactive parity)

| Slice item | File | Lines | Notes |
|------------|------|------:|-------|
| `HeadlessConfig` + `IsHeadless` resolver | `internal/config/headless.go` (NEW) | 38 | Leaf-safe: imports only `os`/`strings`. Precedence: `CENTINELA_HEADLESS` env (authoritative) > `[headless] enabled` > (`detect_ci` AND `CI`). |
| Wire section into Config | `internal/config/config.go` | +1 | `Headless HeadlessConfig` toml:"headless". Zero value = off. |
| Review-prompt short-circuit | `cmd/centinela/hook_context_review_mode.go` | 34 | `if config.IsHeadless(cfg) { return false }` AFTER the nil/done guard, BEFORE `effectiveConfirmationMode` — headless wins over explicit `every_step`. |
| Plan-advisor short-circuit | `cmd/centinela/hook_plan_advisor.go` | 44 | `if config.IsHeadless(cfg) { return nil }` after cfg load, BEFORE the workflow loop — emits nothing and never loads workflows. |

## Slice 2 — Verdict packet + command

| Slice item | File | Lines | Notes |
|------------|------|------:|-------|
| Packet structs (schema `centinela.verdict/v1`) | `internal/verdict/packet.go` (NEW) | 70 | `Counts` is a struct (no marshaled maps). Verify status UPPERCASE; gate status lowercased. |
| Status mappers + tallies | `internal/verdict/mappers.go` (NEW) | 65 | `gateStatus`, `gateLine`, `checkLine`, `gateCounts`, `verifyCounts` (via `VerificationResult.Tally`). |
| Aggregator | `internal/verdict/assemble.go` (NEW) | 90 | `AssembleVerdict`: fail iff `!gates.AllPassed` OR `verify.HasFailures`. Profile via `workflow.EffectiveProfile`; archetype via `workflow.DisplayArchetype`; `Headless` via `config.IsHeadless`; `GeneratedAt` injected (no `time.Now` in-package). |
| On-disk evidence index | `internal/verdict/evidence_index.go` (NEW) | 37 | Iterates `evidence.AllRoles()`, `os.Stat` JSON, `evidence.Read`, sorts by role, returns empty (non-nil) slice when none. |
| `centinela verdict <feature>` cmd | `cmd/centinela/verdict.go` (NEW) | 60 | `SilenceUsage`/`SilenceErrors`. Wires real `gates.RunAll` + `verify.Verify` (`verifyRoot()`+`NewExecRunner()`) + `verdict.EvidenceIndex` + RFC3339 now. `MarshalIndent` to os.Stdout only; concise sentinel error on exit-1 → main prints to stderr, exits 1. |

## Layer / import-graph

`internal/verdict` is deliberately left UNMAPPED in `centinela.toml`
`[gates.import_graph]` — like `internal/verify`/`internal/ui` it is an
aggregator and surfaces only as the non-failing `import_graph` warning.
`internal/config` stays a strict leaf (headless.go imports only os/strings).

## Verification results

- `go build ./...` — clean.
- `go vet ./...` — no issues.
- `gofmt -l internal cmd` — empty.
- `go test ./...` — 1506 passed, 25 packages (green; unchanged from baseline).
- Every new/changed `.go` <=100 lines (max: assemble.go at 90).
- `centinela validate` — import_graph warning now also lists `verdict`
  (expected, non-failing). Coverage gate at 93.9% because new code is untested
  by design — qa-senior writes those tests next.

## Dogfood (binary /tmp/cent-hg built from ./cmd/centinela)

- `verdict <feature>` with evidence → valid JSON on stdout, schema
  `centinela.verdict/v1`, evidence sorted by role, exit 0, clean stderr.
- Forced G1 gate fail → exit 1, full JSON still on stdout, sentinel error on
  stderr only.
- Feature with no role JSON → `"evidence": []` (empty list, not null).
- `CENTINELA_HEADLESS=1 hook context` → review prompt suppressed.

## Handoff

Next role: **qa-senior** — resolver truth table, assembler golden JSON,
evidence index sort/empty, acceptance tests for the 26 scenarios; close the
coverage gate with colocated `_test.go` (<=100 lines each).
