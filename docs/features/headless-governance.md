# Feature: headless-governance

- surface: internal
- status: planned
- roadmap: Phase 6 — Capability-Adaptive Governance
- fixes: prompts assume a human in a chat session; unattended runs stall on
  questions or silently bypass them, and a run's verdict is only ever a
  human-read transcript — there is no machine artifact to review.

## Problem

Centinela's governance signals were built for one consumer: a human in a chat
session. Two gaps block unattended, CI, and fleet use (cf. the sibling Capataz
quota-aware daemon and Magallanes fleet control plane):

1. **Confirmation/advisor prompts assume a human will answer.** The step-review
   prompt (`hook_context.go` → `ui.RenderReviewReady`) and the plan-advisor
   questions (`hook_plan_advisor.go`) are stdout directives aimed at a person.
   The individual knobs to silence them already exist
   (`step_confirmation_mode = auto`, `plan_advisor_mode = off`), but there is
   **no single umbrella signal** — an unattended runner must know and set both,
   and there is no CI auto-detection. (No blocking stdin reads exist, so nothing
   actually *hangs* today; the risk is a runner that doesn't set the knobs gets
   noise it can't act on, or a human forgets one knob.)

2. **There is no machine-readable end-of-run verdict.** Gate results
   (`gates.Result`), verify checks (`verify.VerificationResult`), evidence
   files, and workflow state all exist as structs, but every surface renders
   them through Lipgloss to stdout. There is no `--json` anywhere. A fleet
   controller or CI job that wants to review an agent's work by *evidence*
   (did the gates pass? did verify confirm the claims? what evidence was
   produced?) has to scrape styled terminal text.

## Who's hurting

Anyone running Centinela without a human at the keyboard: CI pipelines gating a
merge, the Capataz daemon driving a feature to completion, a Magallanes fleet
reviewing many agents' work. Today they reverse-engineer governance from
transcripts instead of consuming it as a first-class output.

## The core idea

Make unattended execution a first-class consumer of governance, in two pieces:

- **A headless umbrella** — one resolved signal (`CENTINELA_HEADLESS` env >
  `[headless] enabled` config > opt-in `CI` auto-detect) that, when active,
  forces the prompt-emitting hooks into their silent modes. It composes with
  the existing per-knob settings; it does not replace them.
- **A verdict packet** — a deterministic JSON document aggregating a workflow
  snapshot, gate results, verify checks, and an evidence index, emitted by a new
  `centinela verdict <feature>` command to stdout, with an exit code that
  encodes pass/fail.

## Scope

### In (v1)

- `[headless]` config section: `enabled` (bool), `detect_ci` (bool, opt-in).
- `CENTINELA_HEADLESS` env override (highest precedence).
- A `config.IsHeadless(cfg)` resolver consulted by the review-mode and
  plan-advisor hooks; when headless is active they suppress human-aimed output.
- New `internal/verdict` package: `VerdictPacket` + `AssembleVerdict`,
  Lipgloss-free, deterministic JSON marshal.
- New `centinela verdict <feature>` command: JSON to **stdout**, any status
  text to **stderr**, exit 0 = pass / 1 = fail.
- `--headless` flag on the `verdict` command (for parity / explicit override).

### Out (v1, deferred)

- `--json` flags on `validate` / `verify` (the dedicated `verdict` command is
  the single v1 surface; note these as a follow-up).
- A `fail_on_warning` knob (warnings do not fail v1; exit codes are 0/1 only).
- A `--plain` flag on `status` (non-TTY already renders plain text).
- Changing what any gate or verify check *checks* — verification is constant.
- Any new blocking stdin read or interactive mode.

## Hard invariant (must not regress)

Default off → **zero behavior change**. With no `[headless]` config, no
`CENTINELA_HEADLESS` env, and `detect_ci` unset, every hook resolves
byte-identically to today. `internal/verify` and `internal/gates` are not
modified — the verdict packet only *reads* their results.

## Dependencies

- None (the per-knob settings `step_confirmation_mode` / `plan_advisor_mode`
  and the `gates`/`verify` packages all already exist and are consumed read-only).
