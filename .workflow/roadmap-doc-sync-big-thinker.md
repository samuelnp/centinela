# Big-Thinker Report — roadmap-doc-sync

## Decision
Make `.workflow/roadmap.json` the single source of truth and **generate** `ROADMAP.md`
from it, guarded by a `roadmap_drift` gate that byte-compares the on-disk file to
generator output. User-chosen approach: **enrich the JSON with prose** (per-feature
`description`/`fixes`, per-phase `note`, top-level `intro`) and migrate all existing
hand-written ROADMAP.md narrative into it, so generation is full-fidelity and no prose
is lost. Motivation: this roadmap has drifted from roadmap.json twice while hand-synced.

## Shape
- Schema: extend `Feature`/`Phase`/`Roadmap` in `internal/roadmap` (raw-mutation path
  preserves unknown keys, so defer/promote stays intact).
- Generator: pure `RenderMarkdown(*Roadmap) ([]byte, error)` in `internal/roadmap/`
  (split mdgen.go / mdgen_phase.go / mdgen_feature.go for the ≤100-line rule). NOT in
  `internal/ui` (that is terminal presentation).
- Command: `centinela roadmap generate` (thin orchestrator in `cmd/`).
- Gate: `internal/gates/roadmap_drift.go` + `internal/config/roadmap_drift.go`
  (Enabled + Severity warn|fail), modeled on spec-traceability; wired into
  `RunWithFilter` + `GatesConfig` + `centinela.toml` (ships `enabled=true severity="warn"`).
- Migration: populate JSON from current prose, regenerate ROADMAP.md byte-exact.

## Determinism guarantees
Iterate only ordered slices (Phases/Features/DependsOn), never Go maps. No live
per-feature status glyph in the generated file — live status stays in `centinela roadmap`.
Exactly one trailing newline, LF-only, no trailing whitespace.

## Dominant risk
The ≤100-line G1 rule: every new source + test file is pre-split and budgeted. Coverage
gate is per-package (95%) — colocated `_test.go` files move it, tests under `tests/` do not.

## Handoff
→ feature-specialist (Gherkin spec + edge-case hardening).
