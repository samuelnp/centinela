# docs-step-markdown-first — qa-senior

## Test Inventory

| Tier | File | Covers |
|---|---|---|
| unit (colocated) | internal/docsctx/context_test.go | Load happy path; every missing-required combination aggregated into ONE error naming all (and only) missing paths |
| unit (colocated) | internal/docsctx/context_boundary_test.go | absent changelog stays addressable; 0-byte required input counts as present (documented boundary) |
| unit (colocated) | internal/docsctx/render_test.go | section order + `> source:` lines, absent-changelog hint, trailing-newline trimming, determinism |
| unit (colocated) | cmd/centinela/docs_context_test.go | command wiring, bad-slug rejection, aggregated missing-input error, `docs generate`/`docs validate` unknown → RunE error (exit 1) |
| unit (colocated) | cmd/centinela/merge_no_regen_test.go | clean merge emits no portal-regen notice and creates no docs/project-docs/index.html (replaces deleted merge_regen{,_helper}_test.go; hosts seedCleanMergeRepo) |
| unit (colocated) | internal/orchestration/output_rules_docs_test.go | docs-output rule table: exemption removed (empty outputs fail), docs/ pass, exact README.md pass, ./README.md normalization, absolute-path fail, directory fail, docs/../x cleaned form, plan-file-only pass (pins plan decision 3) |
| unit (colocated) | internal/orchestration/feature_surface_format_test.go | surface parser fix: bare/blockquote/dash/star/indented/mixed-case/underscore all user-facing; mid-line prose, internal values, bold form, no-line all internal |
| unit (colocated) | internal/orchestration/output_rules_test.go (rewritten cases) | docs-specialist real-file error, docs-file rule error, docs/ pass |
| unit (colocated) | internal/orchestration/validate_test.go (fixture swap) | writeEvidence now uses docs/guides/guide.md, no project-docs |
| unit (colocated) | internal/workflow/validate_changelog_test.go | validateChangelog missing/0-byte/whitespace-only/non-empty for BOTH internal and user-facing fixtures (pins changelog-for-all, decision 2) |
| unit (colocated) | internal/workflow/validate_docs_test.go (rewritten) | user-facing artifact gate = changelog only; error names the path; ValidateArtifacts passes non-strict |
| unit (colocated) | internal/workflow/validate_orchestration_docs_test.go (rewritten) | strict user-facing: missing evidence fails; .workflow-only outputs fail naming docs/ or README.md; docs/guides output passes |
| unit (colocated) | internal/workflow/steps_test.go (fixture swap) | Complete-to-done now uses a changelog, no KB bundle |
| unit (colocated) | cmd/centinela hook tests (rewritten) | hook_context_docs_test.go: user-facing nags for specialist evidence file + changelog, evidence file silences docs nag; hook_statusline_rules_test.go: MISSING_DOCS_OUTPUT keys off the evidence file then MISSING_DOCS_EVIDENCE |
| unit (colocated) | cmd/centinela complete tests (fixture swap) | validation-specialist outputs no longer reference the deleted portal |
| unit (colocated) | internal/evidence/fill_companion_test.go (updated) | docs-specialist companion header "Docs Updated" (was "KB Pages") |
| integration | tests/integration/readme_centinela_usage_integration_test.go | HOWTO pins `centinela docs context`; `docs generate`/`docs validate` banned |
| integration | tests/integration/right_size_docs_step_integration_test.go | internal light path unchanged; user-facing strict path demands specialist evidence (portal/KB assertion removed) |
| acceptance | tests/acceptance/docs_step_markdown_first_context_test.go | real binary: docs context happy / missing-aggregation / optional-changelog-hint / idempotent double run |
| acceptance | tests/acceptance/docs_step_markdown_first_pipeline_test.go | real binary: docs generate + docs validate exit non-zero as unknown commands; docgen dir gone; merge.go clean of portal refs |
| acceptance | tests/acceptance/docs_step_markdown_first_evidence_test.go | scenarios 1–4: real doc file passes, README.md alone passes, .workflow-only outputs fail naming docs/ or README.md, ghost path fails real-file check |
| acceptance | tests/acceptance/docs_step_markdown_first_changelog_test.go | scenarios 5–6: user-facing missing changelog names the path; internal changelog-only passes under a strict saved workflow |
| acceptance | tests/acceptance/right_size_docs_step_userfacing_test.go (rewritten) | user-facing artifact gate: changelog pass/fail (KB scenarios superseded) |
| acceptance | tests/acceptance/right_size_docs_step_merge_test.go (rewritten) | merge.go contains no docsPortalRegen/docgen/project-docs; internal/docgen and docs/project-docs deleted |
| acceptance | tests/acceptance/docs_latest_features_getting_started_test.go (rewritten) | getting-started guide teaches docs context + changelog, bans deleted commands (render_story read dropped) |
| acceptance | tests/acceptance/merge_truthful_delivery_scope_test.go (rewritten scenario) | merge from worktree CWD regenerates NO portal in the primary tree |
| acceptance | tests/acceptance/deterministic_artifact_scaffolds_fill_test.go + agent_evidence_contract_acceptance_test.go (updated pins) | "Docs Updated" header; evidence-contract "no exemptions"/"NOT exempt"/"real updated file under" |

