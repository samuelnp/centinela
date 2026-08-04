# Plan: hook-context-panel-diet

**Brief:** [docs/features/hook-context-panel-diet.md](../features/hook-context-panel-diet.md)
**Spec:** [specs/hook-context-panel-diet.feature](../../specs/hook-context-panel-diet.feature)
**Archetype:** canonical

## 0. Re-measured baseline (the brief's numbers, verified against this repo today)

The brief's numbers were **not reproduced** by a bare `centinela hook context`
run right now (single active workflow, plan step): that call emits only
**234 bytes**, of which ~24 are trailing whitespace, because none of the
per-turn nudge panels fire for this exact workflow today (its own feature
brief already exists, and its plan artifacts are not yet complete, so neither
`RenderFeatureBriefNeeded` nor `RenderReviewReady` renders).

Root-caused directly by calling the render functions with the brief's stated
scenario (single workflow, plan step, `FEATURE BRIEF REQUIRED` panel active —
the condition that applies to most real plan-step turns, since a workflow
spends most of its plan step *before* the brief exists):

| Combination | Total bytes | Content (no trailing WS) | Trailing WS | % |
|---|---|---|---|---|
| `RenderContextCapped` + `RenderReviewReady` + `RenderFeatureBriefNeeded` | 2203 | 1249 | 954 | 43.3% |

This confirms the brief's mechanism and magnitude (its 1770/1034/736/41.6%
was measured under a similarly nudge-heavy turn); the exact byte counts drift
with which nudge panel is active and current roadmap counts, but the
40%+ trailing-whitespace tax is real, reproducible, and specific to hook
output that composes a body via `lipgloss.JoinVertical(lipgloss.Left, ...)`.

**Root cause, precisely:** `remove-panel-borders` (commit `3ab3cef`, already
on `main`) deleted the *outer* fixed-width bordered box
(`panelStyle(t).Render(content)` in `internal/ui/panel.go`) that used to
square every panel to a fixed terminal column width — that was the original
"93 columns" mechanism. It did **not** touch the *inner*
`lipgloss.JoinVertical(lipgloss.Left, ...)` calls inside each individual
panel-body constructor (`render_brief.go`, `render_review.go`,
`RenderContextCapped` in `render.go`). `JoinVertical` still right-pads every
line in its argument list to the width of that call's longest line — that
residual, per-panel padding is the entire remaining leak. There is no
bordered rectangle left to square; the padding is dead weight inherited from
a mechanism that no longer exists.

Confirmed via `git show 3ab3cef -- internal/ui/panel.go`: the diff shows
exactly `panelStyle(t).Render(content)` → `renderSystemLine(...) + "\n\n" +
body` — an outer-box removal only.

## 1. Emitter inventory — every `hook_*.go` that writes to stdout/stderr

Grepped every non-test `cmd/centinela/hook_*.go` for `ui.Render*` /
`fmt.Println` / `fmt.Fprintln(os.Std*)`. Classified by whether the callee
render function builds its body via `lipgloss.JoinVertical` (padding-bearing)
and whether that function has any non-hook caller (checked via
`grep -rn "<Func>(" --include='*.go' . | grep -v _test.go`— every function
below has **zero** callers outside its listed hook file):

