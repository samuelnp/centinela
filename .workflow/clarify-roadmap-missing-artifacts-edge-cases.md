# Edge Cases: clarify-roadmap-missing-artifacts

- `ROADMAP.md` can exist while `.workflow/roadmap.json` is missing.
- `.workflow/roadmap.json` can exist but be invalid JSON or drift from `ROADMAP.md`.
- Setup should continue to analysis, quality, and production-readiness prompts only after roadmap JSON is present.
- Agents need both setup artifact templates and per-feature workflow artifact templates to avoid recovery loops.
