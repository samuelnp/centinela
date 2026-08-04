# hook-context-panel-diet — qa-senior

The implementation landed with **zero** tests referencing it: the edge-case
tester's grep found no test anywhere naming `trimTrailingWS` or this feature
slug, and the size guard the plan mandates (`internal/ui/panel_budget_test.go`)
did not exist. Everything below is new.

## Test Inventory

### Colocated unit (`internal/ui`, package `ui`)

This is the tier that moves the per-package coverage gate; `tests/*` tiers do not.

| File | Lines | What it locks down |
|---|---|---|
| `internal/ui/panel_budget_test.go` | 98 | **The size guard (plan §4).** Fixed in-memory `workflow.Workflow{Feature:"sample-feature", CurrentStep:"plan"}`, `StepOrder` nil so `OrderedSteps()` falls back to `DefaultStepOrder`. Renders `RenderContextCapped` + `RenderReviewReady` + `RenderFeatureBriefNeeded` directly. Zero filesystem read, zero `os.Chdir`, zero active-workflow lookup. Three tests: zero-trailing-whitespace, byte budget, repo-state independence. |
| `internal/ui/panel_trim_test.go` | 93 | `trimTrailingWS` table (12 cases) + long-line case + idempotency. |
| `internal/ui/render_trailing_ws_test.go` | 98 | All six trimmed render functions under 11 realistic fixtures; governance-signal survival; the **fixture non-tautology** guard. |
| `internal/ui/render_cli_padding_test.go` | 71 | CLI-surface non-regression: `RenderStatus` / `RenderBlocked` keep their padding; honest note on `RenderRoadmap`. |

Every file is ≤100 lines (G1 applies to `_test.go` under `internal/` and `cmd/`).

**Size guard specifics.** Measured baseline **1195 bytes** (reproduced
independently here; matches the senior-engineer's 1195 and the pre-fix 2004
exactly). Budget **1400** — ~17% headroom for a clarified sentence in a nudge
panel, still far below the ~2004 a reintroduced padding pass would produce. The
failure message names the measured size, the budget, the dated baseline, and
the two distinct remedies (re-measure and raise the constant vs. re-check the
six `trimTrailingWS` wraps), because a guard that fails cryptically costs the
next person an afternoon.

**`trimTrailingWS` table cases.** Empty string; spaces-only; tabs-only; single
line with no trailing newline; already-trimmed; mixed spaces+tabs; one trailing
newline preserved; several trailing newlines preserved; **interior blank
separator stays a line** (`"a  \n   \nb  "` → `"a\n\nb"` — three lines in,
three out, so a `JoinVertical` spacer never silently vanishes); 5000-byte line;
plus the three cases the edge-case tester found unhandled, pinned as **current
behaviour with an explicit `latent:` note explaining why they are unreachable
today** rather than left untested:

- **CRLF** — `"a  \r\nb  \r\n"` is returned *unchanged*; `TrimRight(l, " \t")`
  stops at `\r`. Unreachable: nothing in `internal/ui` or
  `cmd/centinela/hook_context*` emits a literal `\r`.
- **Meaningful trailing tab** — `"col1\tcol2\t"` → `"col1\tcol2"`; a tab used
  as content would be eaten. Unreachable: no panel renders tab-separated columns.
- **NBSP (U+00A0)** — preserved, since the cutset is ASCII space/tab only.
  Written as ` ` escapes, not invisible literal bytes.

Idempotency is asserted over every table case plus the live combined panel
render — the property the six same-line return wraps depend on.

### Unit tier (`tests/unit/hook_context_panel_diet_unit_test.go`)

Same contract through the **exported** package surface only: all six hook
renderers unpadded under a mismatched-width three-workflow fixture; the
per-archetype ladder (spike shows `code 1/2` and no `validate`); `RenderStatus`
still padded.

### Integration tier (`tests/integration/hook_context_panel_diet_integration_test.go`)

Drives the **real binary** end to end. Builds `./cmd/centinela` into
`t.TempDir()`, seeds throwaway project dirs (never the real `.workflow/`), runs
`hook context`:

