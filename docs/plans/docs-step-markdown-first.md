# Plan: docs-step-markdown-first

**Date:** 2026-07-30 · **Phase:** 13 — Lighter Centinela · **Surface:** internal
**Brief:** docs/features/docs-step-markdown-first.md · **Spec:** specs/docs-step-markdown-first.feature

## Design decisions (grounded in current code)

1. **Single source for the real-file rule.** The "outputs must include a real
   doc file" check lives ONLY in `internal/orchestration/output_rules.go`
   (`dispatchRoleOutputs` gains a `RoleDocsSpecialist` case; the top-of-func
   exemption is removed). `validateDocsUserFacing` does NOT re-implement it —
   it collapses to the changelog check, because orchestration evidence
   validation already runs for the docs step via `RequiredRolesForFeature`
   (docs-specialist is required for user-facing features; internal features
   return no roles, so the new rule cannot bite them).
2. **User-facing features now ALSO require the changelog.** `validateDocsOutput`
   requires a non-empty `.workflow/<feature>-changelog.md` for every feature
   (previously user-facing skipped it). Rationale: the rewritten prompt makes
   "write the changelog" a specialist duty, `centinela docs context` prints the
   changelog draft, and delivery already consumes it. Net: gate strengthened,
   never weakened.
3. **Accepted-path rule (fixed by brief):** an output passes if it is an
   existing regular file and (`strings.HasPrefix(path, "docs/")` after
   normalization, or `path == "README.md"`). No carve-outs for
   `docs/features/`/`docs/plans/` — the brief fixes "any path under docs/ or
   README.md" so scaffolded projects without `docs/guides/` pass. The gaming
   vector (listing the plan file) is mitigated by prompt duties, not code.
4. **`centinela docs validate` is DELETED** along with `docs generate`. Its
   whole body is `docgen.ValidateInputs()` (PROJECT.md/ROADMAP.md/roadmap.json
   presence — generator preflight). With the generator gone it validates
   nothing anyone consumes; `centinela docs context` reports its own missing
   inputs with actionable errors. The `docs` root command survives with the
   single `context` subcommand.
5. **`docs context` output is plain markdown on stdout** (pipe-friendly, no
   lipgloss panels around the body): `# Docs Context: <feature>` then sections
   `## Feature brief`, `## Plan`, `## Spec scenarios`, `## Changelog draft`,
   each embedding the file verbatim with a `> source:` line. Brief, plan and
   spec are required — ALL missing ones are reported in one error, non-zero
   exit. Changelog draft is optional: absent → the section prints a hint
   (`no changelog draft yet — run: centinela artifact new <feature> changelog`)
   and exit stays 0.

## File-by-file

### Deleted (code step, slice S3)

- `internal/docgen/` — entire package (38 files incl. colocated tests).
- `cmd/centinela/docs_generate.go`, `docs_generate_test.go`,
  `docs_generate_more_test.go` — command + tests.
- `cmd/centinela/docs_validate.go`, `docs_validate_test.go` — per decision 4.
- `cmd/centinela/docs_helpers_test.go` — fixture builder only used by the two
  deleted command tests.
- `docs/project-docs/` — generated portal + KB output (git rm, whole tree).
- `tests/acceptance/docs_generate_acceptance_test.go`,
  `tests/acceptance/docs_knowledge_base_acceptance_test.go`,
  `tests/acceptance/improve_docs_llm_hybrid_ui_acceptance_test.go`,
  `tests/unit/docs_generate_unit_test.go`,
  `tests/integration/docs_knowledge_base_integration_test.go` — pin deleted
  behavior wholesale (removed via `rm` in the code step so the tree compiles;
  file deletion is not a tests-step write).

### Edited (code step)

- `internal/orchestration/output_rules.go` — remove the
  `if role == RoleDocsSpecialist { return nil }` exemption; add
  `case RoleDocsSpecialist:` to `dispatchRoleOutputs` requiring
  `hasDocsOutput(files)`; error message names the contract:
  "documentation-specialist outputs must include a real updated file under
  docs/ or README.md".
- `internal/orchestration/output_helpers.go` — add `hasDocsOutput(paths)`
  (docs/ prefix or exact README.md). `hasImplementationOutput` keeps its
  `docs/project-docs/` exclusion untouched (legacy senior-engineer evidence in
  downstream repos must not start passing/failing differently).
- `internal/workflow/validate_docs.go` — `validateDocsOutput` calls the
  changelog check for ALL features; delete `validateDocsUserFacing`, `kbDir`,
  and the `orchestration` import. Rename `validateDocsInternal` →
  `validateChangelog`.
- `cmd/centinela/merge.go` — delete `docsPortalRegen` seam, `mergePortalTitle`,
  the regen call + notice, and the `docgen` import.
