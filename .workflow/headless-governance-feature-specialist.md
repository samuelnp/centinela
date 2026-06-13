# Feature-Specialist Report: headless-governance

**Date:** 2026-06-13
**Role:** feature-specialist (plan step, after big-thinker)
**Handoff to:** senior-engineer

## Resolved open questions (big-thinker leans confirmed)

1. **Exit mechanism — sentinel error with SilenceErrors/SilenceUsage (CONFIRMED).**
   The `verdict` command marshals the packet to stdout, then returns a sentinel
   error carrying the exit code (1 on fail). The command sets `SilenceErrors` and
   `SilenceUsage` so cobra prints no usage/error noise and the JSON is the only
   stdout output. `main` translates the sentinel into the process exit code. This
   keeps the command unit-testable (capture stdout, assert returned error) and
   guarantees the JSON reaches stdout before the non-zero exit. NOT `os.Exit`
   inside RunE.

2. **Evidence index breadth — ALL on-disk role evidence for the feature (CONFIRMED).**
   The index enumerates every `.workflow/<feature>-*.json` role file present,
   not just the current step's required roles, so a reviewer sees the full
   produced trail. Missing-but-required roles surface via gates/verify, not
   invented here. Empty when no role JSON files exist.

3. **Verdict gate-filter scope — always full-scan in v1 (CONFIRMED).**
   `verdict` passes `Filter = nil` to `gates.RunWithFilter`, i.e. a comprehensive
   full scan. A `--changed/--full` flag is a documented follow-up; a verdict
   should be complete, not diff-scoped.

4. **`[headless]` config-comment wording (CONFIRMED).**
   The generated/sample `[headless]` section must carry this comment documenting
   the precedence and the "headless wins" contract:

   ```toml
   # [headless] — unattended execution umbrella (CI / daemon / fleet).
   # When headless is active, the step-review and plan-advisor hooks are silenced
   # regardless of step_confirmation_mode or plan_advisor_mode — headless WINS
   # over those explicit per-knob settings (it is an absolute "no human-aimed
   # output" contract). Precedence: CENTINELA_HEADLESS env > enabled >
   # (detect_ci && CI). Default off → byte-identical to a normal human session.
   [headless]
   enabled = false     # force headless on regardless of CI
   detect_ci = false   # opt-in: treat CI=="true"||"1" as headless
   ```

## Behavior Summary

Two deliverables, both additive and zero-config back-compatible:

- **Headless umbrella.** `config.IsHeadless(cfg) bool` (leaf `internal/config`)
  resolves precedence `CENTINELA_HEADLESS` env > `[headless] enabled` >
  (`detect_ci` && `CI=="true"||"1"`). The step-review hook
  (`shouldRenderReviewPrompt`) and plan-advisor hook (`runHookPlanAdvisor`)
  short-circuit to silence BEFORE the per-knob resolver, so headless wins over
  explicit `step_confirmation_mode`/`plan_advisor_mode`. Default off is
  byte-identical to today.

- **Verdict packet.** `internal/verdict.AssembleVerdict(feature, cfg, wf, deps)`
  runs gates (full scan), verify, indexes on-disk evidence, snapshots run info,
  computes a pass/fail summary. `GeneratedAt` is injected (deterministic,
  golden-testable). The `centinela verdict <feature>` command marshals the packet
  to stdout, status text to stderr, exit 0 (pass) / 1 (fail = any gate Fail OR
  `VerificationResult.HasFailures()`). Warnings are reported but non-failing.
  Gate statuses lowercased; verify statuses native UPPERCASE. Schema is
  `centinela.verdict/v1`.

## UX States (CLI surfaces)

- **Loading:** n/a (no spinner; synchronous run).
- **Empty:** feature with no on-disk role JSON → `evidence: []`; valid JSON still
  emitted with a computed summary.
- **Error:** gate Fail or verify failure → `verdict:"fail"`, `exitCode:1`, JSON
  still on stdout, command returns sentinel error → process exit 1.
  Hook side: headless active → review prompt and plan-advisor directive suppressed
  (empty stdout from those hooks).
- **Success:** all gates pass + no verify failures → `verdict:"pass"`,
  `exitCode:0`, JSON on stdout, process exit 0.
- **Stream separation:** JSON → stdout; any human status line → stderr (so a fleet
  consumer pipes stdout to a parser cleanly).

## Out of Scope (v1)

- `--json` on `validate`/`verify` (follow-up).
- `fail_on_warning` knob (exit codes 0/1 only).
- `--plain` on `status`.
- `--changed/--full` on `verdict` (always full-scan in v1).
- Any change to what gates/verify check; any new stdin read.

## Handoff

- **Outputs:** `specs/headless-governance.feature` (25 scenarios, stable titles
  for `// Scenario:` traceability), this report, the plan.
- **To senior-engineer:** implement Slice 1 (headless umbrella) then Slice 2/3
  (verdict package + command) per the big-thinker file table. Per locked design
  `internal/verdict` stays UNMAPPED in the import_graph layer map (non-failing
  warn, like verify/ui) — do NOT add it to the layer config. Inject `GeneratedAt`;
  marshal only structs/slices; sort evidence by role; golden-test the JSON.
