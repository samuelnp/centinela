### Feature-Specialist Report: roadmap-json-contract
**Date:** 2026-07-01

#### Behavior Summary

`centinela roadmap` gains a `--json` flag that emits a deterministic
`RoadmapView` — ordered `phases[]`, each with ordered `features[]` carrying
`name, phase, status, readiness, dependsOn, blockedBy`, plus a top-level
`counts` object scoped exactly like `Summary()` (Backlog/Baseline excluded).
`roadmap ready --json` emits the array of ready feature names in declared
order, and that set is identical to the `readiness:"ready"` features in
`roadmap --json`. A new `roadmap show` command (alias `list`) prints the
persisted typed `Roadmap` as text (unchanged rendering) or, with `--json`,
dumps it verbatim — a different, simpler contract than `RoadmapView` because
it includes non-schedulable phases and carries no derived fields at all. The
critical design lock coming out of this step: `status` and `readiness` are
two independent dimensions in `FeatureView`. For `done`/`in-progress`
features, `readiness` is omitted entirely (`omitempty`) — the signal lives
solely in `status`. Only `planned` features get a `readiness` value of
`"ready"` or `"blocked"`, and only `"blocked"` rows carry a non-empty
`blockedBy`. All JSON paths must be byte-stable (ordered-slice iteration
only) and must preserve today's text output and error/exit-code behavior
exactly when `--json` is absent or when the source `roadmap.json` is
missing/malformed.

#### Gherkin Scenarios

See `specs/roadmap-json-contract.feature` for the full executable spec.
Key scenarios by group:

- **Design lock (readiness convention):** "readiness is empty for a done
  feature", "readiness is empty for an in-progress feature", "readiness is
  'ready' for a planned feature whose dependencies are all done", "readiness
  is 'blocked' for a planned feature with an unmet dependency", and the
  `Scenario Outline` table pinning all four (status, unmet) → (status,
  readiness) combinations, including the `(omitted)` convention for
  done/in-progress.
- **`roadmap --json` happy path:** "roadmap --json emits ordered phases and
  features with counts" (Given/When/Then: given a roadmap with 4 features
  across one status/readiness combination each, when `roadmap --json` runs,
  then it exits 0, is valid JSON, has `phases`/`counts` top-level fields, and
  features/counts match exactly); "Phase with zero features renders as an
  empty features array"; "Non-schedulable phases are excluded from roadmap
  --json".
- **`roadmap ready --json`:** "emits the ready feature names in declared
  order"; "ready --json set is identical to the readiness:ready set in
  roadmap --json" (cross-checks the two JSON surfaces against each other);
  "when nothing is ready emits an empty array, not null".
- **`roadmap show`/`list --json`:** "emits the persisted Roadmap verbatim,
  including non-schedulable phases"; "roadmap list --json is an alias for
  roadmap show --json"; "roadmap show (no flag) prints the same text as
  roadmap (no flag)" — pins the alias and the persisted-verbatim contract.
- **Determinism:** three scenarios (one per JSON surface) asserting
  byte-identical output across two consecutive runs on a fixed on-disk
  roadmap.
- **Empty roadmap:** three scenarios asserting `{"phases":[],"counts":{...
  all zero}}`, ready `"[]"`, and show's persisted-empty-struct verbatim.
- **Missing/malformed source:** five scenarios — missing file for each of
  the three `--json` surfaces plus text mode, and malformed JSON / a
  dependency cycle — all asserting non-zero exit, an stderr message, and no
  partial JSON on stdout.
- **Text-output regression guard:** two scenarios pinning `roadmap` and
  `roadmap ready` (no flag) as byte-identical to today's
  `ui.RenderRoadmap`/`ui.RenderReadyList` output.

#### UX States

| State | `roadmap` (text) | `roadmap --json` | `roadmap ready [--json]` | `roadmap show/list [--json]` |
|---|---|---|---|---|
| Loading | n/a (synchronous CLI, no async loading state) | n/a | n/a | n/a |
| Empty (roadmap has no phases) | unchanged existing empty-roadmap text rendering | `{"phases":[],"counts":{"planned":0,"inProgress":0,"done":0}}` | `[]` | valid JSON of the persisted empty `Roadmap` struct |
| Error (missing `.workflow/roadmap.json`) | non-zero exit, error on stderr | non-zero exit, error on stderr, no stdout JSON | non-zero exit, error on stderr, no stdout JSON | non-zero exit, error on stderr, no stdout JSON |
| Error (malformed JSON / dependency cycle) | non-zero exit, error on stderr (existing `Load()` behavior) | same — `Load()` rejects before any view is built | same | same (show still calls `Load()`) |
| Success (populated roadmap) | unchanged existing text rendering | full `RoadmapView` with `status`/`readiness`/`dependsOn`/`blockedBy`/`counts` | array of ready feature names | persisted `Roadmap` verbatim, including Backlog/Baseline |

#### Out-of-Scope

- Any mutation of the roadmap (`add`/`remove`/`edit`/`move`/`reorder`/phase
  operations) — reserved for `roadmap-crud-add-remove`, `roadmap-edit-move`,
  `roadmap-phase-ops`.
- The `draft` status/readiness dimension and the `Feature.Draft` field — not
  introduced until `roadmap-crud-add-remove`; this contract only exposes
  states that exist today (`planned|in-progress|done`, `ready|blocked`).
- MCP roadmap tools — MCP remains governance-read-only; no new MCP surface.
- Any Magallanes-side code — that lives in the separate `../magallanes` repo
  and is out of scope for this Centinela feature.
- Any change to the persisted `roadmap.json` schema — the view types are a
  pure projection; `Roadmap`/`Phase`/`Feature` are unchanged.

#### Deferred Findings

None. No new gaps surfaced beyond the umbrella design and the two
outstanding questions the big-thinker already flagged (both resolved below).
No `centinela roadmap defer` calls made.

#### Handoff

- **Next role:** senior-engineer
- **Open clarifications (resolved and locked in this step):**
  1. **`readiness` for `done`/`in-progress` rows:** confirmed and locked —
     `FeatureView.Readiness` is the empty string and is omitted from the
     JSON via `json:"readiness,omitempty"` for these rows. `status` alone
     carries the signal for `done`/`in-progress`; `readiness` only takes a
     value (`ready`/`blocked`) for `planned` features. Encoded in the
     Gherkin's design-lock scenarios and the `Scenario Outline` table.
  2. **`blockedBy` serialization:** confirmed — `omitempty`, so it is
     entirely absent (not `[]`, not `null`) for every feature except
     `readiness:"blocked"` rows. This is intentionally asymmetric with
     `dependsOn`, which is always present (even as `[]` when a feature has
     no dependencies) because `dependsOn` is persisted input, not derived
     state.
  3. **`roadmap --json` vs `roadmap show --json` scoping asymmetry:**
     confirmed and made explicit in the spec — `roadmap --json`
     (`RoadmapView`) excludes Backlog/Baseline (schedulable phases only,
     matching `Summary()`/`ready`); `roadmap show --json` (persisted
     `Roadmap` verbatim) includes them, and carries no derived fields
     (`status`/`readiness`/`blockedBy`) at all since it is a direct dump of
     the storage contract.
- No new open questions for senior-engineer; proceed to implement
  `internal/roadmap/view_types.go` + `view.go`, the `--json` flags on
  `roadmap.go`/`roadmap_ready.go`, and the new `roadmap_show.go` per the
  plan.
