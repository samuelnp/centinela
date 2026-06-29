### Feature-Specialist Report: brownfield-setup-detection
**Date:** 2026-06-29

#### Behavior Summary

When Centinela's setup hook fires on `UserPromptSubmit` and `PROJECT.md` is absent, it currently
unconditionally emits the GREENFIELD directive, which cold-interrogates the user with six questions
and ignores any existing source. The fix adds a cheap, root-only detector `analyze.HasSource(root
string) bool` that inspects only the project root — checking for known manifests (go.mod,
package.json, Cargo.toml, Gemfile, pyproject.toml, requirements.txt, Makefile) and/or non-empty
source directories (src, app, lib, cmd, pkg, internal) — then routes accordingly: brownfield repos
receive a new BROWNFIELD directive via `ui.RenderBrownfieldSetupNeeded()` that instructs the agent
to run `centinela analyze` then `centinela synthesize`, enrich the generated draft by reading design
docs, manifests, and i18n files, set `**Project Stage:** existing`, and confirm uncertain fields
with the user before finalizing. Greenfield/empty repos continue through the existing question-based
path unchanged. The detector is forbidden from tree-walking or depending on
`.workflow/analysis.json` (absent on first prompt), keeping it O(1) cost per hook invocation.

#### Acceptance Criteria (Gherkin)

See `specs/brownfield-setup-detection.feature` for the full spec. Scenarios at a glance:

1. Brownfield repo with go.mod emits the brownfield directive (happy path, manifest signal)
2. Brownfield repo with package.json emits the brownfield directive (second manifest variant)
3. Brownfield repo with only a Makefile is detected as brownfield (Makefile-only edge case)
4. Brownfield repo with a populated src/ directory is detected as brownfield (source-dir signal)
5. Greenfield empty repo still emits the existing question-based setup directive (negative path)
6. Empty src/ directory is NOT a brownfield signal (false-positive guard)
7. PROJECT.md already present bypasses both setup directives (unchanged behavior guard)
8. Brownfield directive instructs enrich-then-confirm workflow (directive text content check)
9. Brownfield repo with populated internal/ directory is detected as brownfield (Go-specific)
10. HasSource detector does not walk subdirectories (cost/depth guard)

#### UX States

| State        | Trigger                                       | Surface                                                    |
|--------------|-----------------------------------------------|------------------------------------------------------------|
| n/a          | No loading state; hook is synchronous         | n/a                                                        |
| empty        | No project files detected (true greenfield)   | GREENFIELD directive + six-question prompt in hook output  |
| success      | Manifest or non-empty source dir at root      | BROWNFIELD directive panel via RenderBrownfieldSetupNeeded |
| pass-through | PROJECT.md present                            | No setup directive; hook falls through to roadmap checks   |

#### Edge Cases

- Makefile-only repo detected as brownfield (manifest signal, no source dirs)
- Empty src/ is not a brownfield signal (directory signal requires at least one non-hidden entry)
- PROJECT.md present → no setup directive emitted (hook proceeds to roadmap checks)
- Greenfield empty repo still gets the question-based setup prompt unchanged
- Deep nested source with no root-level signal → greenfield (detector is root-only, no tree-walk)
- Brownfield directive text must NOT contain "Do not answer the user" (greenfield wording)
- synthesize-written PROJECT.md must carry `Project Stage: existing` so projectstage.Parse
  returns Existing and bootstrap is skipped
- Detector cannot depend on .workflow/analysis.json (absent on first prompt of a brownfield repo)

#### Out-of-Scope

- Brownfield onboarding documentation in `docs/architecture/new-project-guide.md` — deliberate
  v1 exclusion; recorded as `brownfield-onboarding-docs` in the deferred roadmap.
- Modifying `centinela analyze` or `centinela synthesize` engines — reused as-is.
- Deep tree-walk or reading file contents during detection — forbidden for cost reasons.
- Extending the manifest list to cover pom.xml, composer.json, build.gradle, etc. — acceptable
  v1 false-negative; source-dir signal catches most such repos. Recorded as deferred breadth.
- Any UI rendering beyond the system panel pattern already established by RenderSetupNeeded.

#### Deferred Findings

- `brownfield-onboarding-docs` — pre-agreed exclusion already recorded by the big-thinker.
- `brownfield-manifest-breadth` — NEW discovery: extending manifest detection to additional
  ecosystems (Maven/pom.xml, Composer/composer.json, Gradle/build.gradle) for broader coverage.

#### Test-Mapping Notes

- Acceptance tests (rollout Step 4) MUST drive the installed binary against a temp directory to
  assert directive output — not stub `ui.RenderBrownfieldSetupNeeded`.
- Scenarios 5 and 6 are regression guards; they must be exercised even if all other scenarios
  pass, because the greenfield path is "byte-for-byte unchanged by design".
- Scenario 7 (PROJECT.md present) maps to the early-return guard already in hook_setup.go and
  requires no new code — it is a non-regression check.
- Scenario 10 (depth guard) can be implemented as a unit test for `analyze.HasSource` by
  creating a nested dir tree with source files two levels deep and no root-level signal.
- QA should assert a parity test: the manifest set used by HasSource exactly matches the keys
  from `internal/analyze/manifests.go:manifestTable` to prevent drift.

#### Handoff

- Next role: senior-engineer
- Open clarifications:
  1. Resolved recommendation: HasSource derives manifest set from manifestTable keys directly
     (or guarded by a parity test), not a mirrored standalone list.
  2. Resolved recommendation: "populated" source dir means any non-hidden entry (no extension
     filtering needed); keep it cheap.
  3. The BROWNFIELD directive's first line must NOT replicate the greenfield "Do not answer the
     user's message" wording — it must read as an invitation to run analyze/synthesize.