- `cmd/centinela/hook_context.go` — docs-step nudge loop: user-facing branch
  stats `.workflow/<f>-documentation-specialist.md` instead of
  `docs/project-docs/index.html`; changelog nudge now applies to every feature
  (matches decision 2).
- `cmd/centinela/hook_statusline_rules.go` — `MISSING_DOCS_OUTPUT` user-facing
  branch checks `fileExists(".workflow/" + f + "-documentation-specialist.md")`.
- `internal/ui/render_review.go` — `RenderDocumentationNeeded` text: run
  `centinela docs context <feature>`, update `docs/guides/` or `README.md`,
  write evidence listing the real files (drop "Then write:
  docs/project-docs/index.html").
- `internal/evidence/artifact_docs.go` — companion md Outcome line: guides/
  README files updated + changelog, not "KB pages and project-docs entries".
- `internal/evidence/companion_skeletons.go` — `RoleDocsSpecialist` headers
  `{"KB Pages", "project-docs Entries", "Outcome"}` →
  `{"Docs Updated", "Changelog", "Outcome"}`.
- `internal/setup/opencode_agents.go` — command list: drop
  `centinela docs validate` and `centinela docs generate --out …`; add
  `centinela docs context <feature>`. Regenerate/edit golden fixtures
  `internal/setup/testdata/golden/{opencode,codex}/AGENTS.md` in lockstep.

### Added (code step, slice S2)

- `internal/docsctx/context.go` — `Load(feature) (Context, error)`: resolves
  the four input paths, aggregates missing-required errors. ≤100 lines.
- `internal/docsctx/render.go` — `Render(Context) string` markdown assembly.
  ≤100 lines.
- `cmd/centinela/docs_context.go` — thin cobra command `docs context
  <feature>` (ExactArgs(1), slug validation via worktree.ValidateFeatureSlug),
  prints `docsctx.Render`.

### Docs / prompts (code step, slice S4 — every `docs/architecture/` edit is
mirrored byte-identical into `internal/scaffold/assets/docs/architecture/`)

- `docs/architecture/documentation-generator-prompt.md` — full rewrite: role =
  markdown-first documentation specialist. Duties: run `centinela docs context
  <feature>` (sole mandated input surface — no repo crawl), update
  `docs/guides/` and/or `README.md` where the feature changed behavior, CLI
  surface, or config (a user-facing feature with genuinely no doc-worthy
  change still updates the closest guide section or README capability list —
  the gate requires ≥1 real file), write `.workflow/<feature>-changelog.md`
  via `centinela artifact new`, evidence pair via the evidence CLI with
  `outputs` listing the real files written. Rules section: role is NO LONGER
  exempt from outputs-must-be-real-files; ≥1 output under `docs/` or
  `README.md`. Internal features: changelog only (unchanged).
- `docs/architecture/evidence-contract.md` — remove the docs-specialist
  exemption paragraph; document the new per-role rule; drop
  `docs/project-docs/` from the senior-engineer non-implementation list ONLY
  in prose if referenced as live guidance (keep list matching code — code
  keeps the prefix, so keep prose too).
- `docs/architecture/senior-engineer-prompt.md` — same `docs/project-docs/`
  mention: leave matching code (no change) unless wording implies the portal
  exists; adjust phrasing only.
- `docs/architecture/artifact-templates.md` — replace the
  "`docs/project-docs/index.html` after the docs step" row with changelog +
  real doc updates.
