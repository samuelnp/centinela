# Plan: `centinela evidence` — Typed Artifact CLI

References: [docs/features/evidence-cli.md](../features/evidence-cli.md),
[docs/architecture/evidence-contract.md](../architecture/evidence-contract.md).

## Problem

Across every step and every role, agents hand-write `.workflow/<feature>-<role>.{json,md}`
and templated companions (`edge-cases.md`, `gatekeeper.md`,
`documentation-specialist.{md,json}`). With no typed CLI, agents reach for
`python3 -c`/heredoc/`jq` to escape, merge, and minify JSON. Output is
unreadable, often invalid, and forces human round-trips. The schema is
documented in [evidence-contract.md](../architecture/evidence-contract.md)
but authoring is still manual.

## Solution

Three dependency-ordered slices. Domain logic lives in `internal/evidence/`
(new); `cmd/centinela/` stays a thin orchestrator. The validator already in
`internal/orchestration/` (`evidence.go`, `output_rules.go`,
`plan_snapshot.go`, `evidence_ux.go`) is reused — not rewritten.

### Slice 1 — `evidence-cli-core` (the blocker for value)

- **`internal/evidence/schema.go` — new.** One Go struct per role
  (`BigThinker`, `FeatureSpecialist`, `SeniorEngineer`, `UXUISpecialist`,
  `QASenior`, `ValidationSpecialist`, `Gatekeeper`, `ProductionReadiness`,
  `DocumentationSpecialist`). Each embeds a common `Meta` (cli_version,
  written_at, role, feature) and exposes `Validate() []FieldError`. Single
  source of truth for both runtime validation and prompt-embedded
  skeletons. ≤100 LOC per file — split per role if needed.
- **`internal/evidence/io.go` — new.** `Read(feature, role)`,
  `WriteAtomic(feature, role, payload)` (temp-file + rename),
  `Lock(feature, role)` (advisory `flock` with short timeout). All writes
  go through `WriteAtomic`; no partial files on disk.
- **`internal/evidence/companion.go` — new.** Pairs each JSON write with
  the `.md` companion (narrative). Single subcommand call writes both.
- **`internal/evidence/validate.go` — new.** Walks `.workflow/<feature>-*.{json,md}`,
  delegates schema validation to `internal/orchestration/evidence.go`'s
  existing rules (no duplicate logic), and emits per-error fix hints in
  the form `centinela evidence set <role> <field> <value>`.
- **`cmd/centinela/evidence.go` — new.** Thin orchestrator registering
  `centinela evidence` with subcommands:
  - `init <feature> <role>` — drop the per-role skeleton.
  - `set <feature> <role> <field> <value>` — typed scalar set.
  - `append <feature> <role> <field> <value>` — append to a list field
    (e.g., `outputs`, `edgeCases`, `inputs`).
  - `read <feature> <role> [--field <name>]` — typed read for agents to
    inspect predecessor evidence without `jq`/`python`.
  - `validate <feature>` — exit non-zero on failure with fix hints.
  - `repair <feature>` — drop orphaned temp files; idempotent.
  - `schema <role>` — print the JSON skeleton for prompt embedding.
- **`cmd/centinela/evidence_*_test.go` — new.** Unit tests per subcommand
  (≤100 LOC each, [[project_g1_applies_to_test_files]]).

### Slice 2 — `evidence-cli-artifacts`

- **`internal/evidence/artifact.go` — new.** Typed templates for
  `edge-cases`, `gatekeeper`, `production-readiness`,
  `documentation-specialist` (md + json pair where applicable).
- **`cmd/centinela/artifact.go` — new.** `centinela artifact new <feature> <kind>`.
  Idempotent — refuses to overwrite without `--force`.
- **`internal/hookpolicy/format_evidence.go` — new.** Pure function
  `FormatEvidence(path string, body []byte) ([]byte, error)` — detects
  `.workflow/*.json` paths scoped to the active feature's prefix,
  pretty-prints, normalizes key order. Returns input unchanged for
  non-evidence paths.
- **`cmd/centinela/hook_postwrite.go` — modify.** After existing
  postwrite logic, call `hookpolicy.FormatEvidence` and write back via
  the atomic writer. Worktree-scoped: only the active feature's
  `.workflow/` prefix is touched.
- **`internal/hookpolicy/format_evidence_test.go` — new.** Includes a
  regression test for [[project_worktree_operational_model]] — other
  features' `.workflow/` files are untouched.

### Slice 3 — `evidence-cli-prompts`

- **`docs/architecture/*-prompt.md` — modify.** Each agent prompt
  (`big-thinker`, `feature-specialist`, `senior-engineer`,
  `ux-ui-specialist`, `qa-senior`, `validation-specialist`,
  `gatekeeper`, `production-readiness`, `documentation-specialist`)
  gets a "Authoring rules" block:
  - Use `centinela evidence init|set|append` to author your JSON.
  - Use `centinela evidence read` to inspect predecessor evidence.
  - Do NOT use `python`, `jq`, heredoc, or `cat <<EOF` to write
    `.workflow/*.json`. The postwrite hook will reformat your output
    and the validator will reject it on schema mismatch.
  - The per-role JSON skeleton is removed from the prompt body and
    fetched at runtime via `centinela evidence schema <role>`
    (single source of truth — Slice 1).
- **`internal/scaffold/assets/docs/architecture/*-prompt.md` — modify.**
  Mirror every change. [[project_scaffold_mirror_partial_parity]] —
  extend the parity test in this PR to cover prompts.
- **`tests/acceptance/prompts_mandate_cli_test.go` — new.** Asserts no
  agent prompt contains forbidden authoring instructions (`python3 -c`,
  heredoc patterns, raw JSON examples). Catches future drift.

## Validation

- `centinela validate` (lint + type + full test suite) per existing gate.
- `go test ./...` end-to-end including new unit and acceptance tests.
- Smoke test: run a fresh feature workflow end-to-end (`start` → `plan`
  → `code` → `tests` → `validate` → `docs`), authoring all evidence
  via the new CLI only. Asserts zero `python`/`heredoc` reaches in
  transcript review.
- Per-package coverage stays ≥95% [[project_coverage_per_package_no_coverpkg]]
  — add colocated `_test.go` files (each ≤100 LOC).

## Compatibility

- AC7: pre-existing `.workflow/*.json` files remain valid. Unknown
  fields preserved on round-trip; missing required fields rejected.
- No change to `centinela start/complete/status/validate` flags or to
  workflow state format.
- Schema bound to binary version via `_meta.cli_version`. Older clients
  reading newer files: tolerated. Older files validated by newer
  binaries: tolerated if all required fields present.

## Rollout

1. Slice 1 merged behind no flag — CLI available, prompts unchanged.
2. Slice 2 merged — postwrite formatter active for all features.
3. Slice 3 merged — prompts mandate the CLI; acceptance test enforces
   the mandate going forward.

Slices 2 and 3 may split into follow-up features if the plan-step
subagents flag sequencing pressure; Slice 1 is the only blocker.
