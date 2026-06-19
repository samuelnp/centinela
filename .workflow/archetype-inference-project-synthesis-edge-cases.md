# Edge Cases: archetype-inference-project-synthesis

## Covered

- **Each archetype inferred from conventional signals** — n-tier (handler/
  service/repository), rails-native (Gemfile+rails+app/*), ecs (systems/
  components/entities), hexagonal (domain/application/ports/infrastructure),
  modular (modules/*/public+internal). Tests: `infer_test.TestInfer_Archetypes`;
  acceptance `TestAccSynth_Rails/ECS/GoNTierEndToEnd`.
- **Empty inventory** (no packages/manifests/language) → `custom` archetype,
  `low` confidence, draft with TODO stubs not fabricated paths. Tests:
  `infer_test` (empty case), `draft_test.TestDraft_EmptyEmitsStubs`.
- **Ambiguous tie** (two archetypes within the margin) → `Ambiguous`, forced
  `low` confidence, banner in the draft. Tests: `infer_test.TestInfer_AmbiguousTie`,
  `draft_test`, acceptance `TestAccSynth_AmbiguousLowConfidence`.
- **Deterministic ranking + byte-stable draft** — re-runs are identical. Tests:
  `infer_test`, `draft_test`, acceptance `TestAccSynth_Deterministic`.
- **Missing analysis.json** → `ErrNoInventory`, command exits non-zero with
  "run centinela analyze first", nothing written. Tests: `load_test`,
  `synthesize_errors_test`, acceptance `TestAccSynth_MissingInventoryFails`.
- **Malformed JSON** and **schema drift** → distinct errors (not ErrNoInventory).
  Tests: `load_test.TestLoad_MalformedJSON/SchemaDrift`, `synthesize_errors_test`.
- **Read error that is not not-found** (path is a directory) → distinct "reading"
  error. Test: `load_test.TestLoad_ReadErrorNotMissing`.
- **Existing PROJECT.md never clobbered** → writes `PROJECT.draft.md`, original
  byte-unchanged, command notes preservation. Tests: `write_test`,
  `synthesize_errors_test`, acceptance `TestAccSynth_ExistingPreserved`.
- **Un-writable output path** (under a regular file) → error, no partial file.
  Test: `write_test.TestWriteDraft_UnwritableDirErrors`.
- **Unmatched layer slot** → `<!-- TODO: confirm -->` row rather than a wrong
  path. Tests: `draft_test.TestMatchPaths_NoHitTODO`.
- **Unknown primary language** → naming-convention TODO row. Test:
  `draft_test.TestDraft_AmbiguousBannerAndUnknownLangNaming`.
- **Full analyze→synthesize pipeline** on a real Go fixture (file-system
  boundary). Test: `tests/integration/synthesize_pipeline_test.go`.

## Residual Risks

- **Heuristic mis-classification** on unconventional layouts (e.g. Centinela's
  own package names) → falls back to `custom`/low honestly; mitigated by the
  confidence + rationale + draft-first confirm/correct model. The `Inferer` is
  an interface, so a richer (or LLM) backend can replace it without API changes.
- **Framework not in the rule table** → that signal is ignored; folder/graph
  signals still apply. Adding a framework is a one-line `rules.go` table edit.
- **Monorepo / multiple manifests** → a single top-level archetype is inferred;
  per-module PROJECT.md is deferred to the brownfield-roadmap-generation /
  follow-up backlog.
- **Remaining PROJECT.md sections needing human input** (Elevator Pitch, Domain
  Language) → emitted as guided stubs; the user fills them before promoting the
  draft to PROJECT.md.
