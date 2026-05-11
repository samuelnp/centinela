# Plan: Per-Feature Knowledge Base in the Docs Step

1. Update `docs/architecture/documentation-generator-prompt.md` to make the
   documentation-specialist responsible for writing
   `docs/project-docs/kb/<feature>.md` against a fixed schema (frontmatter +
   four H2 sections: "What it does", "When you'd use it", "How it behaves",
   "Examples"). Audience: Centinela end-users, plain language, no
   Given/When/Then leakage.
2. Extend `internal/docgen/types.go` with a `KBPage` struct and a `KB`
   field on `Data`.
3. Add `loadKBPages()` in `internal/docgen/load.go` (or a sibling file if
   the 100-line budget is tight) that reads `docs/project-docs/kb/*.md`,
   parses the frontmatter block, and extracts the four H2 sections.
   Missing required sections → return an error naming the file and section.
4. Add `internal/docgen/render_kb.go` with two functions:
   `RenderKBIndex(d *Data)` (grid of cards: one per spec/feature, status badge
   from `loadStates()`, link to its page or "guide not yet written") and
   `RenderKBFeature(p KBPage)` (full page with sidebar back-link to index).
   Split into helper files if either function would push the file past 100
   lines.
5. Wire generation in `internal/docgen/generate.go`: after writing the main
   `index.html`, write `docs/project-docs/kb/index.html` and one
   `docs/project-docs/kb/<feature>.html` per `KBPage`.
6. Update `internal/docgen/render_nav.go` so the main TOC includes a
   "Knowledge Base" link to `kb/index.html`.
7. Tighten `internal/workflow/validate_docs.go`: after the existing
   `index.html` check, also require
   `docs/project-docs/kb/<feature>.md` and `docs/project-docs/kb/<feature>.html`
   for the current feature. Error messages must name the exact missing path.
8. Add unit tests under `internal/docgen/`: frontmatter parse, missing-section
   error, KB index rendering with mixed populated + placeholder cards,
   feature page rendering of all four sections.
9. Add integration test covering `centinela docs generate` end-to-end with a
   sample KB md present.
10. Add an executable acceptance harness under `tests/acceptance/` for the
    new spec scenarios.
11. Author `docs/project-docs/kb/docs-knowledge-base-pages.md` (the first real
    KB md) and run `centinela docs generate` so the workflow completes.
