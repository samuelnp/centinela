# delivery-artifact-generation

## Problem

`completion-delivery-prompt` bridged completion to delivery: `centinela deliver
<feature> --via pr|merge` now pushes the branch and opens a PR. But the PR body
is whatever `gh pr create --fill` scrapes from commit subjects, and the
`CHANGELOG.md` entry is still hand-written or skipped. Meanwhile Centinela
already holds rich, structured evidence for the feature — the feature brief, the
plan, the per-role orchestration evidence, the gatekeeper report, and the
claim-verification results. That evidence is the best possible source for a PR
description and a release-notes line, yet today it is left on the floor and the
delivery artifacts are reconstructed from scratch (inconsistently, or not at
all).

## Who / Why

**Who.** The developer/operator (and the orchestrating agent) delivering a
completed feature from a Centinela-governed worktree, who must produce a PR
description reviewers can trust and a changelog entry the release flow can ship.

**Why.** The moment of delivery is exactly when all the evidence exists and is
freshest. Composing the PR body and the `CHANGELOG` entry *from that evidence*
makes delivery output consistent, traceable back to the plan/gates, and free —
instead of a manual rewrite that drifts from what was actually planned and
verified.

## In Scope

- A read-only composer (new aggregator package `internal/delivery`) that reads
  the evidence Centinela already holds for a feature and renders two artifacts:
  1. **PR body** (Markdown): summary from the feature brief + plan, a
     "What changed / Why" section, an acceptance/spec reference, and a gate
     status line drawn from the gatekeeper report and claim-verification
     results. Always ends with a Centinela provenance footer.
  2. **Changelog entry**: a single Keep-a-Changelog line placed under the
     correct `### Added`/`### Changed`/`### Fixed` subsection of the
     `## [Unreleased]` block in `CHANGELOG.md`. Seeded from the
     `.workflow/<feature>-changelog.md` artifact when present, otherwise
     derived from the brief.
- `centinela deliver <feature> --via pr` uses the composed PR body
  (`gh pr create --body-file …`) instead of `--fill`.
- A way to write/update the changelog entry into `CHANGELOG.md` at delivery
  (idempotent: re-running does not duplicate the line).
- Graceful degradation: when an evidence source is missing, the corresponding
  section is omitted (or marked unknown) and composition still succeeds — it
  never blocks delivery or fabricates a gate result it cannot source.

## Out of Scope

- Changing *when/whether* delivery happens or the `--via` matrix
  (`completion-delivery-prompt` owns that; this only enriches the artifacts).
- Native PR creation for non-GitHub remotes — still `gh`-specific; no body
  composition where there is no PR to open.
- Version bumping / tagging / GitHub Release publishing
  (`automate-semver-release` and the Release workflow own those).
- Multi-feature / aggregate release notes spanning several features
  (`team-dashboard` territory) — one feature's artifacts only.
- Editing commit history or squashing.

## Acceptance Summary

- `centinela deliver <feature> --via pr` opens the PR with a composed,
  evidence-sourced body (summary, what/why, acceptance reference, gate status,
  provenance footer) — not the raw `--fill` commit dump.
- The composed body sources its gate-status line from the gatekeeper report and
  claim-verification results; when those are absent the line is omitted, not
  faked.
- Delivery writes one changelog line under the correct Keep-a-Changelog
  category in the `[Unreleased]` block; re-running delivery does not duplicate
  it.
- With a missing evidence source, composition degrades section-by-section and
  still produces a usable artifact rather than erroring.
- The composer is read-only and lives in `internal/delivery` (aggregator
  layer); `cmd/` stays a thin orchestrator (G7).
