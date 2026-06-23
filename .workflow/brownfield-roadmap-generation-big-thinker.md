### Big-Thinker Report: brownfield-roadmap-generation
**Date:** 2026-06-23

#### Problem

Centinela's roadmap generator is **greenfield-shaped**: every feature it emits is
treated as net-new work that lies ahead of you. When a mature codebase adopts
Centinela through Phase 9 (`analyze` → `synthesize` → `reconstruct`), the team
already has a large body of *working, shipped* capability. The current generator
has no way to say "this already exists, don't re-plan it." The result is a
roadmap that re-proposes functionality the team built years ago, drowning the one
thing they actually want — **the next real gap they can `centinela start`** — in
dozens of phantom "to build" entries.

**Who is hurting:** the brownfield onboarder (a team lead / staff engineer running
the Phase-9 flow on an existing repo) and, by extension, the operating agent that
reads the roadmap and tries to pick up work — it cannot distinguish "done" from
"to do" because *status is not stored in `roadmap.json`*; it is derived at read
time from per-feature workflow state (`roadmap.FeatureStatus`). A freshly
generated brownfield feature has **no** `.workflow/<feature>.json`, so it derives
to `planned` — i.e. everything looks un-started, including the already-built.

**What they do now:** nothing automated. After `reconstruct` writes a spec corpus,
they hand-curate a roadmap, manually deleting/annotating the things that exist —
exactly the error-prone busywork the framework exists to eliminate.

**Why now:** `analyze.Inventory` is frozen and proven, and `reconstruct` shipped a
deterministic `Reconstruction{Targets, ...}` that already enumerates the
already-built surfaces (one target per significant module). The substrate for
"what exists" is sitting on disk. This feature is the last Phase-9 slice that turns
"what the repo is / does" into "what is left to do."

#### Scope (In / Out)

**In (v1):**
- A new aggregator package `internal/brownmap` (working name) and a thin
  `centinela roadmap brownfield` (subcommand of the existing `roadmap` group)
  that reads `analyze.Inventory` + the `reconstruct.Reconstruction` (regenerated
  in-process from the same Inventory, no new file contract) and emits a
  **draft roadmap** that partitions capability into:
  - a single **"Baseline" phase** holding one already-built entry per reconstruct
    target (the things that exist), and
  - one or more **net-new / gap phases** holding the TODOs and user-stated goals.
- The Baseline phase is a **phase-name convention** (exactly like the existing
  `Backlog` / `Phase 0: Bootstrap` conventions in `internal/roadmap`), so its
  entries are excluded from status counts and validate coverage — they are facts,
  not schedulable work — *without* adding a status field to `roadmap.json`.
- Never-clobber output: write to a draft path (`.workflow/roadmap.brownfield.json`
  or stdout/`--out`), never overwrite a real `.workflow/roadmap.json`, mirroring
  `synthesize` (`PROJECT.draft.md`) and `reconstruct` (`.workflow/reconstructed/`).
- Deterministic, byte-stable, no LLM — same philosophy as its three siblings.
- A swappable generator interface (`Brownfielder`-style) with a deterministic
  default backend, the LLM drop-in seam its siblings all established.
- The G2 / `centinela.toml` aggregator-layer registration for the new package
  (+ the read-only `roadmap` edge it needs to emit `roadmap.Roadmap` values).

**Out (v1, defer to roadmap):**
- Framework-specific gap detection (parsing routes/handlers/jobs per framework to
  find missing endpoints) — unbounded across frameworks, like reconstruct deferred
  route extraction.
- LLM-based gap inference / user-goal elicitation — v1 stays deterministic; gaps
  come from `reconstruct`'s `# TODO: confirm` markers + a `--goal` flag, not from
  model reasoning.
- Editing / merging into an existing real `roadmap.json` (promote-into-roadmap).
  v1 emits a draft the user reviews and adopts; in-place merge is a follow-up.
- Re-deriving Baseline "done" status by synthesizing fake completed
  `.workflow/<feature>.json` stubs — rejected (see Dependencies); the phase
  convention is cleaner and reversible.

#### Dependencies & Assumptions

- **Builds on `analyze.Inventory`** (read-only via `analyze.Load`, the frozen
  seam) **and `reconstruct`** — the new package regenerates a
  `reconstruct.Reconstruction` in-process from the same Inventory (`Targets` =
  already-built surfaces; `TodoCount` / per-target TODOs = gap signal). No new
  on-disk contract is required between reconstruct and this feature.
- **Emits `roadmap.Roadmap` / `Phase` / `Feature`** values (the frozen domain
  types in `internal/roadmap/types.go`) and persists via a draft writer — it does
  **not** call `roadmap.Save` against the canonical file.
- **Layer placement — aggregator.** `internal/brownmap` imports `internal/analyze`
  (domain, read-only), `internal/reconstruct` (aggregator, read-only — sibling
  aggregator-to-aggregator import), and `internal/roadmap` (read-only, for the
  `Roadmap`/`Phase`/`Feature` types + the `BacklogPhaseName`-style convention
  constant) + stdlib only. It must not import `cmd/` or `internal/ui`; it is
  imported only by `cmd/` and its result type by `internal/ui` for rendering.
