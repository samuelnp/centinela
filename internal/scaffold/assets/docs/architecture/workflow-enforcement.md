<!-- centinela:doc-version=1 template=docs/architecture/workflow-enforcement.md -->
# Workflow Enforcement System

## Problem
AI agents (and humans) skip steps when given a list. Instructions alone are
not enough. We need mechanical enforcement.

## Three Enforcement Layers

### Layer 1: Workflow Tracker (`.workflow/`)

Every feature gets a state file at `.workflow/<feature-name>.json`:

```json
{
  "feature": "example-feature",
  "startedAt": "2026-03-05T17:00:00Z",
  "currentStep": "code",
  "steps": {
    "plan": {
      "status": "done",
      "artifact": "docs/plans/example-feature.md",
      "completedAt": "2026-03-05T17:00:00Z"
    },
    "code": {
      "status": "in-progress",
      "artifact": null,
      "completedAt": null
    },
    "tests": { "status": "pending" },
    "validate": { "status": "pending" }
  }
}
```

Rules:
- A step can only move to "in-progress" if the previous step is "done".
- A step is "done" only when its required artifact exists on disk.
- The workflow file is the source of truth for current progress.

### Layer 2: Required Artifacts Per Step (ENFORCED by `centinela complete`)

The `complete` command validates artifacts exist on disk BEFORE advancing.
If validation fails, the step stays in-progress and the command exits with an error.

| Step | Validation | What it checks |
|------|-----------|----------------|
| plan | File search | A plan file in `docs/plans/` + a `.feature` file exists in `specs/` |
| code | None | Architecture rules govern this step |
| tests | File search | Test suite files in `tests/unit/` or `tests/integration/` + executable acceptance artifacts in `tests/acceptance/` + `validate.commands` includes acceptance execution + `.workflow/<feature>-edge-cases.md` |
| validate | Gate checks + commands | All built-in gates pass + all `centinela.toml` validate commands exit 0 |

> Note: The exact file extensions and paths checked are project-specific. See PROJECT.md → Folder Structure for the authoritative paths.

### Layer 3: The `centinela` Binary

The `centinela` CLI enforces the workflow with two mechanisms:

1. **Pre-write hook**: Blocks file writes in the wrong workflow step. Runs automatically via agent integrations.
2. **`centinela complete`**: Prevents advancing past a step without its required artifact or passing gates.

## Workflow Commands

The AI agent must use these commands to manage workflow:

```bash
# Start a new feature workflow
centinela start <feature-name>

# Mark current step as done (validates artifact exists, runs gates on validate step)
centinela complete <feature-name>

# Show current status
centinela status <feature-name>

# Show status of all active features
centinela status-all

# Run gate checks and validate commands manually
centinela validate
```

## How the AI Agent Must Behave

BEFORE starting any feature:
1. Run `centinela start <feature-name>`
2. This creates the `.workflow` JSON and sets step to "plan"

BEFORE writing any code file:
1. Check current step in workflow
2. If the target file's layer doesn't match current step, STOP
3. Complete the current step first, then advance

AFTER each step:
1. Run `centinela complete <feature-name>`
2. This validates the artifact exists and advances to next step
3. Output the current workflow status to the user

## Validate Step

When completing the `validate` step, centinela automatically runs:
1. **Built-in gates** — G1 (file size), G11 (i18n if configured in `centinela.toml`)
2. **User commands** — all entries in `centinela.toml → [validate] commands`

Configure your stack's lint/type-check/test commands in `centinela.toml`:

```toml
[validate]
commands = [
  "npx tsc --noEmit",
  "npx vitest run",
  "npx cucumber-js"
]
```

Commands run natively via the OS — no shell scripts required. This works on Windows, macOS, and Linux.

## Dynamic Model Routing (optional)

By default (`[orchestration] routing_mode = "static"`, which is also what an
absent key means) every role's model comes from `[orchestration.models]`
project-wide, and all directive output is unchanged.