1. Four mixed-archetype workflows (canonical short, canonical long, hotfix,
   spike) → zero padded lines in the whole hook payload, panels still present.
2. Per-archetype ladder assertions parsed off the rendered output: spike has no
   `validate`, hotfix has no `plan`, canonical keeps `validate`.
3. Zero-workflow empty state → unpadded, directive intact.

### Acceptance tier (`tests/acceptance/hook_context_panel_diet{,_guard,_helper}_test.go`)

All three files carry `// Acceptance: specs/hook-context-panel-diet.feature`.
Binary built from `./cmd/centinela` into a temp dir (`sync.Once`), all fixtures
local — **no network URL anywhere**.

## Acceptance Wiring — 11/11 scenarios

| Spec scenario | Test |
|---|---|
| Hook context output carries no trailing whitespace | `TestPDHookContextHasNoTrailingWhitespace` |
| A feature-brief-required nudge panel carries no trailing whitespace | `TestPDFeatureBriefPanelHasNoTrailingWhitespace` |
| A review-required nudge panel carries no trailing whitespace | `TestPDReviewReadyPanelHasNoTrailingWhitespace` |
| The active feature, its step, and its progress still appear | `TestPDGovernanceHeaderSurvivesTheDiet` |
| The full step ladder still distinguishes archetypes with different gates | `TestPDStepLadderStillDistinguishesArchetypes` |
| A blocking review-required instruction still appears when due | `TestPDReviewInstructionStillNamesItsFeature` |
| centinela status output is unchanged, trailing whitespace included | `TestPDStatusOutputKeepsItsPadding` |
| A blocked-write refusal panel is unaffected by this feature | `TestPDBlockedWritePanelKeepsItsPadding` |
| The size guard passes at the recorded budget | `TestPDSizeGuardPassesAtRecordedBudget` |
| The size guard fails when panel output grows past budget | `TestPDSizeGuardFailsPastBudget` |
| The size guard does not depend on this repository's live workflow count | `TestPDSizeGuardIgnoresLiveWorkflowCount` |

Verified against the gate's own matcher semantics
(`internal/gates/spec_traceability_{parse,match}.go`: `// Acceptance:` header +
standalone `// Scenario:` lines, whitespace-collapsed, one trailing period
stripped, lowercased) by re-implementing `parseScenarios` / `coveredScenarios` /
`uncovered` over this repo's files: **11 scenarios in the spec, 11 covered,
0 missing.**

The three size-guard scenarios drive the colocated guard **as a subprocess**
(`go test ./internal/ui/ -run …`) rather than re-implementing its assertion in
the acceptance package. Re-implementing it would prove nothing about the check
that actually runs in CI. To make "it should fail, naming the byte length and
the budget it exceeded" observable, the guard reads a test-only env knob
(`CENTINELA_PANEL_DIET_PAD`, no production code touches it); the acceptance
test injects 500 bytes, requires a non-zero exit, and reads both numbers back
out of the failure message (`measured N bytes, budget M`), asserting `N > M`
and `N >= 500`. It reads them rather than hardcoding today's 1695/1400, so a
legitimate re-baseline of the budget does not break this scenario for the wrong
reason.

## Regression Guards — red→green evidence

`trimTrailingWS` was reverted to `return s` in place (with `_ = strings.TrimRight`
to keep the import honest), `go test ./internal/ui/` re-run, then `panel.go`
restored from a scratch backup and confirmed byte-identical
(`git diff --stat internal/ui/panel.go` → empty).

**5 top-level tests went red, covering all six trimmed render functions:**

| Test | Reverted |
|---|---|
| `TestPanelBudgetNoTrailingWhitespace` | FAIL |
| `TestPanelBudgetWithinByteBudget` | FAIL |
| `TestTrimTrailingWSTable` | FAIL (6 of 12 subcases) |
| `TestTrimTrailingWSLongLine` | FAIL |
| `TestHookPanelsCarryNoTrailingWhitespace` | FAIL (10 of 11 subcases — every one of `RenderContextCapped`, `RenderReviewReady`, `RenderFeatureBriefNeeded`, `RenderEdgeCaseReportNeeded`, `RenderDocumentationNeeded`, `RenderChangelogNeeded`) |

