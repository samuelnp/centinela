# Edge-Case Review: fix-setup-hook-template-detection

## Scenarios Reviewed

- `PROJECT.md.template` missing, `PROJECT.md` present, `ROADMAP.md` missing:
  roadmap guidance must still be injected.
- New repo with neither template nor project file and no `centinela.toml`:
  setup hook should remain silent.
- Repo with `centinela.toml` but missing project files:
  setup directive + panel must be shown.
- `ROADMAP.md` present but production readiness prompt missing:
  production-readiness setup guidance must be shown.

## Outcome

- Implemented checks now cover renamed-template projects.
- Added plain `CENTINELA DIRECTIVE` line before boxed panel to improve model
  compliance and reduce styling ambiguity.
