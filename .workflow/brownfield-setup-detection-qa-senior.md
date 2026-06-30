### QA-Senior Report: brownfield-setup-detection
**Date:** 2026-06-29

#### Test Inventory
| Tier        | File | Scenarios |
|-------------|------|-----------|
| unit        | internal/analyze/detect_test.go | `HasSource` table: empty→false, go.mod→true, Makefile-only→true, package.json→true, populated src/→true, empty src/→false, populated internal/→true; `dirHasEntry`: missing path, plain file, empty dir, populated dir |
| unit        | internal/ui/render_setup_test.go | `RenderBrownfieldSetupNeeded` carries BROWNFIELD, Do NOT interrogate, analyze, synthesize, ENRICH, `**Project Stage:** existing`, confirm |
| unit        | cmd/centinela/hook_setup_brownfield_test.go | `runHookSetup` brownfield route (centinela.toml+go.mod, no PROJECT.md) → brownfield directive+panel; greenfield route (empty repo) → setup-required directive, no BROWNFIELD |
| integration | tests/integration/brownfield_setup_detection_integration_test.go | real `analyze.HasSource` + `ui.RenderBrownfieldSetupNeeded` wired: manifest→brownfield+enrich panel; empty src/→greenfield |
| acceptance  | tests/acceptance/brownfield_setup_detection_test.go | built binary `centinela hook setup`: go.mod / package.json / populated src/ / populated internal/ → BROWNFIELD directive (analyze, synthesize, `**Project Stage:** existing`) |
| acceptance  | tests/acceptance/brownfield_setup_detection_more_test.go | Makefile-only→brownfield; Cargo.toml→enrich-then-confirm (no "ignore the user"); empty repo / empty src/ / deep-nested-only→greenfield, no BROWNFIELD; PROJECT.md present→bypass both directives |

All 10 `.feature` scenarios are mapped via `// Scenario:` comments across the two acceptance files (scenarios 1,2,4,9 in the main file; 3,5,6,7,8,10 in the `_more_` file). The spec_traceability gate maps every scenario name.

#### Coverage Gaps
- None at the scenario level — all 10 spec scenarios have an executable acceptance assertion, plus colocated unit coverage for `internal/analyze`, `internal/ui`, and `cmd/centinela` so the 95% TOTAL coverage gate holds.
- `dirHasEntry`'s `os.Open` error branch (a dir that stats as a directory but cannot be opened — a permissions race) is not exercised; it degrades to the safe "not a signal" default. New-code line coverage: `HasSource` 100%, `emitSetupDirective` 100%, `RenderBrownfieldSetupNeeded` 100%, `dirHasEntry` 88.9%. TOTAL 95.1% ≥ 95.0%.

#### Acceptance Wiring
centinela.toml validate.commands already runs the acceptance tier:
```toml
commands = [
  "go test ./...",
  "go test ./tests/acceptance/...",
]
```

#### Deferred Findings
- none — no new gaps deferred from this step. (`brownfield-manifest-breadth` and `brownfield-onboarding-docs` were already recorded by predecessors.)

#### Handoff
- Next role: validation-specialist
- Edge-case report: `.workflow/brownfield-setup-detection-edge-cases.md` (empty-src non-signal, Makefile signal, depth guard / no subdir walk, PROJECT.md bypass, greenfield unchanged, hook early-return guard, dirHasEntry robustness).
