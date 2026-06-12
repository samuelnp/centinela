# Feature: deferred-findings-roadmap-capture

- surface: internal
- status: planned
- roadmap: Phase 5 — Operability & DX
- fixes: out-of-scope discoveries and deferred fixes live only in per-feature prose artifacts and evaporate — they never reach the roadmap, the single planning source of truth

## Problem

Every workflow step produces findings that are deliberately not acted on in
that step, and today all of them die in prose:

- **big-thinker** frames scope and explicitly lists what is *out* for v1 —
  the "Out" bullets in its report go nowhere.
- **feature-specialist** has a mandatory `#### Out-of-Scope` section — a
  bullet list that is never read again after the plan step.
- **senior-engineer** records "Outstanding TODOs" in its handoff section —
  improvements it saw but correctly did not make mid-feature.
- **qa-senior** (with edge-case-tester) writes "Residual Risks" into
  `.workflow/<feature>-edge-cases.md` — known gaps that nobody triages.

The roadmap (`.workflow/roadmap.json`) is the single planning source of
truth, but nothing connects these four capture points to it. Information the
project already paid tokens to discover is lost, and the same gaps get
re-discovered (or shipped) later. The operator's explicit requirement: when
plan-step agents detect something not covered by the current feature, and
when code/tests-step agents surface findings they won't fix immediately,
that information must always land in the roadmap.

## The core idea

Make deferred-finding capture a first-class, mechanically verifiable output
of the four roles, with a CLI path that keeps `centinela roadmap validate`
green by construction.

Two design surfaces, decided at plan:

1. **Capture mechanism.** Direct `roadmap.json` edits are hostile to
   mid-step agents: every appended feature must also appear in
   `roadmap-analysis.json` and `roadmap-quality.json` with overall ≥ 9, or
   `roadmap validate` fails. Candidate designs:
   - (a) `centinela roadmap defer <slug> --summary <text> [--source
     <feature>/<role>]` — atomically appends the finding to a dedicated
     backlog/triage phase in `roadmap.json` *and* writes the matching
     analysis/quality entries, so validate never breaks.
   - (b) a deferred-findings ledger (e.g. `.workflow/deferred-findings.json`)
     written by `roadmap defer`, surfaced by `centinela roadmap`, and
     promoted into a real phase by an explicit `roadmap promote` at triage
     time.
2. **Prompt contract.** The four role prompts (`big-thinker`,
   `feature-specialist`, `senior-engineer`, `qa-senior`) and their byte-
   identical mirrors under `internal/scaffold/assets/docs/architecture/`
   gain a required "Deferred findings" obligation: any out-of-scope
   detection or not-fixed-now finding MUST be recorded via the capture
   command, and the report section references the recorded slugs (or
   states "none").

## Goal

- A `centinela roadmap defer`-style command an agent can run mid-step that
  records a finding without breaking `roadmap validate`.
- Updated prompt contracts for big-thinker, feature-specialist,
  senior-engineer, and qa-senior (plus scaffold mirrors, which are parity-
  tested) making capture mandatory whenever such a finding exists.
- Deferred findings are visible in `centinela roadmap` output so they are
  actually triaged, not just stored.
- Worktree-safe: capture from inside `.worktrees/<feature>/` must not
  corrupt or race the root roadmap artifacts.

## Non-goals (v1)

- **No auto-prioritization or auto-scheduling.** A deferred finding gets a
  slug, a summary, and a source; a human (or a later triage feature)
  decides phase, dependencies, and priority.
- **No validator hard-gate on "did the agent defer everything it should
  have".** That is unverifiable; the contract is prompt-level, the
  mechanics are CLI-level.
- **No change to gates or claim verification.**
- **No retroactive backfill** of the 397-entry legacy memory corpus or old
  Residual Risks sections.
