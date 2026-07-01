### Feature-Specialist Report: roadmap-crud-add-remove
**Date:** 2026-07-01

#### Behavior Summary

This feature lets an operator author a roadmap feature directly instead of only
capturing findings via `defer`/`promote`. `roadmap add <slug> --phase <P>`
appends a new `Feature` with `Draft: true` into a chosen schedulable phase,
running the same validate-then-mutate-then-write-once discipline as `defer`
(slug/collision/phase/dependency-cycle checks all run before the single
`writeRawRoadmap`, so any rejection leaves `roadmap.json` byte-identical).
`roadmap remove`/`rm <slug>` deletes a planned feature, refusing when another
feature (including a draft) still lists it in `dependsOn`, or when its derived
`FeatureStatus` is in-progress/done. `roadmap promote` is generalized to branch
on the slug's *current location*: a Backlog finding still moves into
`--phase` (unchanged); a draft already sitting in a schedulable phase is
finalized **in place** — `Draft` is cleared via `replaceFeatureAt` and the
analysis+quality artifacts are appended, with no phase move. The `Draft`
dimension has exactly one coverage-set exemption hook
(`NonBacklogFeatureSet`: `if f.Draft { continue }`) but four independent
readers must agree with the persisted `f.Draft` field: the coverage set
itself, `classifyFeature`/`ReadySet` (a draft is `State:"draft"`, never
`"ready"`), `Summary()` (drafts are not committed planned work), and
`BuildView`/`buildFeatureView` (`draft:true` + `readiness:"draft"` in
`roadmap --json`). `centinela start <draft>` is refused, mirroring the
existing Backlog refusal. ROADMAP.md renders a deterministic ` *(draft)*`
marker on a draft's bullet line.

#### Gherkin Scenarios

All scenarios live in
[`specs/roadmap-crud-add-remove.feature`](../specs/roadmap-crud-add-remove.feature):

- **add creates a draft in a chosen schedulable phase and validate stays
  PASS** — Given phase "Phase 1: Foundations" has no "new-widget", When
  `roadmap add new-widget --phase "Phase 1: Foundations"` runs, Then the new
  entry has `draft:true`, roadmap.json parses via `Load`, `roadmap validate`
  exits 0, and every untouched phase is byte-identical.
- **add accepts optional description, depends-on, and archetype flags** —
  Given a done `auth-service`, When `add` is run with `--description
  --depends-on auth-service --archetype canonical`, Then all three fields
  land on the new entry alongside `draft:true`.
- **add rejects invalid input and leaves roadmap.json byte-identical**
  (Scenario Outline, 7 rows) — Given a captured "before" snapshot, When add is
  run with an invalid slug / duplicate name / unknown phase / Backlog target /
  Baseline target / unknown `dependsOn` / self-referencing `dependsOn` (cycle),
  Then each exits non-zero with the matching error substring and the file is
  byte-identical to "before".
- **add duplicate name across a different phase reports the owning phase** —
  Given "billing-api" exists in "Phase 2: Growth", When add targets "Phase 1:
  Foundations" with the same slug, Then stderr names "Phase 2: Growth".
- **add against an empty roadmap errors "unknown phase" with no silent phase
  creation** — Given `{"phases":[]}`, When add runs, Then it errors "unknown
  phase" and the file remains exactly `{"phases":[]}`.
- **add against a missing or malformed roadmap.json surfaces an error and
  leaves the file untouched** — Given no file on disk, When add runs, Then it
  exits non-zero and the file is still absent.
- **remove deletes a planned feature and leaves the file valid** — Given a
  planned, non-depended-on feature, When `remove` runs, Then it disappears,
  the file parses via `Load`, and untouched phases are byte-identical.
- **rm is an alias for remove** — same setup, `rm` in place of `remove`,
  identical outcome.
- **remove a feature that does not exist errors "not found"** — Given no such
  feature anywhere, When `remove` runs, Then stderr contains "not found" and
  the file is byte-identical.
- **remove the last feature of a phase leaves the phase with an empty
  features array** — Given a phase with exactly one feature, When it is
  removed, Then the phase still exists with `"features": []`.
- **remove is refused when another feature depends on it, naming the
  dependents** — Given "checkout-ui" depends on "auth-service", When removing
  "auth-service" is attempted, Then stderr names "checkout-ui" and the file is
  untouched.
- **remove is refused when the only dependent is itself a draft** — Given a
  draft "draft-consumer" depends on "auth-service", When removing
  "auth-service" is attempted, Then stderr names "draft-consumer" — drafts are
  real dependents, not exempt from this guard.
- **remove is refused for an in-progress or done feature** (Scenario Outline,
  2 rows) — Given a non-depended-on feature with status in-progress/done, When
  `remove` runs, Then it exits non-zero naming the status and the file is
  untouched.
- **promote finalizes a draft in place — no phase move, draft cleared,
  artifacts written** — Given a draft already in "Phase 1: Foundations", When
  `promote --scores 9,9,9,9,9,9` runs, Then the feature stays in the same
  phase, `draft` is cleared, and analysis+quality gain entries; `roadmap
  validate` exits 0.