With `routing_mode = "dynamic"` the orchestrator may route a role's tier for ONE
feature. `centinela start` and the orchestration hook then emit a single extra
directive line while any of the current step's roles are still un-routed —
naming those roles, their floors, the static defaults, and the command to
record the decision:

```bash
centinela route set <feature> <role> <tier> [--reason "..."]
centinela route show <feature>
```

`route set` is refused for an unknown role or tier, a role no step of this
workflow schedules, a tier below the role's floor (the error names the floor), a
downgrade below the static default with no `--reason`, and any downgrade once
the role's step is underway or completed. Upgrades are allowed anytime. Floors
come from `[orchestration.floors]` (shipped defaults: `gatekeeper = "reasoning"`,
`planner = "balanced"`); an explicit entry replaces the default. Un-routed roles
always fall back to the static chain, and every accepted decision is recorded as
a `route-decision` telemetry event.

Floors bind twice, and the second time is the one that matters: the workflow
state file is agent-writable, so a route hand-written below its role's floor
would bypass `route set` entirely. Model resolution therefore re-checks the
floor and IGNORES any route that fails it — falling back to static, exactly as
it treats a corrupt tier. Two raw keys naming the same role are dropped for the
same reason. Floors govern ROUTES only: a project that lowers a role in
`[orchestration.models]` has made that choice explicitly, and `route show`
labels such a row `ignored` rather than reporting a tier the hook will not emit.

## Skip Rules

All five steps are mandatory. No step can be skipped — this is enforced by the binary.
Domain/core logic, tests, and validate are especially non-negotiable.

## Preserved Custom Sections

## Layer 1: Workflow State (`.workflow/`)

Each feature has `.workflow/<feature>.json` with:

- `currentStep` in `plan|code|tests|validate|docs|done`
- per-step status (`pending|in-progress|done`)
- completion timestamps

This file is the source of truth for progress.


## Layer 2: Artifact Validation (`centinela complete`)

Before advancing, Centinela validates required artifacts:

| Step | Required artifacts |
|------|--------------------|
| plan | `docs/features/<feature>.md`, `docs/plans/<feature>.md`, and at least one `specs/*.feature` |
| code | none (architecture rules apply during implementation) |
| tests | test files in `tests/unit` or `tests/integration`, executable acceptance files in `tests/acceptance`, at least one acceptance execution command in `[validate] commands`, and `.workflow/<feature>-edge-cases.md` |
| validate | gatekeeper report at `.workflow/<feature>-gatekeeper.md` — with a non-empty commands-run record and a current verified revision — and `centinela validate` pass |
| docs | non-empty `.workflow/<feature>-changelog.md`; user-facing features also require the `.workflow/<feature>-documentation-specialist.{md,json}` evidence pair whose `outputs` include a real updated file under `docs/` or `README.md` |

In strict orchestration mode, `plan` evidence from `planner` (or, for a workflow
started before the `planner-v1` contract, from `big-thinker` and
`feature-specialist`) MUST include the feature's own brief
`docs/features/<feature>.md` and its plan `docs/plans/<feature>.md` in its JSON
`inputs` list. Additional inputs are allowed — the rule is *include*, not
*only* — so evidence written before this rule shrank still validates.

Strict orchestration evidence must also be actionable. Required specialist JSON
`outputs` must point to real repo files on disk rather than summary strings.

- `planner` outputs must include a real `docs/plans/...` or `specs/...` artifact (same rule for the retired `big-thinker` / `feature-specialist` on legacy workflows).
- `senior-engineer` outputs must include at least one real non-evidence implementation file.
- `ux-ui-specialist` is required during `code` when `docs/features/<feature>.md` declares `surface: user-facing`; its outputs must include at least one real UI file under configured `ui_paths`, `mobileFirst: true`, and the required UX edge-case tag set.
- `qa-senior` outputs must include at least one real `tests/...` file and `.workflow/<feature>-edge-cases.md`.
- `documentation-specialist` outputs must include at least one real updated file under `docs/` or exactly `README.md`.

If validation fails, the step remains in progress.


## Layer 3: Hook Enforcement

Centinela hooks enforce write discipline and context:

