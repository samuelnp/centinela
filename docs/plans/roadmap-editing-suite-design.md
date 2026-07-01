# Roadmap Editing Suite — Design

> Multi-feature design/decomposition doc. Individual feature plans (written during
> each feature's `plan` step) reference this. Not a single-feature plan itself.

## Context — why this exists

Today `centinela roadmap` can `defer` (append an out-of-scope finding to the
validate-exempt **Backlog**), `promote` (move a Backlog finding into a real phase
behind a quality-evaluator ≥9 pass), plus read-only `generate` / `ready` /
`validate` / `iterate` / `brownfield`. There is **no `add`, `remove`, `edit`,
`move`, or `reorder`** — you cannot deliberately author or curate the roadmap;
you can only capture findings and promote them.

The driver is **Magallanes** (separate Next.js SaaS control plane): it needs a
**"Plan project" page** where a user grows / edits / removes roadmap items, then
launches agents (via Capataz) to implement ready features. That requires two
things Centinela lacks: a full **editing command surface**, and a
**machine-readable contract** Magallanes can shell out to.

**Boundary:** this work is **Centinela-side only** — new CLI commands + a JSON
contract. The Magallanes Plan page is separate later work in `../magallanes`.

## The crux: the ≥9 quality gate

`roadmap validate` requires every **schedulable** feature to have an analysis
entry (role `senior-product-manager`) *and* a quality entry with **overall ≥ 9**
(`internal/roadmap/analysis.go`, `quality.go`; coverage set defined once in
`backlog.go` `NonBacklogFeatureSet`). Naively adding a feature to a real phase
therefore makes `validate` fail and blocks greenfield starts. The design must let
you author features **without** breaking `validate`.

## Decisions (agreed with product owner)

1. Build `add`, `remove`/`rm`, `edit`/`update`, `move`, `reorder`, plus phase ops
   (`phase add`/`rename`/`remove`) and `--json` read output.
2. **Draft/holding state** via a per-feature `Draft bool` field — new features
   land in their real target phase but flagged `draft:true`, which exempts them
   from the ≥9 coverage gate until scored. (Chosen over a "Draft" holding phase,
   which fights "add directly to a chosen phase" and re-breaks `validate` on the
   first `move` out.)
3. **No separate `finalize` verb — generalize `promote`.** `promote` gains an
   in-place mode: for a draft feature already in a schedulable phase,
   `promote <slug> --scores …` clears the draft flag and appends the
   analysis+quality artifacts **without moving** the feature. For a Backlog
   finding, `promote <slug> --phase <p> --scores …` moves-and-scores as today.
4. **`--json`** on read commands so Magallanes consumes a stable contract by
   shelling out (matching how it already calls `centinela validate`/`complete`).
   Mutations stay deterministic CLI. MCP stays governance-read-only (out of scope).

## Command surface

```
roadmap add <slug>     --phase --description --depends-on --archetype   # lands DRAFT
roadmap remove|rm <slug>                                                # guarded
roadmap edit|update <slug> --name --description --depends-on --archetype
roadmap move <slug>    --to-phase [--before|--after <anchor>]
roadmap reorder <slug> [--before|--after <anchor>]
roadmap promote <slug> [--phase <p>] --scores ac,uv,dc,dep,ee,overall [--summary]
                       #  Backlog finding -> move+score;  in-place draft -> finalize+score
roadmap phase add <name> [--note --after] | rename <old> <new> | remove <name> [--force]
roadmap show|list [--json]      # + --json on `roadmap` and `roadmap ready`
```

## Draft mechanism (details)

Add `Draft bool \`json:"draft,omitempty"\`` to `Feature` (`types.go`). Functions
that must learn about it:

| Location | Change |
|---|---|
| `backlog.go` `NonBacklogFeatureSet` | `if f.Draft { continue }` — the **single** place drafts leave the coverage set, so `validate` never demands artifacts for a draft |
| `roadmap.go` `Summary()` | skip/segregate drafts so they aren't counted as committed `planned` work |
| `readiness.go` `DeriveReadiness`/`classifyFeature` | draft → `State:"draft"`, excluded from `ReadySet` |
| `cmd/centinela/start_guard.go` | refuse `start` on a draft (no scores yet → would bypass the ≥9 gate); mirror the existing Backlog refusal |
| `mdgen_feature.go` | render a deterministic ` *(draft)*` marker in ROADMAP.md |

`ValidateDependencies` **still** runs on drafts — a draft must be structurally
sound (legal archetype, deps reference known features, no cycle) even if unscored.

## Machine-readable `--json` contract

- `roadmap --json` → `BuildView(r)` (`RoadmapView`): per-feature `name, phase,
  status (planned|in-progress|done), readiness (ready|blocked|draft), draft,
  dependsOn, blockedBy` plus `counts`. The rich status contract Magallanes drives.
- `roadmap ready --json` → `ReadySet(r)` (array of names).
- `roadmap show --json` → the persisted typed `Roadmap` verbatim (storage contract).

Deterministic/byte-stable (ordered slices only). Marshalling is a one-liner in each
cmd (dashboard/verdict precedent); view construction lives in `internal/roadmap`.

## Mutation invariants (reuse existing raw layer)

All mutations use the **format-preserving raw I/O** (`rawio.go`, `rawmutate.go`,
`rawmove.go`, `rawrender.go`) that `defer`/`promote` already use: read → mutate
`rawDoc` in memory → run `toRoadmap()` + `ValidateDependencies` as the structural
gate → `writeRawRoadmap` **once** (atomic temp+rename, one-feature-per-line,
merge-friendly). A rejected mutation leaves `roadmap.json` byte-identical.

Guards: `remove` rejects a depended-on feature or an in-progress/done one;
`edit --name` rewrites every dependent's `dependsOn` and checks collisions;
`edit`/`move` re-run `ValidateDependencies` so a cycle is rejected before any
write; `move` refuses Backlog/Baseline as source/target (directs to `defer`/
`promote`); `phase remove` refuses a non-empty phase unless `--force`. The one
genuinely new raw complexity is phase insert/remove **reindexing the `dirty` map**
(isolated in a `rawphaseops.go`/`rawphase_struct.go` pair).

## Decomposition — four features, in dependency order

1. **`roadmap-json-contract`** — read-only, zero mutation. `view_types.go`,
   `view.go` (`BuildView`), `--json` on `roadmap`/`ready`, new `roadmap show`.
   Unblocks Magallanes first; lowest risk.
2. **`roadmap-crud-add-remove`** — introduces the **draft field + all its hooks**
   (`backlog`/`readiness`/`Summary`/`mdgen`/start-guard/`draft.go`), the
   generalized raw-feature helpers (`rawfeature_find.go`, `rawfeature_mutate.go`,
   `rawtyped.go`), `add`, `remove`, and the **generalized `promote`** (in-place
   draft finalize via the shared artifact path). Must precede 3 & 4.
3. **`roadmap-edit-move`** — `edit` (+ `rawdeps.go` dependent rewrite), `move`,
   `reorder`, and cycle re-validation across renames/moves.
4. **`roadmap-phase-ops`** — `phase add/rename/remove` and the
   `rawphaseops.go`/`rawphase_struct.go` dirty-reindex machinery (highest-
   complexity raw change), landed last.

## Constraints & tests

- Every source file **and** `_test.go` ≤ 100 lines → most capabilities split
  across 2–3 small files (one verb per file, matching `roadmap_defer.go`).
- Coverage gate runs without `-coverpkg`: new `internal/roadmap` logic must be
  covered by **colocated `internal/roadmap/*_test.go`**; `tests/`-tier files
  contribute zero to the per-package 95% gate. Aim ≥97%.
- Edge cases to cover per feature: empty roadmap, missing `roadmap.json`, add to
  nonexistent phase, duplicate name, remove depended-on / in-progress, rename
  collision, cycle via edited deps, Backlog/Baseline move refusal, draft vs ≥9
  gate, non-draft promote-in-place, non-empty phase remove, concurrent writes
  (inherited atomic write; last-writer-wins, not upgraded to locking).

## Backlog note

The four sub-features above are recorded in the roadmap **Backlog** (the only
non-gate-breaking add path today — dogfooding the gap this suite fixes). Once
feature 2 lands, they can be `promote`d into a real "Roadmap Authoring" phase.
