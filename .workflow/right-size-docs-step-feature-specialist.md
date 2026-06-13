### Feature-Specialist Report: right-size-docs-step
**Date:** 2026-06-12

#### Behavior Summary
The `docs` step becomes surface-aware, mirroring the existing code-step ux-ui gating on `orchestration.IsUserFacingFeature`. A brief that declares `surface: user-facing` keeps today's full flow unchanged — knowledge-base guide (`kb/<feature>.md` + `.html`), portal `index.html`, and the documentation-specialist evidence. Anything else (default/internal, including absence of any surface line) takes a light path: the docs step requires only a one-line `.workflow/<feature>-changelog.md` and skips the KB guide, the per-feature portal regen, and the documentation-specialist evidence. Required-role gating (`RequiredRolesForFeature("docs")`) and artifact validation (`validateDocsOutput`) both branch on surface; the two docs-step hook nags (status-line `MISSING_DOCS_OUTPUT`, context banner) become surface-aware so internal features are not nagged for a portal they no longer produce. The portal stays current by regenerating best-effort at merge time, so the ~130 KB rebuild happens once per delivery rather than once per feature and never fails an otherwise-clean merge. No change to gates or claim verification (the docs step has no gate); default surface is internal; the change is forward-only.

#### Gherkin Scenarios
Reference: `specs/right-size-docs-step.feature` (10 scenarios).

1. **A user-facing feature still requires the documentation-specialist role** — Given a brief declaring `surface: user-facing`, When `RequiredRolesForFeature(feature,"docs")` resolves, Then the result includes `RoleDocsSpecialist`.
2. **An internal feature does not require the documentation-specialist role** — Given a brief with no user-facing surface, When `RequiredRolesForFeature(feature,"docs")` resolves, Then the result excludes `RoleDocsSpecialist`.
3. **A user-facing docs step still requires the knowledge-base guide** — Given a user-facing feature with `kb/<f>.md`, `kb/<f>.html`, and `index.html` present, When `validateDocsOutput` runs, Then it passes.
4. **A user-facing docs step fails without the knowledge-base guide** — Given a user-facing feature missing the KB markdown, When `validateDocsOutput` runs, Then it fails with an error naming the missing knowledge-base guide.
5. **An internal docs step passes with only a one-line changelog** — Given an internal feature with `.workflow/<f>-changelog.md` (non-blank first line) and NO KB guide, When `validateDocsOutput` runs, Then it passes.
6. **An internal docs step fails without a changelog entry** — Given an internal feature with no changelog file, When `validateDocsOutput` runs, Then it fails with an error naming the missing changelog entry.
7. **An internal docs step fails when the changelog entry is blank** — Given an internal feature whose changelog file is empty/whitespace, When `validateDocsOutput` runs, Then it fails with an error naming the changelog entry as empty (distinct from the missing-file error in scenario 6).
8. **A clean merge regenerates the documentation portal** — Given a feature is merged successfully, When the merge completes, Then the merge-time docgen seam is invoked to regenerate the portal.
9. **A portal regeneration failure does not fail a clean merge** — Given a successful merge where portal regeneration fails, When the merge completes, Then the merge still succeeds and a one-line notice is reported (observable, asserts best-effort).
10. **The default surface is internal when none is declared** — Given a brief with no surface line, When the surface is resolved via `IsUserFacingFeature`, Then it is treated as internal (returns false).

#### UX States
| State | Trigger | Surface |
|---|---|---|
| Docs validation passes (user-facing) | `validateDocsOutput`: KB md + KB html + index.html all present | user-facing |
| Docs validation fails — names KB guide | `validateDocsOutput`: KB markdown/page missing on user-facing path | user-facing |
| Docs validation passes (internal) | `validateDocsOutput`: non-blank `.workflow/<f>-changelog.md` present, KB not required | internal |
| Docs validation fails — names missing changelog | `validateDocsOutput`: changelog file absent on internal path | internal |
| Docs validation fails — changelog empty | `validateDocsOutput`: changelog present but blank/whitespace on internal path | internal |
| Merge-time portal-regen notice | `runMerge`: docgen regen fails (inputs absent) → one-line notice, merge still succeeds | n/a (merge-time, surface-independent) |
| Status-line docs nag suppressed for internal | `hook_statusline_rules.go:41`: docs step + internal feature → no `MISSING_DOCS_OUTPUT` for missing portal | internal (surface-aware) |
| Context banner docs nag suppressed for internal | `hook_context.go:72`: docs step + internal feature → no "documentation needed" banner for missing portal | internal (surface-aware) |

#### Out-of-Scope
- No full changelog automation / CHANGELOG.md assembly from per-feature entries (that is delivery-artifact-generation, Phase 10).
- No new surface values — only the existing user-facing / internal split.
- No change to gates or claim verification — the docs step has no gate.
- `centinela docs generate` is NOT removed — it still runs for user-facing features and at merge time; only its per-internal-feature obligation is dropped.

#### Handoff
- **Next role:** senior-engineer
- **Open clarifications (re-flagged for senior-engineer):**
  - The TWO extra index.html nags beyond `validateDocsOutput` MUST also be made surface-aware: `cmd/centinela/hook_statusline_rules.go:41` (`MISSING_DOCS_OUTPUT`) and `cmd/centinela/hook_context.go:72` (context banner). (Note: these live in `cmd/centinela/`, not `internal/workflow/` as the big-thinker line refs implied.)
  - The managed prompt doc `docs/architecture/documentation-generator-prompt.md` and its `internal/scaffold/assets/...` mirror MUST change byte-for-byte together — the scaffold-arch parity acceptance test covers EVERY arch doc and is not allowlisted for this prompt.
  - Merge-regen MUST be best-effort behind a docgen-runner seam: invoked on a clean merge (scenario 8), and a regen failure prints a notice and never fails the merge (scenario 9). `docgen.Generate(outPath, title)` requires PROJECT.md + ROADMAP.md + non-empty `.workflow/roadmap.json`, which may be absent.
  - New `KindChangelog ArtifactKind = "changelog"` in `internal/evidence/artifact.go` (KindsAllowed) + a `case KindChangelog` template emitting the `.workflow/<feature>-changelog.md` one-liner.
</content>
