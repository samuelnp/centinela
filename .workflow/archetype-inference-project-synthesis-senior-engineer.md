# archetype-inference-project-synthesis — senior-engineer

## Files Touched

**New aggregator package `internal/synthesize`**: archetype.go (Archetype/Signal/
Score/Inference types + Reasons), signals.go (lowercased Inventory view +
predicates), rules.go (deterministic scoring table), infer.go (Inferer interface
+ ruleInferer + rank/classify confidence), profiles.go (per-archetype
template data: pattern/G2/G7/reference/layer slots), draft.go (Draft orchestrator
+ header/banner), sections_arch.go (Architecture Choice / Layer Mapping /
Gatekeeper Paths, package-bucketed), sections_meta.go (Tech Stack / Folder /
Locales), sections_naming.go (language-keyed naming), write.go (don't-clobber
WriteDraft → PROJECT.draft.md).

**Contract owner:** `internal/analyze/load.go` — `Load(path)` with ErrNoInventory
/ malformed / schema-drift errors (mirrors `Save`).

**Command:** `cmd/centinela/synthesize.go` (thin, mirrors analyze.go; --in/--out/
--json; actionable missing-inventory error). **UI:** `internal/ui/render_synthesize.go`.

**Config:** `centinela.toml` (synthesize added to the aggregator layer) +
`PROJECT.md` G2 prose (registers synthesize as aggregator + ui→synthesize edge).

## Architecture Compliance

`internal/synthesize` is an **aggregator** (imports the `internal/analyze` domain
read-only + stdlib only), mirroring doctor/insights/calibration/audit — avoids a
forbidden domain→domain edge; `analyze` never imports `synthesize` (no cycle).
Dogfooded `pr-gate`: import_graph = **0 failed**. `cmd/synthesize.go` is a thin
orchestrator (G7 — no business logic). All 15 source files ≤100 lines (G1).

## Type-Safety Notes

`Archetype` is a typed string with enumerated consts; the rule table is data
(`[]rule` with typed match funcs), not control flow, so adding a signal is a
table edit. `Inferer` is an interface (`NewInferer()` → deterministic default) —
the swap seam for a future LLM backend without touching cmd/ or the synthesizer.
No `interface{}`/`any`. JSON decode into the typed `analyze.Inventory`.

## Trade-Offs

- **Deterministic + template, no LLM** — byte-stable, offline-testable, matches
  the analyze substrate. Verified: a Go n-tier fixture infers n-tier/high and
  renders a full draft; this repo's unconventional package names (gates/workflow/
  analyze) honestly infer `custom`/low rather than guess.
- **Never clobbers** — `WriteDraft` writes `PROJECT.draft.md` when `PROJECT.md`
  exists; payload written in one call (no partial file).
- **Honest gaps** — unmatched layer slots + human-only sections emit
  `<!-- TODO: confirm -->`.

## Handoff

→ qa-senior: colocated unit tests for the synthesize package (infer scoring/
confidence/ambiguity over Inventory fixtures; rules/predicates; draft + sections
content assertions; write clobber-safety) for the 95% per-package gate;
`analyze.Load` tests (missing/malformed/drift/round-trip); cmd synthesize +
errors tests; an analyze→synthesize integration test; and the acceptance suite
for the 7 spec scenarios with `// Acceptance:`/`// Scenario:` traceability. No
LLM/network — Inventory fixtures only.
