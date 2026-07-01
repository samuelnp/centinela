# Feature Brief — roadmap-json-contract

> Feature 1 of 4 in the Roadmap Editing Suite. See the umbrella design:
> [docs/plans/roadmap-editing-suite-design.md](../plans/roadmap-editing-suite-design.md).

## Problem — what pain does this solve? Who is the user?

The `roadmap` command family prints **human text only** (except `brownfield
--json`). An external consumer — **Magallanes**, the multi-tenant SaaS control
plane — needs to render a "Plan project" page from a project's roadmap: phases,
features, per-feature status and readiness. Today its only options are to parse
`ROADMAP.md` prose or read `.workflow/roadmap.json` raw (which lacks derived
status/readiness, computed from workflow state at runtime). Magallanes already
shells out to `centinela validate`/`complete`; it needs the same shell-out path
to get a **stable, machine-readable roadmap view**. This feature delivers that
contract. It is read-only and unblocks Magallanes ahead of the mutation features.

## User Stories

- As a **Magallanes backend**, I want `centinela roadmap --json` to return every
  phase and feature with derived `status` and `readiness`, so the Plan page shows
  live progress without re-implementing Centinela's status logic.
- As a **Magallanes backend**, I want `centinela roadmap ready --json` to return
  the machine list of startable feature names, so I can offer "launch agent" only
  on features whose dependencies are met.
- As an **operator/script**, I want `centinela roadmap show --json` to dump the
  persisted roadmap verbatim, so I can diff or archive the raw source of truth.

## Acceptance Criteria (→ Gherkin)

1. `roadmap --json` emits a `RoadmapView`: ordered `phases[]`, each with ordered
   `features[]` carrying `name, phase, status (planned|in-progress|done),
   readiness (ready|blocked), dependsOn[], blockedBy[]`, plus a top-level
   `counts` object. Non-JSON output is byte-for-byte unchanged from today.
2. `roadmap ready --json` emits a JSON array of ready feature names, in declared
   roadmap order, identical set to the text view.
3. `roadmap show` (alias `list`) prints the roadmap as text; `roadmap show
   --json` emits the **persisted** typed `Roadmap` verbatim (storage contract).
4. All JSON output is deterministic / byte-stable across runs and platforms
   (ordered slices only; no map iteration).
5. Exit code 0 on success; non-zero with a stderr message when `roadmap.json` is
   absent or malformed — in both text and `--json` modes.

## Edge Cases

- **Empty roadmap** (`{"phases":[]}`): `--json` emits `{"phases":[],"counts":{…all
  zero}}`; `ready --json` emits `[]` (not `null`).
- **Missing `.workflow/roadmap.json`**: surfaced as an error, non-zero exit, no
  partial JSON on stdout.
- **Malformed JSON / dependency-cycle in source**: `Load()` already rejects;
  the command reports the error rather than emitting a half-built view.
- **Phase with zero features**: rendered as a phase with `features: []`.
- **Non-schedulable phases (Backlog/Baseline)**: included in `show --json` (raw)
  and in the status view, but their features are classified per existing rules
  (excluded from `ready`/counts as today).
- **`ready --json` when nothing is ready**: `[]`.

## Data Model

New view types in `internal/roadmap/` (derived, never persisted):

- `RoadmapView { phases []PhaseView; counts StatusCounts }`
- `PhaseView { name string; features []FeatureView }`
- `FeatureView { name, phase, status, readiness string; dependsOn, blockedBy []string }`
- `StatusCounts { planned, inProgress, done int }`

`status` derives from `FeatureStatus(name)` (workflow state); `readiness`/
`blockedBy` derive from `DeriveReadiness`/`classifyFeature` (`readiness.go`).
The persisted `Roadmap`/`Phase`/`Feature` types are unchanged. **Note:** the
`draft` status/readiness value is introduced later by `roadmap-crud-add-remove`
(which adds the `Draft` field); this feature exposes only the states that exist
today.

## Integration Points

- **Consumer:** Magallanes shells out to `centinela roadmap [--json]`,
  `roadmap ready --json`, `roadmap show --json`. No Centinela→Magallanes coupling.
- **Existing internals reused:** `roadmap.Load()`, `FeatureStatus`,
  `DeriveReadiness`/`ReadySet`, `ui.RenderRoadmap`/`ui.RenderReadyList`.
- **`--json` flag pattern:** mirrors `dashboard`/`verdict` (a `BoolVar` + inline
  `json.MarshalIndent`); view construction lives in `internal/roadmap`, only the
  one-line marshal lives in `cmd/centinela`.

## Risks

- **Determinism regression** (Medium): any map iteration in the view builder
  breaks byte-stability. Mitigation: iterate ordered slices only; add a
  byte-stable marshal test.
- **Contract drift** (Medium): Magallanes will pin to this JSON shape. Mitigation:
  treat field names as a stable contract; additive-only changes downstream.
- **Text-output regression** (Low): must not alter existing human output.
  Mitigation: `--json` gated strictly behind the flag; keep the default path.
- **Coverage gate** (Low): new logic needs colocated `internal/roadmap/*_test.go`
  (tests/ tier doesn't move the per-package 95% gate). Aim ≥97%.

## Decomposition

This is already the smallest slice of the four-feature suite. No further split.
Successor features (do not build here): `roadmap-crud-add-remove` (adds the
`draft` dimension + mutation), `roadmap-edit-move`, `roadmap-phase-ops`.
