### Senior-Engineer Report: code-quality-hardening
**Date:** 2026-06-10

#### Files Touched
| Path | Reason |
|------|--------|
| internal/hookpolicy/format_evidence_order.go | Added `"coverage"` between `mobileFirst` and `handoffTo` in `jsonKeyOrder`, mirroring `internal/evidence/schema.go`; rewrote doc comment to drop the false `format_evidence_test.go` claim and reference the key-order parity test (`TestEvidenceKeyOrderParity`) the tests step adds |
| internal/workflow/state.go | `Load()` reports "no workflow found" only on `errors.Is(err, fs.ErrNotExist)`; read failures wrap with `%w` naming the state file path; parse error now also names the path |
| internal/workflow/active.go | `ActiveWorkflows` no longer silently skips corrupted/unreadable state files — emits `workflow warning: <err>` to stderr and continues listing (least invasive: signature unchanged, all callers unaffected); non-workflow JSONs (evidence, roadmap) still skip silently via the feature-name mismatch check |
| cmd/centinela/start.go | Hard-fail on `config.Load()` error matching `complete.go`; error names centinela.toml (loader wraps it); fails before worktree provisioning and before any workflow state file is written |
| cmd/centinela/hook_context.go | On config error, inject one `config warning: <err>` line into the context output, continue with defaults, exit 0; load moved above the no-workflows early return so the warning always surfaces |
| cmd/centinela/hook_prewrite.go | Replaced silent `cfg, _ :=` with stderr `config warning:` + defaults; exit behavior unchanged |
| cmd/centinela/hook_plan_advisor.go | Replaced silent `cfg, _ :=` with stdout `config warning:` + defaults |
| cmd/centinela/hook_orchestration.go | Already fell back to defaults on error; now also prints `config warning:` so the failure is visible in injected context |
| cmd/centinela/hook_statusline_view.go | Deliberate silent fallback retained (now explicit `cfg, err :=` + comment): the statusline is a single-line protocol surface that cannot carry a warning line; the context hook surfaces the same failure every prompt |
| cmd/centinela/migrate.go | User-facing command: hard-fail on config error, moved load before any migration side effect is applied |
| scripts/check-fmt.sh | New executable format gate: `gofmt -l` over `cmd internal tests`, offenders to stderr + exit 1, silent exit 0 otherwise (mirrors check-coverage.sh env-override style via `FMT_DIRS`) |
| centinela.toml | Appended `./scripts/check-fmt.sh` to `[validate] commands` |
| 28 files (gofmt-only) | `gofmt -w` mechanical reformat, no semantic change: internal/evidence/schema.go, internal/ui/render_gates.go, internal/ui/render_status.go, internal/verify/runner.go, internal/worktree/merger.go + 23 `_test.go` files across internal/, tests/acceptance, tests/integration, tests/unit |

config.Load() call-site audit (all 10 sites): start.go, migrate.go → fail (changed); hook_context.go, hook_prewrite.go, hook_plan_advisor.go, hook_orchestration.go → warn (changed); hook_statusline_view.go → deliberate silent fallback (documented); complete.go, validate.go, verify.go → already fail (unchanged).

#### Architecture Compliance
- Boundary checks passed: no new internal imports anywhere; `internal/workflow` still imports only stdlib (+ leaf packages elsewhere); `internal/hookpolicy` still does NOT import `internal/evidence` (key order stays duplicated by design); cmd/ imports unchanged. `go vet ./...` clean.
- G1 file size: every touched source file ≤ 100 lines (largest: hook_context.go at 91); no exceptions added.
- G7 outer-layer rule: cmd/ changes are error-policy plumbing only (fail vs warn per surface); the missing-vs-corrupted classification logic lives in `internal/workflow.Load`.

#### Type-Safety Notes
- All previously discarded `error` values are now bound and handled; no `_ =` discards remain on `config.Load()` outside the documented statusline fallback (which still binds `err`).
- `errors.Is(err, fs.ErrNotExist)` used instead of `os.IsNotExist` for wrapped-error correctness.
- `config.Load()` returns defaults (nil error) when centinela.toml is absent, so hard-failing commands never break zero-config projects — only genuinely corrupted/unreadable TOML fails.

#### Trade-Offs
- `ActiveWorkflows` warns on stderr from the domain layer rather than changing its signature to return warnings: keeps all nine callers untouched; stderr never corrupts the hooks' stdout protocols. Verified zero spurious warnings against the real `.workflow/` directory (evidence and roadmap JSONs unmarshal cleanly and are rejected by the feature-name check, not by errors).
- Statusline hook keeps silent defaults — a warning line would corrupt the single-line statusline contract; the context hook covers visibility.
- `migrate` now loads config before building/applying plans (fail-early) instead of patching the discard in place at the end, so corrupted TOML cannot half-apply a migration.
- Behavior change: `centinela start` with corrupted TOML now fails (previously proceeded with empty config). Intended per plan; flag for release notes.

#### Handoff
- Next role: qa-senior
- Outstanding TODOs: write `TestEvidenceKeyOrderParity` in `internal/hookpolicy/format_evidence_parity_test.go` (external `package hookpolicy_test`, byte-compare vs `evidence.MarshalJSON`) — the rewritten doc comment references it; unit tests for check-fmt.sh (use `FMT_DIRS` override against a temp tree), start/hook corrupted-TOML scenarios, and the three `workflow.Load` scenarios (missing / corrupted / unreadable — chmod fixture, skip as root); acceptance artifacts for all nine spec scenarios; colocated `_test.go` coverage for changed cmd/ and internal/ lines (per-package 95% gate, no -coverpkg). Verified here: gofmt -l clean, go vet clean, go build OK, go test ./... 1220 passed (no existing tests needed adjusting).
