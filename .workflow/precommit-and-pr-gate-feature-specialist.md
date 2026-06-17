# precommit-and-pr-gate — feature-specialist

## Behavior Summary

`centinela precommit` gates the staged changes and exits non-zero (blocking the
commit) on a fail-severity gate; fast (skips cross-compile by default). An
installer wires an idempotent, non-clobbering `.git/hooks/pre-commit`.
`centinela pr-gate` gates the PR's changed files, emits a deterministic Markdown
verdict + pass/fail exit; CI posts/updates one marker PR comment.

## Acceptance Criteria (Gherkin)

`specs/precommit-and-pr-gate.feature` — 15 scenarios, 1:1 mapped to Go acceptance
tests via `// Scenario: <name>`, modeled on `specs/custom-gate-sdk.feature`
(narrative Feature, Background, comment block, exit-code + determinism rigor).
Uses real cases (oversized staged `.go` file for G1, `git add`).

## UX States

CLI/text + Markdown. States: precommit pass (exit 0) / fail (named gate, exit 1,
commit blocked); installer wrote/idempotent/preserved; pr-gate verdict Markdown
(pass/fail per gate, details), no-PR-context (stdout only).

## Edge Cases

Staged fail blocks + names; clean staged passes; unstaged ignored; not-a-repo /
nothing staged → clean exit 0; build gate skipped by default; warn non-blocking;
installer executable / idempotent / preserves existing hook; uninstall removes
only its block; pr-gate Markdown + exit; pr-gate outside PR → stdout no post;
`fail_on_warning`; custom+audit gates participate; determinism. (14 in evidence.)

## Out-of-Scope

`--post` (CI posts via `gh`); direct GitHub API; doctor hook-status integration;
non-GitHub hosts; rewriting built-in gates.

## Handoff

→ senior-engineer. Implement per `docs/plans/precommit-and-pr-gate.md`;
staged-diff API, skip-build cfg copy, marker-delimited installer, markdown
renderer, and CI snippet are fixed there. Scaffold-assets mirror is a no-op.
