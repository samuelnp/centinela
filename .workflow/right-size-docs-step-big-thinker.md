### Big-Thinker Report: right-size-docs-step
**Date:** 2026-06-12

#### Problem
Every feature's `docs` step demands the same heavy output regardless of reader value:
a plain-language KB guide (`kb/<feature>.md` + `.html`), a full ~130 KB `index.html`
portal regeneration, and a documentation-specialist subagent evidence ceremony. That
is correct for a user-facing feature with an end-user story but pure waste for an
internal refactor, bugfix, or chore. The fix makes the `docs` step surface-aware,
exactly mirroring how the `code` step already gates the `ux-ui-specialist` on
`IsUserFacingFeature`: user-facing keeps the full flow; internal needs only a one-line
changelog, and portal regeneration moves to merge time.

#### Scope
- **In:** surface-conditional `RequiredRolesForFeature("docs")` (policy.go);
  surface-conditional `validateDocsOutput` (validate_docs.go, may import orchestration);
  `.workflow/<feature>-changelog.md` artifact (new `changelog` ArtifactKind + template);
  best-effort merge-time portal regen in `runMerge` (merge.go) via `docgen.Generate`;
  surface-awareness for the two docs-step hook nags (`hook_statusline_rules.go:41`,
  `hook_context.go:72`); managed prompt doc + byte-identical scaffold mirror.
- **Out:** CHANGELOG.md assembly automation (delivery-artifact-generation, Phase 10);
  new surface values; any gate or claim-verification change (docs step has no gate);
  removal of `centinela docs generate`.

#### Dependencies & Assumptions
- `orchestration.IsUserFacingFeature` is the single source of truth for surface;
  absence of `surface: user-facing` ⇒ internal, identical to the code step today.
- `internal/workflow` already imports `internal/orchestration` (validate_orchestration.go),
  so calling `IsUserFacingFeature` from `validateDocsOutput` adds no new import edge.
- `docgen.Generate(outPath, title string) error` is the regen entrypoint; it calls
  `LoadData` → `ValidateInputs`, which REQUIRES PROJECT.md, ROADMAP.md, and a non-empty
  `.workflow/roadmap.json`. Merge regen must therefore be best-effort.
- The scaffold-arch parity acceptance test (`scaffold_arch_parity_acceptance_test.go`)
  iterates EVERY `docs/architecture/*.md` and asserts byte-equality with the mirror
  (allowlist excludes only 5 unrelated docs). The generator prompt is NOT allowlisted,
  so the doc and its mirror must change together byte-for-byte.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|---|---|---|---|
| A user-facing feature silently gets the light path (lost KB) | Med | Low (forward-only; 0 current briefs declare `surface: user-facing`) | Mirrors the existing code-step surface contract; user-facing already declares the surface for ux-ui gating; surface chosen path visible in status/docs output |
| Two docs-step hook checks still hard-require index.html (status-line `MISSING_DOCS_OUTPUT`, context-banner) | Med | High if unaddressed | Make both surface-aware in the same change so internal features are not nagged for a portal they no longer need |
| Merge regen fails because PROJECT.md/ROADMAP.md/roadmap.json absent | High | Med | Best-effort: ignore the regen error, print a one-line notice, never fail an otherwise-clean merge (spec scenario 9) |
| Existing docs tests assume KB+index always required | Med | Med | Those fixtures are user-facing-shaped; keep the user-facing path byte-identical; add internal-path fixtures that declare surface explicitly |
| Scaffold-mirror byte drift on the generator prompt doc | Med | Med | Parity test covers it; edit both copies identically; run the acceptance test |

#### Rollout
- **Step 1:** Surface-conditional role gating (`RequiredRolesForFeature`, policy.go) +
  internal-vs-user-facing artifact check (`validateDocsOutput`, validate_docs.go) +
  new `changelog` ArtifactKind/template/scaffold + make the two docs-step hook nags
  surface-aware.
- **Step 2:** Best-effort merge-time portal regen in `runMerge` (merge.go) behind a
  docgen-runner seam so tests can assert it is called on a clean merge and that a regen
  failure does not fail the merge.
- **Step 3:** Update `docs/architecture/documentation-generator-prompt.md` and its
  scaffold mirror byte-for-byte (user-facing → full KB flow; internal → one-line
  changelog, skip KB/portal/evidence).
