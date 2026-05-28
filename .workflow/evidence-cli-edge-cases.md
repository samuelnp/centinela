### Edge-Case Report: evidence-cli
**Date:** 2026-05-28

#### Risk Matrix

- **Case:** Concurrent writes — lock file orphaned after crash between flock and json rename
- **Impact:** High
- **Likelihood:** Low
- **Why:** `lock.go` opens/creates `<feature>-<role>.lock` then polls `syscall.Flock`. If the process is hard-killed after the lock file is created but before `Unlock` is deferred, the `.lock` file persists with the OS lock released (OS clears flock on fd close), but the file itself is never deleted. `Repair` only removes `.json.tmp` orphans, not `.lock` orphans. Stale lock files accumulate silently and are harmless but leave `.workflow/` polluted.

- **Case:** Deadlock — lock on JSON and companion .md not held atomically
- **Impact:** Medium
- **Likelihood:** Low
- **Why:** `evidence init` calls `Lock(feature, role)` for the JSON but `companion.go` writes the `.md` outside the lock window. Two concurrent `init` calls can interleave: Agent A holds JSON lock, Agent B writes companion `.md`; the companion can belong to Agent B's content while the JSON belongs to Agent A's. No deadlock, but a mismatched pair is silently written.

- **Case:** Schema/version skew — `extra` key collides with a reserved field name added in a future version
- **Impact:** High
- **Likelihood:** Medium
- **Why:** `schema_unmarshal.go` routes any key not in `knownKeys` straight into `Extra`. A future binary that promotes `extra.notes` to a first-class field will silently have both `r.Notes` (from `assignKnown`) and `r.Extra["notes"]` populated on round-trip from an older file. The marshal side only emits `r.Notes` (first-class), so the `extra.notes` value is silently dropped. No test covers promotion-on-round-trip.

- **Case:** `extra` key equals a reserved slug: `extra.feature`, `extra._meta`, `extra.outputs`
- **Impact:** Medium
- **Likelihood:** Low
- **Why:** `setExtra` in `setter.go` does not guard against keys that match `knownKeys`. Writing `centinela evidence set alpha big-thinker extra.feature "x"` stores a second `feature` key in `Extra`. `MarshalJSON` in `schema.go` emits known fields first then `sortedKeys(extra)`, so two `feature` keys land in the output. `json.Unmarshal` on that output keeps only the last one — silent data loss on round-trip.

- **Case:** Partial/aborted run — `writeTempFile` succeeds but `os.Rename` fails (cross-device target or file held open by IDE)
- **Impact:** Medium
- **Likelihood:** Low
- **Why:** `io_write.go` calls `os.Remove(tmp)` on rename failure, but does NOT return the original error wrapped with the remove error. If `Remove` also fails (e.g. permission error), only the `Remove` error is surfaced. The original rename failure reason is lost, making diagnosis harder. The `.tmp` file may also persist if `Remove` fails.

- **Case:** `Repair` sweeps live temp files during concurrent append
- **Impact:** Medium
- **Likelihood:** Low
- **Why:** `repair.go` globs `*.json.tmp` and removes all matches for the feature. The glob happens before the lock is taken by the concurrent writer. If `evidence repair alpha` runs between when Agent B's `writeTempFile` succeeds and when `os.Rename` executes, `Repair` deletes the temp file, then `Rename` fails with ENOENT. The JSON on disk is stale. No mtime guard is implemented (the plan mentioned one but the code does not have it).

- **Case:** Postwrite formatter active when `activeFeature` is empty (cwd outside any worktree)
- **Impact:** Medium
- **Likelihood:** Medium
- **Why:** `FormatEvidence` in `format_evidence.go` returns `(body, false, nil)` when `activeFeature == ""`, so no files are reformatted. This is safe. However, the caller in `hook_postwrite.go` must pass the correct `activeFeature`; if `DetectFeatureFromCwd` returns `""` due to the process running from a non-worktree directory (e.g. the repo root), the hook silently skips reformatting with no diagnostic. An agent that hand-writes JSON from outside a worktree never sees a reformat — invalidating AC4 without any error.

