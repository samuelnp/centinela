# hook-context-panel-diet — senior-engineer

## Files Touched

| File | Change |
|---|---|
| `internal/ui/panel.go` | Added `trimTrailingWS(s string) string` helper (strips trailing spaces/tabs per line) + `strings` import |
| `internal/ui/render.go` | Same-line wrap: `RenderContextCapped`'s return now `trimTrailingWS(renderSystemPanel(...))` |
| `internal/ui/render_review.go` | Same-line wrap of 4 returns: `RenderReviewReady`, `RenderEdgeCaseReportNeeded`, `RenderDocumentationNeeded`, `RenderChangelogNeeded` |
| `internal/ui/render_brief.go` | Same-line wrap: `RenderFeatureBriefNeeded`'s return |

No other files changed. `RenderStatus`, `RenderRoadmap`, `RenderBlocked`, `RenderBlockedStaleBinary`, and every other CLI-only or deferred-emitter renderer are untouched.

## Re-verification of the plan's zero-non-hook-caller claim

Re-ran the plan's grep independently (`grep -rn "<Func>(" --include='*.go' . | grep -v _test.go`) for all six functions before editing. Confirmed: every one has exactly one non-test, non-definition call site, and it is in `cmd/centinela/hook_context.go`:
- `RenderContextCapped` → `hook_context.go:44` (plus its own internal use by `RenderContext`, itself unused outside `render.go`)
- `RenderReviewReady` → `hook_context.go:47`
- `RenderFeatureBriefNeeded` → `hook_context.go:53`
- `RenderEdgeCaseReportNeeded` → `hook_context.go:63`
- `RenderDocumentationNeeded` → `hook_context.go:72`
- `RenderChangelogNeeded` → `hook_context.go:76`

The plan's safety argument holds: no shared caller exists to disambiguate hook-vs-CLI intent for, so wrapping the return unconditionally inside each function is safe.

## Architecture Compliance

- **Archetype:** n-tier (PROJECT.md). `internal/ui` is the presentation layer; changes stay entirely inside it — no new cross-layer imports (only added stdlib `strings` to `panel.go`, already imported by `render.go`/`render_brief.go`).
- **G7 (outer layer):** not implicated — `cmd/centinela/hook_context.go` (outer layer) is unchanged; all logic lives in `internal/ui`.
- **G1 (≤100 lines):** all four touched files measured after edit:
  - `internal/ui/panel.go`: **57 lines** (was 39)
  - `internal/ui/render.go`: **100 lines** (was 100 — net-zero same-line wrap, exactly at the G1 ceiling as the plan anticipated; did not reformat the return across multiple lines, no new import added)
  - `internal/ui/render_review.go`: **54 lines** (was 54 — same-line wraps only)
  - `internal/ui/render_brief.go`: **30 lines** (was 30 — same-line wrap only)
  No file exceeds 100 lines; no G1 exception needed.

## Type-Safety Notes

Go's static typing already enforced correctness here — `trimTrailingWS(s string) string` is a pure, fully-typed string transform with no interface{}/any usage. No new type surface introduced.

## Trade-Offs

- Fix lives at the render-function return point in `internal/ui`, not in the emitter (`hook_context.go`) or via a hook/TTY-mode flag — avoids duplicating post-processing across six `fmt.Println` call sites in the outer layer, and needs no flag because caller intent is already expressed structurally (each function has exactly one caller).
- No content cut: step ladder, nudge panel self-identification (`<feature> · <step>`), and the `docs/features/*.md` skeleton headings in `RenderFeatureBriefNeeded` were all evaluated and kept (see plan §3) — they are load-bearing per-turn governance signal, not decorative padding.
- `RenderStatus`/`RenderRoadmap` and other CLI-facing renderers deliberately keep their own residual `JoinVertical` padding — out of scope per the brief's constraint ("changes what the HOOK emits, not how the CLI renders to a TTY"), confirmed unaffected by measurement below.

## Deferred Findings (not filed — orchestrator to file from primary checkout)

