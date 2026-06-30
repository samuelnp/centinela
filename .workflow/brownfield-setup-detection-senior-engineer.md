### Senior-Engineer Report: brownfield-setup-detection
**Date:** 2026-06-29

#### Files Touched
| Path | Reason |
|------|--------|
| internal/analyze/detect.go | NEW. `HasSource(root string) bool` — cheap, root-only brownfield detector. Reuses `manifestTable` keys (no duplicated filename list) + checks non-empty conventional source dirs. No recursion, no file reads. |
| internal/ui/render_setup.go | Added `RenderBrownfieldSetupNeeded() string` mirroring `RenderSetupNeeded()` panel style (StyleYellow header, StyleMuted steps, StyleRed close, `renderSystemPanel("SETUP", …, toneWarn, body)`). Instructs analyze → synthesize → ENRICH → set `Project Stage: existing` → present+confirm. |
| cmd/centinela/hook_setup.go | Routed the `!hasProject` branch through `emitSetupDirective()` instead of inline greenfield output. Kept ≤100 lines. |
| cmd/centinela/hook_setup_setup.go | NEW. Holds `emitSetupDirective()`: brownfield (`analyze.HasSource(".")`) emits the BROWNFIELD directive + `RenderBrownfieldSetupNeeded()`; greenfield path unchanged. Extracted to keep `hook_setup.go` under the G1 cap. |

#### Architecture Compliance
- Boundary checks passed: n-tier archetype (PROJECT.md → Architecture Choice → n-tier). `cmd/centinela` (outer/CLI layer) imports `internal/analyze` and `internal/ui` — allowed (cmd may import internal/*). `internal/analyze/detect.go` imports only stdlib (`os`, `path/filepath`); imports nothing internal. No reverse (internal → cmd) imports introduced.
- G1 file size: detect.go 60, render_setup.go 90, hook_setup.go 90, hook_setup_setup.go 22 — all ≤100. No G1 exception needed.
- G7 outer-layer rule: no business logic added to the outer layer. The CLI command only routes on a boolean from `internal/analyze`; the detection logic lives in `internal/analyze`, the rendering in `internal/ui`.

#### Type-Safety Notes
- No `interface{}`/`any`, no reflection, no untyped ducks. `HasSource` and `dirHasEntry` take/return concrete `string`/`bool`.
- Errors handled explicitly: `os.Stat`/`os.Open` errors short-circuit to "not a signal" rather than being ignored or panicking. `Readdirnames(1)` error is intentionally non-fatal (the len check is the decision).
- Reuses the existing `manifestTable map[string]manifestEntry` by iterating keys — single source of truth, no parallel string slice to drift.

#### Trade-Offs
- Detector is intentionally shallow (root-only, `Readdirnames(1)`): O(entries at root), safe to run on every UserPromptSubmit. Trade-off: a repo whose only source sits in a deeply nested non-root dir, with no root manifest, reads as greenfield (spec scenario 10 accepts this).
- Source-dir signal requires a non-empty directory (reads a single entry). An empty `src/` is deliberately NOT a signal (spec scenario 6), preventing freshly-scaffolded repos from being misclassified.
- Manifest breadth limited to the existing `manifestTable` set (go.mod, package.json, Cargo.toml, Gemfile, pyproject.toml, requirements.txt, Makefile). Unlisted manifests (pom.xml, composer.json, build.gradle) degrade to greenfield — no regression. Recorded as deferred breadth by predecessors.
- Extracted `emitSetupDirective()` into a sibling file rather than inlining: keeps `hook_setup.go` under the 100-line G1 cap without a G1 exception.

#### Deferred Findings
- `brownfield-onboarding-docs` and `brownfield-manifest-breadth` were already recorded by predecessors (big-thinker / feature-specialist) via `centinela roadmap defer`. No new deferrals from the code step.

#### Handoff
- Next role: qa-senior
- Outstanding TODOs: QA to add unit tests (manifest hit, source-dir hit, empty-dir miss, empty-repo miss, depth guard, manifestTable parity), the `RenderBrownfieldSetupNeeded` content test (analyze/synthesize/ENRICH/`Project Stage:** existing`/confirm literals), and acceptance tests driving the built binary against temp brownfield/greenfield/empty-src repos. Smoke-tested manually: go.mod repo → BROWNFIELD directive; empty-src-only repo → GREENFIELD directive (both exit 0).
