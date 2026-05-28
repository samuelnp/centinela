# Orchestration Evidence: documentation-specialist

- Feature: `evidence-cli`
- Step: docs
- Outcome: Authored a plain-language Knowledge Base entry (`docs/project-docs/kb/evidence-cli.md`) that describes the typed evidence CLI for end users—explaining what the `centinela evidence` and `centinela artifact new` commands do, when to use them, and how they behave across 17 scenarios (init, set, append, read, validate, repair, concurrent writes, artifact templates, atomic safety, schema tolerance, free-form fields, worktree scoping). The KB entry includes concrete command examples. Regenerated `docs/project-docs/index.html` via the deterministic generator, which picked up the new entry and produced `kb/evidence-cli.html` plus updated `kb/index.html`. All authoring followed the contract in `documentation-generator-prompt.md`: no embedded schema stubs (referenced `centinela evidence schema` instead), no `mobileFirst` field (non-UX role), inputs include all required artifact snapshots, outputs list real file paths.
- Handoff: complete