- **`hook-panel-diet-remaining-emitters`** — `hook_prewrite_block.go` (`RenderBlocked`, `RenderBlockedStaleBinary`), `hook_merge.go`, `hook_migrate.go`, and the `hook setup` family (`hook_setup.go`, `hook_setup_setup.go`, `hook_setup_cascade.go`) all build panel bodies via the same `lipgloss.JoinVertical(lipgloss.Left, ...)` pattern and carry the same residual trailing-whitespace padding, but none fire on every `UserPromptSubmit` turn (rare/gated events), so they are outside this feature's measured leak. Same fix shape (`trimTrailingWS` wrap) applies directly as a narrow follow-up.

## Seam for qa-senior (tests step)

- The size guard (`internal/ui/panel_budget_test.go`, per plan §4) is a **tests-step** artifact — `tests/` writes are blocked during the code step, and the plan explicitly scopes the guard's test file to the tests step. It should call `RenderContextCapped` + `RenderReviewReady` + `RenderFeatureBriefNeeded` directly against a fixed, deterministic in-memory fixture (`workflow.Workflow{Feature: "sample-feature", CurrentStep: "plan"}`, canonical 5-step order via `StepOrder` left empty so `OrderedSteps()` falls back to `workflow.DefaultStepOrder`) — never the live repo/filesystem state.
- Two invariants: (1) zero trailing whitespace, exactly, on every line; (2) byte budget with headroom, not exact pinning.
- Also per plan §7 (already audited): re-run `go test ./internal/ui/... ./cmd/centinela/...` after landing the guard and grep the diff for any newly-red substring assertion that happened to depend on trailing spaces — none were found in this plan-step/code-step audit (all existing tests use `strings.Contains`/`mustContain`/`mustNotContain`, which trailing-whitespace trimming cannot break), but re-verify rather than trust.

## Handoff

**Next role: qa-senior.**

**Measured before/after** (own reproduction of the plan's §0 scenario — `RenderContextCapped` + `RenderReviewReady` + `RenderFeatureBriefNeeded` against the fixed fixture `workflow.Workflow{Feature: "sample-feature", CurrentStep: "plan"}`, canonical step order, measured via a throwaway colocated `_test.go` harness built and run, then deleted before evidence/report — not committed):

| | Total bytes | Content (no trailing WS) | Trailing WS | % trailing |
|---|---|---|---|---|
| Before (git-stashed pre-fix code) | 2004 | 1195 | 809 | 40.4% |
| After (this fix) | 1195 | 1195 | 0 | 0.0% |

Content bytes are identical before/after (1195 = 1195), confirming the fix removes only trailing whitespace and zero visible content. Magnitude is consistent with the plan's own measurement (2203/1249/954/43.3%) — the small difference is fixture-shape (different `next` step string, different join separator), not a different mechanism.

Live sanity check in this repo (`echo '{}' | <binary> hook context`, current active workflow at plan step, brief already exists so only `RenderContextCapped` fires): installed binary 234 bytes → scratch (fixed) binary 210 bytes.

**Human-facing output diff (installed vs. scratch `/tmp/cent-diet` built from this worktree):**
- `centinela status hook-context-panel-diet`: **IDENTICAL** (byte-for-byte `diff` empty).
- Blocked-write scenario (`hook prewrite` on a `tests/` file during the `code` step): **IDENTICAL** (byte-for-byte `diff` empty, exit code 2 both), and the output visibly still carries its original trailing-whitespace padding (`RenderBlocked` untouched, as scoped).

**Breakage inventory:** none. `go build ./...`, `go vet ./...` clean. `go test ./... -run xxxNONE` found zero test-compile breaks across the whole module. `go test ./internal/ui/... ./cmd/centinela/...` — 791 tests passed, no failures, no new red substring assertions.

**Deviations from plan:** none. Rollout followed Slice 1 exactly (§6): `panel.go` gets the helper, four call sites in `render_review.go`, one each in `render.go` and `render_brief.go`, same-line wraps only, `render.go` net-zero at exactly 100 lines.
