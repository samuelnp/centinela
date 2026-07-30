# docs-step-markdown-first — senior-engineer

## Files Touched

| File | Change | Slice |
|---|---|---|
| internal/orchestration/output_rules.go | removed docs-specialist exemption; added `RoleDocsSpecialist` dispatch case requiring `hasDocsOutput` | S1 |
| internal/orchestration/output_helpers.go | added `hasDocsOutput` (docs/ prefix or exact README.md); `hasImplementationOutput` untouched | S1 |
| internal/workflow/validate_docs.go | changelog check for ALL features; deleted `validateDocsUserFacing` + `kbDir` + orchestration import; renamed `validateDocsInternal` → `validateChangelog` | S1 |
| cmd/centinela/hook_context.go | docs nudges: user-facing keys off `.workflow/<f>-documentation-specialist.md`; changelog nudge for every feature | S1 |
| cmd/centinela/hook_statusline_rules.go | MISSING_DOCS_OUTPUT user-facing branch checks the specialist evidence file | S1 |
| internal/ui/render_review.go | RenderDocumentationNeeded → docs context + real markdown updates text | S1 |
| internal/evidence/artifact_docs.go | companion Outcome line → guides/README + changelog | S1 |
| internal/evidence/companion_skeletons.go | docs-specialist headers → {Docs Updated, Changelog, Outcome} | S1 |
| internal/docsctx/context.go (NEW) | `Load(feature)`: resolves 4 inputs, aggregated missing-required error | S2 |
| internal/docsctx/render.go (NEW) | `Render(Context)`: plain-markdown assembly, `> source:` lines, changelog hint | S2 |
| cmd/centinela/docs_context.go (NEW) | `docs context <feature>` cobra command, slug validation, SilenceUsage/Errors | S2 |
| cmd/centinela/docs.go | Short text updated; RunE rejects unknown subcommands with exit 1 (cobra non-runnable parents exit 0 otherwise — spec requires failure) | S3 |
| cmd/centinela/merge.go | deleted `docsPortalRegen` seam, `mergePortalTitle`, regen call + notice, docgen import | S3 |
| internal/docgen/ (38 files) | git rm -r | S3 |
| cmd/centinela/docs_generate.go, docs_validate.go + 4 colocated test files | git rm | S3 |
| docs/project-docs/ (whole tree) | git rm -r | S3 |
| tests/acceptance/docs_generate_acceptance_test.go, docs_knowledge_base_acceptance_test.go, improve_docs_llm_hybrid_ui_acceptance_test.go; tests/unit/docs_generate_unit_test.go; tests/integration/docs_knowledge_base_integration_test.go | git rm (wholesale-obsolete per plan) | S3 |
| docs/architecture/documentation-generator-prompt.md | full rewrite: markdown-first specialist, docs context sole input surface, no exemption | S4 |
| docs/architecture/evidence-contract.md | global rule 5 "no exemptions"; docs-specialist per-role rule rewritten | S4 |
| docs/architecture/artifact-templates.md | portal row → changelog + real doc updates; docs-specialist bullet added | S4 |
| docs/architecture/workflow-enforcement.md | docs artifact-table row + docs-specialist bullet | S4 |
| internal/scaffold/assets/docs/architecture/{documentation-generator-prompt,evidence-contract,artifact-templates,workflow-enforcement}.md | byte-identical mirrors (cmp-verified) | S4 |
| CLAUDE.md + internal/scaffold/assets/CLAUDE.md | step-5 text + docs artifact-table row (same edits both) | S4 |
| README.md | DOCART mermaid node → guides/README + changelog | S4 |
| HOWTO.md | docs-step block: docs context replaces validate/generate; `centinela docs` mention kept | S4 |
| docs/guides/getting-started.md, docs/guides/workflow-and-hooks.md | docs validate/generate instructions → new contract + docs context | S4 |
| internal/setup/opencode_agents.go + testdata/golden/{opencode,codex}/AGENTS.md | command list swap, fixtures in lockstep (setup suite green) | S4 |

## Architecture Compliance

- Boundaries: `internal/docsctx` imports stdlib only; `cmd/centinela` imports
  `internal/docsctx` + `internal/worktree` (cmd → internal is the allowed
  n-tier direction). `internal/workflow` dropped its `orchestration` import.
- G1: every added/edited source file ≤100 lines (max touched: hook_context.go
  at 97; docsctx files 62/32).
