### Senior-Engineer Report: evidence-cli
**Date:** 2026-05-28

#### Files Touched

| Path | Reason |
|------|--------|
| `cmd/centinela/evidence.go` | Cobra parent command for the `centinela evidence` subtree (Slice 1). |
| `cmd/centinela/evidence_init.go` | `evidence init <feature> <role>` — drops schema-valid JSON + companion `.md` skeleton (Slice 1). |
| `cmd/centinela/evidence_set.go` | `evidence set` — atomic scalar field write; supports `extra.<key>` (Slice 1). |
| `cmd/centinela/evidence_append.go` | `evidence append` — list-field append with dedup (Slice 1). |
| `cmd/centinela/evidence_read.go` | `evidence read --field` — JSON output for predecessor inspection (Slice 1). |
| `cmd/centinela/evidence_validate.go` | `evidence validate` — non-zero exit + fix-hint stderr (Slice 1). |
| `cmd/centinela/evidence_repair.go` | `evidence repair` — sweeps orphan `.json.tmp` files (Slice 1). |
| `cmd/centinela/evidence_schema.go` | `evidence schema <role>` — prints JSON skeleton for prompt embedding (Slice 1). |
| `cmd/centinela/artifact.go` | `centinela artifact new` — pre-filled stubs for edge-cases / gatekeeper / production-readiness / documentation-specialist (Slice 2). |
| `cmd/centinela/hook_postwrite.go` | Extends the postwrite hook to reformat `.workflow/*.json` written by hand and scope to the active feature (Slice 2). |
| `internal/evidence/schema.go` + `schema_init.go` / `schema_marshal.go` / `schema_unmarshal.go` / `schema_validate.go` | Typed evidence document, stable-key MarshalJSON, validation rules, role-specific gates (Slice 1). |
| `internal/evidence/io.go` + `io_write.go` | Write-temp-then-rename atomic write with `_meta.cli_version` and `written_at` injection. |
| `internal/evidence/lock.go` | `flock`-based advisory lock to serialize concurrent appends (Slice 1, fixes scenario "Concurrent writes serialize"). |
| `internal/evidence/appender.go` / `setter.go` | Field-level mutation primitives consumed by `evidence_{append,set}.go`. |
| `internal/evidence/repair.go` | Orphan tmp sweeper, idempotent. |
| `internal/evidence/roles.go` | Canonical role list + per-role rule lookup. |
| `internal/evidence/companion.go` | Companion `.md` skeleton emission for `init`. |
| `internal/evidence/fixhints.go` | Synthesizes the "centinela evidence append … " fix hints on validate failure. |
| `internal/evidence/orchestration_bridge.go` | Hands typed validate results back to `internal/orchestration` so existing strict-mode checks remain authoritative. |
| `internal/evidence/artifact.go` + `artifact_{write,templates,edge_cases,gatekeeper,prodready,docs}.go` | Pluggable artifact template registry for Slice 2's `centinela artifact new`. |
| `internal/hookpolicy/format_evidence.go` + `format_evidence_order.go` | Postwrite reformatter — re-emits JSON with stable key order, scoped to the active feature. |
| `docs/architecture/*-prompt.md` (×9) | Authoring rules block + `centinela evidence schema <role>` placeholder replacing the embedded JSON skeleton (Slice 3, sources only — mirror updated separately). |
| `docs/architecture/production-readiness-prompt.md.template` | Generic template variant of the authoring rules (Slice 3). |
| `internal/scaffold/assets/docs/architecture/*-prompt.md` + `…production-readiness-prompt.md.template` | Mirror sync, byte-identical to canonical (Slice 3). |
| `specs/evidence-cli.feature` | Acceptance criteria covering all 15 scenarios (Slice 1 baseline). |

#### Architecture Compliance

- **n-tier boundaries (PROJECT.md → Architecture Choice):**
  - `cmd/centinela/*` (outer layer) imports `internal/evidence` and `internal/orchestration` — allowed.
  - `internal/evidence/*` imports only `encoding/json`, `os`, `path/filepath`, `sync`, `syscall`, and `errors` — no upward imports.
  - `internal/hookpolicy/format_evidence*.go` reuses the same `internal/evidence` marshaller for stable-order reformatting — no new outbound deps.
- **G1 (≤100 LOC per file):** every new/modified Go file in this slice is ≤100 LOC. Verified by `internal/gates` against the diff during prior `centinela validate` runs.
- **G2 (n-tier outer layer):** the new `cmd/centinela/evidence_*.go` files contain only Cobra wiring (`cobra.Command`, arg parsing, exit-code mapping); all logic lives in `internal/evidence`.
- **G7 (no business logic in outer):** verified — see above.

#### Type-Safety Notes

- The evidence document is a struct (`internal/evidence.Document`) with typed fields; no `map[string]any` in the public API. The free-form `extra.<key>` slot is the single explicit `map[string]json.RawMessage` carrier with documented semantics.
- `MarshalJSON` / `UnmarshalJSON` enforce stable key order without resorting to `interface{}` round-trips (see `schema_marshal.go`).
- Role names are a `string`-typed constant set (`internal/evidence/roles.go`) consumed by every command — unknown roles fail at parse time, not at I/O time.
- Errors are wrapped with `%w` throughout (atomic write, lock acquisition, validate) so the orchestration layer can `errors.Is` / `errors.As` cleanly.

#### Trade-Offs

- **`flock` vs. `lockfile` library** — chose POSIX `flock(2)` via `syscall.Flock` to avoid a new dependency; non-POSIX support is not in scope (centinela ships only on macOS/Linux per PROJECT.md).
- **Atomic write via `os.Rename` on the same filesystem** — rejected double-write + checksum scheme; rename is the simplest crash-safe primitive that meets the spec scenario "Atomic write survives a crash mid-append". `evidence repair` cleans orphan `.tmp` files left from a hard kill.
- **Stable key order via custom `MarshalJSON`** — rejected `json.Encoder` field-order patches and the `tidwall/sjson` library; a small explicit emitter keeps the output deterministic and the dependency surface zero.
- **Postwrite reformat is in-process** — kept the hook in-process to avoid an extra fork; the reformatter reuses the same marshaller, so hand-written JSON always normalises to canonical bytes.
- **Scope postwrite to the active feature via worktree CWD** — the hook resolves the active feature from the workflow state of the worktree it was invoked in. Touching `.workflow/<other-feature>-*.json` from a different worktree is a no-op (spec scenario "Postwrite formatter is scoped to the active feature").
- **Prompt mandate is asserted via a new acceptance test** — rather than a runtime check inside the agents (which the LLM can ignore), Slice 3 freezes the invariant as a test (`prompts_mandate_cli_acceptance_test.go`). Drift caught at CI time, not at runtime.

#### Handoff

- Next role: `qa-senior`
- Outstanding TODOs:
  - QA must add unit/integration tests in `tests/unit` and `tests/integration` to bring the new behaviours under the per-package coverage gate (the acceptance harness only exercises end-to-end paths).
  - QA must extend `.workflow/evidence-cli-edge-cases.md` with the corner cases enumerated above (orphan tmp, lock contention, scoping, schema skew).
  - One known gap: the CLI does not yet ship a verb to author the `.workflow/<feature>-<role>.md` companion in tandem with the JSON beyond `evidence init`'s stub. Authoring the substantive narrative still uses the Write tool. Tracked for a follow-up slice.
