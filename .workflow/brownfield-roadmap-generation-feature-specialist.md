### Feature-Specialist Report: brownfield-roadmap-generation
**Date:** 2026-06-23

#### Behavior Summary

`centinela roadmap brownfield` is a deterministic, no-LLM subcommand under the
existing `roadmap` group. It loads the frozen `analyze.Inventory` from
`.workflow/analysis.json`, regenerates a `reconstruct.Reconstruction` in-process
from that same Inventory, and emits a **draft** `roadmap.Roadmap` partitioned into
(a) a single **Baseline** phase — one feature per already-built reconstruct target,
identified by a `BaselinePhaseName` phase-name convention that exempts it from
status counts and validate coverage exactly like the existing `Backlog` phase, with
no schema change to `roadmap.json` — and (b) one or more **net-new gap** phases,
seeded from per-target `# TODO: confirm` markers and any repeatable `--goal`
strings. It writes the draft to a distinct draft path (or stdout/`--json`) and
**never** overwrites a canonical `.workflow/roadmap.json`, mirroring how
`synthesize` and `reconstruct` refuse to clobber hand-authored files. Output is
byte-stable: a second run over an unchanged Inventory produces a byte-identical
draft. A missing/old inventory surfaces `analyze.ErrNoInventory` with an actionable
"run `centinela analyze` first" message, a non-zero exit, and zero files written.

#### Gherkin Scenarios   (reference specs/brownfield-roadmap-generation.feature)

The acceptance contract is the ten scenarios in
`specs/brownfield-roadmap-generation.feature`. Each Given/When/Then maps to an
executable assertion the qa-senior will make real (exit code, draft file existence,
Baseline phase presence, canonical `roadmap.json` byte-equality, byte-identical
re-run, gap/goal feature presence, summary text):

1. **Built repo → Baseline phase** — valid `analysis.json` ⇒ exit 0, draft written,
   draft contains a Baseline phase (by convention name) with ≥1 already-built
   feature.
2. **Never clobbers canonical roadmap.json** — an existing hand-authored
   `.workflow/roadmap.json` is left byte-for-byte unchanged; the draft goes to the
   draft path; summary reports which path it wrote.
3. **Gap phase from reconstruct TODOs** — TODO-bearing targets ⇒ ≥1 gap phase
   distinct from Baseline; each TODO target appears as a schedulable gap feature.
4. **`--goal "<text>"` adds a net-new gap feature** — goal-derived feature appears
   in a gap phase, never in Baseline.
5. **Baseline excluded from status counts + coverage** — `Summary` and the
   non-schedulable coverage set both exclude Baseline features, via the same
   predicate mechanism that exempts Backlog.
6. **Determinism** — two runs on an unchanged Inventory into the same draft path ⇒
   both exit 0 and the draft file is byte-identical.
7. **Missing inventory** — no inventory ⇒ non-zero exit, "run `centinela analyze`
   first" message, no draft written.
8. **Empty/doc-only inventory** — no behavioral packages ⇒ exit 0, summary reports
   0 baseline / 0 gaps, no malformed empty roadmap.
9. **No TODOs and no goals** — Baseline-only draft, no gap phase, summary hints to
   supply `--goal`.
10. **Summary contract** — stdout reports baseline count, gap count, and draft path.

#### UX States  (table)

| State | CLI condition | Observable behavior |
|-------|---------------|---------------------|
| Loading | n/a (synchronous, no network/LLM) | Single in-process pass; no spinner/progress needed |
| Empty | Inventory present but no significant surfaces (doc-only / no behavioral packages) | Exit 0; summary reports `0 baseline / 0 gaps`; valid (non-malformed) draft or explicit "nothing to draft" — no crash |
| Empty (no gaps) | Built surfaces but no TODO markers and no `--goal` | Exit 0; Baseline-only draft; summary hints "no gaps detected — supply --goal to add net-new work" |
| Error | Missing/old `analysis.json` (`analyze.ErrNoInventory`) | Exit ≠0; stderr message "run `centinela analyze` first"; **no files written** |
| Success | Valid inventory with targets (+ optional TODOs/goals) | Exit 0; draft written to draft path (canonical `roadmap.json` untouched); summary prints baseline count, gap count, draft path |

#### Out-of-Scope

Restated from the plan and big-thinker "Out (v1)":

- **Framework-specific gap detection** — parsing routes/handlers/jobs per framework
  to infer missing endpoints. Unbounded across frameworks (mirrors reconstruct's
  deferred route extraction). Gaps come only from reconstruct `# TODO: confirm`
  markers + `--goal`.
- **LLM-based gap inference / goal elicitation** — v1 is deterministic; no model
  reasoning. The `Brownfielder` interface is the swap seam for a future LLM backend
  without touching `cmd/`.
- **In-place merge / promote into an existing real `roadmap.json`** — v1 emits a
  draft the user reviews and adopts; in-place merge is a follow-up.
- **Faking "done" workflow state** — no synthesized completed
  `.workflow/<feature>.json` stubs to fake Baseline "done" status; the phase-name
  convention is cleaner, honest, and reversible.

#### Deferred Findings

None. The two natural follow-ups (framework-specific gap detection and
promote-brownfield-draft-into-roadmap) are already captured in this feature's "Out"
scope, consistent with the big-thinker handoff; no genuinely new out-of-scope
discovery surfaced during acceptance-contract authoring. The plan is complete and
self-consistent with the brief and big-thinker report, so no plan refinement was
needed.

#### Handoff — Next role: senior-engineer

- **Acceptance contract:** `specs/brownfield-roadmap-generation.feature` (10
  scenarios) is the binding contract. Keep Then-steps assertable: exit code, draft
  file existence at the draft path, Baseline phase present by convention name,
  canonical `roadmap.json` byte-unchanged, byte-identical re-run, gap/goal features
  present in a non-Baseline phase, summary text (baseline/gap counts + draft path).
- **Key implementation seams to honor** (from the plan): the `Brownfielder`
  interface + `NewBrownfielder` deterministic default; `BaselinePhaseName` +
  `isBaselinePhaseName`/`IsBaselinePhaseName` mirroring `backlog.go`, wired into the
  same four Backlog skip sites (prefer a single shared `isNonSchedulablePhase`
  helper); never-clobber draft writer (atomic temp-file+rename, refuse the canonical
  `RoadmapFile`); the `aggregator` layer `allow += "aggregator"` toml edit + PROJECT.md
  G2 prose landing **with** the code; all source files ≤100 lines.
- **Open decisions for senior-engineer to lock** (from big-thinker handoff): the
  exact draft path string and flag set (`--in`/`--out`/`--json`/`--goal`); the
  `BaselinePhaseName` string (and whether it carries a status glyph); whether
  per-target TODO counts need a tiny accessor added to reconstruct (no logic dup).