| Hook file | Render calls | Padding-bearing? | Frequency | In scope here? |
|---|---|---|---|---|
| `hook_context.go` | `RenderContextCapped`, `RenderReviewReady`, `RenderFeatureBriefNeeded`, `RenderEdgeCaseReportNeeded`, `RenderDocumentationNeeded`, `RenderChangelogNeeded`, `RenderSuccess` | Yes (all but `RenderSuccess`, single-line) | **Every `UserPromptSubmit`** | **Yes — this is the measured leak** |
| `hook_context_roadmap.go` | `RenderRoadmapSummary` | No (single line, `renderSystemLine` only) | Digest-deduped, per session | No — already clean |
| `hook_postwrite.go` | `RenderTag` | No (single line) | Every write | No — already clean |
| `hook_prewrite_block.go` | `RenderBlocked`, `RenderBlockedStaleBinary` | Yes | Only on a blocked write (rare) | **No — deferred** (same root cause, not the measured "every turn" leak; also dual-audience — human terminal echo AND agent tool-feedback — needs its own risk read) |
| `hook_merge.go` | `RenderMergeStewardNeeded` | Yes | Only when merge-steward fires | **No — deferred** |
| `hook_migrate.go` | `RenderMigrationNeeded` | Yes | Only pre-migration | **No — deferred** |
| `hook_setup.go`, `hook_setup_setup.go`, `hook_setup_cascade.go` | `RenderRoadmapNeeded`, `RenderRoadmapAnalysisNeeded`, `RenderRoadmapQualityNeeded`, `RenderProductionReadinessSetupNeeded`, `RenderBrownfieldSetupNeeded`, `RenderSetupNeeded`, `RenderRoadmapJSONNeeded`, `RenderRoadmapCheckpoint` | Yes | Only pre-roadmap / setup phase | **No — deferred** |
| `hook_statusline.go` | `RenderStatusLine` | No (measured 111B live, 0 trailing) | Every prompt (statusline) | No — already clean, matches brief |
| `hook_session.go` | `RenderSessionRehydration` | Verified clean live (1093B total / 1091B stripped ≈ 0 trailing) | Once per session | No — brief already scopes this out, confirmed |
| `hook_plan_advisor.go`, `hook_orchestration.go`, `hook_autostart.go`, `hook_prewrite.go` | plain strings / `planadvisor.Directive` | No (no `ui.Render*` body composition) | n/a | No |

**Scope decision:** fix only `hook_context.go`'s six padding-bearing calls —
this is the entire measured "every turn" leak, and every other padding-bearing
emitter fires on rare, gated events, not on every prompt. Fixing them too
would be free of *measured* justification (an inflated plan); they share the
identical root cause and identical fix shape, so a follow-up feature can apply
the same one-line wrap to each. Logged as a deferred finding below.

## 2. Decision 1 — where to cut the padding

**At the render-function return point, in `internal/ui`, not at the emitter
(`cmd/centinela/hook_context.go`) and not via a hook/TTY-mode flag.**

Why not the emitter: `hook_context.go` would need to post-process six
different `fmt.Println(ui.Render...(...))` call sites with the same
string-processing logic duplicated (or piped through a shared local helper) —
strictly worse than fixing it once where the string is built.

Why not a hook-specific render *mode* (a bool/enum threaded through
`renderSystemPanel`): **unnecessary, because it is already unambiguous.**
Every one of the six functions being touched
(`RenderContextCapped`, `RenderReviewReady`, `RenderFeatureBriefNeeded`,
`RenderEdgeCaseReportNeeded`, `RenderDocumentationNeeded`,
`RenderChangelogNeeded`) has **zero callers outside `hook_context.go`**
(verified by grep, see §1). There is no shared caller to disambiguate for —
the "caller's intent" the brief asks about is already expressed structurally,
by which function you call, not by a flag on a shared one. `renderSystemPanel`
itself (the shared low-level primitive) and `RenderStatus` /
`RenderRoadmap` / other CLI-facing renderers are **not touched** — they keep
emitting exactly the bytes they do today, including their own residual
`JoinVertical` padding (confirmed live: `centinela status` output today still
right-pads `Feature`, `Started`, `Profile` etc. to the widest row). That
padding is arguably just as dead as the hook-side padding (no border remains
there either), but the brief's constraint is explicit — *"changes what the
HOOK emits, not how the CLI renders to a TTY"* — so leaving `RenderStatus`
alone is a deliberate, conservative choice, not an oversight, and it keeps the
diff minimal and trivially revertible.

**Mechanism:** one new helper in `internal/ui/panel.go`:

