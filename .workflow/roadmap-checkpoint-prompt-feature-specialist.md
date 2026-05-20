### Feature-Specialist Report: roadmap-checkpoint-prompt

**Date:** 2026-05-20

#### Behavior Summary

After the project initiator has produced every roadmap-defining artifact
(`PROJECT.md`, `ROADMAP.md`, valid `.workflow/roadmap.json`,
`.workflow/roadmap-analysis.{md,json}`, `.workflow/roadmap-quality.{md,json}`,
and `docs/architecture/production-readiness-prompt.md`),
`centinela hook setup` emits exactly one new directive — `CENTINELA
DIRECTIVE: roadmap checkpoint …` — followed by a system panel that names the
first incomplete Phase 0 bootstrap feature and asks the user to choose
"keep iterating on the roadmap" or "start implementing
`<feature-name>`". Decision logic lives entirely in
`internal/roadmapcheckpoint` and is purely a function of (a) presence and
freshness of `.workflow/roadmap-checkpoint.json` (the iterate marker), (b)
the mtimes of the roadmap-defining artifacts, (c) bootstrap completeness,
(d) the first non-done Phase 0 feature, and (e) whether
`.workflow/<first-phase-0-feature>.json` (the workflow file written by
`centinela start`) already exists. The directive is one-shot: writing the
marker suppresses it; editing any roadmap-defining artifact later makes the
marker stale and re-fires it exactly once until a new marker is written.
Starting the first Phase 0 feature creates `.workflow/<feature>.json`, which
also suppresses the directive without a marker write. The host hook
`cmd/centinela/hook_setup.go` stays a thin orchestrator that simply loads
the roadmap, calls `roadmapcheckpoint.Decide`, and on `DecisionEmit` or
`DecisionStale` prints the directive line plus
`ui.RenderRoadmapCheckpoint(first)`. This new branch runs strictly after
every existing setup directive, so the earlier "missing artifact" prompts
keep their precedence and never coexist with the checkpoint on the same
run.

#### Gherkin Scenarios

Specified in `specs/roadmap-checkpoint-prompt.feature`:

- **Happy path emits the checkpoint directive when no marker exists** —
  full artifact set + no marker + an incomplete first Phase 0 feature +
  no workflow file → directive line and panel naming the first incomplete
  Phase 0 feature; exit zero.
- **Suppressed when the marker is fresh against all roadmap artifacts** —
  marker `at` ≥ max(mtime of every required artifact) → no directive, no
  panel.
- **Stale marker re-fires when ROADMAP.md was modified after the marker** —
  ROADMAP.md mtime > marker `at` → directive re-emitted.
- **Stale marker re-fires when any roadmap supporting artifact was modified
  after the marker** — `.workflow/roadmap-analysis.json` mtime > marker
  `at` → directive re-emitted (covers all required-artifact axes).
- **Suppressed when bootstrap is already complete** — every Phase 0
  feature has status `done` → never emit (no panel, no marker write).
- **Suppressed when no Phase 0 bootstrap features exist in the roadmap** —
  `FirstIncompleteBootstrap` returns `("", false)` → no directive.
- **Suppressed when the workflow file for the first Phase 0 feature
  already exists** — user already picked "start"; the natural
  `.workflow/<feature>.json` signal suppresses the directive without a
  marker write.
- **Order of precedence — missing roadmap-defining artifact lets the
  existing setup directives fire instead** — ROADMAP.md missing → the
  existing "roadmap required" directive fires; checkpoint is not
  evaluated.
- **Order of precedence — invalid roadmap.json yields the roadmap-json
  directive, not the checkpoint** — `roadmap.Load` returns an error → the
  existing `roadmapJSONDirective` fires; checkpoint is skipped.
- **Multiple Phase 0 features, only the first is done — picks the second
  as target** — `FirstIncompleteBootstrap` walks `BootstrapFeatures` in
  declared order and returns the first non-done one.
- **Malformed marker JSON is treated as missing and re-emits without
  crashing** — `json.Unmarshal` error on the marker file falls through to
  emit; the binary does not panic or exit non-zero.
- **Marker `at` field unparseable as RFC 3339 is treated as stale and
  re-emits** — `time.Parse(time.RFC3339, …)` error on `at` triggers a
  stale outcome; renderer + directive still print.

Each scenario maps to executable assertions: integration tests scaffold a
temp project root, set artifact mtimes via `os.Chtimes`, drive
`runHookSetup` (or its in-process equivalent), and assert the captured
stdout contains — or does NOT contain — the directive line and the
feature name from the rendered panel. Unit tests for
`roadmapcheckpoint.Decide` cover each branch with a fake filesystem /
fixture roadmap.

#### UX States

The feature has no graphical surface; the only output channel is
`centinela hook setup`'s stdout, consumed by Claude as a system message.
The relevant states are panel-render states emitted by the CLI hook:

