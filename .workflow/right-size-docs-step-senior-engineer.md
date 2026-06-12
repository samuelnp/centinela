# Senior-Engineer Report: right-size-docs-step

**Date:** 2026-06-12

Made the `docs` step surface-aware, mirroring the code step's ux-ui gating on
`orchestration.IsUserFacingFeature`. User-facing features keep the full
KB+portal+documentation-specialist flow unchanged; internal features (default)
require only a one-line `.workflow/<feature>-changelog.md`. The portal is
regenerated best-effort at merge time.

## Files Touched

### Source (implementation)
- `internal/orchestration/policy.go` — `RequiredRolesForFeature`: for
  `step == "docs"` and a non-user-facing feature, return no roles (drops
  documentation-specialist). Code-step ux-ui gating untouched.
- `internal/workflow/validate_docs.go` — `validateDocsOutput` now branches on
  `orchestration.IsUserFacingFeature`: `validateDocsUserFacing` (portal + KB md +
  KB html, wording unchanged) vs `validateDocsInternal` (changelog exists +
  non-blank first line; distinct "missing" and "empty" errors naming the file).
- `internal/evidence/artifact.go` — added `KindChangelog` to `KindsAllowed`.
- `internal/evidence/artifact_templates.go` — `case KindChangelog` → `single(...)`.
- `internal/evidence/artifact_changelog.go` (new) — `changelogBody` one-line stub.
- `cmd/centinela/merge.go` — package-level `docsPortalRegen` seam + `mergePortalTitle`;
  after a clean merge, call it and on error print `notice: portal regen skipped: <err>`
  and continue (never fails the merge).
- `cmd/centinela/hook_statusline_rules.go` — docs-step `MISSING_DOCS_OUTPUT` nag is
  surface-aware: user-facing → missing portal; internal → missing changelog.
- `cmd/centinela/hook_context.go` — docs banner surface-aware: user-facing →
  `RenderDocumentationNeeded`; internal → `RenderChangelogNeeded`.
- `internal/ui/render_review.go` — new `RenderChangelogNeeded` panel.
- `docs/architecture/documentation-generator-prompt.md` + its scaffold mirror
  `internal/scaffold/assets/...` — added a "Surface-aware docs step" section,
  edited byte-for-byte identically.

### Tests adjusted (fixture intent preserved — NO assertion weakened)
Each fixture below was always meant to exercise the *user-facing* docs path but
declared no surface, so under the new default (internal) it took the light path.
Fix in every case: declare `surface: user-facing` in the fixture brief.
- `internal/orchestration/validate_test.go` — `TestRequiredRolesAndValidateStep`:
  write `docs/features/f.md` with `surface: user-facing` up front so the docs step
  still requires documentation-specialist evidence.
- `internal/workflow/validate_docs_test.go` — `TestValidateDocsOutput`: user-facing
  brief so the KB+portal contract is asserted.
- `internal/workflow/validate_orchestration_docs_test.go` —
  `TestValidateArtifactsDocsStrictOrchestration`: user-facing brief.
- `internal/workflow/steps_test.go` — `TestCompleteTransitionsToDone`: brief now
  `# b\nsurface: user-facing` so the existing full KB bundle satisfies docs.
- `cmd/centinela/hook_context_docs_test.go` — `TestRunHookContextDocsReminder`:
  user-facing brief so the portal banner still fires.
- `cmd/centinela/hook_orchestration_docs_test.go` —
  `TestRunHookOrchestrationIncludesDocsRole`: user-facing brief so the docs role
  directive still appears.
- `cmd/centinela/hook_statusline_rules_test.go` — `TestStatuslineRulesDocs`:
  user-facing brief so the portal→evidence gate sequence is exercised.

## Architecture Compliance

- **Boundary checks (G2):** `internal/workflow` → `internal/orchestration` is a
  pre-existing legal edge (validate_orchestration.go). `cmd` may import
  `internal/docgen`, `internal/orchestration`, `internal/ui` (all already used).
  `internal/orchestration` imports nothing new. `internal/evidence` edges unchanged.
  No cycles.