```go
// trimTrailingWS strips trailing spaces/tabs from every line of a hook-only
// panel render. lipgloss.JoinVertical(lipgloss.Left, ...) right-pads every
// line in a block to the width of that block's longest line — a leftover of
// the fixed-width bordered-box era (removed in remove-panel-borders,
// 3ab3cef) with no remaining visual purpose, since these panels render as
// flat text with no rectangle left to square.
func trimTrailingWS(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.Join(lines, "\n")
}
```

Applied as a **same-line wrap of the existing return statement** in each of
the six functions, e.g. in `render.go`:

```go
// before: return renderSystemPanel("HOOK", "ACTIVE WORKFLOWS", toneInfo, body)
return trimTrailingWS(renderSystemPanel("HOOK", "ACTIVE WORKFLOWS", toneInfo, body))
```

**G1 constraint — read before touching `render.go`:** `internal/ui/render.go`
is **exactly 100 lines today** (the G1 ceiling). `panel.go` is 39,
`render_review.go` is 54, `render_brief.go` is 30 — all with headroom.
`trimTrailingWS` lives in `panel.go` (same package `ui`, no new import
needed anywhere). The wrap in `render.go` must stay a same-line edit of the
existing `return` statement — do **not** reformat it across multiple lines,
and do not add an import (there is nothing to import). Net line delta for
`render.go`: **0**. If a future edit to `render.go` needs even one more line
for any reason, split first, exactly as `token-diet` did for
`hook_context.go` (98→ lines) and `capability.go` (100 lines) — this file is
already at the same cliff edge.

## 3. Decision 2 — content redundancy audit

**Conclusion: no further content is safe to cut. The padding removal is the
entire win.**

Checked each candidate concretely:

- **The ACTIVE WORKFLOWS step ladder** (`▶ plan · ○ code · ○ tests · ○
  validate · ○ docs`) looks redundant against the header's
  `<feature>  <step> <done>/<total>` — until archetype is accounted for.
  `internal/workflow/archetype_order.go` defines **four different step
  orders**: canonical (5: plan/code/tests/validate/docs), hotfix (3:
  code/tests/validate — no plan), refactor (4: no docs), spike (2: plan/code
  — no validate, ungated by construction). The header's `<done>/<total>`
  count alone cannot disambiguate "code 0/3" hotfix from "code 1/4" refactor
  from a canonical workflow with 3 steps skipped some other way — the ladder
  is the only per-turn signal of **which concrete steps this workflow will
  run**, including whether a validate gate applies at all (spike). Cutting it
  risks the model assuming a gate exists when the archetype has none, or
  vice versa. Not cut.
- **The nudge panels' repeated `<feature> · <step>`** (e.g.
  `RenderReviewReady`'s `"⏸ "+feature+" · "+step+" artifacts complete"`)
  duplicates the ACTIVE WORKFLOWS header above it for a *single* workflow.
  But `hook_context.go` renders nudge panels in a flat loop across **all**
  active workflows (cap 5) — each panel is not nested under its workflow's
  header, so with >1 active workflow, a panel's self-identification is the
  only way to know which workflow it is about. Not cut (and not scaled by
  padding removal alone, so out of the measured leak anyway).
- **The `docs/features/<feature>.md` skeleton headings inside
  `RenderFeatureBriefNeeded`** (`## Problem`, `## User Stories`, etc.) look
  like static ritual the model has seen before — but this panel fires
  precisely when a brief does **not** exist yet, i.e. for a model that has
  *not* necessarily seen this checklist this session (a fresh context, a
  different agent). Removing it trades a one-time nudge that fully specifies
  the required document shape for a bare "write the brief" instruction the
  model would have to reconstruct from memory or CLAUDE.md. Not cut.

If the honest answer were "nothing more is safe to cut," this is that answer:
the padding was the win; the surviving content is load-bearing.

## 4. Decision 3 — the size guard