- G7 (no gate weakened): user-facing docs step now requires changelog AND
  real-file specialist outputs; internal path behavior-equivalent (rename
  only). `hasImplementationOutput` keeps its `docs/project-docs/` exclusion
  so legacy senior-engineer evidence downstream is unaffected.
- Scaffold parity: all four edited arch docs cmp-identical to their
  `internal/scaffold/assets/` mirrors; senior-engineer-prompt.md left
  unchanged (its prefix list still matches code).

## Type-Safety Notes

- No `interface{}`/`any`; `docsctx.Section`/`Context` are fully typed structs.
- `Load` returns `(Context, error)` with a zero Context on failure; missing
  required inputs aggregate into one error naming all paths.
- `go build ./...`, `go vet ./internal/...` clean.

## Trade-Offs

- `docsCmd` gained a small RunE: cobra prints help with exit 0 for unknown
  subcommands of non-runnable parents (Args validators are skipped), which
  would break the "HTML pipeline is gone" spec scenario. The RunE makes
  `docs generate`/`docs validate` exit 1 with
  `unknown command "..." for "centinela docs"`.
- Kept `.workflow` path literal in docsctx (matches hook_context idiom)
  rather than importing workflow.WorkflowDir — avoids a package dependency
  for one constant.
- Doc-version header stays `1` on the rewritten prompt (bumping
  CurrentDocVersion is global; deferred — see below).

## Deferred Findings

- `migrate-doc-version-refresh` (recorded via `centinela roadmap defer`,
  source docs-step-markdown-first/senior-engineer): migrate drift detection
  is doc-version-only, so downstream repos on version 1 never receive the
  rewritten documentation-generator prompt (or any same-version content
  rewrite); needs content-hash or version-bump strategy.

## Handoff

- **Next role:** qa-senior.
- **Compile-broken inventory** (`go test ./... -run xxxNONE`): exactly ONE
  package fails to build — `cmd/centinela`, from
  `merge_regen_test.go` (6× undefined `docsPortalRegen`) and
  `merge_truthful_test.go` (3× undefined `docsPortalRegen`).
  `merge_regen_helper_test.go` compiles but is orphaned (plan: delete both
  regen files, de-stub truthful in tests step). Every other package compiles.
- **Expected runtime failures** (pinning tests, rewrite in tests step):
  - internal/workflow: `TestCompleteTransitionsToDone` (steps_test.go),
    `TestValidateDocsOutput` (validate_docs_test.go),
    `TestValidateArtifactsDocsStrictOrchestration`
    (validate_orchestration_docs_test.go) — all per plan table.
  - internal/orchestration: `TestValidateActionableOutputsRoleRules`
    (output_rules_test.go) — pins the deleted exemption. NOTE: plan table
    lists validate_test.go; the actual breaking file is output_rules_test.go.
  - internal/evidence: `TestCompanionSkeletonPerRoleHeaders`
    (fill_companion_test.go) — pins old "KB Pages" headers. NOT in the plan
    table; same class (companion header swap), needs rewrite too.
  - tests/acceptance (all confirmed failing, per plan table):
    `TestRDSCleanMergeRegeneratesPortal`,
    `TestRDSPortalRegenFailureDoesNotFailMerge`,
    `TestRDSUserFacingPassesWithKnowledgeBase`,
    `TestRDSUserFacingFailsWithoutKnowledgeBase`,
    `TestAccMergePortalRegenTargetsPrimaryTree`,
    `TestDocsLatestFeaturesAndGettingStartedStayInSync`.
  - tests/integration: `TestReadmeCentinelaUsageLinksLandingPageHowto` —
    confirmed failing (HOWTO no longer pins `docs generate`); rewrite to pin
    `centinela docs context`.
  - cmd/centinela (once compile fixed): hook_context_docs_test.go,
    hook_context_docs_internal_test.go, hook_statusline_docs_internal_test.go,
    hook_statusline_rules_test.go (project-docs fixture — NOT in plan table,
    verify it), complete_branches_test.go, complete_cmd_test.go still
    reference docs/project-docs fixtures; per plan they get evidence-file
    fixtures.
- **Outstanding TODOs:** legacy-spec scenario rewrites (5 .feature files) in
  tests step; new tests per plan (docsctx colocated ≥97%, docs_context_test,
  orchestration docs-output rule tests, workflow changelog-for-all,
  docs_step_markdown_first_acceptance_test).
- Dogfood: `/tmp/centinela-dsmf docs context docs-step-markdown-first` exits
  0 with all four sections; bogus slug exits 1 naming all three missing
  paths; `docs generate` exits 1 as unknown command.
