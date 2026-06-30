# brownfield-setup-detection — documentation-specialist

## KB Pages

No new KB page was required for this feature. The brownfield detection behaviour is captured in the CHANGELOG and is a refinement of the existing setup hook — not a standalone user-facing command warranting a standalone KB article.

## project-docs Entries

- `docs/project-docs/index.html` — regenerated via `centinela docs generate` (52.7 KB); reflects the full current feature set including brownfield-setup-detection.

## Outcome

All documentation artifacts for the `brownfield-setup-detection` feature have been produced:

1. `.workflow/brownfield-setup-detection-changelog.md` — changelog artifact describing the feature in release-note terms.
2. `CHANGELOG.md` — new bullet appended to `## [Unreleased] / ### Added` describing `feat(setup): detect brownfield projects`.
3. `docs/project-docs/index.html` — refreshed by `centinela docs generate`; file present and up to date.
4. `.workflow/brownfield-setup-detection-documentation-specialist.md` — this report.

Notable documentation decisions:
- The greenfield path is explicitly noted as unchanged; only repos with detected source files (manifests or non-empty src/app/lib/cmd/pkg/internal dirs) trigger the new brownfield directive.
- `Project Stage: existing` is the required field value for brownfield-detected setup; documented in the changelog bullet.
- No new architecture doc was needed: the feature is a routing change inside the existing setup hook, not a new architectural layer.