- **Case:** `isActiveFeatureEvidence` — symlinked `.workflow/` path does not match `workflow.WorkflowDir`
- **Impact:** Medium
- **Likelihood:** Low
- **Why:** `format_evidence.go` checks `strings.HasSuffix(dir, workflow.WorkflowDir)` after `filepath.ToSlash`. If the worktree is accessed via a symlink (`ln -s .worktrees/alpha alt-alpha`), `filepath.Dir` returns the symlink path while `workflow.WorkflowDir` is resolved to the canonical path. `HasSuffix` fails; the formatter silently skips all files in that session.

- **Case:** Path traversal in feature arg — `../` or absolute path
- **Impact:** High
- **Likelihood:** Low
- **Why:** `pathFor` in `io.go` calls `filepath.Join(workflow.WorkflowDir, fmt.Sprintf("%s-%s.json", feature, role))`. Go's `filepath.Join` cleans `..` components, but a feature name of `../../etc/passwd` resolves to `../../etc/passwd-big-thinker.json` relative to `workflow.WorkflowDir`. If `WorkflowDir` is a relative path (e.g. `.workflow`), the final path could escape the repo. No validation of the feature slug exists in `ParseRole` or in any `cmd/` entrypoint before calling `pathFor`.

- **Case:** Very long feature names or unicode slugs causing filesystem errors
- **Impact:** Low
- **Likelihood:** Low
- **Why:** macOS HFS+ has a 255-byte filename limit. A feature name of 240+ chars causes `os.OpenFile` to fail with ENAMETOOLONG on the lock and temp files. The error message mentions the internal path, not a user-friendly hint. Unicode feature names with NFC/NFD normalization differences (macOS normalizes to NFD) could cause `IsKnownRole` lookups to succeed but `isActiveFeatureEvidence`'s `HasPrefix` to fail.

- **Case:** `AppendField` dedup is case-sensitive; whitespace-padded values treated as different
- **Impact:** Low
- **Likelihood:** Medium
- **Why:** `appendUnique` in `appender.go` uses `item == value` (exact string equality). `centinela evidence append alpha big-thinker outputs " foo.md"` (leading space) produces a duplicate entry distinct from `"foo.md"`. Agent prompts or shell expansion can easily inject leading/trailing whitespace.

- **Case:** `centinela evidence read --field extra.<key>` on a non-existent key returns an error, not empty
- **Impact:** Low
- **Likelihood:** Medium
- **Why:** `ReadField` in `appender.go` returns `fmt.Errorf("extra key %q not set", ...)` for missing extra keys. This is correct but the exit code from the CLI command must be non-zero. If the Cobra command maps this to exit 0 (e.g. empty stdout), agents that check exit code will incorrectly treat "not set" as success.

- **Case:** `centinela evidence schema <role>` — HTML escape in JSON skeleton output
- **Impact:** Low
- **Likelihood:** Low
- **Why:** `SchemaSkeleton` in `repair.go` calls `skel.MarshalJSON()` which uses a manual byte buffer (no `json.Encoder`). There is no HTML escape risk from the custom emitter. However, if the skeleton is ever routed through `json.Indent` (as the hook does), `<`, `>`, `&` in string values are escaped to `<` etc. Field values like `"<feature-slug>"` would render as `"<feature-slug>"` in the formatted output, breaking prompt embedding readability.

- **Case:** `prompts_mandate_cli_acceptance_test.go` false positive — fenced code block labeled "what NOT to do" contains `python3 -c`
- **Impact:** Medium
- **Likelihood:** High
- **Why:** The acceptance test in Slice 3 is described as scanning for `python3 -c` in every prompt file. If a prompt includes a "forbidden patterns" example (e.g. showing what an agent must NOT do), the literal string `python3 -c` appears in the file and the test fails. The feature brief explicitly flags this risk. No evidence that the test implementation uses context-aware scanning (e.g. skipping `<!-- NOT DO -->` sections).

- **Case:** Scaffold mirror parity — case-only filename differences on macOS case-insensitive FS
- **Impact:** Medium
- **Likelihood:** Low
- **Why:** macOS APFS is case-insensitive by default. A parity test that compares `docs/architecture/Big-Thinker-prompt.md` vs `internal/scaffold/assets/docs/architecture/big-thinker-prompt.md` passes on macOS (same inode) but fails on Linux CI (different files). The parity test must compare byte contents, not just file existence.

