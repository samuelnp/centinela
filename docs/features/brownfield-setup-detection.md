# brownfield-setup-detection

## Problem

When a project with existing source but no `PROJECT.md` starts a Centinela
session, the setup hook always emits the GREENFIELD directive.
`internal/ui/render_setup.go` `RenderSetupNeeded()` literally tells the agent
"Do not answer the user's message" and to ask 6 cold setup questions, reading
only `PROJECT.md.template`. It never inspects the existing codebase, so a repo
full of source is interrogated as if it were empty. Real symptom: a user
onboarding an existing Expo/React-Native game was asked to hand-describe a
stack that was sitting right there in `package.json` and the design docs.

Centinela already ships a brownfield engine that nobody routes to:
`centinela analyze` (scans repo → `.workflow/analysis.json`) and
`centinela synthesize` (drafts `PROJECT.md` / `PROJECT.draft.md` from that
inventory, inferring the archetype). The fix is detection + routing, not
building brownfield support.

## User Stories

- As a user onboarding an existing codebase, I am not cold-interrogated about a
  stack the tooling can already read from my manifests and design docs.
- As a user with an empty/greenfield repo, I still get the question-based setup
  flow unchanged.
- As an agent, when the repo has source I am told to `analyze` + `synthesize`,
  then enrich the draft, then confirm with the user — instead of asking the
  6 cold questions.

## Acceptance Criteria

- A cheap, root-only detector `analyze.HasSource(root string) bool` reports
  whether the repo already contains source (known manifests and/or populated
  source dirs), without a full tree-walk.
- When `PROJECT.md` is missing AND `HasSource(".")` is true, the setup hook
  emits a BROWNFIELD directive and `ui.RenderBrownfieldSetupNeeded()` instead of
  the greenfield setup.
- When `PROJECT.md` is missing AND `HasSource(".")` is false, the existing
  greenfield directive + `ui.RenderSetupNeeded()` is emitted unchanged.
- The brownfield directive instructs the agent to: run `centinela analyze` then
  `centinela synthesize`; enrich the draft by reading key source (design docs,
  manifests, i18n); set `**Project Stage:** existing`; present the drafted
  `PROJECT.md` and confirm uncertain fields before finalizing.
- `cmd/centinela/hook_setup.go` stays within the G1 budget (≤100 lines, ≤130
  only with a justified exception).

## Edge Cases

- Empty/greenfield repo (no manifests, no populated source dirs) still gets the
  question-based greenfield setup.
- A brownfield repo whose only signal is a `Makefile` is detected as source.
- A repo with an empty `src/` (no files) is NOT treated as source on the
  directory signal alone.
- `synthesize` runs before `PROJECT.md` exists, so it writes `PROJECT.md`
  (not `PROJECT.draft.md`); that file must carry `Project Stage: existing` so
  `projectstage.Parse` returns `Existing` and bootstrap is skipped.
- The detector runs on every `UserPromptSubmit`, so it must be cheap
  (single root readdir) — it cannot depend on `.workflow/analysis.json`, which
  does not exist on the first prompt of a brownfield repo.

## Out of Scope (v1)

- Brownfield onboarding documentation in `new-project-guide.md` — deliberate
  follow-up, recorded as a deferred roadmap item.
