### Big-Thinker Report: code-quality-hardening
**Date:** 2026-06-09

#### Problem
- A full-repo quality review found four defects that the current mechanical gates miss: (1) `internal/hookpolicy/format_evidence_order.go` duplicates the evidence `jsonKeyOrder` but omits `"coverage"`, so the postwrite hook reorders keys in coverage-bearing evidence files and breaks the byte-stable serialization invariant — and its doc comment claims a parity test that does not exist; (2) `gofmt -l` flags 23 files (including 5 non-test sources) and nothing in `[validate] commands` gates formatting; (3) `cmd/centinela/start.go:37` and `cmd/centinela/hook_context.go:37` silently swallow `config.Load()` errors while `complete.go:31` hard-fails, so a corrupted `centinela.toml` lets a feature start with empty config and then fail at complete; (4) `internal/workflow/state.go` `Load()` masks all read failures (including permission errors) as "no workflow found". All four were independently verified against source before this report.

#### Scope
- In: add `"coverage"` to the hookpolicy `jsonKeyOrder` (between `mobileFirst` and `handoffTo`); correct the false doc comment; add a real behavior-level parity test (`package hookpolicy_test`, marshal via `evidence.MarshalJSON`, byte-compare against the hook formatter); `gofmt -w` the 23 flagged files; add `scripts/check-fmt.sh` and append it to `[validate] commands`; hard-fail `start` on config load errors matching `complete.go`; warn-and-continue in `hook_context.go` (hooks must not break the host session); distinguish `fs.ErrNotExist` from read failure in `workflow.Load()` and name the file path in parse/read errors; audit other `config.Load()` call sites for the same silent-discard pattern.
- Out: CWD-relative path architecture refactor; test-suite quality / coverage-padding cleanup; consolidation of the three shell-exec wrappers; any new built-in gate type (the format check rides `[validate] commands`).

#### Dependencies & Assumptions
- Canonical key order lives in `internal/evidence/schema.go:40-43`; hookpolicy mirror is `internal/hookpolicy/format_evidence_order.go:14-17`.
- Correction to the brief/plan rationale: today **neither** `internal/evidence` imports `internal/hookpolicy` nor vice versa (verified by grep); the existing doc comment's cycle claim is wrong as a current fact. The external `hookpolicy_test` package is still the right home for the parity test — it stays cycle-proof if `evidence` ever does grow a hookpolicy dependency.
- Correction to the brief: `workflow.Load()` does NOT mask parse failures — `state.go:46` already returns `"invalid workflow file: %w"`. Only the `os.ReadFile` branch (`state.go:40-43`) masks everything as "no workflow found". The fix still applies (the parse error also lacks the file path the spec requires), but the defect is narrower than stated.
- `[validate] commands` in `centinela.toml` currently has three entries (go test, acceptance, check-coverage.sh); `./scripts/check-coverage.sh` is the precedent for a wrapper-script command, since gofmt itself always exits 0.
- G1 applies to test files too: the new parity test and any extracted hook_context helper must stay ≤100 lines.
- Per-package coverage gate (no -coverpkg): new behavior in `cmd/` and `internal/` needs colocated `_test.go` files to hold the 95% line.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Hard-failing `start` on corrupted TOML is a behavior change | Users with broken config blocked at start instead of at complete | Medium | Intended fix; call out in release notes; error message names `centinela.toml` |
| Mass `gofmt -w` touches 23 files, inflating the diff | Review noise; merge conflicts with sibling worktrees | High | Land formatting as its own commit; no semantic change; check-fmt.sh keeps tree clean after |
| New check-fmt.sh validate command fails in CI on stragglers | Red CI for the whole repo | Low | Run the script locally on the formatted tree before wiring it into `[validate]` |
| hook_context warning path pushes file over 100 lines (G1) | Gate failure | Medium | Extract warning helper into its own ≤100-line file |
| Parity test placed in internal test package re-creates a future cycle hazard | Build break if evidence ever imports hookpolicy | Low | Use external `package hookpolicy_test` as planned |
| Existing on-disk evidence written with the buggy order | Byte-diff churn when hook reformats older files | Low | Reorder is one-time and converges to canonical order; no data loss |

#### Rollout
- Step 1: Fix `jsonKeyOrder` drift + correct the doc comment + add the behavior-level parity test (smallest correct slice).
- Step 2: `gofmt -w` all 23 flagged files in a standalone commit; add `scripts/check-fmt.sh`; append it to `[validate] commands`; confirm `centinela validate` is green.
- Step 3: Align config-error policy — `start.go` hard-fails like `complete.go`; `hook_context.go` injects a `config warning:` line and exits 0; audit remaining `config.Load()` call sites.
- Step 4: `workflow.Load()` — `errors.Is(err, fs.ErrNotExist)` → "no workflow found"; otherwise wrap with `%w` naming the file path; review `ActiveWorkflows` callers so corrupted state surfaces rather than being skipped.
- Step 5: Acceptance tests for the seven spec scenarios; full validate.

#### Handoff
- Next role: feature-specialist
- Outstanding questions: (a) Should the corrected doc comment in `format_evidence_order.go` drop the (currently false) import-cycle rationale entirely and justify the duplication purely on layer thinness? (b) For `ActiveWorkflows`, is a stderr warning sufficient for corrupted state files, or should listing hard-fail? (c) Does the gofmt mass-format commit need coordination with other open worktrees to limit conflicts?
