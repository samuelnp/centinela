# Plan: brownfield-setup-detection

## Problem

The setup hook's `!hasProject` branch (`cmd/centinela/hook_setup.go:34-37`)
unconditionally emits the greenfield directive + `ui.RenderSetupNeeded()`, which
tells the agent to ignore the user and ask 6 cold setup questions reading only
`PROJECT.md.template`. It never inspects the existing codebase, so a repo full of
source is onboarded as if empty. Centinela already has the brownfield engine
(`centinela analyze` → `.workflow/analysis.json`, `centinela synthesize` →
drafts `PROJECT.md`, inferring the archetype) but nothing routes to it on first
prompt. This feature adds the missing detection + routing.

## Scope

### In
- `analyze.HasSource(root string) bool` — a cheap, root-only detector in a new
  file `internal/analyze/detect.go`. Single root readdir checking for known
  manifests (`go.mod`, `package.json`, `Cargo.toml`, `Gemfile`, `pyproject.toml`,
  `requirements.txt`, `Makefile` — the same set `internal/analyze/manifests.go`
  already knows) and/or populated source dirs (`src`, `app`, `lib`, `cmd`,
  `pkg`, `internal`).
- Route the `!hasProject` branch on `analyze.HasSource(".")`: brownfield → new
  directive line + `ui.RenderBrownfieldSetupNeeded()`; greenfield → existing
  path unchanged.
- `ui.RenderBrownfieldSetupNeeded()` in `internal/ui/render_setup.go`, mirroring
  `RenderSetupNeeded()` style (`lipgloss.JoinVertical` + `renderSystemPanel`,
  `StyleYellow`/`StyleMuted`/`StyleRed`, `toneWarn`). Directive: do NOT
  cold-interrogate; (a) run `centinela analyze` then `centinela synthesize`,
  (b) ENRICH the draft by reading key source (design docs, manifests, i18n) to
  correct inferred guesses and fill gaps, (c) set `**Project Stage:** existing`,
  (d) present the drafted `PROJECT.md` and confirm uncertain fields, then
  finalize.

### Out
- Brownfield onboarding documentation in `new-project-guide.md` — deliberate
  follow-up, OUT of v1. Recorded as a deferred roadmap item.
- Any change to the `analyze`/`synthesize` engines themselves. They are reused
  as-is; this feature only detects and routes.
- Replacing or fully tree-walking in the hook. Detection stays a single root
  readdir for cost reasons (see Risks).

## Dependencies & Assumptions

- Reuses `centinela analyze` (`internal/analyze`) and `centinela synthesize`
  (`internal/synthesize`) verbatim — both already shipped and tested.
- `internal/projectstage` parses `Project Stage:` from `PROJECT.md`; the
  directive relies on the agent writing `**Project Stage:** existing` so
  `projectstage.Parse` → `Existing` and bootstrap is skipped.
- The manifest set lives in `internal/analyze/manifests.go` (`manifestTable`).
  `HasSource` should share that knowledge to avoid drift — either reference the
  table keys or keep a small mirrored list with a test asserting parity.
- Centinela's own CLI directive strings are English product output (consistent
  with the existing hardcoded `render_setup.go`); the i18n "no hardcoded
  strings" rule targets governed *user* projects, not Centinela's CLI. Follow
  the existing pattern.
- `hook_setup.go` is 91 lines today; adding a branch must keep it ≤100 (G1).
  Extract a small helper (e.g. `setupDirective()` returning directive+panel) if
  the branch pushes it over.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Detector too expensive — runs on every `UserPromptSubmit` | High | Medium | Single root `os.ReadDir`/`os.Stat` only; no recursion, no `analyze.Analyze()` tree-walk, no reading file contents. Cannot depend on `.workflow/analysis.json` (absent on first prompt). |
| Manifest list drifts from `manifestTable` | Medium | Medium | Source `HasSource` from `manifestTable` keys, or add a parity unit test asserting the two sets match. |
| False positive: empty `src/` dir flips greenfield repo to brownfield | Medium | Medium | Treat a source dir as a signal only if it is non-empty; manifests are the strong signal. Cover with an edge-case test. |
| False negative: source repo with an unlisted manifest only (e.g. `pom.xml`, `composer.json`) read as greenfield | Low | Medium | Acceptable for v1 — still degrades to the current (greenfield) behaviour, no regression. Source-dir signal catches most such repos. Note as deferred breadth. |
| `hook_setup.go` exceeds 100 lines (G1) | Medium | Medium | Extract the branch into a helper; keep `runHookSetup` lean. Gatekeeper checks file size. |
| Regression to greenfield onboarding | High | Low | Greenfield path is byte-for-byte unchanged; only the `HasSource==true` case is new. Add a test asserting the empty-repo path still emits `RenderSetupNeeded`. |
| `synthesize` writes `PROJECT.draft.md` instead of `PROJECT.md` if one already exists | Low | Low | In the brownfield-onboarding case `PROJECT.md` is absent by definition (we are inside `!hasProject`), so `synthesize` writes `PROJECT.md`. Directive reminds the agent to confirm. |

## Rollout

- Step 1 (smallest correct slice): Add `analyze.HasSource(root)` +
  `internal/analyze/detect.go` with unit tests (manifest hit, source-dir hit,
  empty-dir miss, empty-repo miss, manifest-table parity).
- Step 2: Add `ui.RenderBrownfieldSetupNeeded()` to `render_setup.go` mirroring
  the existing panel style; unit-test that it contains the analyze/synthesize +
  `Project Stage: existing` + confirm instructions.
- Step 3: Route the `!hasProject` branch in `hook_setup.go` on
  `analyze.HasSource(".")`; extract a helper if needed to stay ≤100 lines.
  Integration-test both branches (brownfield repo → brownfield directive;
  empty repo → greenfield directive).
- Step 4 (acceptance): drive the built binary against a temp brownfield repo
  (only a `Makefile`, or a `package.json`) and assert the BROWNFIELD directive;
  drive against an empty repo and assert the greenfield directive.
- Deferred: brownfield onboarding docs in `new-project-guide.md`.

## Deferred Findings

- `brownfield-onboarding-docs` — document the brownfield onboarding path
  (analyze → synthesize → enrich → confirm) in
  `docs/architecture/new-project-guide.md`. Deliberate v1 exclusion, recorded
  via `centinela roadmap defer`.

## Handoff

- Next role: feature-specialist.
- Outstanding questions:
  1. Should `HasSource` derive its manifest set directly from `manifestTable`
     (single source of truth) or keep a mirrored list guarded by a parity test?
     (Recommendation: derive from the table.)
  2. Minimum signal for a "populated" source dir — any non-dotfile entry, or at
     least one file with a recognized source extension? (Recommendation: any
     non-empty, non-hidden entry; keep it cheap.)
  3. Confirm the brownfield directive's first line wording so it does not read
     as the greenfield "Do not answer the user" block — it should invite the
     agent to run analyze/synthesize, not freeze.
