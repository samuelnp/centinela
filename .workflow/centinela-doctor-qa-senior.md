# QA-Senior Report — centinela-doctor

## Test inventory
- Colocated unit (`internal/doctor/`, drive 95% per-package): one file per check + helpers, registry, context, git, fix/branches; all ≤100 lines. Plus `internal/ui/render_doctor_test.go`, `cmd/centinela/doctor_test.go`.
- Acceptance (`tests/acceptance/centinela_doctor_*`): all 38 scenarios, 1:1, split across ≤100-line files.
- Integration (`tests/integration/`): full doctor run + --fix round-trip (drift+glyph+orphaned tmp → fix → all OK → idempotent).
- Edge-case report: `.workflow/centinela-doctor-edge-cases.md`.

## Results (independently re-verified by orchestrator)
- 1970 tests pass (27 packages); gofmt/vet clean; every internal/cmd _test.go ≤100 lines.
- Total coverage 95.2% (precise ~95.06, genuine margin); internal/doctor 97.5%; render_doctor 100%.
- spec-traceability: no doctor scenarios uncovered. validate --full: all gates pass.

## Testability
Used overridable package vars gitRunner/versionRunner + temp repos under t.TempDir() for hermetic, deterministic checks (no shelling out / network).

## Source bugs
None — implementation correct as written. All failures were test-fixture issues, resolved in tests.

Handoff → validation-specialist.
