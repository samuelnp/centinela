# Feature-Specialist Report — delivery-artifact-generation

## Behavior Summary

`centinela deliver <feature> --via pr` is enriched to compose its delivery
artifacts from the structured evidence Centinela already holds, instead of
scraping commit subjects (`gh pr create --fill`) and leaving `CHANGELOG.md`
hand-written.

Two artifacts are composed by a new read-only aggregator package
`internal/delivery` (pure, no I/O):

1. **PR body** (Markdown), passed via `gh pr create --body-file`: an ordered set
   of sections — summary (from brief `## Problem` / plan `Problem & Goal`),
   "What changed / Why" (brief `## Who / Why` + plan summary), an acceptance
   reference (brief `## Acceptance Summary` + a pointer to
   `specs/<feature>.feature`), a gate-status line (gatekeeper verdict +
   optional verification tally), and a fixed Centinela provenance footer.
2. **Changelog line**: one Keep-a-Changelog bullet inserted idempotently under
   the correct `### Added`/`### Changed`/`### Fixed` of the `## [Unreleased]`
   block of `CHANGELOG.md`.

Composition is read-only, degrades section-by-section when a source is missing,
never fabricates a gate result, and never blocks or changes *whether* delivery
happens. The delivery decision and `--via` matrix are unchanged
(`completion-delivery-prompt` owns those). `cmd/` stays a thin orchestrator (G7):
all file reads/writes and the temp-file lifecycle live in `cmd/`; the package is
given text and returns text.

## Acceptance Criteria (Gherkin)

Authored in `specs/delivery-artifact-generation.feature`. Each scenario maps to
executable assertions over: terminal output, the `gh pr create` invocation args
(`--body-file` present, `--fill` absent), the body-file contents, and
`CHANGELOG.md` on disk.

| Scenario | Guarantees |
|----------|------------|
| PR delivery composes the body from evidence | `--body-file` not `--fill`; summary / what-why / acceptance-ref / gate-status / provenance footer all present; pushes; exit 0 |
| Inserts exactly one Keep-a-Changelog line under correct subsection | one bullet under `### Added`; no other subsection touched; nothing outside `[Unreleased]` modified |
| Re-running delivery does not duplicate the line | idempotent; `CHANGELOG.md` unchanged on second run; exit 0 |
| feat-shaped -> Added | category selection |
| fix-shaped -> Fixed | category selection |
| other shape -> Changed | category selection default |
| Missing gatekeeper report omits its section | graceful degradation; footer still present; delivery succeeds |
| Gate status never faked | no passing gate claimed when none can be sourced; line omitted |
| Brief and plan both absent | summary/what-why/acceptance omitted; footer present; PR still opened |
| No origin remote | PR delivery refused; nothing pushed; no PR; non-zero exit |
| gh absent/unauthenticated | branch pushed; honest manual instructions; no PR claimed; non-zero exit |

## UX States

Centinela is a CLI; the surfaces are terminal output, the composed PR body, and
`CHANGELOG.md`. There is no graphical UI, so visual/interaction states are n/a.

| State | Terminal output | PR body | CHANGELOG.md |
|-------|-----------------|---------|--------------|
| Happy path (full evidence, gh present) | push + opened-PR URL; exit 0 | all sections + footer | one new `[Unreleased]` bullet |
| Re-run (idempotent) | push + PR; exit 0 | composed fresh | unchanged (no duplicate) |
| Missing one evidence source | push + PR; exit 0 | that section omitted, footer present | unchanged behavior |
| No passing gate sourceable | push + PR; exit 0 | gate-status line omitted | unchanged behavior |
| No origin remote | refusal message; exit non-zero | not built | untouched |
| gh absent | push + honest manual PR instructions; exit non-zero, no PR claimed | composed body still written to file | one new `[Unreleased]` bullet (written before push) |
| Loading / empty / focus / hover / disabled (graphical) | n/a | n/a | n/a |

## Edge Cases

See `.workflow/delivery-artifact-generation-edge-cases.md` (authored at the tests
step) for the full enumeration. The acceptance contract guarantees, at minimum:

- PR body composed via `--body-file`, never `--fill`.
- Exactly one changelog bullet per delivery, idempotent on re-run, scoped to
  `[Unreleased]`; released sections untouched.
- Category mapping: `feat:` -> Added, `fix:` -> Fixed, else -> Changed.
- Section-by-section omission on any missing source; provenance footer always
  present.
- Gate status omitted, never fabricated, when no verdict/tally can be sourced.
- Brief+plan both absent still yields a usable PR (footer-only) and opens the PR.
- No-origin refusal and gh-absent honest-failure paths unchanged.

## Out-of-Scope

- Changing *when/whether* delivery happens or the `--via` matrix
  (`completion-delivery-prompt`).
- Non-GitHub PR creation (remains `gh`-specific; no body where there is no PR).
- Version bump / tag / GitHub Release (`automate-semver-release`).
- Multi-feature / aggregate release notes (`team-dashboard` territory).
- Editing commit history or squashing.

Deferred findings: none newly discovered. The big-thinker already deferred
`centinela-changelog-subcommand`; not re-deferred here.

## Handoff

To **senior-engineer**. Implement per `docs/plans/delivery-artifact-generation.md`.
Hold the line on:

- **Pure composer**: zero I/O in `internal/delivery`; it receives file bodies as
  strings and returns strings. All reads/writes and the temp-file lifecycle stay
  in `cmd/centinela/deliver_artifacts.go` (G7 thin orchestrator, every file
  <= 100 lines).
- **Omit over fabricate** for gate status: drive it primarily from the static
  gatekeeper-report verdict; make any verification tally optional behind a short
  timeout and omit on error — never print a passing gate it cannot source.
- **Changelog idempotency** scoped to `[Unreleased]`: dedupe by normalized bullet
  text; never touch released sections; create a missing `### <Category>` in the
  canonical `Added -> Changed -> Fixed` order.
- **Land G2 + import_graph mapping in Slice A**: PROJECT.md aggregator-layer
  paragraph for `internal/delivery` + `centinela.toml` aggregator `paths` entry,
  so the gate is green from the first slice.
- **Preserve the existing deliver seams** (`gitDeliver`/`ghAvailable`/`ghCreatePR`);
  only extend `ghCreatePR` to accept a body path and drop `--fill`. Reuse and
  extend `deliver_pr_test.go`/`deliver_test.go` so the no-origin and gh-absent
  honest-failure paths stay unchanged.
- **Acceptance test**: drive the built binary against a local bare `origin` (no
  real network push) with a faked `gh` on `PATH`; assert the create invocation
  received `--body-file` whose contents include the composed sections + footer,
  and that `CHANGELOG.md` gained exactly one `[Unreleased]` line (idempotent on
  re-run). Wire the acceptance execution into `validate.commands`.
