# Edge Cases: docs-step-markdown-first

## Risk Matrix

| Case | Impact | Likelihood | Why |
|---|---|---|---|
| Surface line in blockquote/bullet/bold form is not detected | High | Medium | `normalizeSurface` (internal/orchestration/feature_surface.go) only matches lines starting literally with `surface:`. Real briefs use `> surface: user-facing` (docs/features/landing-page.md) and `- surface: internal`; probe confirms `> surface: user-facing`, `- surface: ...`, `**Surface:** ...` all normalize to "". Such user-facing features silently get the internal contract — no docs-specialist evidence, the new real-file gate never applies. Pre-existing parser, but this feature makes it THE gate discriminator. |
| Legacy exempt evidence retro-gated (no docs-step contract pin) | Medium | Medium | plan/validate steps pin legacy contracts (`RequiredEvidenceRoles`); docs does not. A user-facing workflow started pre-change with docs-specialist evidence written under the old exemption (descriptive strings, portal-only outputs) fails `complete` under the new binary — late, after the full gate run. Repair path exists (edit outputs via evidence CLI) but is manual. |
| Gate satisfied by an existing untouched file (gaming) | Medium | High | `hasDocsOutput` checks existence + prefix only — no mtime/git-diff. `outputs: ["docs/plans/<feature>.md"]` passes; error text "real updated file" overclaims ("updated" is not checked). ACCEPTED RISK per plan decision 3 — mitigated by prompt duties, documented by the planned plan-file-only-pass test. |
| Symlink under docs/ pointing outside the repo | Low | Low | `os.Stat` follows symlinks, so `docs/x.md -> /etc/hosts` satisfies both the real-file and docs-prefix checks. Requires a repo write to plant the symlink — same trust class as decision 3's accepted gaming vector. |
| Path traversal `docs/../secret` | None | — | `normalizeOutputPath` runs `filepath.Clean`: `docs/../x` → `x` (no docs/ prefix, fails), `docs/../README.md` → `README.md` (legitimately passes). No escape: absolute and `../` paths can never gain the `docs/` prefix. |
| Absolute path to a genuinely updated docs file | Medium | Medium | Agents habitually emit absolute paths. `/…/worktree/docs/guides/x.md` stats fine but lacks the `docs/` prefix → gate fails despite real work, with an error implying no doc was touched. Conservative direction (no bypass) but a confusing false negative. |
| Output entry is a directory (`docs/`, `docs/guides`) | Low | Low | `missingOutputFiles` treats `info.IsDir()` as missing → "actionable outputs must be real files; missing: docs/guides". Correctly rejected; message says "missing" for a thing that exists, mildly confusing. |
| `./README.md`, padded, `docs//x.md` variants | None | — | normalization (TrimSpace, TrimPrefix `./`, Clean, ToSlash) handles all three; `readme.md` on case-insensitive APFS stats OK but fails the exact `README.md` compare — conservative fail. |
| Whitespace-only / empty changelog | None | — | `validateChangelog` scans for any non-whitespace line; empty and whitespace-only both fail with actionable errors. Covered. |
| Changelog with a single line > 64KB | Low | Low | `bufio.Scanner` default token cap; `Scan()` fails, error is ignored, function reports "changelog entry is empty" for a non-empty file. Misleading failure only. |
| `docs context` is CWD-relative | Medium | Medium | `docsctx.Load` resolves `docs/features/...` etc. against the process CWD. Run from repo root while the feature lives in `.worktrees/<f>/` (known agent trap, memory: worktree CWD) or from a subdir → all-missing error whose hint ("written in the plan step") misdiagnoses. Unreadable (EACCES) files are indistinguishable from missing. |
| Empty (0-byte) brief/plan/spec | Low | Low | `readSection` marks any readable file Present → `docs context` exits 0 rendering empty sections. Boundary between "missing" (error) and "empty" (silent pass) is undocumented. |
| Bad slug / injection into `docs context` | None | — | `worktree.ValidateFeatureSlug` (strict kebab regex) rejects `../x`, spaces, metacharacters before any path is built. Dogfooded exit 1. |
| `docs context` re-run / `artifact new changelog` re-run | None | — | context is read-only and deterministic; `WriteArtifact` refuses to overwrite without `--force` (`ErrArtifactExists`). Residual: `--force` clobbers a specialist-written changelog (known class). |
| Stray docs-specialist evidence on an internal feature | None | — | `ValidateRoles` returns early for empty role sets and only iterates required roles; internal features return no docs roles, so invalid stray evidence cannot bite (plan decision 1 confirmed in code). |
| Downstream repos still carrying docs/project-docs/ | Medium | High | Merge no longer regenerates the portal; downstream trees keep a silently-stale index.html forever. `hasImplementationOutput` deliberately keeps the exclusion so legacy senior-engineer evidence is unaffected — but in THIS repo the tree was deleted, so legacy evidence listing `docs/project-docs/index.html` would now fail as missing if ever re-validated (completed workflows are not re-validated in practice). |
| Section spoofing in `docs context` output | Low | Low | Files embed verbatim with no fencing; a brief containing `## Changelog draft` or prompt-injection text flows straight into the specialist's context. Consumer is an LLM reading trusted repo files — same trust boundary as before. |

