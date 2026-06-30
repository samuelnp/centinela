### Big-Thinker Report: brownfield-setup-detection
**Date:** 2026-06-29

#### Problem
When a project with existing source but no `PROJECT.md` starts a Centinela
session, the setup hook always emits the GREENFIELD directive. The `!hasProject`
branch in `cmd/centinela/hook_setup.go` (lines 34-37) unconditionally calls
`ui.RenderSetupNeeded()`, which tells the agent "Do not answer the user's
message" and to ask 6 cold setup questions reading only `PROJECT.md.template`.
It never inspects the codebase, so a repo full of source is interrogated as if
empty (real symptom: an existing Expo/React-Native game whose stack was sitting
in `package.json` + design docs got hand-described by the user). Centinela
already ships the brownfield engine — `centinela analyze` (repo →
`.workflow/analysis.json`) and `centinela synthesize` (drafts `PROJECT.md` and
infers archetype) — but nothing routes to it on first prompt. The fix is
detection + routing, not building brownfield support.

#### Scope
- In:
  - `analyze.HasSource(root string) bool` — cheap, root-only detector in a new
    `internal/analyze/detect.go` (single root readdir; manifests from the same
    set `manifests.go` knows + populated source dirs `src/app/lib/cmd/pkg/internal`).
  - Route `hook_setup.go` `!hasProject` branch on `analyze.HasSource(".")`:
    brownfield → new directive + `ui.RenderBrownfieldSetupNeeded()`; greenfield →
    existing path unchanged. Keep file ≤100 lines (extract helper if needed).
  - `ui.RenderBrownfieldSetupNeeded()` mirroring `RenderSetupNeeded()` style;
    directive = analyze → synthesize → enrich draft from source → set
    `**Project Stage:** existing` → present + confirm → finalize.
- Out:
  - Brownfield onboarding docs in `new-project-guide.md` (deliberate v1
    exclusion; deferred to roadmap).
  - Any change to the `analyze`/`synthesize` engines (reused verbatim).
  - Full tree-walk in the hook (cost — see Risks).

#### Dependencies & Assumptions
- Reuses `internal/analyze` and `internal/synthesize` as-is (both shipped/tested).
- Manifest knowledge lives in `manifestTable` (`internal/analyze/manifests.go`);
  `HasSource` should derive from it or be parity-tested to avoid drift.
- `internal/projectstage.Parse` maps `Project Stage: existing` → `Existing`,
  skipping bootstrap — the directive must instruct the agent to write it.
- In the `!hasProject` case `PROJECT.md` is absent, so `synthesize` writes
  `PROJECT.md` (not `PROJECT.draft.md`).
- Centinela's own CLI directive strings are English product output (consistent
  with the already-hardcoded `render_setup.go`); the i18n rule targets governed
  user projects, not Centinela's CLI.
- `hook_setup.go` is 91 lines today; a new branch must stay ≤100 (G1).

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Detector too expensive (runs every UserPromptSubmit) | High | Medium | Single root `ReadDir`/`Stat`, no recursion, no `Analyze()`, no file reads; cannot depend on `.workflow/analysis.json` (absent on first prompt) |
| Manifest list drifts from `manifestTable` | Medium | Medium | Derive `HasSource` from the table keys, or add a parity unit test |
| False positive: empty `src/` flips greenfield → brownfield | Medium | Medium | Source dir counts only if non-empty; manifests are the strong signal; edge-case test |
| False negative: unlisted manifest only (`pom.xml`, `composer.json`) | Low | Medium | Acceptable v1 — degrades to current greenfield behaviour, no regression; source-dir signal catches most |
| `hook_setup.go` exceeds 100 lines (G1) | Medium | Medium | Extract branch into a helper; keep `runHookSetup` lean |
| Regression to greenfield onboarding | High | Low | Greenfield path byte-for-byte unchanged; only `HasSource==true` is new; test empty-repo still emits `RenderSetupNeeded` |

#### Rollout
- Step 1: `analyze.HasSource(root)` + `internal/analyze/detect.go` with unit
  tests (manifest hit, source-dir hit, empty-dir miss, empty-repo miss, parity).
- Step 2: `ui.RenderBrownfieldSetupNeeded()` mirroring the panel style; unit-test
  it contains analyze/synthesize + `Project Stage: existing` + confirm steps.
- Step 3: Route the `!hasProject` branch on `analyze.HasSource(".")`; extract a
  helper to stay ≤100 lines; integration-test both branches.
- Step 4 (acceptance): drive the built binary against a temp brownfield repo
  (Makefile-only, package.json) → BROWNFIELD directive; empty repo → greenfield.
- Deferred: brownfield onboarding docs in `new-project-guide.md`.

#### Deferred Findings
- Recorded `brownfield-onboarding-docs` via `centinela roadmap defer` — document
  the analyze → synthesize → enrich → confirm path in
  `docs/architecture/new-project-guide.md`. Deliberate v1 exclusion.

#### Handoff
- Next role: feature-specialist.
- Outstanding questions:
  1. `HasSource` manifest set — derive from `manifestTable` (preferred) or
     mirrored list + parity test?
  2. "Populated" source dir minimum — any non-hidden entry (preferred, cheap) or
     a recognized source extension?
  3. Confirm the brownfield directive's opening line invites running
     analyze/synthesize rather than reusing the greenfield "Do not answer" freeze.
