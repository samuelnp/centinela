# docs-step-markdown-first — planner

### Planner Report: docs-step-markdown-first
**Date:** 2026-07-30

#### Problem

The docs step's expensive path hurts every team running a user-facing feature:
a fresh-context documentation-specialist is prompted to crawl the whole repo
(PROJECT.md, ROADMAP.md, all briefs/plans/specs, every `kb/*.md`) and then
drive `internal/docgen` (~1,200 lines across 38 files) to render a portal
(`docs/project-docs/index.html`, 63KB in this repo) plus per-feature KB
markdown+HTML pairs that nobody reads. Cost is ~40–70K tokens per user-facing
feature and grows with project size, while the documentation humans actually
read — `docs/guides/` and `README.md` — is never touched by the step. The gate
(`validateDocsUserFacing`) enforces the unread artifacts (index.html +
kb/<f>.md + kb/<f>.html) and `output_rules.go` explicitly EXEMPTS the
documentation-specialist from the outputs-must-be-real-files rule, so the one
role whose whole job is producing files is the one role allowed to fabricate
its output list. Why now: Phase 13 ("Lighter Centinela") targets exactly this
class of waste, and the follow-up `docstring-gate` is blocked on this
redesign.

#### Scope

**In (v1):**
- Delete the HTML pipeline: `internal/docgen` (whole package),
  `centinela docs generate`, `centinela docs validate` (decided: delete — its
  body is only the generator preflight `docgen.ValidateInputs`), the
  `kb/<feature>.md → .html` contract, the `docs/project-docs/index.html` gate,
  merge-time portal regeneration (`docsPortalRegen` seam in
  `cmd/centinela/merge.go`), and the committed `docs/project-docs/` output.
- Re-gate user-facing docs: remove the docs-specialist exemption in
  `internal/orchestration/output_rules.go`; new dispatch rule requires ≥1
  existing output file under `docs/` or exactly `README.md`. The rule lives
  ONLY in orchestration; `validateDocsOutput` collapses to the changelog
  check for all features (user-facing now also requires the non-empty
  changelog — strengthening, per the "no gate weakened" constraint).
- New `centinela docs context <feature>`: prints brief, plan, spec scenarios,
  and changelog draft as plain markdown; brief/plan/spec required (aggregated
  error), changelog optional with an `artifact new` hint. Logic in a new
  `internal/docsctx` package; thin cobra command.
- Rewrite `documentation-generator-prompt.md` (+ byte-identical scaffold
  mirror), update evidence-contract/artifact-templates/workflow-enforcement/
  CLAUDE.md/README.md/HOWTO.md/guides, `internal/setup/opencode_agents.go`
  command list, and the opencode/codex AGENTS.md golden fixtures.
- Hook surfaces: `hook_context.go`, `hook_statusline_rules.go`
  (MISSING_DOCS_OUTPUT now keys off the documentation-specialist evidence
  file), `render_review.go` nudge text.
- Rewrite/delete the ~831 lines of tests pinning the old pipeline; new spec +
  acceptance/unit coverage for the new contract.

**Out (v1):** see Out-of-Scope below.

#### Dependencies & Assumptions

- **Internal modules:** `internal/orchestration` (policy.go role gating stays
  as-is — docs-specialist required only for user-facing; output_rules/
  output_helpers change), `internal/workflow/validate_docs.go`,
  `internal/evidence` (artifact_docs.go companion, companion_skeletons.go
  headers, existing `KindChangelog` artifact reused as the "changelog draft"),
  `internal/setup` (AGENTS content + golden fixtures), `internal/ui`,
  `cmd/centinela` (merge, hooks, docs command family).
- **Prior features built on:** right-size-docs-step (surface-aware docs step —
  its internal/changelog path is kept verbatim),
  enforce-actionable-orchestration-evidence (the real-files rule this feature
  extends to the last exempt role), evidence-cli, parallel-feature-worktrees.
- **Assumptions:** `IsUserFacingFeature` (brief `surface:` line) remains the
  surface switch; `worktree.ValidateFeatureSlug` is reusable for the new
  command; deleting `internal/docgen` breaks no non-test importer outside
  the files inventoried (verified by grep: cmd docs_generate/docs_validate/
  merge only); coverage gate is per-package so removing the well-covered
  docgen package cannot drop another package below 95%; `rm` of obsolete
  test files during the code step is not blocked by prewrite hooks (bash
  deletion, not a tool write).

