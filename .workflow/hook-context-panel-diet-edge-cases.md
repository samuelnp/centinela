# Edge Cases: hook-context-panel-diet

Methodology: built `/tmp/cent-ec` from this worktree (`go build -o /tmp/cent-ec
./cmd/centinela`) and `/tmp/cent-before` from a `git worktree add --detach
7513283` (commit immediately before this feature's plan/code commits), then
drove both against synthetic `.workflow/*.json` fixtures in a scratch temp
directory (never the real repo's `.workflow/`). A throwaway
`internal/ui/zz_scratch_*_test.go` unit harness (built, run, then deleted —
not committed) exercised `trimTrailingWS` directly and the render functions
under a forced-color `lipgloss` profile. `git status` confirms no scratch
artifacts survive.

## Covered

- **Idempotency and byte-level behaviour of `trimTrailingWS`** (probed via
  `go test`, 17 cases): empty string, whitespace-only string (collapses to
  `""`, not removed as a line), single line with no trailing newline, already
  -trimmed input, a final trailing newline (`"abc\n"` stays `"abc\n"` —
  required by callers that expect one), multiple trailing newlines (all
  preserved), a 5000-byte line, CJK content, emoji content, box-drawing
  characters. Every case is idempotent: `trimTrailingWS(trimTrailingWS(s)) ==
  trimTrailingWS(s)`.
- **Blank-separator lines do not collapse.** `"a  \n   \nb  "` (a
  whitespace-only middle line, the exact shape `JoinVertical` produces for a
  spacer `""` between padded siblings) trims to `"a\n\nb"` — three lines in,
  three lines out, the middle one now genuinely empty instead of
  space-padded. `strings.Split`/`strings.Join` on `"\n"` never merges or
  drops a line just because `TrimRight` emptied it. Verified this holds for
  the real per-panel spacer lines (`""` args to `lipgloss.JoinVertical` in
  `render_review.go` / `render_brief.go`) via live capture below — panels keep
  their blank lines between sections.
- **CRLF is a real, currently-unreachable gap.** `trimTrailingWS` splits on
  `"\n"` only and trims `" \t"` only. Fed `"a  \r\nb  \r\n"`, it returns the
  string **unchanged** — the trailing spaces before each `\r` are NOT
  stripped, because `TrimRight(l, " \t")` stops at `\r`, which isn't in the
  cutset. Grepped `internal/ui/*.go` (non-test) and `cmd/centinela/hook_context*.go`
  for a literal `\r`: zero matches. Nothing in the touched render path emits
  CRLF today, so this is a latent gap, not a live bug — see Residual Risks.
- **A "meaningful" trailing tab would also be silently eaten.**
  `"col1\tcol2\t"` → `"col1\tcol2"`. Grepped for literal tab usage in
  `internal/ui/render*.go` (non-test): zero matches — no current panel uses
  tab-separated columns, so this is dead code path today, not a live bug.
- **Non-breaking space (U+00A0) is correctly left alone** — `TrimRight`'s
  cutset is the literal ASCII space/tab, so `"abc  "` is returned
  unchanged. Grepped for NBSP literals in the touched files: zero matches, so
  this never fires today either way; noted because a future nudge panel that
  pastes user-facing prose containing NBSP would keep it (correct — NBSP is
  visible content, not padding) but the reverse (NBSP-as-padding) can't
  happen since nothing generates it.
- **ANSI interaction — probed under a forced `termenv.TrueColor` profile,
  not reasoned from the "no color today" fact alone.** Isolated
  `lipgloss.JoinVertical` padding two colored spans of different width:
  `JOINED_RAW="\x1b[1ma\x1b[0m                              \n\x1b[1ma much
  longer colored line here\x1b[0m"` — lipgloss closes the SGR span with
  `\x1b[0m` on the *short* line and appends the pad as **plain, unstyled
  spaces after the reset**, not inside it. `trimTrailingWS` strips exactly
  those trailing plain spaces and stops at the `m` of the reset code (not a
  space/tab), leaving `\x1b[1ma\x1b[0m` intact — no dangling escape, no
  stripped reset. Re-ran `RenderContextCapped` and `RenderFeatureBriefNeeded`
  themselves under the same forced profile: every line still ends cleanly on
  a reset sequence. Separately confirmed under a **real PTY**
  (`TERM=xterm-256color`, `FORCE_COLOR=1`, via a Python `pty.openpty()`
  harness) that this repo's actual `hook context` and `status` output still
  carry **zero** `\x1b` bytes end-to-end — lipgloss's own auto-detection
  never turns color on here regardless of forcing `FORCE_COLOR` (it isn't a
  signal lipgloss's `termenv` checks). Net: even if color were ever turned on
  for this path, the fix is safe by construction, not by accident.
- **CJK/emoji width.** `trimTrailingWS` operates on runes via
  `strings.TrimRight`, not on terminal display columns, so double-width CJK
  content is never mis-split mid-character; confirmed on `"你好世界   \n測試行
  "` and `"✅ done  \n⚠ warn  "` — content byte-identical minus the trailing
  ASCII spaces.
- **Governance signal, live-probed across 7 distinct workflow states**
  (temp `.workflow/*.json` fixtures, real binary, real render path — not
  reasoning from the source alone):
  1. Single canonical workflow at `plan`, no brief yet — ladder shows
     `▶ plan · ○ code · ○ tests · ○ validate · ○ docs`, `FEATURE BRIEF
     REQUIRED` panel fires. Zero padded lines.
  2. **Hotfix + spike + canonical, mixed, with `stepOrder` correctly
     persisted** (as `centinela start --archetype <x>` actually writes it):
     spike shows `code 1/2` with a 2-step ladder (`plan · code`, no
     `validate`); hotfix shows `code 0/3` with a 3-step ladder (`code ·
     tests · validate`, no `plan`); canonical keeps its 5-step ladder. This
     is the spec's "archetypes with different gates" scenario, verified
     against the real render function, not a fixture-only guess.
  3. **7 active workflows** (cap is 5) — `ACTIVE WORKFLOWS` correctly shows
     the 5 most-recent plus a clean, unpadded `+2 more active` line. **But
     the nudge-panel loops in `hook_context.go` (`RenderReviewReady`,
     `RenderFeatureBriefNeeded`, etc.) iterate the *uncapped* `wfs` slice, not
     `shown`** — a workflow beyond the "+2 more" cutoff can still fire a
     nudge panel that never appears in the ladder above it. This is
     pre-existing behavior (not introduced by this feature) but it directly
     validates the plan's D2 justification for keeping each nudge panel's
     `<feature> · <step>` self-identification: with more than 5 active
     workflows, that self-ID is the *only* way to know which workflow an
     orphaned nudge panel is about.
  4. **Zero active workflows** — `CENTINELA DIRECTIVE: no active workflow...`
     plus `RenderSuccess("No active workflows.")`, a single-line render
     (`renderSystemLine`, no `JoinVertical`) that was never padded to begin
     with. Correct, unaffected by this feature.
  5. **A `done` workflow** — confirmed `ActiveWorkflows` excludes
     `currentStep: "done"` entirely (by design, pre-existing), so it never
     reaches the render path. Not a regression surface for this feature.
  6. **A degraded/unmodellable state file** (parseable JSON, unknown shape —
     e.g. `"steps"` typed as a string instead of an object, with a future
     `schemaVersion`): `Load` falls back to `degradedWorkflow`, which sets
     `CurrentStep` from the raw `currentStep` field but leaves `StepOrder`
     **nil**. `OrderedSteps()` then falls back to the full canonical 5-step
     order regardless of the file's real (unreadable) archetype. Rendered
     live: `degraded-feat  code 1/5` with a full 5-step ladder and zero
     padding — trimming itself is fine, but the ladder's own claimed
     precision ("the only per-turn signal of which concrete steps this
     workflow will run", plan §3) silently degrades to "assume canonical" for
     any state file this binary can partially-but-not-fully read. Pre-existing
     (`OrderedSteps`/`degradedWorkflow`), not caused by `trimTrailingWS` — see
     Residual Risks.
  7. **Genuinely corrupt JSON** (unparseable, not just future-schema): `Load`
     errors, `ActiveWorkflows` logs a `workflow warning:` to stderr and
     **drops the workflow from the active list entirely** — `hook context`
     then reports "no active workflow" even though a real, blocked-write
     -relevant workflow exists on disk. Also pre-existing, also not caused by
     this feature, but it is the sharpest version of "does the governance
     signal survive a degraded state file" the brief asked about, and the
     honest answer for this one input shape is no.
- **CLI-surface regression check, before/after binaries, identical
  fixtures:** `status <feature>` — byte-identical. `roadmap` (exit 1, no
  roadmap.json in scratch dir) — byte-identical. `doctor` (exit 1) —
  byte-identical. `hook prewrite` blocked-write refusal (`tests/unit/foo_test.go`
  during `plan` step) — exit 2 both, byte-identical, panel still carries its
  original padding (`RenderBlocked` untouched, confirmed live). `validate`
  — both binaries hit the same 10s timeout on the same first few lines of
  output (no network/gate config in the scratch dir), byte-identical up to
  the timeout. `hook context` on the same fixture: 1564 → 935 bytes
  (40.2% reduction), diff shows only trailing-space removal, no content
  loss — independently reproduces the senior-engineer's 2004→1195 measurement
  on a different fixture shape.

## Missing or Weak Scenarios

- **The size guard from plan §4 (`internal/ui/panel_budget_test.go`) does not
  exist in the tree yet.** Grepped `internal/ui`, `cmd/centinela`, and
  `tests/acceptance` for `budget`/`Budget`: no hits outside unrelated cost
  files. This was called out in the plan as the third slice and in the
  senior-engineer handoff as a tests-step artifact — it is the single
  concrete "measurement guard" this feature exists to add, and right now
  nothing stops a future edit from reintroducing `JoinVertical` padding
  undetected.
- **No test anywhere asserts the six render functions produce zero trailing
  whitespace.** `internal/ui/panel_test.go` only checks for border characters
  and content substrings; `hook_context_*_test.go` and the `token_diet_*`
  acceptance files check specific substrings (roadmap digest, feature names,
  edge-case filenames) but never `line == strings.TrimRight(line, " \t")`.
  The 11 Gherkin scenarios in `specs/hook-context-panel-diet.feature` have
  **zero** matching acceptance test file — grepped
  `tests/acceptance/*.go` and `internal/ui/*_test.go` for
  `hook-context-panel-diet`/`panel-diet`/`trimTrailingWS`: only my own
  now-deleted scratch file matched.
- **No unit test locks in the CRLF/tab/NBSP boundary behavior** documented
  above. Not urgent (unreachable today) but worth one small table-driven test
  so a future change to panel content that introduces a literal `\t` or CRLF
  doesn't silently regress without anyone noticing the interaction.
- **No test exercises the >5-active-workflow nudge-panel/ladder mismatch**
  (uncapped `wfs` loop vs. capped `shown` ladder) — pre-existing, but since
  this feature's plan explicitly leans on nudge-panel self-identification as
  the safety argument for *not* cutting that content (D2), a test that
  pins "a nudge panel beyond the display cap still names its feature" would
  make that argument enforceable instead of just asserted.

## Proposed/Added Tests (Unit/Integration/Acceptance)

**Unit** (`internal/ui`, highest priority for qa-senior):
1. `TestTrimTrailingWS_TableDriven` — the 17-case table from this probe
   (empty, whitespace-only, no-trailing-newline, already-trimmed, blank
   -separator-preserved, CRLF-not-stripped, trailing-tab-stripped,
   NBSP-preserved, box-drawing, CJK, emoji, long line, trailing-newline
   -preserved, multi-trailing-newline-preserved, idempotency assertion on
   every case).
2. `TestSixRenderFunctionsCarryNoTrailingWhitespace` — call each of
   `RenderContextCapped`, `RenderReviewReady`, `RenderFeatureBriefNeeded`,
   `RenderEdgeCaseReportNeeded`, `RenderDocumentationNeeded`,
   `RenderChangelogNeeded` with realistic multi-workflow / multi-archetype
   fixtures (not just one workflow — the padding only manifests when sibling
   lines differ in width) and assert every line equals its own
   `strings.TrimRight(line, " \t")`.
3. `internal/ui/panel_budget_test.go` per plan §4 — the size-guard test that
   is currently missing entirely: zero-trailing-whitespace + `<= 1400` byte
   budget on the fixed `sample-feature`/`plan` fixture, comment recording the
   measured baseline and date.

**Integration** (`cmd/centinela`):
4. A `hook_context` test with an archetype mix (hotfix + spike + canonical)
   asserting per-archetype ladder length (spike has no `validate`, hotfix has
   no `plan`) AND zero trailing whitespace on the combined output — closest
   existing file is `hook_context_branches_test.go`; extend or add
   alongside it.
5. A degraded-state-file test: write a `.workflow/<feature>.json` with a
   future `schemaVersion` and a type-mismatched field, assert the workflow
   still surfaces (feature name + some step signal present) rather than
   silently vanishing, matching `ActiveWorkflows`' documented intent — this
   guards against future changes turning "degrade gracefully" into "drop
   silently" for schema-forward files, which is the one governance-signal
   -loss case in this whole audit that's controllable (unlike the genuinely
   -corrupt-JSON case, which is a hard parse failure).

**Acceptance** (`tests/acceptance`, tied to `specs/hook-context-panel-diet.feature`):
6. A new `hook_context_panel_diet_test.go` implementing the file's 11
   scenarios end-to-end against the built binary — none currently exist.
   Minimum bar: the "no trailing whitespace" scenarios (3), the archetype
   -ladder scenario, the `centinela status` byte-identity scenario, and the
   blocked-write byte-identity scenario, all of which this report already
   probed manually and can be captured directly as Go test bodies using the
   same fixture shapes above.

## Residual Risks

- **CRLF / literal-tab / NBSP handling is untested and unreachable today,
  not fixed.** If a future panel ever assembles content from an external
  source (e.g., a pasted error message, a subprocess's output) that contains
  `\r\n` or a literal trailing tab used as meaningful content, `trimTrailingWS`
  will silently do the wrong thing (leave whitespace-before-`\r`, or eat a
  meaningful trailing tab) with no test to catch it. Mitigation: the unit
  test proposed above pins current behavior so a future change is a visible
  diff, not a silent one; genuinely handling CRLF would mean splitting on
  `\r\n|\n` — not done here since nothing in-tree needs it, and doing it
  speculatively would be scope creep beyond what's measured.
- **The archetype-specific step ladder trusts `wf.StepOrder`, not
  `wf.Archetype`, and degrades to canonical (with `validate`) when
  `StepOrder` is empty** — true for both a genuinely pre-`stepOrder`-field
  legacy file and for any file `degradedWorkflow` reconstructs. This means
  the plan's own safety argument for keeping the ladder ("the only per-turn
  signal of which concrete steps run, including whether validate applies at
  all") is not reliable for degraded state — a spike/hotfix workflow that
  lands in the degraded path will show a full canonical ladder including a
  `validate` step it doesn't actually gate on. Pre-existing
  (`OrderedSteps`/`degradedWorkflow`, unrelated to `trimTrailingWS`), out of
  this feature's stated scope (presentation only, no gate/directive
  semantics change), but flagged because it's exactly the failure mode
  the brief's "governance signal is not negotiable" constraint warns about.
- **A genuinely-corrupt `.workflow/<feature>.json` makes the workflow
  disappear from hook context entirely** (`ActiveWorkflows` drops it after
  logging to stderr, which the model never sees). Also pre-existing, also
  out of this feature's stated scope, but it is the actual worst case for
  "governance signal survives a degraded state file" — the signal does not
  survive; it silently vanishes, and the model would see `CENTINELA
  DIRECTIVE: no active workflow` instead of a warning that state is
  unreadable.
- **The >5-active-workflow nudge-panel/ladder mismatch** (nudge loops iterate
  uncapped `wfs`, the ladder shows only `shown`) is pre-existing and
  arguably intentional (nudges are more urgent than the display cap), but
  it's undocumented behavior that a reader of the plan's D2 section would
  not expect from "the panel is not nested under its workflow's header" — the
  header might not even be *shown*, not just structurally separate.
- **Two consecutive panels never get a blank-line separator** (confirmed
  byte-identical before/after this feature — `fmt.Println` per panel, no
  render function appends a trailing blank line) — panels run directly
  together (`...○ docs\n 🛡️👁️  HOOK  FEATURE BRIEF REQUIRED`). Not a
  regression (identical before/after), but worth noting since it's adjacent
  to "did trimming collapse a visual separator" — it did not; this gap was
  already there.

## Deferred Findings

Not filed via `centinela roadmap defer` per instructions — listed here for
the orchestrator to file from the primary checkout:

- **`hook-panel-diet-remaining-emitters`** (already identified by the plan
  and senior-engineer, independently reconfirmed by re-grepping every
  `cmd/centinela/hook_*.go` for `ui.Render*`/`fmt.Println`/`fmt.Fprintln
  (os.Std*)`): `hook_prewrite_block.go` (`RenderBlocked`,
  `RenderBlockedStaleBinary`), `hook_merge.go` (`RenderMergeStewardNeeded`),
  `hook_migrate.go` (`RenderMigrationNeeded`), and the `hook setup` family
  (`hook_setup.go`, `hook_setup_setup.go`, `hook_setup_cascade.go` —
  `RenderRoadmapNeeded`, `RenderRoadmapAnalysisNeeded`,
  `RenderRoadmapQualityNeeded`, `RenderProductionReadinessSetupNeeded`,
  `RenderBrownfieldSetupNeeded`, `RenderSetupNeeded`, `RenderRoadmapJSONNeeded`,
  `RenderRoadmapCheckpoint`) all build bodies via the same
  `lipgloss.JoinVertical(lipgloss.Left, ...)` pattern and carry the same
  residual padding, but none fire on every `UserPromptSubmit` turn — confirmed
  live for `RenderBlocked` (still padded, byte-identical before/after this
  feature). Same fix shape (`trimTrailingWS` wrap) applies directly as a
  narrow follow-up.
- **`archetype-ladder-trusts-degraded-steporder`** (new finding from this
  audit, not previously logged anywhere in this feature's plan/senior
  -engineer artifacts): the step ladder silently falls back to the canonical
  5-step order — including a `validate` gate — for any workflow whose
  `StepOrder` field is empty, which includes every `degradedWorkflow`
  -reconstructed file regardless of its real archetype. Scoped fix would live
  in `degradedWorkflow` (persist a best-effort `stepOrder` from the raw
  `archetype` field if present) or in `OrderedSteps` (fall back via
  `ArchetypeStepOrder(wf.Archetype)` before falling back to
  `DefaultStepOrder`) — outside this feature's presentation-only scope.
- **`corrupt-workflow-file-silently-drops-from-active-list`** (new finding):
  `ActiveWorkflows` logs a `workflow warning:` to stderr on a parse failure
  and excludes the workflow from the returned list, so `hook context` reports
  "no active workflow" for a feature that is, in fact, mid-workflow with an
  unreadable state file. A minimal fix would surface a degraded-but-present
  entry (as `degradedWorkflow` already does for the future-schema case) for
  genuinely malformed JSON too, so the model at least learns "a workflow
  exists here and its state file is broken" instead of "start a new
  workflow." Outside this feature's scope (workflow/state semantics, not
  presentation).