**Tests that stayed green under the revert — each correctly so, stated rather
than hidden:**

- `TestPanelBudgetIndependentOfRepoState` — asserts determinism across repo
  state, not trimming. Identity is equally deterministic. Correct.
- `TestTrimTrailingWSIdempotent` — identity is trivially idempotent. Correct;
  it guards double-wrapping, not the trim itself.
- `TestTrimTrailingWSTable` subcases `empty string`, `already trimmed`,
  `one/several trailing newlines`, `CRLF`, `NBSP` — inputs where the correct
  output *is* the input. Correct by construction.
- `TestCLIRenderersKeepTheirPadding`, `TestRenderRoadmapUnaffected`,
  `TestDietFixturesActuallyPadWhenUntrimmed` — these assert the **opposite**
  direction (padding must survive / raw `JoinVertical` must pad). Going red on
  the revert would mean they were wired backwards.

**Anti-tautology guard.** `TestDietFixturesActuallyPadWhenUntrimmed` composes
one fixture's body exactly as `RenderContextCapped` does but without the trim
and asserts that raw form *is* padded. If someone ever normalises the fixture
feature names to equal widths, `JoinVertical` stops padding and the
zero-trailing-whitespace assertions would pass with the fix reverted — that
test fails first and says exactly that.

**Scope-boundary guard.** `TestCLIRenderersKeepTheirPadding` (colocated),
`TestCLIRenderStatusStillPadded` (unit tier), `TestPDStatusOutputKeepsItsPadding`
and `TestPDBlockedWritePanelKeepsItsPadding` (acceptance, real binary) all pin
that `RenderStatus` / `RenderBlocked` keep every padded byte. Their real target
is the most likely accidental over-reach: moving the trim down into the shared
`renderSystemPanel` primitive, which would silently change `centinela status`
and blocked-write output for every user. The comment says so, and says the
cleanup is a deliberate separate decision, not a fix to apply when the test
goes red.

**Honest exception:** `RenderRoadmap` has *no* padding to preserve — it
composes with `strings.Join`, never `lipgloss.JoinVertical`, so it emitted zero
trailing whitespace before this feature and emits zero after. Asserting "still
padded" for it would be a false test. `TestRenderRoadmapUnaffected` records
that instead, and the shared-primitive regression it would otherwise guard is
already covered by `TestCLIRenderersKeepTheirPadding`.

## Coverage Gaps

One profiled run: `go test ./... -coverprofile=coverage.out` → **exit 0, 48
packages ok, zero failures**, `tests/acceptance` 459.976s (no `test timed out`
panic this time, despite another session running its own full suite against a
sibling worktree concurrently — the known
`acceptance-tier-exceeds-default-test-timeout` flake did not fire). Then
`COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh` →
**coverage gate passed: 97.3% >= 95.0%**.

Per-package, for the packages this step touches:

- `internal/ui` — **99.4%** (the package this feature touches)
- `cmd/centinela` — **96.3%** (unchanged; the hook emitter was not edited)

