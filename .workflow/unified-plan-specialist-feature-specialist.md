### Feature-Specialist Report: unified-plan-specialist
**Date:** 2026-07-29

#### Behavior Summary
The plan step collapses from two sequential subagents (`big-thinker` at the
reasoning tier, `feature-specialist` at the balanced tier) into one `planner`
subagent at the reasoning tier that produces a single evidence pair
(`.workflow/<feature>-planner.{md,json}`) carrying both lenses — strategy
first, then spec — as ordered sections of one report. Back-compat is
state-dated, not file-presence-based: `centinela start` pins
`PlanContract: "planner-v1"` into `.workflow/<feature>.json`; a workflow with
that pin requires and accepts only the planner pair, and REFUSES a
hand-authored legacy pair even if both legacy files exist on disk. A workflow
with an empty `PlanContract` (started before this feature shipped) keeps
requiring and accepting the exact pre-migration two-role set, including its
partial-set failure mode. One resolver, `workflow.RequiredEvidenceRoles`, is
the single source of truth consulted by the hook directive, the `complete`
gate, and claim verification, so no surface can name a different required set
than another (the precedent-flagged PR #83 CRITICAL class). Legacy
model-override config keys (`big-thinker`/`feature-specialist`) keep resolving
and alias to `planner`; their deprecation is surfaced once, at `doctor` and at
`start`, never on every hook turn. The plan-advisor, statusline, OpenCode
agent config, and the merged `planner-prompt.md` doc (both lenses, ≤130
lines) complete the surface-level renames. Nothing about the after-plan human
confirmation stop, or the O(N) `docs/features/*.md` plan-snapshot input rule,
changes.

#### Gherkin Scenarios
See `specs/unified-plan-specialist.feature` (19 scenarios/outlines). Groups:
- **Fresh workflow, one role:** directive names exactly one `planner` role at
  reasoning tier and lists only the planner evidence pair; planner evidence
  with both artifacts passes `complete`; a hand-authored legacy pair is
  refused on a `planner-v1` workflow (contract-aware refusal); `evidence init`
  for a retired legacy role errors naming `planner` on a pinned workflow AND
  on a feature with no workflow state at all.
- **Legacy in-flight workflow:** empty `PlanContract` still requires and
  accepts the complete two-role legacy set verbatim; a partial legacy set
  (big-thinker only) still fails, naming the set THIS workflow requires plus
  a one-line contract annotation — never offering `planner` as an escape
  hatch it cannot satisfy; `evidence init` for a legacy role succeeds on an
  unpinned workflow.
- **Directive/gate agreement:** a Scenario Outline asserts the hook directive
  and the `complete` gate name the identical role set for both a pinned and
  an unpinned workflow — the single-resolver invariant (D3).
- **Guided profile:** no subagent evidence required, directive still names
  the single `planner` role, no evidence requirement printed.
- **Config aliasing:** a legacy `big-thinker` model override still resolves
  under `planner`; the deprecation notice appears at `doctor` and once at
  `start`, and is NOT repeated on every plan-step hook context call.
- **Plan-advisor:** header reads "One planner agent, two lenses: strategy
  first, then spec."; questions tagged `[strategy]`/`[spec]`, never the old
  role slugs; no instruction to delegate to two agents.
- **Planner prompt doc:** both lens headings present, strategy before spec;
  single `## Purpose`/`## Prompt Template`/`## Required Artifact`; one CLI
  authoring-rules block, not two; ≤130 lines or the documented aggregate
  ratchet (<215 total); legacy prompt docs no longer exist.
- **Scaffold parity:** the new prompt doc is mirrored byte-identically in
  `internal/scaffold/assets/docs/architecture/`; the two legacy mirrors are
  gone.
- **Statusline:** shows `planner` for a `planner-v1` workflow's plan step;
  still shows the legacy pair for an unpinned workflow.

#### UX States
This is a CLI-only feature (no graphical UI surface); the states below are
the operator-visible CLI surfaces this spec exercises, in place of the
loading/empty/error/success rows the template expects for UI features.

| Surface | Trigger | Behavior |
|---------|---------|----------|
| Hook directive (plan step) | `centinela` requests hook context during plan | Names exactly one role (`planner` if pinned; the legacy pair if unpinned), its tier, and the exact evidence file paths required — never both |
| `complete` block (fresh, forged legacy) | `centinela complete <f>` on a `planner-v1` workflow with only legacy files | Hard-blocked; message names `planner` as required, never accepts the legacy pair |
| `complete` block (legacy, partial set) | `centinela complete <f>` on an unpinned workflow with one of two legacy files | Hard-blocked; message names the legacy pair as *this workflow's* required set plus a one-line contract annotation |
| `evidence init <f> <legacy-role>` | Operator requests a retired-role stub | Succeeds silently on an unpinned workflow; errors naming `planner` on a pinned workflow or when no workflow state exists |
| `centinela doctor` | Deprecated `[orchestration.models]` keys present | One check row suggesting migration to `planner`; not blocking |
| `centinela start` | Legacy model-override keys present in `centinela.toml` | Prints the same deprecation notice once; never repeats it on subsequent hook turns |
| `centinela status` / statusline | Viewing a feature at the plan step | Shows `planner` for a `planner-v1` workflow; shows the legacy pair for an unpinned workflow |
| Plan-advisor questions | Plan step, advisor invoked | Two lens tags (`[strategy]`, `[spec]`) under one header naming one agent |

#### Out-of-Scope
- Any change to the code/tests/validate/docs roles or their evidence rules
  (`adversarial-validate-verifier` owns the validate-step role change).
- The after-plan human confirmation stop (`workflow.step_confirmation_mode`)
  — kept exactly as-is; it is the assurance mechanism at this step.
- The O(N) `RequiredPlanInputs` plan-snapshot glob's *content* rule — stays
  byte-for-byte identical; `planner` inherits it unchanged (`token-diet`'s
  scope, not this feature's).
- Renaming or deleting the `RoleBigThinker` / `RoleFeatureSpecial` constants
  — they survive for legacy evidence, config keys, and `brownmap` roadmap
  provenance (`Source.Role: "big-thinker"`), left alone.
- A Codex-equivalent role-agent registry — Codex has no per-role agent
  config surface today; already tracked as a deferred finding, not built here.
- Rewriting `docs/project-docs/kb/*.md` historical pages — they describe what
  shipped at the time and would be falsified by a retroactive edit.

#### Deferred Findings
No genuinely new gaps surfaced while writing the acceptance spec. The three
deferred findings this feature owns were already captured by big-thinker and
are not duplicated here:
- `codex-claude-role-agent-registry`
- `managed-agent-retirement-sweep`
- `prompt-doc-budget-ratchet`

#### Handoff
- Next role: senior-engineer
- Open clarifications:
  1. D6's fallback (per-file budget exception + aggregate <215-line ratchet)
     is conditional on the merged prompt doc not fitting in 130 lines after
     de-duplication — the code step should attempt the strict 130-line
     target first and only invoke the fallback (recording the reason in its
     own evidence) if it genuinely cannot fit without trimming lens content.
  2. Confirmed for the spec: the brief's Acceptance Criterion 4 ("accepts
     EITHER the planner pair OR the complete legacy set") is superseded by
     the plan's D2 state-dated contract rule, per the big-thinker report.
     The spec above is written to D2, not to the brief's literal AC4 text —
     flagged as a known brief/plan divergence, not a new one.