#### Risks (table: Risk | Impact | Likelihood | Mitigation)

| Risk | Impact | Likelihood | Mitigation |
|---|---|---|---|
| ~831 lines of pinning tests (13+ files across 4 tiers) break at once; code step ends with a non-compiling test tree | High | Certain | Planned two-phase sequencing: wholesale-obsolete test files deleted via `rm` in code step; rewrites land in tests step. Run `go test ./... -run xxxNONE` at end of code step to inventory every compile break (go build hides them) |
| Installed binary lags worktree: new `docs context` / gate logic untested against reality | Medium | High | Dogfood via `/tmp` binary built from `./cmd/centinela` before the validate gate; this feature's own brief is internal (no `surface:` line) so its own docs step uses the unchanged changelog path under both binaries — no deadlock |
| Scaffold-mirror / golden-fixture drift (parity test covers only 8 arch docs; AGENTS.md goldens pin `docs generate`) | Medium | High | Explicit lockstep edits: every `docs/architecture/` file + `internal/scaffold/assets/` mirror + `internal/setup/testdata/golden/{opencode,codex}/AGENTS.md` + `internal/scaffold/assets/CLAUDE.md` are in the file-by-file plan |
| Regression on right-size-docs-step: internal changelog contract accidentally altered, or user-facing legacy evidence (outputs=`docs/project-docs/index.html`) breaks downstream mid-flight workflows | High | Medium | Internal path kept verbatim (rename only); the new rule accepts any existing `docs/` file, so legacy evidence still passes in repos where the portal file exists; nudge text changes only |
| Scaffolded projects without `docs/guides/` fail the new gate | Medium | Low | Rule accepts ANY `docs/` path or `README.md` (fixed by brief); spec scenario covers README-only projects |
| Gaming vector: specialist lists `docs/plans/<f>.md` (always exists) and passes | Low | Medium | Accepted per brief decision; mitigated by rewritten prompt duties; README drift automation is an explicit follow-up |
| `merge_truthful_*` tests stub the removed `docsPortalRegen` seam; merge-path coverage dips in `cmd/centinela` | Medium | Medium | Tests-step rewrite drops stubs and adds a no-portal-regen assertion; keep cmd package ≥97% |
| Legacy specs (right-size-docs-step, docs-knowledge-base-pages, generate-html-project-docs, improve-docs-llm-hybrid-ui, docs-latest-features-getting-started) now describe deleted behavior; spec-traceability/`DetectSpecConflicts` friction | Medium | Medium | Rewrite contradicted scenarios alongside their acceptance rewrites in the tests step; fallback to code step if hooks block `specs/` writes then |
| New `internal/docsctx` package misses the per-package 95% coverage gate | Medium | Low | Colocated `_test.go` planned with the package; target ≥97% |

#### Rollout

1. **S1 — gate swap** (smallest correct slice): output_rules exemption
   removal + `hasDocsOutput` + `validate_docs.go` collapse + hook/nudge text.
   Binary stays fully functional; old commands still present.
2. **S2 — `centinela docs context`**: `internal/docsctx` + cobra wiring;
   dogfood with /tmp binary.
3. **S3 — deletion**: docgen, docs generate/validate, merge regen,
   `docs/project-docs/`, wholesale-obsolete test files.
4. **S4 — prose**: prompt rewrite + mirrors + goldens + CLAUDE/README/HOWTO/
   guides.
5. **Tests step**: rewrites per plan table, new unit/acceptance tests,
   legacy-spec scenario cleanup.
6. **Can wait**: docstring-gate (next roadmap feature), README drift
   automation, doc-site pipeline, migrate-time pruning of stale
   `docs/project-docs/` downstream (deferred, see below).

#### Behavior Summary

After this feature, completing the docs step for a user-facing feature
requires a non-empty `.workflow/<feature>-changelog.md` plus
documentation-specialist evidence whose `outputs` include at least one
existing file under `docs/` or `README.md` — the role loses its exemption
from the real-files rule, so the step can only pass when human-readable
markdown was actually updated. Internal features keep the one-line changelog
contract unchanged. The specialist gets its inputs from the new
`centinela docs context <feature>`, which prints the feature brief, plan,
spec scenarios, and changelog draft as plain markdown (erroring with all
missing required inputs at once), replacing the mandated repo crawl. The
entire HTML pipeline is gone: `centinela docs generate` and
`centinela docs validate` are unknown commands, `internal/docgen` is deleted,
merge no longer regenerates a portal, and `docs/project-docs/` disappears
from the repo.