One caveat stated plainly: a robustness edit to
`tests/acceptance/hook_context_panel_diet_guard_test.go` (reading the byte
counts out of the guard's failure message instead of hardcoding 1695/1400)
landed while that profiled run was already executing, so the acceptance package
in the profile may have been compiled from the pre-edit file. `tests/*` tiers
contribute no statements to the coverage profile, so the 97.3% figure is
unaffected. All 11 acceptance scenarios were re-run on the final tree
afterwards and pass; no full-suite re-run was performed.

Remaining honest gaps:

- **The >5-active-workflow nudge/ladder mismatch is still untested.**
  `hook_context.go` loops the *uncapped* `wfs` for nudge panels but the capped
  `shown` for the ladder, so a workflow past the "+N more active" cutoff can
  fire a nudge panel with no header above it. Pre-existing. The plan's D2
  argument for keeping each nudge panel's `<feature> · <step>` self-ID leans on
  exactly this, and `TestPDReviewInstructionStillNamesItsFeature` pins the
  self-ID itself — but not the >5 case that makes it load-bearing. Left as-is:
  the cap semantics are workflow behaviour, outside a presentation-only feature.
- **ANSI/colour behaviour is not asserted in a committed test.** The edge-case
  tester probed it under a forced `termenv.TrueColor` profile and under a real
  PTY (lipgloss appends pad as plain spaces *after* the SGR reset, so trimming
  stops at the `m` and never eats a reset). Reproducing that needs a
  process-wide lipgloss profile mutation, which would make the `ui` package's
  tests order-dependent. The finding is recorded, not encoded.
- **`internal/ui` is at 99.4%, not 100%.** The uncovered remainder is
  pre-existing and unrelated to the touched files.

## Deferred Findings

Not filed via `centinela roadmap defer` (per instructions — orchestrator to
file from the primary checkout).

1. **`hook-panel-diet-remaining-emitters`** — carried forward from plan §8, the
   senior-engineer report, and the edge-case report. `hook_prewrite_block.go`
   (`RenderBlocked`, `RenderBlockedStaleBinary`), `hook_merge.go`
   (`RenderMergeStewardNeeded`), `hook_migrate.go` (`RenderMigrationNeeded`),
   and the `hook setup` family (`hook_setup.go`, `hook_setup_setup.go`,
   `hook_setup_cascade.go`) all build bodies via the same
   `lipgloss.JoinVertical(lipgloss.Left, …)` and carry identical residual
   padding, but none fires on every `UserPromptSubmit` turn. Same one-line fix
   shape. **Note for whoever picks this up:** the tests added here deliberately
   assert that `RenderBlocked` *is* still padded, so that follow-up must update
   `TestCLIRenderersKeepTheirPadding`, `TestCLIRenderStatusStillPadded` and
   `TestPDBlockedWritePanelKeepsItsPadding` as part of its diff — which is the
   point: the change becomes visible and deliberate instead of silent.
2. **`archetype-ladder-trusts-degraded-steporder`** — PRE-EXISTING, OUT OF
   SCOPE, not fixed here. `OrderedSteps()` falls back to the canonical 5-step
   order whenever `StepOrder` is empty, which includes every
   `degradedWorkflow`-reconstructed state file regardless of its real
   archetype. A degraded spike/hotfix therefore renders a phantom `validate`
   step it does not gate on, undermining the plan's own D2 justification for
   keeping the ladder. Fix would live in `degradedWorkflow` (persist a
   best-effort `stepOrder` from the raw `archetype`) or in `OrderedSteps` (fall
   back via `ArchetypeStepOrder(wf.Archetype)` first). Workflow-state
   semantics, not presentation.
3. **`corrupt-workflow-file-silently-drops-from-active-list`** — PRE-EXISTING,
   OUT OF SCOPE, not fixed here. On unparseable JSON, `ActiveWorkflows` logs a
   `workflow warning:` to stderr (which the model never sees) and drops the
   workflow entirely, so `hook context` reports "no active workflow" for a
   feature that is mid-workflow with an unreadable state file. The governance
   signal does not degrade — it vanishes. A minimal fix would surface a
   degraded-but-present entry, as the future-schema path already does.

## Handoff

**Next role: gatekeeper** (adversarial verifier).

What changed in this step: **9 new test files, 0 production files.**
`internal/ui/panel.go` was temporarily mutated for the red→green proof and
restored byte-identically — `git status` shows no modification to any
non-test file.

Worth an independent adversarial look:

- **The size guard's env knob.** `CENTINELA_PANEL_DIET_PAD` is read by a
  `_test.go` file only. Confirm no production code path reads it and that an
  unset/garbage value is a no-op (it is: `strconv.Atoi` error → skip).
- **The 1400 budget.** Re-derive it: measure the fixture yourself, check
  1195 + headroom, and confirm reintroducing padding (~2004 bytes) still trips it.
- **Traceability.** Re-run the real gate rather than trusting the 11/11 above.
- **The red→green claim.** Re-do the revert independently; the failing set
  should be the 5 tests named above and no more.