- **Case:** `centinela artifact new` — unknown kind suggestion list ordering is non-deterministic
- **Impact:** Low
- **Likelihood:** Low
- **Why:** `ParseKind` in `artifact.go` builds the allowed list via `KindsAllowed()` which returns a fixed slice, then sorts it. Ordering is stable. However, the error message format `(allowed: [edge-cases gatekeeper production-readiness documentation-specialist])` uses Go's default `%v` slice format which omits commas — differs from the spec scenario which expects a comma-separated list.

- **Case:** `centinela artifact new` — `documentation-specialist` kind writes both `.md` and `.json` but overwrite guard only checks `.md`
- **Impact:** Medium
- **Likelihood:** Low
- **Why:** `KindDocumentationSpecialist` creates two files. If the `.md` does not exist but the `.json` does (partial prior run), the guard passes and the `.json` is silently overwritten. No test covers the partial-existence case for two-file artifact kinds.

#### Missing or Weak Scenarios

The following cases from the risk matrix have NO coverage in `specs/evidence-cli.feature`:

1. Lock file orphan after crash — no scenario for `centinela evidence repair` cleaning `.lock` files (only `.json.tmp` is covered).
2. `extra` key colliding with reserved field name (e.g. `extra.feature`) — Scenario "Free-form attachments use the extra slot" only tests `extra.note` (benign key); no test for `extra.feature` or `extra._meta`.
3. Feature arg path traversal (`../` slug) — no scenario; would be added to `tests/unit/evidence_io_test.go`.
4. Append with leading/trailing whitespace in value — not covered; would be added to `tests/unit/evidence_appender_test.go`.
5. `prompts_mandate_cli` false-positive with fenced "forbidden" example — not covered; would be added to `tests/acceptance/prompts_mandate_cli_test.go` with an allowlist for code blocks marked as negative examples.
6. `artifact new documentation-specialist` partial existence (`.json` present, `.md` absent) — no scenario; would be added to `tests/integration/artifact_test.go`.
7. Postwrite formatter with empty `activeFeature` (cwd outside worktree) — Scenario 12 covers scoping between two active features but not the "no active feature" path.

#### Proposed/Added Tests

**Unit:**
- Test `SetField` with `extra.feature`, `extra._meta`, `extra.outputs` as keys asserts they are stored in `Extra` without corrupting typed fields on round-trip. File: `internal/evidence/setter_branches_test.go` (extend existing).
- Test `AppendField` with values containing leading/trailing whitespace asserts they are NOT deduped against trimmed equivalents. File: `internal/evidence/appender_test.go` (extend existing).
- Test `pathFor` with `../` and absolute-path feature slugs asserts the resulting path stays within `workflow.WorkflowDir`. File: `internal/evidence/io_errors_test.go` (extend existing).

**Integration:**
- Test `Repair` does not delete a `.json.tmp` file that was just created by a concurrent `WriteAtomic` call (simulate via a goroutine hold). File: `tests/integration/evidence_repair_race_test.go` (new).
- Test `artifact new documentation-specialist` when `.json` exists but `.md` does not: command exits non-zero without overwriting the existing `.json`. File: `tests/integration/artifact_partial_test.go` (new).

**Acceptance:**
- Extend `tests/acceptance/prompts_mandate_cli_test.go` to skip lines inside fenced code blocks whose info string contains `forbidden`, `do-not`, or `example-bad`, preventing false positives from negative-example sections. File: `tests/acceptance/prompts_mandate_cli_test.go` (modify).

#### Residual Risks

- **Lock file accumulation**: `.lock` files are never cleaned up. After many crashed invocations `.workflow/` fills with inert lock files. Mitigation: extend `Repair` to also glob `<feature>-*.lock` and remove them; add a test.
- **`extra` key collision with future promoted fields**: No runtime guard prevents `extra.notes` from shadowing a field added in v2. Mitigation: document the reservation in `setter.go` and add a `reservedExtraKeys` check that rejects keys matching any entry in `jsonKeyOrder`.
- **Postwrite hook skips silently when cwd is outside worktree**: AC4 is unobservable when the hook fires from a non-worktree cwd. Mitigation: emit a `[centinela] postwrite: no active feature detected — skipping reformat` debug line to stderr when `activeFeature` is empty and the path matches `.workflow/*.json`.
