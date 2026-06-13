# Plan: right-size-docs-step

Make the `docs` step surface-aware by mirroring the existing code-step
ux-ui-specialist gating. Reuse `IsUserFacingFeature`. Internal features produce a
one-line changelog instead of the KB/portal/evidence bundle; portal regen moves
to merge time. No changes to gates or claim verification (the docs step has no
gate).

## Layer compliance (G2)

- Surface-conditional role gating in `internal/orchestration/policy.go`
  (`RequiredRolesForFeature`) — reuse `IsUserFacingFeature` (same package).
- Surface-conditional artifact check in `internal/workflow/validate_docs.go`
  (`validateDocsOutput`) — needs to call `orchestration.IsUserFacingFeature`;
  confirm internal/workflow may import internal/orchestration (it already does
  via validate_orchestration.go → orchestration). 
- Merge-time portal regen in `cmd/centinela/merge.go` calling `internal/docgen`.
- Untouched: internal/verify, internal/gates, complete.go ship gate, the code
  step gating.

## 1. Surface-conditional documentation-specialist evidence

In `RequiredRolesForFeature(feature, step)` (policy.go), mirror the code-step
pattern: for `step == "docs"`, include `RoleDocsSpecialist` ONLY when
`IsUserFacingFeature(feature)`. Internal docs step → no documentation-specialist
evidence required (the orchestration validator early-returns on empty roles).

## 2. Surface-conditional docs artifacts

In `validateDocsOutput(feature)`:
- **user-facing:** unchanged — require `docs/project-docs/index.html`,
  `kb/<feature>.md`, `kb/<feature>.html`.
- **internal:** require `.workflow/<feature>-changelog.md` (exists + non-empty,
  a one-line entry); do NOT require KB md/html or index.html.
- Keep the function small; split a `validateDocsInternal`/`validateDocsUserFacing`
  helper if needed for G1.

## 3. The internal changelog artifact

- Internal docs step requires `.workflow/<feature>-changelog.md`: a single
  human-readable line summarizing the change (e.g. `fix: …` / `refactor: …`).
- Provide a scaffold via `centinela artifact new <feature> changelog` (mirror the
  edge-cases artifact stub pattern) so it is mechanically creatable, OR document
  writing it directly. Validation: file exists and has a non-blank first line.
- (Assembling these into CHANGELOG.md is delivery-artifact-generation's job —
  out of scope. v1 just produces the per-feature one-liner.)

## 4. Merge-time portal regen

- In `runMerge` (merge.go), AFTER a successful merge + validate, regenerate the
  portal: call `docgen.Generate("docs/project-docs/index.html", <title>)`.
- Best-effort: if docgen inputs are missing (e.g. no roadmap), emit a one-line
  notice and continue — a portal-regen failure must not fail an otherwise-clean
  merge. This keeps the portal current without per-feature regen.

## 5. Docs-specialist prompt (managed doc)

- Update `docs/architecture/documentation-generator-prompt.md` (and its scaffold
  mirror `internal/scaffold/assets/...`) to say: user-facing → full KB flow;
  internal → write the one-line changelog and skip KB/portal/evidence. (Parity
  test covers this mirror — update both.)

## Key decisions (resolved here; big-thinker to confirm against code)

1. **Default surface = internal** (absence of `surface: user-facing` ⇒ internal),
   identical to the code step today. CONSEQUENCE: this relaxes the docs step for
   every feature that does NOT declare user-facing — the intended change. Risk: a
   user-facing feature that forgot the declaration gets the light path. Mitigated
   by: user-facing features already declare `surface: user-facing` for the code
   step's ux-ui gating, and status/docs output should make the chosen path
   visible. Big-thinker: confirm no existing user-facing feature relies on the
   docs step WITHOUT declaring the surface.
2. **Internal artifact = `.workflow/<feature>-changelog.md` one-liner**, not an
   append to CHANGELOG.md (fuzzy to validate). Mechanically checkable.
3. **index.html not required at the internal docs step**; merge-time regen keeps
   it fresh. Big-thinker: confirm nothing else hard-requires a per-feature
   index.html refresh.

## Test plan

- Unit (colocated, per-package):
  - `RequiredRolesForFeature`: docs + user-facing → includes RoleDocsSpecialist;
    docs + internal → excludes it; code-step ux-ui gating unchanged.
  - `validateDocsOutput`: user-facing requires KB+html+index (unchanged);
    internal requires the changelog one-liner and passes WITHOUT KB; internal
    with a blank changelog fails.
  - merge-time regen invoked (inject a docgen runner seam; assert it's called on
    a clean merge; assert a regen failure does NOT fail the merge).
- Integration (`tests/integration`): an internal feature reaches docs-done with
  only a changelog one-liner; a user-facing feature still needs the full bundle.
- Acceptance (`tests/acceptance/right_size_docs_step_test.go`): per-scenario, with
  the `// Acceptance:` + `// Scenario:` comments closing the spec-traceability
  gate on this feature's own spec.

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| A user-facing feature silently gets the light path (lost KB) | Med | Mirrors existing code-step surface contract; user-facing already declares the surface; make the chosen path visible in status/docs output. |
| Portal goes stale (internal features don't regen) | Med | Merge-time regen refreshes it once per delivery; release workflow can also regen. |
| Existing docs tests assume KB always required | Med | Those features/tests are user-facing or fixtures that declare surface; update fixtures to set surface explicitly; keep user-facing path byte-identical. |
| Merge regen failure breaks a clean merge | High | Best-effort with notice; never fail the merge on portal regen. |
| Scaffold-mirror drift on the prompt doc | Low | Update both copies; the parity test guards it. |

## Rollout

1. Surface-conditional role gating (policy.go) + the internal-vs-user-facing
   artifact check (validate_docs.go) + changelog artifact scaffold.
2. Merge-time portal regen (merge.go, best-effort).
3. Managed prompt doc update (+ scaffold mirror).
4. Tests: the surface matrix, the changelog requirement, merge-regen seam, and
   the acceptance dogfood.
