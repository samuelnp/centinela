# Big-Thinker Report — delivery-artifact-generation

## Problem

`centinela deliver --via pr` opens a PR with `gh pr create --fill` (scraped
commit subjects) and leaves `CHANGELOG.md` hand-written or empty. Centinela
already holds rich, structured, freshest-at-delivery evidence (brief, plan,
per-role orchestration evidence, gatekeeper report, claim-verification results)
but discards it at exactly the moment it is most useful. This feature composes a
PR body and a Keep-a-Changelog line from that evidence, read-only and with
graceful degradation, without changing *when/whether* delivery happens.

## Scope

**In:**
- New read-only aggregator package `internal/delivery` that composes (pure, no
  writes): a Markdown PR body (summary, what/why, acceptance reference, gate
  status, provenance footer) and a single Keep-a-Changelog entry.
- `deliver --via pr` uses the composed body via `gh pr create --body-file`
  (drops `--fill`).
- Idempotent insertion of one changelog line under the correct
  `### Added`/`### Changed`/`### Fixed` of the `## [Unreleased]` block.
- Section-by-section graceful degradation; never fabricates a gate result; never
  blocks delivery.
- PROJECT.md G2 paragraph + `centinela.toml` import_graph mapping for the new
  aggregator package; cmd/ stays a thin orchestrator (G7).

**Out:**
- Changing the delivery decision / `--via` matrix (owned by
  `completion-delivery-prompt`).
- Non-GitHub PR creation (still `gh`-specific).
- Version bump / tag / GitHub Release (owned by `automate-semver-release`).
- Multi-feature aggregate release notes (`team-dashboard` territory).
- Editing commit history / squashing.

## Dependencies & Assumptions

- Builds directly on `completion-delivery-prompt` (`cmd/centinela/deliver_pr.go`,
  `internal/gitutil`). The `gitDeliver`/`ghAvailable`/`ghCreatePR` seams are
  preserved and only minimally extended.
- Reads via `evidence.Read` (JSON) + `evidence.ReadCompanion` (`.md`),
  `verify.VerificationResult` (consumed, not run, inside the package), and direct
  `os.ReadFile` for brief/plan/CHANGELOG.
- Assumes `CHANGELOG.md` keeps the Keep-a-Changelog shape with a
  `## [Unreleased]` block and `### Added/Changed/Fixed` subsections (verified on
  disk).
- Assumes `gh pr create --body-file` is available in all supported `gh` versions.
- The aggregator layer's import_graph `allow` already includes `aggregator`
  (from `brownmap`), so aggregator→domain edges are permitted.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|------------|
| Regressing the existing deliver/push/gh flow | High | Medium | Preserve seams; only extend `ghCreatePR` with a body path; reuse + extend existing deliver tests. |
| `gh --body-file` portability (temp file, quoting) | Medium | Low | `os.CreateTemp` + absolute path + `defer Remove`; exec args, no shell. |
| CHANGELOG idempotency / duplicate/misplaced lines | Medium | Medium | Pure `InsertEntry` transform, dedupe by normalized bullet scoped to `[Unreleased]`; golden tests. |
| import_graph gate failure for new package | Medium | Medium | Land G2 + `centinela.toml` mapping in Slice A; `centinela validate` per slice. |
| Brief/plan heading drift breaks extraction | Low | Medium | Tolerant heading scan: missing heading → omit, never error. |
| Slow/fragile double test run at delivery | Medium | Medium | Drive gate status from static gatekeeper report; verification tally optional behind short timeout, omit on error. |

## Rollout

1. **Slice A — changelog only:** `changelog.go` + `changelog_insert.go` (pure,
   unit-tested) + cmd `writeChangelog` wired before push; G2/import_graph mapping.
2. **Slice B — PR body:** `delivery.go`/`prbody.go`/`sections.go`/`extract.go`;
   wire `buildPRBody` + `gh --body-file`, drop `--fill`.
3. **Slice C — gate-status + degradation polish + acceptance binary test** (local
   bare `origin`, faked `gh`). Each slice ships green.

## Deferred Findings

- `centinela-changelog-subcommand` — a standalone `centinela changelog <feature>`
  to (re)generate/insert the line independent of PR delivery, giving the merge
  path changelog parity. Out of scope here; captured for later.

## Handoff

To **feature-specialist**: produce the file-by-file implementation plan and
`specs/delivery-artifact-generation.feature`. Enforce: pure composer (zero I/O in
`internal/delivery`); omit-over-fabricate for gate status; changelog idempotency
scoped to `[Unreleased]`; G2 + import_graph mapping in Slice A; local bare
`origin` for the acceptance test (no real network push).
