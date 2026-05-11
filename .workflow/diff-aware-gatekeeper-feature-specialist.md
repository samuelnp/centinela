# Orchestration Evidence: feature-specialist

- Feature: `diff-aware-gatekeeper`
- Step: `plan`
- Outcome: Authored the Gherkin acceptance criteria covering the
  twelve observable behaviors that v1 must guarantee: auto-default
  per environment, CI-forces-full, branch-introduced violations
  reported, full-mode still reports historical violations, untracked
  files included, configurable diff base, i18n short-circuit when no
  locale file changed, i18n still runs when one did, graceful degrade
  in non-git directories, flag overrides (`--changed` / `--full`),
  mutually exclusive flag rejection, and user-command pass-through.
  Each scenario is phrased against the validate output header and gate
  report lines so the tests step can assert deterministically without
  coupling to specific prose. Specified the new TOML keys
  (`diff_mode`, `diff_base`) and their normalization rules so config
  edge cases are explicit. Specified that the validate output header
  must always indicate which mode ran, so users cannot be confused
  about whether they got a strict or scoped scan.
- Inputs: `docs/features/diff-aware-gatekeeper.md`,
  `docs/plans/diff-aware-gatekeeper.md`,
  `internal/gates/gates.go`, `internal/gates/file_size.go`,
  `internal/gates/i18n_keys.go`, `cmd/centinela/validate.go`,
  `internal/config/config.go`, `.workflow/diff-aware-gatekeeper-big-thinker.md`.
- Outputs: `specs/diff-aware-gatekeeper.feature`,
  `docs/plans/diff-aware-gatekeeper.md`.
- Handoff: `senior-engineer`
