# QA-Senior Report — capability-calibration

## Test inventory
- Colocated unit (internal/calibration 100%): friction/classify/strictness/calibrate _test.go; internal/ui/render_calibration_test.go; internal/telemetry/model_field_test.go (Model round-trip/omitempty/legacy/all Record* stamp); cmd/centinela/{telemetry_model,calibrate}_test.go. All ≤100 lines.
- Acceptance (tests/acceptance/capability_calibration_*): all 27 scenarios, 1:1, split ≤100-line files.
- Integration (tests/integration/calibration_roundtrip_test.go): stamp→read→Calibrate → assert classifications + unattributed-last.
- Edge-case report .workflow/capability-calibration-edge-cases.md.

## Results (orchestrator re-verified)
- 2104 tests pass (29 packages); gofmt/vet clean; every internal/cmd _test.go ≤100 lines.
- internal/calibration 100%; total 95.2% (genuine margin).
- spec-traceability: no calibration scenario uncovered. validate --full: all gates pass.
- Verified classify rule, boundary inclusivity, clamps, guards, Model field round-trip.

## Source bugs
None — implementation correct; only tests added.

Handoff → validation-specialist.
