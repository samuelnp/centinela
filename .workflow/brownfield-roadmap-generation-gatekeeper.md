### Gatekeeper Report: brownfield-roadmap-generation

**Date:** 2026-06-24
**Status:** SAFE

#### Analyzed Specs

- `specs/brownfield-roadmap-generation.feature` (this feature — the new draft command + Baseline/gap partitioning)
- `specs/deferred-findings-roadmap-capture.feature` (Backlog phase append + validate-exempt coverage)
- `specs/roadmap-doc-sync.feature` (ROADMAP.md generation, Backlog phase rendering)
- `specs/roadmap-parallel-readiness.feature` / `roadmap-checkpoint-prompt.feature` (DeriveReadiness / feature-start gating)
- `specs/roadmap-quality-overall-threshold.feature` / `roadmap-senior-pm-analysis.feature` (ValidateAnalysis/ValidateQuality coverage set)
- `specs/clarify-roadmap-missing-artifacts.feature`, `specs/fix-roadmap-write-blocked.feature`, `specs/spec-reconstruction.feature` (reconstruct/roadmap adjacency)
- Scanned for `Baseline` collisions: `specs/audit-baseline-ratchet.feature`, `specs/custom-gate-sdk.feature`, `specs/precommit-and-pr-gate.feature`, `specs/raise-test-coverage-90.feature`

#### Findings

- **Shared-domain edit — no-op for non-Baseline roadmaps (PRIMARY RISK):** Resolved / No conflict.
  Affected specs: `deferred-findings-roadmap-capture`, `roadmap-doc-sync`, `roadmap-parallel-readiness`, `roadmap-quality-overall-threshold`.
  Risk: rerouting `Summary` (roadmap.go:62), `NonBacklogFeatureSet` (backlog.go:65), and `DeriveReadiness` (readiness.go:20) through `isNonSchedulablePhase` could alter behavior for the real `.workflow/roadmap.json`.
  Verification: `isNonSchedulablePhase(name) = isBacklogPhaseName(name) || isBaselinePhaseName(name)` is purely additive. The canonical `.workflow/roadmap.json` contains zero phases named "Baseline" (phases are "Phase 0..10" + "Backlog"); `isBaselinePhaseName` returns false for every one, so the new disjunct never fires → identical output to the prior `isBacklogPhaseName`-only path. Backlog semantics in deferred-findings/roadmap-doc-sync are untouched (`isBacklogPhaseName` still independently gates `BacklogFeatures`/`IsBacklogFeature`/mdgen rendering).
  Suggestion: none.

- **Coverage-set integrity for ValidateAnalysis/ValidateQuality:** No conflict.
  `NonBacklogFeatureSet` (backlog.go) is confirmed as the single coverage set behind both validators via `analysis.go:61` (`RoadmapFeatureSet → NonBacklogFeatureSet`), shared by `ValidateAnalysis` and `ValidateQuality`. Excluding a Baseline phase can only drop features that live in a phase literally named "Baseline"; no such phase exists in the canonical roadmap, so no real, schedulable feature is dropped from required coverage. The exclusion is the same predicate mechanism that already exempts Backlog (matches the feature spec, scenario "Baseline features are excluded ...").

- **Naming-collision risk (Baseline as schedulable work):** Low — recorded, not blocking.
  A future curated roadmap that legitimately named a *schedulable* phase "Baseline" would have its features silently exempted from status counts, validate coverage, and readiness. The name is reserved by convention (`roadmap.BaselinePhaseName`), case-insensitive/trimmed match. No existing spec or the canonical roadmap uses "Baseline" as a phase name; the `Baseline` tokens in `audit-baseline-ratchet`/`custom-gate-sdk`/`precommit-and-pr-gate`/`raise-test-coverage-90` refer to the unrelated `centinela audit baseline` mechanical-gate violation snapshot, not a roadmap phase. No action needed now.

- **Draft never clobbers canonical roadmap.json:** No conflict.
  `internal/brownmap/write.go` `WriteDraft` hard-refuses when `filepath.Clean(path) == filepath.Clean(roadmap.RoadmapFile)` and writes atomically (temp + rename). Matches feature scenario "never clobbers an existing canonical roadmap.json" (byte-for-byte unchanged). `centinela roadmap generate` (mdgen.go) renders phases as-is with no Baseline special-casing, and since the draft is never written to the canonical file, ROADMAP.md generation/drift is unaffected.

- **Layering / import cycle (brownmap aggregator):** No conflict.
  `internal/brownmap` imports only `analyze` (domain), `roadmap` (domain), `reconstruct` (aggregator); no `cmd/` or `internal/ui`. `reconstruct`/`analyze`/`roadmap` contain no reverse import of `brownmap` (only a comment mention), so no cycle. `centinela.toml` aggregator `allow` correctly extended to `["domain","leaf","aggregator"]` with `internal/brownmap/**` in paths, matching PROJECT.md G2.

#### Deferred Findings

- none (no findings warranted deferral; the naming-collision risk is documentary and below the defer threshold).

#### Recommendation

- **SAFE** — the shared `internal/roadmap` edits are a verified no-op for the real (non-Baseline) canonical roadmap, the coverage set drops no schedulable feature, the draft writer cannot clobber the canonical file, and no existing spec conflicts. `go build ./...` succeeds and `go test ./internal/roadmap/... ./internal/reconstruct/... ./internal/brownmap/...` (233 pass) plus `./internal/gates/...` (160 pass) are green.
