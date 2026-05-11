# Orchestration Evidence: documentation-specialist

- Feature: `diff-aware-gatekeeper`
- Step: `docs`
- Outcome: Documented the user-facing surface of the feature in both
  the README and the gatekeepers reference. Mirrored the gatekeepers
  update into the scaffold asset so new projects get the same guidance
  on `centinela init`. Generated the project HTML docs to publish the
  refreshed reference.

  **README.md**
  - New "Diff-aware mode" subsection under "Gate Checks" covering the
    auto/CI defaults, the `[validate] diff_mode` / `diff_base` TOML
    keys, the `--changed` / `--full` flags and their precedence, the
    change-set construction (tracked + untracked), G1 and G11
    semantics under diff mode, the explicit out-of-scope (user
    `[validate] commands`), the degrade paths, and the
    `CI=true|1` detection caveat.
  - `centinela.toml Reference` block extended with the new
    `diff_mode` and `diff_base` keys plus inline comments pointing
    back to the Diff-aware mode section.

  **docs/architecture/gatekeepers.md** + scaffold mirror
  - Item 3 of the "Gate Enforcement" list updated to spell out which
    gates (G1, G11) honor the diff filter, which config knobs drive
    it, the per-invocation overrides, the CI/local defaults, and the
    fact that user `[validate] commands` are not scoped.
  - Same edit applied byte-identically to
    `internal/scaffold/assets/docs/architecture/gatekeepers.md` so
    `centinela init` ships the up-to-date reference. Verified via
    `diff` — only the project-specific "Preserved Custom Sections"
    block differs, as expected.

  **HTML project docs**
  - Regenerated via `centinela docs generate` into
    `docs/project-docs/index.html` so the docs index reflects the
    new feature brief and plan. Validated via
    `centinela docs validate` before generation.

  **Out of scope here** (already addressed in earlier steps):
  - Plan file (`docs/plans/diff-aware-gatekeeper.md`) and feature
    brief (`docs/features/diff-aware-gatekeeper.md`) were written
    at plan step and remain canonical.
  - Edge-cases catalog and qa-senior evidence live under
    `.workflow/`.
  - Gherkin spec at `specs/diff-aware-gatekeeper.feature`.

- Inputs: `README.md`,
  `docs/architecture/gatekeepers.md`,
  `internal/scaffold/assets/docs/architecture/gatekeepers.md`,
  `docs/features/diff-aware-gatekeeper.md`,
  `docs/plans/diff-aware-gatekeeper.md`,
  `specs/diff-aware-gatekeeper.feature`,
  `.workflow/diff-aware-gatekeeper-validation-specialist.md`.
- Outputs:
  `README.md` (edited — Diff-aware mode subsection + toml reference),
  `docs/architecture/gatekeepers.md` (edited — gate-enforcement item 3),
  `internal/scaffold/assets/docs/architecture/gatekeepers.md`
    (edited — same scaffold mirror),
  `docs/project-docs/index.html` (regenerated).
- Handoff: feature complete.
