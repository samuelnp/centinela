# Edge Cases: spec-reconstruction

## Covered

- **Missing/old inventory** — `--in` points at a non-existent file → wraps
  `analyze.ErrNoInventory` with "run `centinela analyze` first", exits non-zero,
  writes no corpus. (`reconstruct_errors_test.go`, `TestAccRecon_NoInventoryFails`)
- **Empty / doc-only inventory** — no behavioral packages (`docs`, `readme`) →
  zero targets selected, exit 0, no empty `.feature` written.
  (`select_test.go:TestSelect_EmptyAndDocOnlyZeroTargets`, `TestAccRecon_DocOnlyZeroTargets`)
- **Polyglot inventory, empty Go graph** — targets still derived from the
  express manifest + `src/api/users` package; Go-graph absence degrades
  gracefully. (`select_test.go:TestSelect_PolyglotEmptyGraphFromManifest`,
  `TestAccRecon_PolyglotEmptyGraph`)
- **Hand-authored spec exists** — a canonical `specs/<slug>.feature` is skipped
  (recorded in `Skipped`, both feature + brief suppressed), left byte-for-byte
  unchanged, and reported. (`write_test.go`, `TestAccRecon_SkipsHandAuthoredSpec`)
- **Re-run determinism** — two runs on an unchanged Inventory produce
  byte-identical files. (`reconstructor_test.go`, `TestReconstructPipeline`,
  `TestAccRecon_Deterministic`)
- **Slug collisions** — two packages mapping to the same stem are
  disambiguated deterministically in pre-sort order. (`slug_test.go`,
  `select_test.go`)
- **Bounded corpus** — a package list beyond `maxTargets` (50) is capped so the
  review set stays reviewable. (`select_test.go`)
- **Exclusion precedence** — test-only / generated / vendored / config-leaf
  packages are excluded even when they would otherwise promote; exclusion wins.
  (`select_exclude_test.go`)
- **Target with no inferable behavior** — still yields a `Feature:` + a single
  `# TODO: confirm` scenario stub, never empty, never an assertion.
  (`feature_test.go`)
- **Generated Gherkin validity** — every emitted `.feature` carries a `Feature:`
  line + ≥1 `Scenario:` line and parses with the real `spec_traceability`
  parser. (`feature_test.go`, `TestAccRecon_GeneratedFeaturesParse`)

## Residual Risks

- **Thin skeletons** — a `Feature:` + lone TODO per module carries limited value
  until the LLM seam fills behavior; mitigated by role-aware target selection.
  The behavioral lift is deferred to the swappable `Reconstructor` backend.
- **Framework route/flow extraction** — real HTTP route / call-flow extraction
  is out of v1 scope; deferred to the roadmap as `brownfield-route-flow-extraction`.
- **`internal/ui` / `cmd/centinela` per-package coverage** — these sit slightly
  below 95% in isolation (91.8% / 93.4%); the total gate passes at 95.2% on the
  strength of `internal/reconstruct` at 99.3%. The new reconstruct code itself is
  fully exercised.
