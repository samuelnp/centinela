### Senior-Engineer Report: roadmap-checkpoint-prompt
**Date:** 2026-05-23

#### Files Touched
| Path | Reason |
|------|--------|
| `internal/roadmapcheckpoint/osfs.go` | New os-backed `FS` adapter (`NewOSFS`) so the pure `Decide` function can run against real disk. Implements `Stat`/`ReadFile`/`Exists` over the `os` package. |
| `internal/ui/render_roadmap_checkpoint.go` | New `RenderRoadmapCheckpoint(featureName)` panel — pure formatting of the two-option checkpoint prompt with recovery hints. |
| `cmd/centinela/hook_setup.go` | Wired the checkpoint branch after the production-readiness check; reuses the already-loaded `*roadmap.Roadmap`; delegates all decision logic to a thin `emitRoadmapCheckpoint` helper. |
| `cmd/centinela/roadmap_iterate.go` | New `centinela roadmap iterate` subcommand whose `RunE` calls `WriteMarker`, giving Claude a marker-write path to persist the "iterate" choice. |

#### Architecture Compliance
- Boundary checks passed:
  - `internal/roadmapcheckpoint/osfs.go` imports only `os` + `time` (stdlib). No cmd import; the adapter satisfies the package-local `FS` interface.
  - `internal/ui/render_roadmap_checkpoint.go` imports only `lipgloss`. It receives the resolved `featureName string` and does NOT import `roadmapcheckpoint` — no business logic in the UI layer (G7).
  - `cmd/centinela` is the only place that wires `roadmapcheckpoint` + `ui` together (outer layer / thin orchestration). `go build ./...` and `go vet ./...` pass.
- G1 file size: every touched file ≤ 100 lines — osfs.go 43, render_roadmap_checkpoint.go 32, hook_setup.go 91, roadmap_iterate.go 29.
- G7 outer-layer rule: `emitRoadmapCheckpoint` contains zero decision logic — it calls `FirstIncompleteBootstrap` + `Decide`, switches on the returned `Decision`, and prints. All emit/suppress/stale rules stay in `internal/roadmapcheckpoint`.

#### Type-Safety Notes
- No `interface{}`/`any`. The `FS` contract is the existing narrow typed interface; `NewOSFS` returns `FS` while the concrete `osFS` struct stays unexported.
- `Decide` returns the typed `Decision` enum; the hook switches on `DecisionSuppressed` explicitly and emits on everything else (Emit/Stale), so a future Decision constant fails closed to "emit" rather than silently mis-classifying — but the only non-suppress values today are Emit/Stale, both of which must emit, matching the spec.
- `emitRoadmapCheckpoint` takes a concrete `*roadmap.Roadmap`; `FirstIncompleteBootstrap` already nil-guards.

#### Trade-Offs
- Reused the `r, err := roadmap.Load()` result that the JSON-validity branch already computes instead of loading the roadmap twice. Smallest correct change and avoids a second disk read.
- Switched the suppress check to `if d == DecisionSuppressed { return }` (emit on the complement) rather than an explicit `case DecisionEmit, DecisionStale`. Keeps the helper short and means the single non-emitting state is the one named. Documented above as a fail-open choice.
- Recovery hint hard-codes the literal commands `centinela roadmap iterate` and `centinela start <feature>` in the panel so Claude has zero-effort wiring (per big-thinker recommendation). i18n is None per PROJECT.md, so English literals are allowed.
- Did NOT add a "start" marker — workflow-file presence (`.workflow/<feature>.json`) remains the canonical "start" signal, already handled inside `Decide`.

#### Handoff
- Next role: qa-senior
- Outstanding TODOs:
  - No feature tests written (owned by the tests step). The `roadmapcheckpoint` package currently has `[no test files]`.
  - qa-senior should cover both `DecisionEmit` and `DecisionStale` through the integration hook since the UX is identical but the directive must fire for each; and assert idempotency (second `runHookSetup` with a fresh marker stays silent).