- `docs/architecture/workflow-enforcement.md` — docs-step artifact table row.
- `CLAUDE.md` + `internal/scaffold/assets/CLAUDE.md` — step 5 description
  ("Documentation specialist output + generated docs HTML" → "updated
  docs/guides/ or README.md + changelog") and the docs row of the artifact
  table.
- `README.md` — mermaid node `DOCART["docs/project-docs/index.html"]` → real
  doc updates; any `docs generate` mention.
- `HOWTO.md` — lines 135–143: replace `docs validate`/`docs generate` block
  with `docs context`; keep `centinela docs` present (pinned by
  `readme_centinela_usage_integration_test.go`, which is rewritten in the
  tests step to pin `centinela docs context`).
- `docs/guides/getting-started.md`, `docs/guides/workflow-and-hooks.md` —
  replace `docs validate`/`docs generate` instructions with the new docs-step
  contract and `docs context`.

## Test strategy

**Existing tests that break, and where their updates land:**

| File | Break | Disposition |
|---|---|---|
| `internal/docgen/*_test.go` | package deleted | deleted with package (code step) |
| `cmd/centinela/docs_generate{,_more}_test.go`, `docs_validate_test.go`, `docs_helpers_test.go` | target deleted | deleted (code step) |
| `cmd/centinela/merge_regen_test.go`, `merge_regen_helper_test.go` | `docsPortalRegen` seam gone | deleted; replaced by a no-portal-regen assertion (tests step) |
| `cmd/centinela/merge_truthful_test.go` | stubs `docsPortalRegen` | edited: drop the stub lines (tests step) |
| `cmd/centinela/hook_context_docs_test.go`, `hook_context_docs_internal_test.go`, `hook_statusline_docs_internal_test.go` | fixtures build `docs/project-docs` | rewritten against evidence-file fixtures (tests step) |
| `cmd/centinela/complete_branches_test.go`, `complete_cmd_test.go` | `docs/project-docs/index.html` fixture + legacy evidence outputs | fixture swapped for another real file (tests step) |
| `internal/workflow/validate_docs_test.go`, `validate_orchestration_docs_test.go`, `steps_test.go` | KB/portal assertions | rewritten: changelog-for-all + orchestration rule (tests step) |
| `internal/orchestration/validate_test.go` | outputs pin `docs/project-docs/index.html` | still a legal docs/ path but file fixture kept; extend with non-docs-output failure case (tests step) |
| `tests/acceptance/right_size_docs_step_userfacing_test.go` | KB required assertions | rewritten to new user-facing contract (tests step) |
| `tests/acceptance/right_size_docs_step_merge_test.go` | asserts `docsPortalRegen(repo)` source text | rewritten: asserts merge.go does NOT reference portal regen (tests step) |
| `tests/acceptance/docs_latest_features_getting_started_test.go` | reads `internal/docgen/render_story.go` | rewritten: guide-only sync assertions (tests step) |
| `tests/integration/readme_centinela_usage_integration_test.go` | HOWTO pin `centinela docs generate` | rewritten to pin `centinela docs context` (tests step) |
| `tests/acceptance/merge_truthful_delivery_scope_test.go` | asserts portal `index.html` exists post-merge | assertion dropped (tests step) |

**Legacy specs whose scenarios now contradict behavior** (rewrite the
contradicted scenarios in the tests step alongside their acceptance tests;
if hooks block `specs/` writes at that step, fall back to the code step):
`specs/right-size-docs-step.feature`, `specs/docs-knowledge-base-pages.feature`,
`specs/generate-html-project-docs.feature`,
`specs/improve-docs-llm-hybrid-ui.feature`,
`specs/docs-latest-features-getting-started.feature`.

**New tests (tests step):**

- `internal/docsctx/context_test.go`, `render_test.go` — colocated (coverage
  gate is per-package; target ≥97%): happy path, each missing-required
  combination, absent-changelog hint.
- `cmd/centinela/docs_context_test.go` — command wiring, bad slug, exit codes.
- `internal/orchestration/` docs-output rule tests — exemption removed,
  docs/ pass, README.md pass, plan-file-only pass (documents decision 3),
  no-real-file fail.
- `internal/workflow/` — user-facing now requires changelog; internal
  unchanged.
- `tests/acceptance/docs_step_markdown_first_acceptance_test.go` (split into
  ≤100-line files as needed) — `// Acceptance:
  specs/docs-step-markdown-first.feature` + one `// Scenario:` per spec
  scenario; drives the real binary for `docs context` and asserts
  `docs generate` is an unknown command.
- `validate.commands` already runs the acceptance tier; confirm the new file
  is picked up (no config change expected).

## Rollout (ordering inside the workflow; `centinela complete` auto-commits per step)

1. **S1 — gate swap** (smallest correct slice): `output_rules.go` +
   `output_helpers.go` + `validate_docs.go` + hook/statusline/render text.
   Tree compiles; old commands still exist.
2. **S2 — `docs context`**: `internal/docsctx` + cobra command. Dogfood with a
   /tmp-built binary (`go build -o /tmp/centinela ./cmd/centinela`) — the
   installed binary lags the worktree.
3. **S3 — deletion**: docgen package, docs generate/validate commands, merge
   portal regen, `docs/project-docs/` tree, wholesale-obsolete test files
   (`go build ./...` AND `go test ./... -run xxxNONE` to inventory remaining
   compile breaks; broken pinning tests are expected until the tests step).
4. **S4 — prose**: prompt rewrite + all mirrors + golden fixtures + CLAUDE/
   README/HOWTO/guides.
5. **Tests step**: rewrites per table above + new tests + legacy-spec scenario
   cleanup; suite green, per-package coverage ≥97% on touched packages.
6. **Can wait** (out of scope): docstring-gate, README drift automation,
   Docusaurus/godoc pipeline, migrate-time pruning of stale
   `docs/project-docs/` in downstream repos (deferred).

## Own-workflow note

This feature's brief declares no `surface:` line → internal → its own docs
step needs only the changelog (old and new binaries agree), so the
installed-binary lag cannot deadlock this workflow. Create
`.workflow/docs-step-markdown-first-changelog.md` early via
`centinela artifact new`.
