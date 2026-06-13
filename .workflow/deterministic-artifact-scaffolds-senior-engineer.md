# deterministic-artifact-scaffolds — senior-engineer

## Files Touched

Slice 1 — inputs pre-fill:
- `internal/orchestration/plan_snapshot.go` (66): renamed `requiredPlanInputs` → exported `RequiredPlanInputs` (same body); updated caller at line 14; added doc comment. Package stays a leaf.
- `internal/evidence/plan_inputs.go` (16, NEW): `PlanInputs(feature, role)` delegates to `orchestration.RequiredPlanInputs` for big-thinker + feature-specialist, else nil.
- `cmd/centinela/evidence_init.go` (68): after `Skeleton(...)`, `if pre := evidence.PlanInputs(...); pre != nil { skel.Inputs = pre }`. `Skeleton` itself untouched (repair/docs templates stay clean).

Slice 2 — FILL marker + companion skeletons:
- `internal/evidence/fill.go` (12, NEW): `FillMarker = "<FILL: %s>"` const + `FillSlot(desc)`.
- `internal/evidence/companion_skeletons.go` (38, NEW): `companionHeaders map[Role][]string` (LOCKED per-role headers) + `companionSkeleton(feature, role) (string, bool)` rendering `## <header>\n\n<FILL: lower>\n\n`.
- `internal/evidence/companion.go` (41): `DefaultCompanionTemplate` now role-aware — skeleton when known, one-line fallback otherwise.

Slice 3 — artifact new body upgrade:
- `internal/evidence/artifact_derive.go` (23, NEW): `analyzedSpecsList()` globs `specs/*.feature` (sorted; FILL-slot row when none).
- `internal/evidence/artifact_gatekeeper.go` (23): Analyzed Specs pre-filled via `analyzedSpecsList()`; Findings/Recommendation → `FillSlot`. `**Status:** SAFE` / `**Date:**` lines kept verbatim.
- `internal/evidence/artifact_edge_cases.go` (18), `artifact_prodready.go` (26), `artifact_changelog.go` (8): italic prose → `FillSlot`. prodready Status/Date unchanged.
- `internal/evidence/artifact_templates.go` (37): corrected the "Pure — does no I/O" comment (gatekeeper now globs).

## Architecture Compliance

G1: every touched/new .go file ≤100 lines (max 68). G2: `internal/orchestration` stays a leaf (rename only, no new import); `internal/evidence → internal/orchestration` edge already existed. G7: `cmd/` change is a 3-line wiring delta. `go vet` and `gofmt -l` clean.

## Type-Safety Notes

No `any`/`interface{}` introduced. `PlanInputs` returns `[]string` (nil for non-plan roles, distinguishing "no pre-fill" from "empty list" so `Skeleton`'s `[]string{}` default survives). No `<FILL:` ever enters a JSON list field — confirmed by dogfood grep.

## Trade-Offs

Pre-fill lives in the `evidence init` command path, NOT in `Skeleton`, so `SchemaSkeleton` (repair) and `docsSpecialistPair` are not poisoned with glob results. `outputs`/`edgeCases` deliberately left empty (would fail the real-file validator at init time). gatekeeper body trades purity for a deterministic mechanical pre-fill; comment updated honestly.

## Test Results

`go build ./...` clean. `go vet ./...` clean. `gofmt -l internal cmd` empty. `go test ./...`: 1465 passed in 24 packages (no breakage — existing companion tests use `Contains`, so the body upgrade did not break them). Dogfood (`/tmp/cent-dass`): big-thinker inputs pre-filled with both feature docs; senior-engineer inputs empty; gatekeeper Analyzed Specs globbed+sorted (and FILL-slot row when no specs); no `<FILL:` in any JSON.

## Handoff

qa-senior. Tests asserting these behaviors should be authored in the tests step. No `tests/`-tier fixes were required (none existed for the old behavior).