#### Gherkin Scenarios (referencing specs/docs-step-markdown-first.feature)

Written to `specs/docs-step-markdown-first.feature` — 10 scenarios:

1. **User-facing docs step passes with a real updated doc file** (happy path)
2. **README.md alone satisfies the real-file rule in a project without
   docs/guides** (scaffolded-project constraint)
3. **User-facing docs step fails when no output is a real doc file**
   (negative — exemption removed)
4. **Documentation-specialist outputs must be real files** (negative —
   missing path rejected)
5. **User-facing docs step fails without a changelog entry** (negative —
   strengthened gate)
6. **Internal docs step keeps the one-line changelog contract** (regression
   guard for right-size-docs-step)
7. **docs context prints the curated feature-scale inputs** (happy path)
8. **docs context reports every missing required input at once** (negative)
9. **docs context treats the changelog draft as optional** (edge)
10. **The HTML pipeline is gone** (removal proof: unknown command + no merge
    regen)

#### UX States (table)

| Surface | Loading | Empty | Error | Success |
|---|---|---|---|---|
| `centinela docs context <f>` | n/a (instant, local reads) | Changelog section prints hint: "no changelog draft yet — run: centinela artifact new <f> changelog" (exit 0) | One aggregated error naming ALL missing required files (brief/plan/spec), non-zero exit; invalid slug error | Plain markdown to stdout: `# Docs Context: <f>` + Feature brief / Plan / Spec scenarios / Changelog draft sections with `> source:` lines |
| Docs-step gate (`centinela complete` / `validate`) | n/a | — | "documentation-specialist outputs must include a real updated file under docs/ or README.md"; changelog missing/empty errors name the path and the `artifact new` fix | Step advances; auto-commit as usual |
| Hook nudges (context/statusline) | n/a | — | `MISSING_DOCS_OUTPUT` → run-documentation-specialist (user-facing, keyed on evidence file) or write-changelog | "REVIEW REQUIRED" pause panel unchanged |
| `centinela docs generate` / `docs validate` | n/a | n/a | Cobra unknown-command error (intentional) | n/a |

#### Out-of-Scope

- Docstring enforcement (`centinela docs lint`, changed-files ratchet) —
  follow-up feature `docstring-gate` (pre-agreed, already on roadmap).
- README drift automation (deterministic README-vs-CLI check) — pre-agreed
  exclusion.
- Docusaurus/godoc site generation — pre-agreed exclusion.
- Pruning stale generated `docs/project-docs/` output from DOWNSTREAM
  projects via `centinela migrate` (new discovery — deferred below; this
  feature only deletes the output in Centinela's own repo).
- Any change to the internal-feature changelog contract or to
  `IsUserFacingFeature` surface detection.
- Semantic quality checks on the updated docs (the gate checks real files,
  not prose quality).

#### Deferred Findings

- `migrate-prune-project-docs` — recorded via `centinela roadmap defer`
  (source: docs-step-markdown-first/planner): `centinela migrate` leaves
  stale generated `docs/project-docs/` output in downstream projects after
  the HTML pipeline removal; add a migrate cleanup that prunes it.

#### Handoff

- **Next role:** senior-engineer.
- Follow the slice order S1→S4 in `docs/plans/docs-step-markdown-first.md`;
  the plan's file-by-file section is the authoritative inventory.
- **Outstanding questions:**
  1. Legacy-spec cleanup timing: rewriting the contradicted scenarios of the
     five legacy `.feature` files is planned for the tests step; if prewrite
     hooks block `specs/` writes there, do it at the end of the code step and
     note it in evidence.
  2. `internal/orchestration/validate_test.go` uses
     `docs/project-docs/index.html` as a fixture — it remains a legal `docs/`
     path (fixture file is created by the test), so only extend, don't
     necessarily rename; engineer's choice.
  3. Confirm no `validate.commands` entry in `centinela.toml` references
     `docs generate`/`docs validate` (grep found none; re-check before S3).
- Create `.workflow/docs-step-markdown-first-changelog.md` early
  (`centinela artifact new docs-step-markdown-first changelog`) — final
  docs-step completion requires it.
