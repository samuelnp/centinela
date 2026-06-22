# Feature-Specialist Report — spec-reconstruction

## Behavior Summary

`centinela reconstruct` is a deterministic Go command (no in-process LLM) that
reads the frozen `analyze.Inventory` (`.workflow/analysis.json`) and emits, into
a review dir (`.workflow/reconstructed/`), one `specs/<slug>.feature` Gherkin
skeleton + one `docs/features/<slug>.md` brief stub per selected
module/surface. Behavior the structural scan cannot know is rendered as explicit
`# TODO: confirm` markers — never a fabricated assertion. The operating agent
(human/LLM) later fills in the real Given/When/Then. Output is byte-stable
across re-runs, never clobbers a hand-authored `specs/<slug>.feature`, and a
missing/old inventory surfaces `analyze.ErrNoInventory` with a "run `centinela
analyze` first" message and a non-zero exit. Generation lives behind a swappable
`Reconstructor` interface (`NewReconstructor()` default) so an LLM backend can
drop in without touching `cmd/`.

## Gherkin Scenarios

Authored in `specs/spec-reconstruction.feature` (1 `Feature:` + 9 `Scenario:`
lines; verified parser-compatible with `internal/gates/spec_traceability_parse.go`
— a `Feature:` line plus lines matching `^\s+Scenario:`). Each scenario name is
unique and traceable to an acceptance test via a `// Scenario:` comment.

1. **A valid inventory reconstructs feature skeletons and brief stubs into the
   review dir** — happy path; >=1 `.feature` + >=1 brief stub with TODO markers
   (AC1).
2. **Every generated feature parses with the spec traceability scenario parser**
   — generated Gherkin validity against the real parser (AC2).
3. **Unknowable behavior is emitted as an explicit TODO confirm and never
   fabricated** — honest gaps, no invented assertions (AC3).
4. **Re-running reconstruct on an unchanged inventory produces byte-identical
   output** — determinism (AC4).
5. **A hand-authored spec is never clobbered and is reported as skipped** —
   skip-if-exists, byte-for-byte preservation, reported (AC5).
6. **Running reconstruct without an inventory fails with guidance and writes
   nothing** — `ErrNoInventory`, non-zero exit, no files written (AC6).
7. **The summary reports targets selected files written and total TODO markers**
   — stdout summary contract (AC7).
8. **An empty doc-only inventory selects zero targets and writes no empty
   feature** — empty-inventory edge: 0 targets, exit 0 (edge case).
9. **A polyglot inventory with an empty Go graph still selects manifest and
   package targets** — graceful degradation for non-Go inventories (edge case).

AC8 (`Reconstructor` interface seam) and AC9 (<=100-line files / aggregator
layer) are structural/architectural guarantees enforced by the senior-engineer
and gatekeeper, not behaviorally observable from the CLI, so they are not
expressed as runtime scenarios.

## UX States (CLI — stdout)

| State | Trigger | stdout / exit |
|-------|---------|---------------|
| Loading | n/a | Single synchronous pass; no progress UI (deterministic, fast, local). |
| Empty | Valid inventory with no behavioral packages | Summary reports `0 targets selected`; exit 0; no empty `.feature` emitted. |
| Error | Missing/old inventory (`ErrNoInventory`) | Error message "run `centinela analyze` first"; non-zero exit; nothing written. |
| Success | Valid inventory, >=1 target | Summary: targets selected, files written, files skipped, total TODO markers; exit 0; files in review dir. |
| Partial (skip) | Target whose canonical `specs/<slug>.feature` exists | That target reported as skipped; real spec untouched; remaining targets written; exit 0. |

Rendering is presentation-only in `internal/ui/render_reconstruct.go`;
`reconstruct` returns the typed `Reconstruction` (no business logic in the outer
layer / `cmd`).

## Out-of-Scope

- Framework-specific HTTP route / call-flow extraction across web frameworks
  (already deferred as `brownfield-route-flow-extraction`; not re-deferred here).
- Any in-process LLM inference (explicitly excluded by the decided approach;
  only the `Reconstructor` seam is provided).
- Changes to `internal/analyze` (the `Load`/`ErrNoInventory` seam is reused).
- New persisted schema (inputs are the frozen Inventory; outputs are text files).
- The downstream `brownfield-roadmap-generation` consumer itself.

## Deferred Findings

None new. Route/flow extraction is already captured on the roadmap as
`brownfield-route-flow-extraction` (per big-thinker). No additional out-of-scope
discoveries during spec authoring.

## Handoff

→ **senior-engineer.** Implement `internal/reconstruct/` (aggregator) per the
plan's source-file split, `cmd/centinela/reconstruct.go`, and
`internal/ui/render_reconstruct.go`. Every scenario in
`specs/spec-reconstruction.feature` must map 1:1 to an acceptance test carrying a
`// Scenario:` traceability comment (normalize: trim, collapse whitespace, strip
trailing period, lowercase — per `spec_traceability_parse.go`). The acceptance
suite must also parse every generated review-dir `.feature` with the real
`spec_traceability` parser to prove Gherkin validity (scenario 2). Honor the
`Reconstructor` interface seam (AC8) and the <=100-line / aggregator-layer
constraints (AC9) so the gatekeeper passes.