## Missing or Weak Scenarios

- Spec has no scenario for the surface-detection formats (`>`/`-`/bold) even
  though surface parsing now decides which docs contract applies.
- No scenario covers legacy (pre-change) docs-specialist evidence meeting the
  new binary — the only migration-path state transition this feature creates.
- Spec scenario "outputs must be real files" covers a missing path but not a
  directory entry, an absolute path, or a `docs/../` traversal form.
- Changelog-empty scenario covers absence only; whitespace-only content and
  the 64KB scanner boundary are untested.
- No scenario pins `docs context` behavior for empty-but-present inputs or a
  wrong-CWD invocation.

## Proposed/Added Tests

Unit (colocated, tests step):
- internal/orchestration: table test for `hasDocsOutput`/dispatch — pass:
  `docs/guides/x.md`, `README.md`, `./README.md`, ` docs/x.md ` (padded),
  `docs/plans/<f>.md` (documents decision 3), `docs/../README.md`; fail:
  `docs/../secret.md`, absolute path to an existing docs file, `docs/guides`
  (directory → "missing"), `readme.md`, `.workflow/...`-only, empty outputs.
- internal/workflow: `validateChangelog` — missing, empty, whitespace-only,
  non-empty for BOTH user-facing and internal fixtures (pins changelog-for-all).
- internal/docsctx: each missing-required combination aggregated into ONE
  error naming all paths; absent-changelog hint; empty-but-present brief
  renders exit-0 with empty section (pins the boundary); body trailing-newline
  trimming.
- internal/orchestration feature_surface: add format cases `> surface:
  user-facing`, `- surface: user-facing`, `**Surface:** user-facing` — will
  FAIL today; pair with the deferred fix or pin current behavior explicitly.

Integration:
- `artifact new <f> changelog` twice → second run exits non-zero with
  ErrArtifactExists and the first file's bytes are untouched.
- `docs context` run with CWD at a subdirectory → non-zero, error names the
  three repo-relative paths (pins the CWD contract).

Acceptance (docs_step_markdown_first_acceptance_test.go):
- Real binary: `docs context` happy path + all-missing aggregation +
  optional-changelog hint (per spec); `docs generate`/`docs validate` exit 1
  as unknown commands; `docs context <f>` run twice yields byte-identical
  stdout (idempotency).
- Docs-step complete for a user-facing fixture whose specialist evidence
  lists only `.workflow/` paths → fails naming "docs/ or README.md".

## Residual Risks

- Decision-3 gaming (existing untouched docs file, incl. the plan file) —
  accepted by plan; enforcement is prompt-duty, not code.
- Symlink-under-docs escape — same trust class as above; not worth a code
  gate while outputs are agent-authored repo paths.
- `--force` on `artifact new changelog` can clobber real specialist content —
  operator error class, documented in memory.
- Prompt-injection via verbatim-embedded repo files in `docs context` —
  unchanged trust boundary (repo files were always the specialist's input).
- Legacy evidence in THIS repo referencing deleted `docs/project-docs/` paths
  would fail if re-validated; completed workflows are not re-validated.

## Deferred Findings

- `surface-line-format-detection` — surface: lines in blockquote/bullet/bold
  briefs undetected; user-facing features silently drop to the internal docs
  contract (recorded via roadmap defer).
- `stale-project-docs-pruning` — downstream repos keep a stale generated
  portal after merge; needs migrate-time pruning (plan "can wait", now
  recorded).
- `changelog-scanner-line-limit` — >64KB single-line changelog misreported
  as "empty" because the bufio.Scanner error is ignored (recorded).