- **G2 import-graph implications (must be documented, like its siblings):**
  - `centinela.toml`: add `internal/brownmap/**` to the `aggregator` layer
    `paths`. **Caveat:** the aggregator layer currently `allow = ["domain",
    "leaf"]`. `roadmap` is a **domain** package (it imports `internal/workflow`),
    so `brownmap → roadmap` and `brownmap → analyze` are already covered by
    `allow=["domain"]`. The **new** wrinkle is `brownmap → reconstruct`
    (aggregator→aggregator), which the current `allow=["domain","leaf"]` does
    **not** permit. Resolve by either (a) adding `"aggregator"` to the aggregator
    layer's `allow` (cleanest; lets aggregators compose), or (b) having `brownmap`
    re-run reconstruct's *logic* without importing it — rejected, that duplicates
    the rule table. Recommend (a), documented in PROJECT.md G2 prose exactly as
    `synthesize`/`reconstruct` were.
  - PROJECT.md: register `internal/brownmap` in the G2 prose + folder structure +
    layer/gatekeeper tables, stating its three read-only edges and the no-cycle
    argument (`analyze`/`reconstruct`/`roadmap` never import `brownmap`).
- **Assumption:** the operating agent adopts the draft by reviewing it and copying
  the net-new phases into the real `roadmap.json`; Baseline entries are
  informational and need no workflow state.
- **Scaffold mirror:** the `centinela.toml` aggregator-paths edit lives in this
  repo's config; mirror into `internal/scaffold/assets` only if the parity test
  covers that toml (the synthesize/reconstruct plans both noted this — verify).

#### Risks (Risk | Impact | Likelihood | Mitigation)

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Clobbering a real `roadmap.json` (data loss of a curated roadmap) | High | Med | Draft-only output to a distinct path / stdout; never call `roadmap.Save` on the canonical file; acceptance test asserts an existing `roadmap.json` is byte-unchanged. |
| Baseline entries leak into status counts / validate coverage (re-planned anyway) | High | Med | Reuse the proven phase-name-convention mechanism (`isBacklogPhaseName` pattern): a `BaselinePhaseName` constant + a single `isBaselinePhaseName` predicate wired into `Summary`, the `NonBacklog` coverage set, readiness, and the render skip — one place, mirroring Backlog. |
| Aggregator→aggregator edge (`brownmap → reconstruct`) trips the `import_graph` gate | Med | High (if unplanned) | Add `"aggregator"` to the aggregator layer `allow` in `centinela.toml`; document in PROJECT.md G2; an unmapped package only warns, a stray edge fails — so the mapping must land with the code. |
| Thin/low-value output — Baseline that just mirrors reconstruct targets feels like noise | Med | Med | Baseline is the *point* (it's what prevents re-planning); value is concentrated in the gap phases seeded from real `# TODO: confirm` markers + `--goal`. Bound and sort the Baseline list (reuse reconstruct's sorted Targets). |
| Greenfield regression — the existing `roadmap generate`/bootstrap path changes behavior | High | Low | New subcommand + new package; zero edits to existing `roadmap_generate.go` / bootstrap logic. The only shared edit is the additive phase-convention predicate, guarded by tests on the existing Backlog/Bootstrap conventions. |
| Determinism drift (map-iteration order in emitted phases) | Med | Low | Pure string/struct assembly over reconstruct's already-sorted `Targets`; byte-stable acceptance test on re-run, mirroring siblings. |
| Missing/old inventory | Low | Med | Surface `analyze.ErrNoInventory` with "run `centinela analyze` first"; non-crashing non-zero exit, no files written — identical to reconstruct's contract. |

#### Rollout

Smallest correct slice first, each independently shippable:

1. **Baseline-only draft.** Read Inventory → regenerate `Reconstruction` →
   emit a `Roadmap` with a single `Baseline` phase (one entry per target) +
   never-clobber draft writer + the `BaselinePhaseName` convention predicate +
   its wiring into status/coverage exemption. This alone fixes the core defect
   ("already-built isn't re-planned"). Ship the G2 / toml mapping here.
2. **Gap phase from reconstruct TODOs.** Add a net-new phase seeded from
   per-target `# TODO: confirm` signals, so the team sees the real incomplete
   areas as schedulable features.
3. **User-stated goals.** A `--goal` (repeatable) flag appends explicit net-new
   gap features the scan can't infer — deterministic, no LLM.
4. **UI render + summary** (`internal/ui/render_brownfield.go`): baseline count,
   gap count, draft path written — presentation only.

(Promote-into-existing-roadmap and framework gap detection are explicitly out;
defer to roadmap if discovered as worth doing.)

#### Deferred Findings
 - none (the two natural follow-ups — framework-specific gap detection and
   promote-brownfield-draft-into-roadmap — are already implied by this feature's
   "Out" scope and the existing reconstruct route-extraction deferral; no
   genuinely *new* out-of-scope discovery surfaced.)

#### Handoff
 - Next role: feature-specialist. Open questions to nail down: exact package +
   subcommand naming (`internal/brownmap` vs `internal/brownfield`;
   `centinela roadmap brownfield` vs a top-level verb); the `BaselinePhaseName`
   string and whether it carries a status glyph; the draft output path/flags
   (`--in`/`--out`/`--json`/`--goal`); whether to add `"aggregator"` to the
   aggregator `allow` (recommended) vs avoid the reconstruct import; and the file
   split keeping every source file ≤100 lines.