| State          | Trigger                                                                                                                                                | Surface                                                                                                                                                                  |
|----------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| emit-fresh     | All roadmap-defining artifacts exist, no marker, bootstrap incomplete, no `.workflow/<first>.json`.                                                     | `CENTINELA DIRECTIVE: roadmap checkpoint …` line + `ui.RenderRoadmapCheckpoint(first)` panel naming the first incomplete Phase 0 feature and the literal `centinela start <feature>` command. |
| emit-stale     | Marker exists; any required artifact mtime > marker `at`, OR marker malformed JSON, OR marker `at` unparseable.                                         | Same directive line + panel as `emit-fresh`. Internally classified as `DecisionStale` so tests can distinguish, but UX is identical.                                     |
| suppressed     | Marker fresh, OR bootstrap already complete, OR no Phase 0 features, OR `.workflow/<first>.json` exists, OR a higher-precedence setup directive fired. | No output from the checkpoint branch. Stdout from earlier setup branches may be present in the "higher-precedence" case.                                                 |
| loading/empty/error | n/a — synchronous CLI; no list view; no graphical error state. Decoding errors on the marker fall through to `emit-stale` rather than rendering an error panel, per the plan. | n/a                                                                                                                                                                       |

#### Out-of-Scope

- Persisting checkpoint history across projects or across the user's
  entire machine — the marker lives in `.workflow/` and is per-project.
- Automatically re-running senior-PM analysis or quality scoring when
  `ROADMAP.md` changes — the user still drives those steps explicitly.
- Multi-phase checkpoints — only Phase 0 is gated here; later phases are
  out of scope.
- Programmatically detecting user "iterate" vs "start" answers — the
  orchestrator (Claude) writes the marker on "iterate" or invokes
  `centinela start <feature>` on "start".
- Writing a "start" marker for symmetry — workflow-file presence is the
  canonical "start" signal; no second marker.
- Cross-project checkpoint telemetry or auditing.
- Any GUI / TUI surface; the binary only prints text to stdout.
- Changing the structure or copy of the existing setup directives —
  the checkpoint slots in after them; precedence is preserved.

#### Edge Cases

1. Marker present and fresh → suppressed (no directive, no panel).
2. Marker present but stale (`ROADMAP.md` mtime > marker `at`) →
   re-emit.
3. Marker present but stale (any of
   `.workflow/roadmap-analysis.{md,json}` or
   `.workflow/roadmap-quality.{md,json}` or `.workflow/roadmap.json`
   modified after `at`) → re-emit.
4. Bootstrap already complete on first run → never emit; never write a
   marker.
5. First Phase 0 feature absent from roadmap (no Phase 0 phase or empty
   phase) → `FirstIncompleteBootstrap` returns `("", false)` → suppress.
6. Workflow file for first Phase 0 feature exists
   (`.workflow/<feature>.json`) → suppress (user already chose "start";
   no marker required).
7. Marker file malformed JSON → treat as missing/stale; directive
   re-emits; binary does not crash; non-zero exit is NOT triggered.
8. Marker `at` field unparseable as RFC 3339 → treat as stale; re-emit.
9. Multiple Phase 0 features, only the first is `done` → pick the second
   non-done one as target; `FirstIncompleteBootstrap` walks declared
   order.
10. Mtime granularity / no-op editor save bumps mtime of a required
    artifact → marker becomes stale and the prompt re-fires. Documented
    trade-off, preferred over silent drift; tests cover the re-fire
    explicitly.
11. Concurrent hook invocations both writing the marker → last writer
    wins. The marker is single-field semantic (`{choice:"iterate",at:…}`);
    no data-integrity issue because both writes produce equivalent
    content.
12. `roadmap.json` invalid (`roadmap.Load` returns an error) → existing
    `roadmapJSONDirective` fires first; checkpoint is not evaluated this
    run, by construction of the hook's order of precedence.

#### Handoff

- Next role: senior-engineer
- Open clarifications:
  - **Directive line wording**: spec pins the prefix
    `CENTINELA DIRECTIVE: roadmap checkpoint` so tests have a stable
    string to assert; the senior-engineer chooses the exact tail copy
    that follows the prefix to match the existing setup-directive style.
  - **Panel copy**: `ui.RenderRoadmapCheckpoint(featureName string)` is
    expected to surface both choices ("iterate" vs "start") and include
    the literal `centinela start <feature-name>` command so Claude has
    zero-effort wiring; senior-engineer to finalize the prose.
  - **Marker payload extension**: per the big-thinker's recommendation
    no "start" marker is written; if the senior-engineer chooses to add
    one for audit, it MUST NOT change the suppression behavior — the
    workflow file remains canonical for "start".
  - **Malformed-marker error surface**: the spec treats malformed JSON
    and unparseable `at` as stale (re-emit). The senior-engineer may
    additionally log a one-line warning to stderr, but the directive
    must still emit and the hook must still exit zero.