- **promote of a Backlog finding still moves it into --phase (unchanged
  behavior)** — Given a Backlog finding, When `promote --phase … --scores …`
  runs, Then it leaves Backlog and lands in the target phase — today's path,
  untouched by this feature.
- **promote of a non-draft, non-Backlog slug is a clear error** — Given a
  plain planned feature, When `promote` is attempted, Then it errors naming
  the slug as neither a draft nor a Backlog finding, file untouched.
- **promote of a draft with overall score below 9 is refused, draft flag left
  intact** — Given a draft, When `promote --scores 9,9,9,9,9,8` runs, Then it
  errors "overall score must be at least 9", `draft` stays `true`, and no
  artifact file changes.
- **a freshly-added draft simultaneously satisfies all four draft readers**
  (the invariant scenario) — Given a fresh `add`, Then `roadmap validate`
  stays PASS (reader 1: coverage-set exemption), the draft is absent from
  `roadmap ready --json` (reader 2: `classifyFeature`/`ReadySet`),
  `counts.planned` in `roadmap --json` excludes it (reader 3: `Summary`), the
  same JSON's entry has `draft:true` and `readiness:"draft"` (reader 4:
  `BuildView`), and `centinela start <draft>` is refused mentioning "draft".
- **ROADMAP.md renders a deterministic " *(draft)*" marker for a draft
  feature** — regenerating the doc twice produces byte-identical output.
- **roadmap --json is byte-identical across two consecutive runs after
  add/remove/promote** — determinism holds with the new field present.
- **existing non-draft roadmap --json output is unchanged by the draft
  extension** — no non-draft entry gains a `draft` field or
  `readiness:"draft"`; the roadmap-json-contract shape is otherwise preserved.

#### UX States

| State   | Trigger                                                                 | CLI surface                                                                                       | JSON surface (`roadmap --json`)                                  |
|---------|--------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------|--------------------------------------------------------------------|
| empty   | `add` against `{"phases":[]}`                                          | Non-zero exit; stderr "unknown phase …"; no phase auto-created                                       | n/a — command fails before any view is built                       |
| error   | `add` unknown phase / Backlog / Baseline target                         | Non-zero exit; stderr "unknown phase %q; known phases: …"                                            | n/a                                                                 |
| error   | `remove` on a depended-on or in-progress/done feature                    | Non-zero exit; stderr names dependents or the blocking status                                        | n/a                                                                 |
| success | `add` into a valid schedulable phase                                    | `ui.RenderSuccess` confirmation naming the slug and phase                                             | new entry has `"draft": true`                                       |
| draft-in-json | any draft feature present                                           | n/a (text mode has no dedicated draft rendering beyond ROADMAP.md's marker)                          | entry has `"draft": true`, `"readiness": "draft"`, absent from `ready --json`, excluded from `counts.planned` |
| remove-guard refusal | `remove` blocked by a dependent or status                        | Non-zero exit; stderr lists the offending dependent name(s) or status; file untouched                | n/a                                                                 |
| success | `promote` in-place finalize                                              | `ui.RenderSuccess`-style confirmation; no "Remember to sync ROADMAP.md" phase-move language needed (no move occurred) | `draft` field removed from the entry entirely (`omitempty`)         |

#### Out-of-Scope

- `edit`/`update`, `move`, `reorder` on an existing feature — deferred to
  `roadmap-edit-move` (recorded in the umbrella design's Backlog).
- `phase add`/`rename`/`remove` — deferred to `roadmap-phase-ops`.
- Silent phase auto-creation on `add` against an unknown/empty roadmap —
  explicitly refused; that capability, if ever wanted, belongs to
  `roadmap-phase-ops`.
- Any locking upgrade to the raw I/O layer — concurrent writers keep today's
  atomic temp+rename, one-feature-per-line, last-writer-wins guarantee.
- Backlog authoring itself (`defer`) — unchanged; `add` explicitly refuses
  Backlog/Baseline as a target phase.

#### Deferred Findings

None. No genuinely new gap surfaced while translating the brief and plan into
executable Gherkin — the brief's scope fence and the big-thinker's four-reader
invariant fully account for every scenario needed. (`--source
roadmap-crud-add-remove/feature-specialist`)

#### Handoff

- **Next role:** senior-engineer.
- **Open clarifications (forwarded from big-thinker, none blocking):**
  - Confirm `buildFeatureView` sets `fv.Readiness = "draft"` and `fv.Draft =
    true` explicitly for drafts (today it only copies Readiness for
    `ready|blocked`) — the invariant scenario in the spec asserts this
    directly; a naive extension that forgets the draft branch will fail it
    immediately.
  - Confirm `appendToPhase` becomes a thin delegator to the new
    `appendFeatureToPhase(target, Feature{Name:slug, DependsOn:[]string{}})`
    so existing Backlog-promote tests remain the regression fence, rather than
    forking a second raw-append path.
  - Confirm the `promote` branch decision reads the slug's current location
    from the raw doc (Backlog phase membership vs. `Draft` flag on a
    schedulable-phase entry) rather than any new flag — the three promote
    scenarios in the spec assert exactly this three-way branch.
  - No other design changes recommended; the plan's file list, reuse map, and
    line-budget split are consistent with the scenarios written here.
