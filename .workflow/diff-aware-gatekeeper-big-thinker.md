# Orchestration Evidence: big-thinker

- Feature: `diff-aware-gatekeeper`
- Step: `plan`
- Outcome: Framed the problem as inner-loop pain on large repos: full
  gate scans are slow and report historical violations that are not
  the developer's to fix. Scoped v1 to file-scoped built-in gates
  (G1, G11). Chose "auto" as the default policy — diff-aware locally,
  full in CI — because it keeps the ship gate strict while making the
  dev loop fast, and respects the existing CI=true convention without
  inventing a new signal. Excluded user `[validate] commands` from the
  diff scope to keep the contract with external tooling unchanged.
  Excluded G7 because it is a manual gatekeeper-subagent review, not
  an automated check, and the subagent already operates on diff.
  Included untracked files in the change set so new violations cannot
  hide behind a missing `git add`.
- Inputs: `docs/features/diff-aware-gatekeeper.md`,
  `docs/plans/diff-aware-gatekeeper.md`,
  `specs/diff-aware-gatekeeper.feature`, the current
  `internal/gates/file_size.go` and `internal/gates/i18n_keys.go`,
  `centinela.toml`, README "Gate Checks" section.
- Outputs: `docs/features/diff-aware-gatekeeper.md`,
  `docs/plans/diff-aware-gatekeeper.md`.
- Handoff: `feature-specialist`
