### Big-Thinker Report: codex-support
**Date:** 2026-06-30

#### Problem
Codex users are locked out of Centinela: prewrite enforcement, postwrite tags,
and prompt-context injection are wired only for Claude Code and OpenCode (plus
advisory-only Aider). Codex's hook system mirrors Claude Code's (lifecycle hooks
in `.codex/config.toml`, stdin-JSON I/O, exit-2 blocking, `additionalContext`
injection, native `AGENTS.md` rules file), so it is a **full first-class**
harness. The `host-harness-adapters` registry exists to absorb exactly this:
add one adapter, reuse the existing managed-file/sync seams, change no behavior
for the three existing harnesses.

#### Scope (In / Out)
**In:**
- `internal/setup/adapter_codex.go` (`codexAdapter`, all three capabilities).
- `internal/setup/codex_config.go` — fully-managed `.codex/config.toml` via the
  EXISTING `planManagedFile` seam, Kind `SyncKindPrewriteHook`, carrying a
  `# centinela:managed-version=` header (already recognized by the seam).
- `applyItem` route for `.codex/config.toml` → `writeManagedCodexConfig`.
- Register `"codex"` in `orderedAgents` + `registry`.
- `cmd/centinela` wiring: extract shared `applyManagedSetup`, add `setupCodex`,
  dispatch case, update `--agent` help strings.
- Golden + capability parity coverage for codex; colocated unit tests; an
  acceptance init→migrate no-drift idempotency test.

**Out:**
- Global `~/.codex/config.toml` (project-local `.codex/` only).
- Widening `composites["both"]` (stays `{claude, opencode}`; codex is
  `--agent codex` only).
- `internal/orchestration` Codex runner enum (separate model-routing concern,
  already exists).
- Editing `agentsContent` / bumping `setupDocVersion` (owned by Phase-11
  `agents-md-canonical-surface`).

#### Dependencies & Assumptions
- Depends on `host-harness-adapters` (the `HarnessAdapter` interface, registry,
  capability-parity test, `planManagedFile`/`writeManaged` seams) — all present.
- Assumes Codex command hooks are declared as inline `[[hooks.<Event>]]` tables
  in project-local `.codex/config.toml`, command form `["centinela","hook",…]`,
  PreToolUse blocking via exit 2 — per the June 2026 doc findings in the brief.
  The EXACT TOML hook-table key/field schema is the one residual unknown to
  confirm against Codex docs in the code step (see Risks).
- `internal/setup` = Infrastructure layer; `cmd/centinela` = thin outer layer
  (G7). The new `applyManagedSetup` helper is presentation glue only.
- Golden fixtures live under `internal/setup/testdata/golden/<agent>/`, not
  repo-root `testdata/`.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|------------|
| Codex `.codex/config.toml` hook-table schema differs from the sketch (key names/command field shape) | Hooks silently don't fire — enforcement absent | Medium | Confirm exact schema against Codex docs in the code step before finalizing the body constant; the body is a single managed const, cheap to correct; acceptance test asserts presence, a manual Codex smoke is advisable |
| init_agent.go breaches the 100-line cap when adding a third near-duplicate setup func | G1 gate fails | High (without mitigation) | REQUIRED extraction of `applyManagedSetup(agent,label)`; collapse all three setup funcs to one-liners |
| TOML body constant pushes `codex_config.go` over 100 lines | G1 gate fails | Medium | Split the body literal into `codex_config_body.go` (const-only), mirroring `opencode_plugin.go` |
| Accidentally editing `agentsContent` to mention Codex | Bumps setupDocVersion, ripples all golden fixtures, forces every project to re-migrate | Low | Explicit non-edit in the plan; reuse `planAgentsFile()` as-is |
| Coverage measured per-package; tests/ tier files don't move the gate | Coverage gate dips below 95% | Medium | Colocate `_test.go` in `internal/setup` and `cmd/centinela`; aim ≥97% |
| Pre-existing custom `.codex/config.toml` in a user repo | User content clobbered | Low | `planManagedFile` routes unmanaged content to `manual-review` (never written); covered by a unit test |

#### Rollout (smallest correct slice first)
1. Pure `internal/setup` core: `codex_config.go` + `adapter_codex.go` + register
   + `applyItem` branch. Parity tests (golden + capability) turn green; no `cmd`
   change yet, no behavior change for claude/opencode/aider.
2. `cmd/` wiring: extract `applyManagedSetup`, add `setupCodex` + dispatch case +
   flag help.
3. Tests: golden fixtures, colocated unit tests, acceptance init→migrate no-drift.

#### Deferred Findings
None new. The AGENTS.md harness-specific "## OpenCode Integration" heading
generalization is already covered by the existing roadmap item
`agents-md-canonical-surface` (Phase 11, ROADMAP.md:172) — do not duplicate.

#### Handoff
Next role: **feature-specialist** — translate this plan into the `.feature`
Gherkin spec in `specs/` and per-AC acceptance criteria, especially: (1) codex
registered + `--agent codex` resolves; (2) init writes managed `.codex/config.toml`
(managed header) + AGENTS.md; (3) golden/capability parity for three harnesses;
(4) greenfield init→migrate reports no pending drift.