**What it measures:** the byte output of the same three-function combination
used in §0 (`RenderContextCapped` + `RenderReviewReady` +
`RenderFeatureBriefNeeded`), called **directly against a fixed, deterministic
in-memory fixture** — not the live repo, not a filesystem-backed active-workflow
count. Fixture: a single synthetic `workflow.Workflow{Feature: "sample-feature",
CurrentStep: "plan"}` with the canonical 5-step order, fed straight to the
three render functions (no `t.TempDir()`, no `.workflow/roadmap.json`, no
`os.Chdir`). This is deliberately **not** keyed to this repo's real active
workflow count or its 130 feature briefs — either would flake as the repo
grows or as other engineers' worktrees carry different active-workflow state.

**Budget:** measured post-fix baseline for this exact fixture is **1249
bytes** (the "no trailing whitespace" figure from §0 — after the fix, total
bytes *equals* that figure, since trailing-whitespace was the only delta).
Guard asserts two things, both zero-tolerance where possible:

1. **Zero trailing whitespace**, exactly: for every line in the combined
   output, `line == strings.TrimRight(line, " \t")`. No slack — this is the
   sharp, deterministic invariant the whole feature exists to establish, and
   it can't drift for incidental reasons (unlike a byte count, which shifts
   if copy changes).
2. **Byte budget with headroom, not exact pinning:** `len(combined) <= 1400`
   bytes (1249 measured + ~12% headroom for legitimate small copy changes,
   e.g. a clarified sentence in a nudge panel) but tight enough that
   reintroducing the `JoinVertical` padding (which would roughly double the
   byte count per §0's before/after) fails loudly. A comment at the constant
   states the measured baseline and the date, per the "prompt-doc-budget"
   style precedent (Backlog: `prompt-doc-budget-ratchet` — no implementation
   exists yet to reuse; this guard is the first of its kind in this repo, so
   it sets rather than follows a pattern).

**Where it lives:** `internal/ui/panel_budget_test.go` (new, colocated,
package `ui`, calling the render functions directly — cheapest and most
deterministic; no binary build, no filesystem, no `os.Chdir`). It fails a
normal `go test ./internal/ui/...` run, which `centinela validate` already
runs — no new gate wiring needed.

**Why not the built binary / `tests/acceptance/`:** `tests/` and `specs/`
writes are blocked during the code step (Centinela hook policy — new test
files land in the tests step). The `internal/ui` colocated test is written in
the tests step alongside the rest of this feature's test suite; nothing here
requires it to exist before code lands. (The spec scenario for the guard is
still written now, in `specs/hook-context-panel-diet.feature`, since specs
are a plan-step artifact — the guard's *test file* is a tests-step artifact.)

## 5. Decision 4 — ANSI/colour

**No work needed. Verified independently, not just trusted from the brief.**

- `grep -rn "NO_COLOR\|ColorProfile\|SetColorProfile\|termenv\|lipgloss.NewRenderer"` across the whole non-test tree: **zero matches**. There is no explicit color-profile forcing anywhere in this codebase.
- Live measurement, piped (as a real hook invocation is): `centinela hook context < /dev/null | od -c | grep -c '033'` → **0** escape sequences.
- Live measurement, **under a real PTY** (via a `pty.openpty()` harness, not just a non-tty pipe — this is the check the brief asked me not to skip): still **0** escape sequences, and no `\x1b` byte anywhere in the captured output.

So `lipgloss`'s default auto-detection is already resolving to a no-color
profile in this environment regardless of TTY attachment (most likely `TERM`
in this sandbox not signaling color support) — but the operative fact for
this feature is simpler and doesn't depend on that explanation: **the code
path already contains zero ANSI bytes, on both the hook path and under a
real terminal**, so there is nothing to strip and nothing to force off. Adding
color-stripping logic would be solving a problem that doesn't exist on this
path today.

## 6. Rollout sequence

1. **Slice 1 (padding removal) — independently revertible.**
   `internal/ui/panel.go`: add `trimTrailingWS`.
   `internal/ui/render.go`: wrap `RenderContextCapped`'s return (same-line).
   `internal/ui/render_review.go`: wrap `RenderReviewReady`,
   `RenderEdgeCaseReportNeeded`, `RenderDocumentationNeeded`,
   `RenderChangelogNeeded`'s returns (same-line each).
   `internal/ui/render_brief.go`: wrap `RenderFeatureBriefNeeded`'s return
   (same-line). No other files change. This slice alone delivers the
   measured ~40%+ reduction and is a single, small, revertible commit.
2. **Slice 2 (size guard).** `internal/ui/panel_budget_test.go`, written in
   the tests step, asserting the zero-trailing-whitespace + byte-budget
   invariants from §4 against the fixed fixture.
3. **Slice 3 (existing-test audit).** Run the full suite; the search
   documented in §7 below found no exact byte-identity assertions to update,
   but the tests step must still run `go test ./internal/ui/... ./cmd/centinela/...`
   and grep the diff for any newly-red substring assertion that happened to
   depend on trailing spaces (e.g. a test asserting a specific column width) —
   none were found in this plan-step audit, but the tests-step agent should
   not skip re-running this check after the code lands.

No slice touches `RenderStatus`, `RenderRoadmap`, or any other CLI-only
renderer — human-facing `centinela status` / `validate` / roadmap output is
byte-for-byte unchanged by this feature.

## 7. Byte-identity test search (breakage list)

Searched exhaustively for the "byte-identity assertions" the brief expected
`token-diet` to have added, across `cmd/centinela/*_test.go`,
`internal/ui/*_test.go`, and every `tests/acceptance/token_diet_*.go` file:
patterns tried — `mustEqual`, `assertEqual`, `require.Equal`, `out != `,
`len(out) ==`, `len(got) ==`, hardcoded byte counts (`1092`, `1770`, `1034`,
`736`), `.golden` fixtures. **Result: none exist.** Every current test on
hook/panel output (`hook_context_panel_test.go`,
`hook_context_docs_test.go`, `hook_context_more_test.go`,
`internal/ui/render_core_test.go`, `internal/ui/panel_test.go`,
`tests/acceptance/token_diet_hook_render_test.go`,
`tests/acceptance/improve_centinela_render_ui_test.go`) uses
`strings.Contains` / `mustContain` / `mustNotContain` substring checks on
specific phrases (`"Documentation output missing"`, `"Roadmap: "`, feature
names, step names, `"BLOCKED WRITE"`). **Trailing-whitespace trimming cannot
break a substring check** — trimming the end of a line never removes
characters that appear earlier in that line or on other lines. This is the
honest finding, correcting the brief's assumption: there is no byte-identity
breakage list to update, because there is no byte-identity test suite yet.
The closest precedent (`remove-panel-borders`'s
`TestRenderSystemPanelHasNoBorder`, `TestRenderBlockedHasNoBorder` in
`internal/ui/render_core_test.go`) is also substring-based (`hasBorder` via
`strings.ContainsAny`) and is unaffected — trailing spaces are not border
characters.

## 8. Deferred finding

`hook_prewrite_block.go` (`RenderBlocked`, `RenderBlockedStaleBinary`),
`hook_merge.go`, `hook_migrate.go`, and the `hook setup` family
(`hook_setup.go`, `hook_setup_setup.go`, `hook_setup_cascade.go`) all build
panel bodies via the same `lipgloss.JoinVertical(lipgloss.Left, ...)` pattern
and carry the same residual padding — but none of them fire on every turn (see
§1's frequency column), so none are part of this feature's *measured* leak.
Same fix shape (`trimTrailingWS` wrap) would apply directly. Recommend a
follow-up backlog item, `hook-panel-diet-remaining-emitters`, scoped
narrowly to those six files, deferred via `centinela roadmap defer` rather
than folded into this feature (keeps this slice's blast radius matched to
what was actually measured).
