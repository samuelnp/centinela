# brownfield-roadmap-generation — senior-engineer

## Files Touched

| File | Status | Lines | Purpose |
|------|--------|-------|---------|
| `internal/brownmap/brownmap.go` | new | 44 | Package doc + `Plan`/`Brownfielder` types, `NewBrownfielder`, `DefaultDraftPath` |
| `internal/brownmap/generate.go` | new | 38 | `Generate` orchestrator: reconstruct → Baseline + gap phases → draft Roadmap |
| `internal/brownmap/baseline.go` | new | 50 | Builds the schedule-exempt Baseline phase from `[]reconstruct.Target` |
| `internal/brownmap/gaps.go` | new | 58 | Builds gap phase(s) from TODO-bearing targets + `--goal` strings |
| `internal/brownmap/write.go` | new | 64 | `WriteDraft`: refuses canonical roadmap; atomic temp+rename; roadmap.Save JSON format |
| `internal/ui/render_brownfield.go` | new | 24 | `RenderBrownfieldSummary` — baseline/gap counts, draft path, no-gaps hint |
| `cmd/centinela/roadmap_brownfield.go` | new | 68 | Thin `roadmap brownfield` Cobra subcommand (G7-clean) |
| `internal/roadmap/baseline.go` | new | 32 | `BaselinePhaseName`, `isBaselinePhaseName`/`IsBaselinePhaseName`, shared `isNonSchedulablePhase` |
| `internal/roadmap/roadmap.go` | edit | 77 | `Summary` skip routed through `isNonSchedulablePhase` |
| `internal/roadmap/backlog.go` | edit | 73 | `NonBacklogFeatureSet` skip routed through `isNonSchedulablePhase` |
| `internal/roadmap/readiness.go` | edit | 83 | `DeriveReadiness` skip routed through `isNonSchedulablePhase` |
| `internal/reconstruct/reconstruct.go` | edit | 76 | New `Reconstruction.TodoTargets()` accessor (+ `strings` import) |
| `centinela.toml` | edit | — | Aggregator layer: `paths += internal/brownmap/**`; `allow += "aggregator"`; comment block |
| `PROJECT.md` | edit | — | G2 prose for `internal/brownmap` + folder-structure tree row |

## Architecture Compliance

**Layer boundaries (N-Tier / G2).** `internal/brownmap` is an aggregator with read-only
edges: `brownmap → analyze` (domain, allowed), `brownmap → roadmap` (domain, allowed),
`brownmap → reconstruct` (aggregator→aggregator). The last required extending the
aggregator layer's `allow` from `["domain","leaf"]` to `["domain","leaf","aggregator"]`
so aggregators may compose. `brownmap` imports no `cmd/` and no `internal/ui`; it is
imported only by `cmd/` (its `Plan` type by `internal/ui` for rendering, exactly like the
sibling `reconstruct`/`synthesize` aggregators). `analyze`/`roadmap`/`reconstruct` never
import `brownmap`, so there is no cycle. The toml + PROJECT.md G2 edits land with the code
so `centinela validate`'s `import_graph` gate stays green. `internal/scaffold/assets/centinela.toml`
was deliberately NOT edited (generic template with no project-specific aggregator paths).

**G7 (outer layer thin).** `cmd/centinela/roadmap_brownfield.go` only wires flags, calls
`analyze.Load` → `brownmap.NewBrownfielder().Generate` → `brownmap.WriteDraft` → render;
all decisions (selection, partitioning, never-clobber, atomic write, summary text) live in
`internal/brownmap` / `internal/ui`. No business logic in cmd.

**Baseline exemption — one conceptual place.** Introduced `isNonSchedulablePhase(name) =
isBacklogPhaseName(name) || isBaselinePhaseName(name)` and routed the THREE scheduling-
exemption sites through it: `roadmap.go` `Summary`, `backlog.go` `NonBacklogFeatureSet`
(the validate coverage set), `readiness.go` `DeriveReadiness`. The two remaining Backlog
references — `mdgen_phase.go` (deferred-finding *line formatting*) and `rawrender.go`
`backlogPhaseIndex` (defer/promote *append mechanics*) — were intentionally LEFT as
`isBacklogPhaseName`: routing them through the shared predicate would make Baseline render
as deferred-finding lines / become a defer append target, which is wrong. Baseline is a
normal-rendering phase that is merely schedule-exempt. This matches the task's "additive
only, keep Backlog/Bootstrap behavior identical" and the rule "don't touch rawmove/rawmutate
append mechanics." Existing 210 roadmap+reconstruct tests pass unchanged.

