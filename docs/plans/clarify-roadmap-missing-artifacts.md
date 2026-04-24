# Plan: Clarify Missing Roadmap Artifacts

1. Update roadmap-related command errors so missing or invalid `.workflow/roadmap.json`
   is named explicitly with a concrete recovery hint.
2. Extend setup hook guidance with a dedicated roadmap JSON panel that shows the exact
   file path and required JSON shape.
3. Add scaffolded documentation covering setup artifacts and per-feature workflow
   artifact templates, then link it from setup guidance and README.
4. Add tests for start, roadmap, hook setup, and the new artifact-template docs.
