# Feature: docs-step-markdown-first

**Phase:** 13 — Lighter Centinela
**Archetype:** canonical
**Depends on:** none (blocks: docstring-gate)

## Problem

The docs step's expensive path pays a fresh-context documentation-specialist
to crawl the repo (PROJECT.md, ROADMAP.md, all briefs/plans/specs, every
`kb/*.md`) and feed a deterministic HTML generator (`internal/docgen`) whose
portal output nobody reads. The genuinely useful human documentation already
lives in `docs/guides/` and README.md, which the step never touches. Cost:
~40–70K tokens per user-facing feature, growing with project size, for
artifacts (KB html, portal index.html) with no readers.

## Expected outcome

The docs step produces documentation humans actually read, at feature-scale
token cost:

1. **HTML pipeline deleted** — `internal/docgen`, `centinela docs generate`,
   the `kb/<feature>.md → .html` contract, the `docs/project-docs/index.html`
   gate, and merge-time portal regeneration are removed. Generated
   `docs/project-docs/` output is deleted from the repo.
2. **Gate on real doc updates** — for user-facing features,
   `validateDocsOutput` requires the documentation-specialist evidence
   `outputs` to include at least one real updated file under `docs/` or
   `README.md` (the role loses its exemption from the outputs-must-be-real-
   files rule). Internal features keep the one-line changelog contract
   unchanged.
3. **Curated context** — new `centinela docs context <feature>` prints the
   feature-scale inputs the specialist needs (feature brief, plan, spec
   scenarios, changelog draft) so the prompt stops mandating a repo crawl.
4. **Prompt rewritten** — documentation-generator-prompt.md (and its scaffold
   mirror + harness golden fixtures) redefines the role's duties: update
   `docs/guides/` and README.md where the feature changed behavior, CLI
   surface, or config; write the changelog; no KB, no portal, no "narrative
   synthesis".

## Out of scope

- Docstring enforcement (`centinela docs lint`, changed-files ratchet) —
  separate follow-up feature `docstring-gate`.
- README drift automation (deterministic README-vs-CLI check).
- Docusaurus/godoc site generation — the redesigned step produces the
  markdown + (later) docstrings such a pipeline would consume; the pipeline
  itself is not built here.

## Constraints

- No gate weakened: the step must still refuse to complete without evidence
  of real documentation work (real-file outputs replace the html-exists
  check, which was weaker).
- Scaffold mirror parity: every edited `docs/architecture/` prompt must be
  updated in lockstep with `internal/scaffold/assets/`, and harness golden
  fixtures (opencode/codex AGENTS.md) regenerated.
- Scaffolded projects with no `docs/guides/` must still pass: the real-file
  outputs rule accepts any path under `docs/` or README.md.
