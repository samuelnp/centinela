# adoption-baseline — senior-engineer

## Files Touched

| File | Layer | Role | Lines |
|------|-------|------|------:|
| `internal/audit/adopt.go` | aggregator | `Adopt(cfg, force) (Outcome, error)` + `Outcome` type + `Baseline.Total()`; owns the skip-if-exists rule (G7) | 45 |
| `internal/ui/render_adopt.go` | presentation | `RenderAdoption(audit.Outcome) string` — human adoption report (header, per-gate counts, total, ratchet-to-zero framing, 0-finding case) | 32 |
| `cmd/centinela/adopt.go` | cmd (outer) | thin Cobra `adoptCmd` registered on `rootCmd`; `--force`/`--json` flags; maps `Outcome` → exit/render only | 57 |
| `cmd/centinela/adopt_render.go` | cmd (outer) | `adoptVerdict` JSON struct + `printAdoptJSON`; mirrors `audit_render.go` | 52 |

All four ≤100 lines. No other files modified (no config, no gate logic, no baseline format change).

## Architecture Compliance

- **G2 (import graph):** zero new edges. `internal/audit/adopt.go` only uses same-package `Load`/`Record`/`Save` + the already-imported `internal/config`. `internal/ui/render_adopt.go` uses `internal/audit`, an edge `render_audit.go` already establishes. `cmd/**` may import aggregator/leaf/presentation — already satisfied. `go build ./...` and `go vet ./...` clean confirm no cycle/edge violation.
- **G1 (file size):** 45 / 32 / 57 / 52 lines — all comfortably ≤100, no G1 exception needed.
- **G7 (no business logic in outer layer):** the skip-if-exists decision (`exists && !force` → no write) lives entirely in `audit.Adopt`. The cmd only branches on `o.Skipped`/`adoptJSON` to choose render vs error/exit code; it performs no `Load`/existence reasoning of its own. `Adopt` provably never calls `Save` on the skip path (early `return Outcome{Skipped:true, Path:path}`), so the file is left untouched.

## Type-Safety Notes

Strict Go throughout: no `any`, no `interface{}` in new code, no reflection. `Outcome` is a concrete struct; `adoptVerdict` uses concrete field types (`map[string]int`, not `map[string]any`). `perGateCounts` returns a non-nil empty map so `--json` on skip renders `"per_gate": {}` rather than `null`. JSON marshalling reuses the stdlib `encoding/json` + struct tags pattern from `audit_render.go`.

## Trade-Offs

- **Exit code on skip:** per the senior-engineer brief (and matching the spec's "exits with a non-zero code" assertions for both text and `--json` skip), `adopt` exits non-zero on skip-without-force in BOTH modes. This is a deliberate divergence from the plan's earlier Decision #4 prose ("Exit 0 on both adopt and skip"); the spec scenarios (lines 65 and 113) and the build brief are authoritative and require non-zero on skip. Implemented accordingly: `printAdoptJSON` prints the skipped verdict then returns a non-nil error.
- **`Baseline.Total()` exported helper:** added an exported method on `Baseline` rather than duplicating the cmd's `countFingerprints` free function, so both `render_adopt.go` and `adopt_render.go` reuse one canonical counter. `countFingerprints` in `audit_baseline.go` is left untouched to avoid churn outside this feature's surface.
- **No merge/edit of existing entries:** `--force` fully re-records via `Save` (replace), never merges — matches out-of-scope in the feature-specialist report.

## Type-Safety / Build Evidence

```
$ go build ./...        → Success
$ go vet ./...          → No issues found
$ gofmt -l internal/audit internal/ui cmd/centinela   → (empty)
line counts: 45 internal/audit/adopt.go | 32 internal/ui/render_adopt.go | 57 cmd/centinela/adopt.go | 52 cmd/centinela/adopt_render.go
over-100 scan (diff + untracked .go)  → (none)
$ centinela evidence validate adoption-baseline  → evidence ok for "adoption-baseline"
```

Smoke test (temp binary `/tmp/cent-adopt`, throwaway git repo; plus a colocated byte-identity Go check removed afterward):
- (a) first `adopt` → wrote `.workflow/audit-baseline.json`, printed report, **exit 0**.
- (b) second `adopt` (no force) → `baseline already exists … use --force to overwrite`, **exit 1**, SHA before==after (**byte-unchanged**).
- (c) `adopt --force` → re-recorded, printed report, **exit 0**.
- (d) `adopt --json` on existing → `{adopted:false, skipped:true, per_gate:{}}`, **exit 1**, file byte-unchanged; `adopt --json --force` → `{adopted:true, skipped:false}`, **exit 0**.
- byte-identity: `Adopt` output == `Record`+`Save` reference output (equal bytes); skip path leaves bytes unchanged.
- worktree left clean: no stray `audit-baseline.json`; `git status` shows only the four new impl files.

## Handoff

Next role: **qa-senior**. Implement the unit/renderer/cmd tests + acceptance per `docs/plans/adoption-baseline.md` test strategy, tracing each spec scenario with `// Scenario: <name>`. Key seams to assert:
- `audit.Adopt`: skip-if-exists returns `Skipped:true` **without writing** (assert file mtime/bytes unchanged); `force=true` overwrites; fresh repo records the full set; load/save error propagation.
- Determinism: `Adopt` output byte-identical to `Record`+`Save` on the same repo.
- `ui.RenderAdoption`: per-gate counts + total + ratchet-to-zero framing on non-empty; `"0 accepted findings — nothing to ratchet."` on empty.
- cmd: `adopt` on existing baseline **exits non-zero** with "use --force" (text); `--force` rewrites exit 0; `--json` shape `{adopted, skipped, path, total, per_gate}` with empty `per_gate:{}` on skip and **non-zero exit** on skip in JSON mode.
- Keep test files ≤100 lines (G1 applies to `_test.go` under `internal/` and `cmd/`); colocate `_test.go` to cover the cmd/internal code for the per-package coverage gate (no `-coverpkg`).

Deferred findings: none.
