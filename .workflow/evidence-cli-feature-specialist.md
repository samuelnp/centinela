### Feature-Specialist Report: evidence-cli
**Date:** 2026-05-28

#### Behavior Summary

`centinela evidence` is a typed CLI layer that replaces all manual `.workflow/<feature>-<role>.json` authoring. Agents call `centinela evidence init|set|append|read|validate|repair|schema` instead of reaching for `python3 -c`, heredocs, or `jq`. Every JSON write is atomic (temp-file + rename), every write is pretty-printed with stable key order, and a companion `.md` narrative is produced in the same call. `centinela evidence validate <feature>` cross-checks the on-disk files against `internal/orchestration/` rules and emits per-field fix hints in the form `centinela evidence set/append …`. `centinela artifact new <feature> <kind>` generates pre-filled templates for the non-role artifacts (`edge-cases`, `gatekeeper`, `production-readiness`, `documentation-specialist`). A PostToolUse hook auto-reformats any `.workflow/*.json` written outside the CLI, scoped strictly to the active feature's prefix so sibling worktrees are never touched. Slice 3 rewrites every agent prompt to mandate the CLI and adds an acceptance test that asserts no prompt embeds forbidden authoring patterns (`python3 -c`, `<<EOF`, raw JSON examples); a scaffold-mirror parity test extends coverage to prompt files.

#### Gherkin Scenarios

All 14 scenarios in `specs/evidence-cli.feature` are referenced below. Coverage assessment follows each entry.

1. **Init drops a schema-valid skeleton** — `centinela evidence init alpha big-thinker` writes `.workflow/alpha-big-thinker.json` with all required fields, pretty-printed, with `_meta.cli_version` set. Covered.
2. **Set writes a scalar field atomically** — `centinela evidence set alpha big-thinker status done` updates the field; no temp file remains. Covers AC3. Covered.
3. **Append extends a list field without duplicating entries** — double-appending the same path yields exactly one entry in `outputs`. Covers idempotent dedup. Covered.
4. **Read returns a single field for predecessor inspection** — `--field outputs` returns the JSON-encoded list; exit 0. Covers the agent introspection path. Covered.
5. **Validate exits non-zero with a fix hint on missing field** — stderr contains the exact `centinela evidence append …` fix command. Covers AC2. Covered.
6. **Atomic write survives a crash mid-append** — original JSON unchanged; `repair` removes orphan temp file. Covers AC3 crash path. Covered.
7. **Concurrent writes serialize via advisory lock** — both appends succeed; deduplication means `foo.md` appears once. Covers the concurrency edge case. Covered.
8. **Schema version skew preserves unknown fields** — older `extra.legacy_field` is retained by newer binary; validation passes. Covers AC7 / version-skew edge. Covered.
9. **Free-form attachments use the extra slot** — `extra.note` round-trips through `set` and `validate` without rejection. Covers the free-form slot. Covered.
10. **Artifact templates drop pre-filled stubs** — `centinela artifact new alpha edge-cases` creates `.workflow/alpha-edge-cases.md`; idempotent guard on re-run. Covers AC5. Covered.
11. **Postwrite hook reformats hand-written evidence** — minified JSON written via Write tool is rewritten pretty-printed; other features untouched. Covers AC4 + worktree scoping. Covered.
12. **Postwrite formatter is scoped to the active feature** — `beta-big-thinker.json` is NOT reformatted when CWD belongs to `alpha`. Covers worktree isolation. Covered.
13. **Agent prompts forbid hand-written JSON** — acceptance scan finds no `python3 -c`, no `<<EOF` near `.workflow`, every prompt references `centinela evidence`. Covers AC6. Covered.
14. **Scaffold mirror parity covers prompts** — edited prompt in `docs/architecture/` must equal mirror in `internal/scaffold/assets/`; test fails with diff otherwise. Covers AC6 mirror side. Covered.

**Gaps worth filling before senior-engineer starts:**

- No negative scenario for `centinela evidence init` when the feature does not exist in `.centinela/` (unknown feature slug). Should exit non-zero with a clear message.
- No scenario for `centinela evidence read` on a non-existent role file (the file has not been init'd yet). Currently there is no Given/When/Then for the "file absent" error path.
- No scenario for `centinela artifact new` with an unknown kind (e.g. `centinela artifact new alpha bogus-kind`); the guard should reject it with a usage error.
- Scenario 7 (concurrent writes) asserts `foo.md` appears exactly once, but that is the dedup constraint from Scenario 3 — the concurrency scenario should instead assert both sequential appends of two *different* paths both land, to avoid conflating dedup semantics with locking semantics.

#### UX States

This is a CLI feature; there is no UI surface. States are expressed as exit codes and stderr/stdout messages.

| State   | Trigger                                      | Surface                                             |
|---------|----------------------------------------------|-----------------------------------------------------|
| success | Valid subcommand completes                   | stdout (read/schema) or silent; exit 0              |
| empty   | `validate` on a freshly-started feature      | stdout: "no evidence files found for feature"       |
| error   | Missing/malformed field, lock timeout, unknown role | stderr: fix-hint line; exit non-zero           |
| loading | n/a (synchronous CLI)                        | n/a                                                 |

#### Out-of-Scope

- GUI or TUI interface for authoring evidence.
- Migration tool that backfills `_meta.cli_version` into pre-existing JSON files (files remain valid but `_meta` will be absent).
- Cross-feature evidence aggregation or reporting dashboards.
- Automatic resolution of conflicting concurrent writes (v1 errors and asks the agent to retry).
- Removing or renaming existing `centinela complete / validate / status / start` flags.
- Storing evidence in any location other than `.workflow/`.

#### Handoff

- Next role: senior-engineer
- Open clarifications:
  1. Confirm whether `centinela evidence init` should fail loudly or silently no-op when a `.workflow/<feature>-<role>.json` already exists (overwrite vs. guard).
  2. Confirm dedup policy for `append`: case-sensitive exact match only, or normalised path comparison?
  3. Should `centinela evidence repair` only drop temp files, or also attempt to recover truncated JSON from the temp file?
  4. For the advisory lock, confirm the timeout value (the plan says "short timeout") — 2s? 5s? This should be a named constant so prompts can cite it.