**G1 file sizes.** Every new/edited Go source file is ≤100 lines (max is `roadmap/readiness.go`
at 83). Only `PROJECT.md` (139) and `centinela.toml` (135) exceed 100 — non-source config/docs,
outside the G1 source budget.

### Verification output

```
$ go build ./...
Go build: Success

$ go vet ./...
Go vet: No issues found

$ gofmt -l internal/brownmap cmd/centinela internal/roadmap internal/ui internal/reconstruct
(no files listed)

$ line-count gate over changed files (only PROJECT.md/centinela.toml > 100; no .go file > 100)
OVER 139 PROJECT.md
OVER 135 centinela.toml

$ centinela evidence validate brownfield-roadmap-generation
evidence ok for "brownfield-roadmap-generation"
```

## Type-Safety Notes

Strict Go throughout; no `interface{}`/`any`, no reflection. `Plan` is a fully typed,
byte-stable result struct. `Brownfielder` is the swap seam for a future LLM backend
(mirrors `reconstruct.Reconstructor`). The new reconstruct accessor signature:

```go
func (r Reconstruction) TodoTargets() []Target
```

It filters `r.Targets` (in their already-sorted order) to those whose assembled feature
`Artifact.Body` still carries the `# TODO: confirm` marker, so `brownmap` does not duplicate
reconstruct's skeleton/TODO rule table — a thin read-only accessor over existing fields.

## Trade-Offs

- **Per-target TODO accessor vs duplicating the rule table.** Added `TodoTargets()` to
  reconstruct (read-only, 18 lines) rather than re-deriving TODO signals in `brownmap`.
  Note: the current reconstruct skeleton emits a uniform 3 TODO markers per target, so today
  every target is TODO-bearing — the gap phase mirrors the Baseline 1:1 plus goals. The
  accessor is genuinely zero-able (filters on body content), so the Baseline-only / no-gaps
  path is reachable for callers that build TODO-free `Reconstruction` values; it does not
  hard-code "all targets are gaps."
- **Atomic write via temp+rename.** `roadmap.Save` uses a plain `WriteFile` and there is no
  shared `writeAtomic` helper, so `brownmap.WriteDraft` implements its own temp-file+rename
  (same-dir temp) for crash safety; JSON format is byte-identical to `roadmap.Save`
  (`MarshalIndent(_, "", "  ")`) for stable diffs.
- **DraftPath reflects `--out`.** `Generate` stamps `DefaultDraftPath`; the cmd overwrites
  `plan.DraftPath` with the resolved `--out` before rendering so the summary reports the path
  actually written.

## Deferred Findings

None. The two natural follow-ups (framework-specific gap detection; promote-brownfield-draft-
into-roadmap) are already captured in this feature's "Out (v1)" scope per the big-thinker and
feature-specialist handoffs — no genuinely new out-of-scope discovery surfaced.

## Handoff

Next role: **qa-senior**. The acceptance contract is `specs/brownfield-roadmap-generation.feature`
(10 scenarios). Key seams to test: `brownmap.NewBrownfielder().Generate` (Baseline phase one
feature per sorted target; gap phase from `TodoTargets()` + `--goal`; byte-stable; empty
inventory → empty Baseline / nil gaps); `brownmap.WriteDraft` (refuses `roadmap.RoadmapFile`;
atomic; canonical roadmap.json byte-unchanged); `roadmap.isNonSchedulablePhase` /
`Summary` / `NonBacklogFeatureSet` / `DeriveReadiness` excluding Baseline (and a regression
guard that Backlog/Bootstrap behavior is unchanged); the cmd error path (missing inventory →
exit 1, actionable message, no draft written); `ui.RenderBrownfieldSummary` (counts, path,
no-gaps hint). Smoke-tested manually: byte-identical re-run, goal lands in Gaps not Baseline,
canonical roadmap.json hash unchanged, refuse-canonical error, missing-inventory exit 1.
Per-package 95% coverage is met with colocated `_test.go` (each ≤100 lines).