- **Step 4:** Unit (surface matrix, changelog requirement, regen seam) + integration +
  acceptance (per-scenario, dogfooding this feature's own internal light path).

#### Handoff
- **Next role:** feature-specialist
- **Outstanding questions (resolved):**
  1. **Default-surface safety (Q1): SAFE.** The code step ALREADY treats absence of
     `surface: user-facing` as internal (`RequiredRolesForFeature` line 39 gates ux-ui on
     `IsUserFacingFeature`, which returns false when the line is absent), so this is
     consistency, not invention. Survey of all 78 feature briefs: **ZERO declare a real
     `surface: user-facing` frontmatter line** (the 3 raw grep hits were prose/table text,
     including this feature's own `surface: internal` brief). 68 briefs have no surface line
     at all; 9 declare `surface: internal` explicitly. Therefore NO existing feature relies
     on the docs step while omitting the declaration — the change is purely forward-looking
     (only affects features built after it ships; already-merged docs are untouched). Future
     user-facing features must add `surface: user-facing`, which they already do for the
     code step's ux-ui gating.
  2. **Import legality (Q2): YES — the check lives in `validateDocsOutput`
     (internal/workflow/validate_docs.go) calling `orchestration.IsUserFacingFeature`.** No
     new forbidden edge or cycle: `internal/workflow` already depends on
     `internal/orchestration` (validate_orchestration.go imports it; orchestration does not
     import workflow). The role-gating half lives natively in
     `orchestration.RequiredRolesForFeature` (same package as `IsUserFacingFeature`).
  3. **index.html (Q3): not hard-required ELSEWHERE except two hook nags that must be made
     surface-aware** — `hook_statusline_rules.go:41` (returns `MISSING_DOCS_OUTPUT`) and
     `hook_context.go:72` (banner). The complete-step gate routes only through
     `validateDocsOutput`; no test always-stats index.html outside docs/orchestration tests
     that write their own fixtures. **Merge-regen feasible:** signature
     `docgen.Generate(outPath, title string) error`; call
     `docgen.Generate("docs/project-docs/index.html", "Centinela Project Docs")`. It needs
     PROJECT.md + ROADMAP.md + non-empty `.workflow/roadmap.json` (via `ValidateInputs`),
     which may be absent → confirms the best-effort path (notice + continue, never fail the
     merge).
  4. **Changelog artifact (Q4): yes, `.workflow/<feature>-changelog.md` one-liner is the
     right minimal artifact.** The scaffold mechanism to mirror is `centinela artifact new
     <feature> <kind>`: add `KindChangelog ArtifactKind = "changelog"` to
     `internal/evidence/artifact.go` (KindsAllowed), a `case KindChangelog` in
     `RenderTemplate` (artifact_templates.go) emitting `single(artifactPath(feature,
     "changelog.md"), …)`. Validation = file exists AND first non-blank line is non-empty —
     sound and mechanically checkable (matches spec scenarios 6/7).
  5. **Dogfood (Q5): YES — recommend building a fresh binary at the docs step and using it
     to complete docs for this feature.** right-size-docs-step is itself `surface: internal`,
     so under its own new rules its docs step needs ONLY
     `.workflow/right-size-docs-step-changelog.md` (a one-liner, e.g.
     `refactor: make the docs step surface-aware so internal features ship a one-line
     changelog instead of the full KB/portal/evidence bundle`) — NOT the KB guide, portal,
     or documentation-specialist evidence. Safe because the freshly built binary's
     `validateDocsOutput` + `RequiredRolesForFeature("docs")` accept a changelog-only
     internal docs step (the installed binary at the repo root predates this feature and
     would still demand the full bundle). This is the real artifact this feature produces at
     its docs step. Per the project memory "Dogfood a new CLI command before the validate
     gate," build a `/tmp` binary from `./cmd/centinela` and use it for the docs-step
     `centinela complete`.

- **Spec scenarios needing adjustment:** none functionally — all 10 scenarios map cleanly to
  the design. One note for feature-specialist: scenarios 4 ("fails naming the missing
  knowledge-base guide") and 6/7 (changelog missing/blank) pin the validator's error
  message wording, so `validateDocsOutput` must name the KB guide on the user-facing path
  and name the changelog on the internal path. Scenario 8 ("portal regenerated") and 9
  ("regen failure does not fail merge") require the docgen-runner seam in `runMerge` so both
  are testable without a real roadmap.
