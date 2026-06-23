# Feature Brief — brownfield-roadmap-generation

> Phase 9: Brownfield Onboarding, final slice. Where `analyze` captures *what the
> repo is*, `synthesize` drafts *PROJECT.md*, and `reconstruct` reconstructs the
> *behavioral contract*, `brownfield-roadmap-generation` produces the **roadmap**:
> a deterministic draft that records already-built capability as a **Baseline**
> (so it is never re-planned) and surfaces the net-new work and gaps (TODOs,
> incomplete areas, user-stated goals) as schedulable features the team can
> immediately `centinela start`. The fourth consumer of the frozen
> `analyze.Inventory`, same "deterministic skeleton + swappable inference seam"
> philosophy — **not** an in-process LLM call.

## Problem

Centinela's roadmap generator is greenfield-shaped: it assumes every feature is
still ahead of you. A mature codebase that has just been through
`analyze`/`synthesize`/`reconstruct` already contains a large body of working,
shipped capability — yet the generator re-proposes all of it as "to build",
burying the one thing the team wants (the next real gap) under phantom entries.
The root cause is structural: a feature's status is **not stored** in
`roadmap.json`; it is derived at read time from per-feature workflow state
(`roadmap.FeatureStatus`). A generated brownfield feature has no
`.workflow/<feature>.json`, so it always derives to `planned`. "Already built"
therefore cannot be expressed with a status field — it needs a different
representation.

## Who / why

The brownfield onboarder (team lead / staff engineer adopting Centinela on an
existing repo) and the operating agent that reads the roadmap to pick up work.
Both need a roadmap that cleanly separates "this exists, leave it" from "this is
the next gap." **Why now:** the Inventory contract is frozen and proven, and
`reconstruct` already enumerates the already-built surfaces as sorted `Targets`
with `# TODO: confirm` gap markers — the exact substrate this feature needs.

## In / Out scope

**In (v1):**
- New aggregator package (working name `internal/brownmap`) + thin
  `centinela roadmap brownfield` command.
- Reads `analyze.Inventory` (read-only) and regenerates a
  `reconstruct.Reconstruction` in-process (no new file contract).
- Emits a **draft `roadmap.Roadmap`** partitioned into a single **Baseline phase**
  (already-built, one entry per target) + net-new/gap phases (from reconstruct
  TODOs and `--goal` flags).
- Baseline is a **phase-name convention** (mirroring the existing `Backlog` /
  `Phase 0: Bootstrap` conventions): excluded from status counts and validate
  coverage, with no new field added to `roadmap.json`.
- Never clobbers a real `roadmap.json` — draft-only output, like `synthesize`'s
  `PROJECT.draft.md` and `reconstruct`'s review dir.
- Deterministic, byte-stable, no LLM; swappable generator interface with a
  deterministic default backend.
- G2 / `centinela.toml` aggregator-layer registration for the new package.

**Out (defer to roadmap):**
- Framework-specific gap detection (route/handler/job parsing) — unbounded across
  frameworks, like reconstruct's deferred route extraction.
- LLM-based gap inference / goal elicitation — v1 is deterministic.
- In-place merge/promote into an existing real `roadmap.json` — v1 emits a draft.
- Synthesizing fake "done" `.workflow/<feature>.json` stubs to fake Baseline
  status — rejected in favor of the phase convention.

## How it works (mechanism)

1. **Load** `.workflow/analysis.json` via `analyze.Load`; surface
   `analyze.ErrNoInventory` with "run `centinela analyze` first" (non-crashing,
   exit ≠0, no files written).
2. **Regenerate** the `reconstruct.Reconstruction` from the Inventory (sorted
   `Targets` = already-built surfaces; per-target `# TODO: confirm` count = gap
   signal). No new edge into reconstruct internals; uses its public result.
3. **Partition** into a `Baseline` phase (one `Feature` per target, `Source`
   recording provenance) and gap phase(s) seeded from TODO-bearing targets plus
   any `--goal` entries.
4. **Write draft, never clobber** — emit to a draft path (proposed
   `.workflow/roadmap.brownfield.json`) or stdout; refuse to overwrite the real
   `roadmap.json`. Print a concise summary (baseline count, gap count, draft path).
5. **Deterministic re-run** — byte-identical output for an unchanged Inventory
   (sorted targets, stable phase/feature order, no map-iteration order).

## Acceptance summary

1. `centinela roadmap brownfield` against a repo with valid analysis.json writes
   a draft roadmap containing a **Baseline** phase with ≥1 already-built entry and
   ≥1 net-new/gap phase (when TODOs or `--goal`s exist).
2. Baseline-phase features are **excluded** from status counts
   (`roadmap.Summary`) and validate coverage (the `NonBacklog`-equivalent set),
   via a single `isBaselinePhaseName` predicate mirroring `isBacklogPhaseName`.
3. The command **never clobbers** an existing `.workflow/roadmap.json`: it writes
   a draft / stdout and says which it did; an existing roadmap is byte-unchanged.
4. Output is **deterministic** — a second run on an unchanged Inventory produces
   byte-identical draft output.
5. Missing/old inventory surfaces `analyze.ErrNoInventory` with "run `centinela
   analyze` first" and a non-crashing non-zero exit; no files written.
6. The generator is an **interface** with a deterministic default backend, so an
   LLM backend can drop in without touching `cmd/`.
7. `internal/brownmap/**` is mapped as an **aggregator** in `centinela.toml` +
   PROJECT.md G2, with the `aggregator` `allow` extended to permit the
   `brownmap → reconstruct` aggregator-to-aggregator edge; no cross-layer import
   violation; all new source files ≤100 lines.

## Edge cases

- **No inventory** → `ErrNoInventory`, actionable message, exit ≠0, no writes.
- **Empty/doc-only inventory** (no targets) → empty Baseline; exit 0; summary
  reports "0 baseline / 0 gaps" rather than an empty malformed roadmap.
- **No TODOs and no `--goal`** → Baseline-only draft; explicitly reports "no gaps
  detected — supply --goal to add net-new work."
- **Existing `.workflow/roadmap.json` present** → never overwritten; draft goes to
  the draft path; reported as such.
- **Re-run after draft exists** → byte-identical draft overwrite; no spurious diff.
- **Polyglot / non-Go inventory** → still partitions from `inv.Packages` /
  manifests via reconstruct's degrade-gracefully target selection.
- **Baseline + gap slug collision** → a target that is both built and has TODOs
  appears once in Baseline (it exists) and its gap is a distinct net-new feature,
  deterministically named so the two never collide.