## Coverage Gaps

Spec scenario coverage for specs/docs-step-markdown-first.feature: **10/10**.

| # | Scenario | Test |
|---|---|---|
| 1 | User-facing docs step passes with a real updated doc file | TestDSMFUserFacingPassesWithRealDocFile |
| 2 | README.md alone satisfies the real-file rule… | TestDSMFReadmeAloneSatisfiesRealFileRule |
| 3 | User-facing docs step fails when no output is a real doc file | TestDSMFUserFacingFailsWithoutRealDocOutput |
| 4 | Documentation-specialist outputs must be real files | TestDSMFSpecialistOutputsMustBeRealFiles |
| 5 | User-facing docs step fails without a changelog entry | TestDSMFUserFacingFailsWithoutChangelog |
| 6 | Internal docs step keeps the one-line changelog contract | TestDSMFInternalKeepsChangelogOnlyContract |
| 7 | docs context prints the curated feature-scale inputs | TestDSMFDocsContextPrintsCuratedInputs (+ idempotency) |
| 8 | docs context reports every missing required input at once | TestDSMFDocsContextAggregatesMissingInputs |
| 9 | docs context treats the changelog draft as optional | TestDSMFDocsContextChangelogOptionalHint |
| 10 | The HTML pipeline is gone | TestDSMFHTMLPipelineCommandsGone + TestDSMFNoMergeTimePortalRegenCodePath + TestAccMergeDoesNotRegeneratePortalInPrimaryTree + TestRunMergeCleanDoesNotRegeneratePortal |

Remaining gap (not spec-scenario): legacy .feature files still describe the
deleted pipeline — the prewrite hook blocks specs/ writes during the tests
step, so their scenario rewrites were NOT performed (deferred, see below).
The rewritten acceptance tests reference specs/docs-step-markdown-first.feature
instead, so no test asserts the contradicted legacy scenarios any more.

## Acceptance Wiring

centinela.toml `validate.commands` runs `go test ./... -coverprofile=coverage.out`
— tests/acceptance is under `./...`, so the acceptance tier (including the four
new docs_step_markdown_first_* files) executes in the single profiled validate
run. No config change needed. Binary-driving tests build from ./cmd/centinela
into a temp dir (buildCent) — never the installed binary — and use local
temp git repos only (no network URLs).

## Regression Guards

- merge_no_regen_test.go (behavioral) + right_size_docs_step_merge_test.go /
  docs_step_markdown_first_pipeline_test.go (source scan): portal regen cannot
  silently return to merge.
- feature_surface_format_test.go: marker-form surface lines can no longer be
  silently dropped; mid-line prose negative keeps the line-start discipline.
- output_rules_docs_test.go pins the accepted decision-3 gaming vector as a
  comment-documented PASS so any future tightening is a conscious change.
- readme/HOWTO/guide tests ban `centinela docs generate|validate` from
  resurfacing in prose.
- validate_changelog_test.go pins changelog-for-all so a surface regression
  cannot re-exempt user-facing features.

## Surface Parser Before/After Audit

Grep of `docs/features/*.md` for `^[> *-]*surface:` (and bold forms — none
exist): 16 briefs declare a surface. Before the fix, marker-form lines parsed
to "": `> surface: user-facing` (landing-page.md) was misclassified internal;
seven `- surface: internal` briefs parsed to "" but classified internal anyway.
After the fix, **only docs/features/landing-page.md changes classification**
(internal → user-facing), which is its genuine declaration. All `- surface:
internal` briefs and all bare-form briefs keep their classification. No other
brief changes.

## Deferred Findings

- `legacy-docs-spec-scenarios-cleanup` (recorded via `centinela roadmap defer`,
  source docs-step-markdown-first/qa-senior): the 5 legacy specs (+ the
  merge-truthful-delivery portal scenario) still describe the deleted pipeline;
  the prewrite hook blocks specs/ writes during the tests step, so the
  contradicted scenarios need a plan-step-capable follow-up.
- `surface-line-format-detection` (deferred by the edge-case report) is being
  **FIXED in this feature** — `normalizeSurface` now strips leading `>`, `-`,
  `*` markers, with the full format matrix pinned by
  internal/orchestration/feature_surface_format_test.go. The roadmap entry can
  be closed when this feature merges.
- `stale-project-docs-pruning` and `changelog-scanner-line-limit` remain
  deferred (edge-case report), unchanged here.

## Handoff

- **Next role:** validation-specialist.
- Suite: single profiled run `go test ./... -coverprofile=coverage.out` +
  `COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh`;
  `./scripts/check-fmt.sh` exit 0.
- Touched-package coverage ≥97% target: internal/docsctx at 100.0%; per-package
  numbers for cmd/centinela, internal/orchestration, internal/workflow,
  internal/evidence recorded by the coverage gate run.
- Legacy spec files intentionally untouched (hook-blocked) — do not treat the
  stale scenarios as live acceptance criteria.
