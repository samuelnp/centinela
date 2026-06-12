# Edge Cases: workflow-archetypes

## Covered

Each entry: **risk** — how it is tested.

- **Spike order has no validate step (ungated by absence, not a bypass).**
  Risk: spike could be (mis)read as a verification bypass; if it shipped via an
  `if archetype == "spike"` branch, the gate could be skipped for any feature.
  Tested: `internal/workflow/archetype_safety_test.go` pins spike order ==
  `[plan, code]` and asserts it contains no `validate`; the same test pins that
  hotfix/refactor/canonical DO contain `validate`. Acceptance
  `TestWA_SpikeNeverReachesShipGate` / `TestWA_ShipGatedArchetypeReachesValidate`
  assert the observable property the ship gate keys on (order contains validate ⇔
  gate fires). The gate stays step-keyed in complete.go — no archetype branch.

- **Merge still validates a promoted spike (executeValidation is step-agnostic).**
  Risk: a spike promoted toward shipping must not escape validation.
  Tested/argued: the merge validation path is step-agnostic — it does not read
  the archetype; re-validation at merge is independent of the spike's missing
  in-workflow validate step. Pinned indirectly by the safety property (the gate
  keys on the step, never on the label) plus the integration order assertions.

- **Hotfix starts at `code`; no plan artifact required.**
  Risk: a hotfix forced through plan/docs is the anti-pattern the feature removes.
  Tested: `tests/integration/workflow_archetypes_integration_test.go` runs the real
  `start --archetype hotfix` and asserts persisted order `[code, tests, validate]`
  with `CurrentStep` first = `code`. `plan` is absent, so `validatePlan` (plan-step
  only) never runs. Acceptance `TestWA_HotfixOrder` asserts plan+docs are absent.

- **Explicit flag overrides roadmap + bootstrap.**
  Risk: a roadmap-pinned archetype (or bootstrap order) silently winning over an
  operator's explicit `--archetype` would run the wrong track.
  Tested: `cmd/centinela/start_archetype_test.go`
  `TestResolveArchetypeOrder_FlagOverridesRoadmap` (flag beats a roadmap pinning a
  different archetype), `_RoadmapWhenNoFlag`, `_FallsThroughToCanonical` cover the
  full precedence table. Acceptance `TestWA_FlagOverridesRoadmapArchetype`.

- **Unknown archetype rejected (flag path AND roadmap-load path).**
  Risk: a typo silently coerced to canonical would run the wrong track; a bad
  roadmap should fail fast.
  Tested: `internal/workflow/archetype_test.go` `TestValidateArchetype` (error
  names value + field); `archetypeOrderByName("bogus")` rejected
  (`start_archetype_test.go`); roadmap path
  `internal/roadmap/archetype_test.go`
  `TestValidateDependencies_RejectsUnknownArchetype` (error names the feature).
  Acceptance `TestWA_UnknownArchetypeRejected`.

- **Archetype is orthogonal to enforcement profile.**
  Risk: coupling sequence (archetype) to strictness (profile) would make some
  combinations impossible or leak one axis into the other.
  Tested: `internal/workflow/archetype_state_test.go`
  `TestArchetype_OrthogonalToProfile` and acceptance
  `TestWA_ArchetypeIndependentOfProfile`: spike order + strict profile →
  StepOrder is the spike order AND EnforcementProfile is strict.

- **ArchetypeStepOrder returns clones, not aliases (mutation safety).**
  Risk: a caller mutating a returned order could corrupt `DefaultStepOrder` and
  poison every later resolution.
  Tested: `internal/workflow/archetype_order_test.go`
  `TestArchetypeStepOrder_ReturnsClonesNotAliases` mutates the returned slice and
  asserts `DefaultStepOrder` and a second call are unaffected.

- **Status shows order[0] / the real archetype, not a hardcoded "plan".**
  Risk: the start.go:84 bug printed "plan" as the current step for archetypes
  starting at `code`; status must reflect the resolved track.
  Tested: integration asserts persisted `CurrentStep`/order[0] = `code` for
  hotfix and `plan` for spike; `internal/ui/render_status_archetype_test.go` and
  acceptance `TestWA_StatusShowsArchetype` assert the archetype line + the
  "spike — no ship gate" annotation render.

## Residual Risks

- **Promoted-spike merge re-validation is asserted as a property, not driven
  end-to-end here.** The merge command's step-agnostic validation is covered by
  the existing merge test suites; this feature adds no archetype awareness to it,
  so no new archetype-specific merge path exists to break. Mitigation: the safety
  test pins that the gate keys on the step, so a promoted spike that gains a
  validate step (canonical/hotfix/refactor on re-classification) is gated normally.
- **`workflowOrderForFeature` greenfield bootstrap branch** (95.2% line cov) is
  exercised by pre-existing `start_guard_*_test.go`; the archetype resolver wraps
  it unchanged, so bootstrap precedence is inherited, not re-implemented.
