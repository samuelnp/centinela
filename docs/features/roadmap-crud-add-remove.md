# Feature Brief — roadmap-crud-add-remove

> Feature 2 of 4 in the Roadmap Editing Suite. Umbrella design:
> [docs/plans/roadmap-editing-suite-design.md](../plans/roadmap-editing-suite-design.md).
> This brief isolates the acceptance criteria and edge cases that belong to THIS
> feature only (not roadmap-edit-move / roadmap-phase-ops).

## Problem — what pain, who

There is no way to author a roadmap feature: `defer` only captures Backlog
findings and `promote` only moves them behind a ≥9 gate. An operator (and the
Magallanes Plan page) needs to **add** a new feature directly, **remove** one,
and **finalize** a drafted feature — without breaking `roadmap validate`. This
feature delivers the create/remove half plus the draft lifecycle that makes
direct authoring gate-safe. Builds on `roadmap-json-contract` (shipped).

## Scope (this feature ONLY)
- **In:** `Draft bool` field + its hooks; `roadmap add`; `roadmap remove`/`rm`;
  generalized `promote` (in-place draft finalize); the generalized raw-feature
  helpers (`rawfeature_find.go`, `rawfeature_mutate.go`, `rawtyped.go`);
  extending the JSON view (`FeatureView.Draft` + `readiness:"draft"`) to expose
  the new dimension feature 1 deliberately deferred.
- **Out (successor features):** `edit`/`update`, `move`, `reorder` → roadmap-edit-move;
  `phase add`/`rename`/`remove` → roadmap-phase-ops.

## User Stories
- As an operator, `roadmap add <slug> --phase P` records a new feature as a
  **draft** in phase P so `roadmap validate` stays green until I score it.
- As an operator, `roadmap remove <slug>` deletes a planned feature, but is
  refused if another feature depends on it or it is already in progress/done.
- As an operator, `roadmap promote <draft-slug> --scores …` finalizes an
  in-place draft (clears the flag, writes analysis+quality) without moving it.
- As the Magallanes backend, `roadmap --json` now reports `draft` features with
  `readiness:"draft"` so the Plan page can show unscored items distinctly.

## Acceptance Criteria (THIS feature → Gherkin)
1. `roadmap add <slug> --phase <P> [--description --depends-on --archetype]`
   appends a feature to phase P with `draft:true`; `roadmap validate` stays
   PASS (the draft is exempt from the ≥9 analysis/quality coverage set).
2. `add` rejects: an invalid slug (non-kebab-case), a duplicate name (reporting
   the owning phase), an unknown/Backlog/Baseline target phase, a `dependsOn`
   referencing an unknown feature, and a `dependsOn` that would create a cycle
   (`ValidateDependencies` runs on the draft) — each leaving `roadmap.json`
   byte-identical.
3. `roadmap remove <slug>` (alias `rm`) deletes the feature and leaves the file
   valid; refuses when (a) another feature lists it in `dependsOn` (names the
   dependents) or (b) its `FeatureStatus` is in-progress/done.
4. `roadmap promote <slug> --scores ac,uv,dc,dep,ee,overall` on a feature that
   is a **draft already in a schedulable phase** clears `draft` in place (no
   phase move) and appends the analysis + quality artifacts; on a **Backlog**
   finding it retains today's move-into-`--phase` behavior. Non-draft, non-
   Backlog slug → clear error.
5. A draft feature is excluded from `roadmap ready`, is not counted as committed
   work in `Summary`, renders with a deterministic ` *(draft)*` marker in
   ROADMAP.md, and `centinela start <draft>` is refused (no scores yet → would
   bypass the ≥9 gate), mirroring the Backlog refusal.
6. `roadmap --json` emits `draft:true` + `readiness:"draft"` for draft features;
   output stays deterministic/byte-stable; existing non-draft output unchanged.

## Edge Cases (THIS feature ONLY)
- Empty roadmap (`{"phases":[]}`): `add` errors "unknown phase …" (no silent
  phase creation — that's roadmap-phase-ops).
- Missing/malformed `roadmap.json`: surfaced as error, file untouched.
- `add` duplicate across a different phase → collision reported with the owning
  phase.
- `add` targeting Backlog/Baseline → refused (drafts live in schedulable phases;
  Backlog authoring stays `defer`).
- `remove` a feature that does not exist → clear "not found".
- `remove` the last feature of a phase → phase remains with `features: []`
  (phase removal is roadmap-phase-ops).
- `remove` a feature depended on by a draft → still refused (drafts are real
  dependents).
- `promote` a draft whose scores include overall<9 → refused by existing score
  validation; draft flag left set.
- `promote` a Backlog finding with `--scores` → unchanged move-and-score path.
- Concurrent writers: inherited atomic temp+rename + one-feature-per-line render
  (last-writer-wins, not upgraded to locking) — same guarantee as defer/promote.

## Data Model
- `Feature` gains `Draft bool \`json:"draft,omitempty"\``. Single coverage-set
  hook in `NonBacklogFeatureSet` (`if f.Draft { continue }`). New `draft.go`
  (`IsDraftFeature`, `DraftFeatures`). `FeatureView` (from feature 1) gains
  `Draft bool` + a `readiness:"draft"` value. No other persisted-schema change.

## Integration Points
- Reuses the raw I/O layer (`rawio`/`rawmutate`/`rawmove`) and
  `ValidateDependencies`; generalizes `appendToPhase`/`removeBacklogFeature`
  into `rawfeature_*`. `promote` reuses its artifact-append path for finalize.
- Consumer: Magallanes reads the extended `roadmap --json`.

## Risks
- **Coverage-set leak** (High): if the draft exemption is added anywhere other
  than the single `NonBacklogFeatureSet` hook, semantics drift. Mitigation: one
  hook, unit-tested both ways.
- **Mutation not byte-preserving** (Med): new raw helpers must round-trip
  untouched phases; assert exact rendered bytes in tests.
- **promote branch ambiguity** (Med): Backlog-move vs in-place-finalize must be
  chosen by the slug's current location, not a flag; test both branches.
- **Per-package coverage** (Low): colocated `internal/roadmap/*_test.go`; ≥97%.

## Decomposition
Smallest coherent slice of the suite that stands alone. No further split; the
raw helpers it introduces are consumed by roadmap-edit-move and roadmap-phase-ops.