- `centinela hook prewrite` blocks out-of-step writes.
- `centinela hook postwrite` emits compact workflow tags.
- `centinela hook setup` injects setup guidance when required files are missing.
- `centinela hook context` injects active workflow context.

These hooks are wired by `centinela init` for Claude and OpenCode integrations.


## Required Agent Behavior

1. Start every feature with `centinela start <feature>`.
2. Do work only for the active step.
3. After producing artifacts, run `centinela complete <feature>`.
4. Respect `workflow.step_confirmation_mode` for review prompts:
   - `every_step` (the strict-profile default): require explicit user confirmation for each step.
   - `after_plan` (the guided-profile default, and so the zero-config default): require confirmation only for plan -> code transition.
   - `auto`: no review prompt; still run `centinela complete <feature>` explicitly.
5. During `plan`, Centinela runs plan-advisor mode by default:
   - `workflow.plan_advisor_mode = "missing_info"` asks only missing high-value questions.
   - `workflow.plan_advisor_mode = "always"` always asks an advisor round.
   - `workflow.plan_advisor_mode = "off"` disables advisor prompting.
   - `workflow.plan_question_limit` caps each advisor round and defaults to `4`.
   - advisor context uses current feature artifacts first, then roadmap dependencies, then same-phase siblings.
   - related edge-case lessons and roadmap quality notes can shape questions, but the hook emits summarized context instead of raw file dumps.


## Enforcement Profiles: process scales, proof does not

The zero-config default is `guided`. A profile changes how much **process** a
feature pays for; it never changes what is **verified**.

| Behavior | Side of the line | strict | guided | outcome |
|---|---|---|---|---|
| Step confirmation cadence | PROCESS | every step | after plan | auto |
| Orchestration evidence bundle (`strict-subagents-v1`) | PROCESS — a record of *who* produced the work | required | — | — |
| Prewrite step-gating | PROCESS | on | on | off |
| Greenfield cascade: `ROADMAP.md`, roadmap analysis/quality, production-readiness *prompt doc* | PROCESS | blocking | advisory | advisory |
| Greenfield `PROJECT.md` + `.workflow/roadmap.json` | PROCESS floor | required | required | required |
| `start` refusals from roadmap content (Backlog, draft, unmet dependency, missing `Phase 0: Bootstrap`, bootstrap incomplete) | PROCESS floor | refuse | refuse | refuse |
| `centinela validate` gate set (file_size, import_graph, security, spec_traceability, build, docstring, coverage, fmt) | **PROOF** | identical | identical | identical |
| Full test suite + acceptance execution | **PROOF** | identical | identical | identical |
| `VerificationFresh` (×2, before and after the suite) | **PROOF** | identical | identical | identical |
| Claim verification | **PROOF** | identical | identical | identical |
| Gatekeeper report + grounded `adversarial-v1` verdict | **PROOF** | identical | identical | identical |
| Production-readiness gate (`**Status:** BLOCKING` blocks `complete`) | **PROOF**, config-driven | identical | identical | identical |

`cmd/centinela/complete_validate_gates.go` carries the invariant in its header
and a source-level test asserts the file contains no profile identifier at all,
so a future profile branch there fails a test rather than a review.

**Back-compat is state-dated, not clock-dated.** Every workflow created after
the default flipped carries `profileContract: "guided-default-v1"`. The
nothing-configured tier of `EffectiveProfile` returns `guided` only for a
workflow carrying that pin; a workflow started earlier — or in flight when the
binary was upgraded — resolves to `strict` for its whole life. No user config is
ever rewritten; `centinela doctor` emits an advisory (never an error) on a
project with workflows and no explicit `enforcement_profile`.

**The self-graded quality threshold is gone, in every profile.** The old
`overall >= 9` refusal on `.workflow/roadmap-quality.json` and on
`roadmap promote --scores` was removed outright: a number assigned by the party
it constrains was never evidence, so deleting it relaxes no proof. Scores are
still recorded, still range-checked 1-10, and `centinela roadmap validate`
reports low ones as advice with exit 0.