- **G1 line counts (all ≤100):** policy.go 49, validate_docs.go 58, artifact.go 62,
  artifact_templates.go 36, artifact_changelog.go 8, merge.go 74,
  hook_statusline_rules.go 65, hook_context.go 98, render_review.go 54.
- **UNTOUCHED (confirmed):** `internal/verify/*`, `internal/gates/*`, the
  `complete.go` ship gate, and the code-step ux-ui gating (only ADDED a docs branch
  beside it in `RequiredRolesForFeature`).
- **Prompt-doc parity:** `diff docs/architecture/documentation-generator-prompt.md
  internal/scaffold/assets/docs/architecture/documentation-generator-prompt.md` is
  EMPTY (Files are identical) — scaffold-arch parity acceptance test holds.

## Type-Safety Notes
- No `interface{}`/`any`. Error wrapping per house style; changelog/portal-notice
  messages name the exact path and remediation command.
- `docsPortalRegen` is a typed `func() error` package var (seam), default-wired to
  `docgen.Generate`.

## Trade-Offs
- Added `RenderChangelogNeeded` (one small UI panel) rather than overloading the
  existing portal banner, to keep messages accurate for the internal path.
- Merge regen uses a fixed title constant (`mergePortalTitle = "Centinela Project
  Documentation"`), matching `docs generate`'s default.
- Default surface = internal is forward-only; 0 existing briefs declare
  `surface: user-facing` (per big-thinker survey), so no merged docs are affected.

## Handoff → qa-senior

Tests to author (colocated unit + integration + acceptance), assertions to make:

1. **Surface matrix — `RequiredRolesForFeature("docs")`:** user-facing brief →
   includes `RoleDocsSpecialist`; internal/no-surface brief → excludes it (empty).
   Code-step ux-ui gating still: code + user-facing → includes ux-ui; code +
   internal → not.
2. **`validateDocsOutput` user-facing:** passes with portal+kb.md+kb.html; fails
   naming the KB guide when kb.md absent (and KB page when kb.html absent).
3. **`validateDocsOutput` internal — changelog missing:** fails with error naming
   the changelog file (the "missing" branch).
4. **`validateDocsOutput` internal — changelog blank:** file present but
   whitespace-only → fails with the distinct "empty" error (separate from missing).
5. **`validateDocsOutput` internal — happy path:** non-blank changelog present and
   NO KB → passes.
6. **Merge-regen seam (merge.go):** swap `docsPortalRegen` with a spy; assert it is
   invoked on a clean merge (scenario 8); assert a regen error prints the
   `notice: portal regen skipped:` line and the merge still succeeds (scenario 9).
7. **Two hook nags:** `statusBlockAndNext` docs step — user-facing missing portal →
   `MISSING_DOCS_OUTPUT`/`run-documentation-specialist`; internal missing changelog
   → `MISSING_DOCS_OUTPUT`/`write-changelog`; internal with changelog present →
   advances past the nag. `runHookContext` — user-facing → `RenderDocumentationNeeded`;
   internal missing changelog → `RenderChangelogNeeded`; internal with changelog → silent.
8. **`KindChangelog` template:** `centinela artifact new <f> changelog` writes
   `.workflow/<f>-changelog.md` with a non-blank first line.
9. **Acceptance** (`tests/acceptance/right_size_docs_step_test.go`): per-scenario
   with `// Acceptance:` + `// Scenario:` comments closing spec traceability on all
   10 scenarios, including scenario 10 (default surface = internal via
   `IsUserFacingFeature` returning false on an undeclared brief).

**Existing fixtures already adjusted by senior-engineer (do not re-touch):** the 7
test files listed under "Tests adjusted" — they now declare `surface: user-facing`
so their original user-facing intent is preserved. qa-senior should add NEW
internal-path fixtures rather than flip these.
