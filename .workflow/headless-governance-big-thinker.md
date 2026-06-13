# Plan: headless-governance

### Big-Thinker Report: headless-governance

**Date:** 2026-06-13

#### Problem

Centinela's governance is built for one consumer — a human in a chat session.
Two concrete gaps block unattended / CI / fleet execution (Capataz, Magallanes):

1. **No umbrella non-interactive signal.** Every confirmation/advisor prompt is
   a stdout *directive* emitted by a hook. The per-knob settings to silence them
   exist (`step_confirmation_mode = auto`, `plan_advisor_mode = off`), but a
   runner must know and set each one independently, and there is no CI
   auto-detect. Surfaces:
   - Step-review prompt: `cmd/centinela/hook_context.go` →
     `shouldRenderReviewPrompt` → `effectiveConfirmationMode`
     (`cmd/centinela/hook_context_review_mode.go`).
   - Plan-advisor questions: `cmd/centinela/hook_plan_advisor.go` →
     `planadvisor.Directive`, gated by `plan_advisor_mode`.
   - (`status` already renders plain text to non-TTY; `init`/`merge` are
     already file-based. No blocking stdin reads exist anywhere.)
2. **No machine-readable verdict.** `gates.Result`, `verify.VerificationResult`,
   evidence JSON, and `workflow.Workflow` all exist as structs but are only ever
   Lipgloss-rendered to stdout. There is no `--json` surface. Reviewing an
   agent's work by *evidence* requires scraping styled terminal text.

#### Scope (In / Out)

**In (v1):**
- `[headless]` config section (`enabled`, `detect_ci`) + `CENTINELA_HEADLESS`
  env override.
- `config.IsHeadless(cfg)` resolver, consulted by the two prompt hooks.
- `internal/verdict` package: `VerdictPacket` + `AssembleVerdict`, deterministic
  JSON, Lipgloss-free.
- `centinela verdict <feature>` command: JSON to stdout, status text to stderr,
  exit 0 pass / 1 fail. `--headless` flag for explicit override.

**Out (v1):** `--json` on `validate`/`verify` (follow-up); `fail_on_warning`
knob (exit 0/1 only); `status --plain`; any change to what gates/verify check;
any new stdin read.

#### Dependencies & Assumptions

- **Dependencies:** none. The per-knob confirmation/advisor settings and the
  `gates`/`verify`/`evidence`/`orchestration`/`workflow` packages all exist and
  are consumed read-only.
- **Assumptions:**
  - The `CI` env var (`"true"`/`"1"`) is the universal CI signal — reuse the
    exact test already in `cmd/centinela/validate_mode.go` (`currentEnv`), but
    gated behind opt-in `detect_ci` because `CI` is set in many local shells.
  - Timestamps are injected into `AssembleVerdict` (not `time.Now()` inside) so
    the packet is deterministic and testable.
  - Verify is run with the same `Deps` construction as `cmd/centinela/verify.go`
    (`verifyRoot()` + `NewExecRunner()`).

#### Divergence from proposed design (with reasons)

I largely **adopt** the orchestrator's direction. Divergences/decisions:

1. **Headless lives in a new `[headless]` config section, not `[workflow]`** —
   adopted as proposed. It's a cross-cutting *execution mode*, not a
   workflow-step knob; a dedicated section reads cleanly and keeps
   `WorkflowConfig` from sprawling.

2. **`IsHeadless` lives in `internal/config`, NOT `internal/workflow`.** The
   resolver reads only env + config (no workflow state needed) and the two hooks
   that consult it already import `config`. `config` is the leaf layer (imports
   nothing internal), so a pure env+config helper belongs there. This keeps the
   hooks thin (G7) and adds no new import edges.

3. **Headless precedence: headless WINS over an explicit `step_confirmation_mode`
   / `plan_advisor_mode`.** The orchestrator left this open and leaned "headless
   wins"; I confirm it, with a clean implementation: headless does NOT mutate or
   override the per-knob resolver. Instead each hook short-circuits *before*
   consulting the per-knob resolver:
   - `shouldRenderReviewPrompt`: `if config.IsHeadless(cfg) { return false }`
     at the top.
   - `runHookPlanAdvisor`: `if config.IsHeadless(cfg) { return nil }` before the
     loop.
   This is the smallest, most auditable change and makes "headless = unattended,
   no human-aimed output" an absolute contract independent of the other knobs.
   Documented in the brief.

4. **CI auto-detect is opt-in (`detect_ci`), not always-on** — adopted. `CI` is
   set in many local dev shells; always-on would silently change behavior for
   existing users running tests locally. Opt-in preserves the zero-config
   back-compat guarantee.

5. **Exit codes 0/1 only in v1** — adopted. Warnings do not fail. A
   `fail_on_warning` knob is a clean follow-up; encoding warn-as-2 now would
   commit us to a tri-state before we know fleet consumers want it.

6. **ONE surface in v1: the dedicated `verdict` command** — adopted. `--json` on
   `validate`/`verify` is a thin follow-up that can reuse the same packet structs
   (or sub-marshals). Shipping one well-shaped surface first keeps scope tight.

7. **The `--headless` flag only exists on the `verdict` command, not globally.**
   Hooks have no flags (they're invoked by Claude Code with a fixed argv), so the
   flag is meaningless for them — they rely on env/config. The `verdict` command
   gets `--headless` purely so a caller can force-resolve headless for that
   invocation (precedence: flag > env > config > detect_ci). For v1 the
   `verdict` command's behavior does not actually branch on headless (it always
   emits JSON), so the flag is wired to the resolver for forward-compat /
   provenance reporting in the packet's `headless` field rather than gating
   output. *Decision:* keep the flag, surface the resolved headless state in the
   packet (`run.headless`), but do not let it alter the JSON shape in v1.

#### The verdict packet

```
// internal/verdict — domain package.
// May import config, workflow, gates, verify, evidence, orchestration, gitdiff.
// Must NOT import internal/ui (Lipgloss-free) or cmd/.

type Summary struct {
    Verdict   string `json:"verdict"`   // "pass" | "fail"
    ExitCode  int    `json:"exitCode"`  // 0 | 1
    Gates     Counts `json:"gates"`
    Verify    Counts `json:"verify"`
}

type Counts struct {
    Pass int `json:"pass"`
    Fail int `json:"fail"`
    Warn int `json:"warn"`
    Skip int `json:"skip"`
}

type RunInfo struct {
    Feature     string `json:"feature"`
    Step        string `json:"step"`
    Profile     string `json:"profile"`
    Archetype   string `json:"archetype"`
    DriverModel string `json:"driverModel,omitempty"`
    Headless    bool   `json:"headless"`
    GeneratedAt string `json:"generatedAt"` // RFC3339, injected
}

type GateResult struct {
    Name    string   `json:"name"`
    Status  string   `json:"status"`  // "pass"|"fail"|"warn"|"skip"
    Message string   `json:"message,omitempty"`
    Details []string `json:"details,omitempty"`
}

type VerifyCheck struct {
    Role   string `json:"role"`
    Claim  string `json:"claim"`
    Status string `json:"status"`  // PASS|FAIL|SKIP|WARN|CONFIG-ERROR|TIMEOUT
    Detail string `json:"detail,omitempty"`
}

type EvidenceEntry struct {
    Role        string `json:"role"`
    Status      string `json:"status"`
    HandoffTo   string `json:"handoffTo,omitempty"`
    GeneratedAt string `json:"generatedAt,omitempty"`
    Path        string `json:"path"`
}

type VerdictPacket struct {
    Schema   string          `json:"schema"`   // "centinela.verdict/v1"
    Run      RunInfo         `json:"run"`
    Summary  Summary         `json:"summary"`
    Gates    []GateResult    `json:"gates"`
    Verify   []VerifyCheck   `json:"verify"`
    Evidence []EvidenceEntry `json:"evidence"`
}

// Deps mirror cmd/centinela/verify.go's construction so the assembler stays
// pure (no os.Getwd / time.Now inside).
type Deps struct {
    GeneratedAt string        // RFC3339, injected by the caller
    Headless    bool          // resolved headless state, injected
    Filter      *gitdiff.Set  // gate filter; nil = full scan
    VerifyDeps  verify.Deps   // Root + Runner, built by cmd
}

func AssembleVerdict(feature string, cfg *config.Config, wf *workflow.Workflow,
    deps Deps) (*VerdictPacket, error)
```

`AssembleVerdict`:
1. `gates.RunWithFilter(cfg, deps.Filter)` → map each `gates.Result` to a
   `GateResult` (status enum → lowercase string via a small mapper).
2. `verify.Verify(feature, wf.CurrentStep, cfg, deps.VerifyDeps)` → map each
   `Check` to a `VerifyCheck`.
3. Enumerate evidence: for each `orchestration.RequiredRolesForFeature(feature,
   wf.CurrentStep)` *and* every `.workflow/<feature>-<role>.json` on disk, call
   `evidence.Read` and record an `EvidenceEntry`. (Decision: index *all*
   on-disk role evidence for the feature, not just the current step's required
   roles, so a reviewer sees the full produced trail. Missing-but-required roles
   are still reflected via the verify checks / gate results, not invented here.)
4. Snapshot `RunInfo` from `wf` + `workflow.EffectiveProfile(wf, cfg)` +
   `deps.Headless` + `deps.GeneratedAt`.
5. Compute `Summary`: `verdict = "fail"` iff any gate Failed OR
   `verify.VerificationResult.HasFailures()`; else `"pass"`. `ExitCode` 1/0
   accordingly. `Counts` tallied from gates and from `VerificationResult.Tally()`.

**Deterministic JSON:** struct field order is the marshal order for Go's
`encoding/json`; gates/verify/evidence slices are emitted in their natural
deterministic order (gates in `RunWithFilter` order, verify in role/check order,
evidence sorted by role name). A `MarshalIndent(packet, "", "  ")` in the
command yields stable, diffable output. No maps are marshaled directly.

#### Example packet

```json
{
  "schema": "centinela.verdict/v1",
  "run": {
    "feature": "headless-governance",
    "step": "validate",
    "profile": "strict",
    "archetype": "canonical",
    "headless": true,
    "generatedAt": "2026-06-13T10:00:00Z"
  },
  "summary": {
    "verdict": "fail",
    "exitCode": 1,
    "gates": { "pass": 5, "fail": 1, "warn": 0, "skip": 0 },
    "verify": { "pass": 3, "fail": 0, "warn": 1, "skip": 0 }
  },
  "gates": [
    { "name": "file-size", "status": "pass" },
    { "name": "import_graph", "status": "fail",
      "message": "forbidden import",
      "details": ["internal/config imports internal/ui"] }
  ],
  "verify": [
    { "role": "validation-specialist", "claim": "tests-pass",
      "status": "PASS" },
    { "role": "validation-specialist", "claim": "coverage",
      "status": "WARN", "detail": "claimed 95.0, measured 94.2" }
  ],
  "evidence": [
    { "role": "big-thinker", "status": "done",
      "handoffTo": "feature-specialist",
      "generatedAt": "2026-06-13T10:00:00Z",
      "path": ".workflow/headless-governance-big-thinker.json" }
  ]
}
```

#### Exit-code semantics

- `0` — `verdict = "pass"`: no gate `Fail` and no blocking verify failure.
- `1` — `verdict = "fail"`: at least one gate `Fail` OR
  `VerificationResult.HasFailures()` (FAIL / CONFIG-ERROR / TIMEOUT).
- Warnings (gate `Warn`, verify `WARN`) are reported in the packet but do NOT
  affect the exit code in v1. `fail_on_warning` is a deferred follow-up.
- The JSON packet is always written to **stdout** even on exit 1 (a fleet
  consumer must read the failing verdict). Any human-styled status line goes to
  **stderr**.

#### Implementation plan — files (package · ≤100-line budget)

New files:

| File | Pkg | Budget | Contents |
|------|-----|-------:|----------|
| `internal/config/headless.go` | config | ~35 | `HeadlessConfig{Enabled, DetectCI}` struct (toml `enabled`, `detect_ci`); `IsHeadless(cfg) bool` resolver (env > enabled > detect_ci&CI). |
| `internal/verdict/packet.go` | verdict | ~70 | `VerdictPacket`, `RunInfo`, `Summary`, `Counts`, `GateResult`, `VerifyCheck`, `EvidenceEntry`, `Deps` struct defs + json tags. |
| `internal/verdict/assemble.go` | verdict | ~85 | `AssembleVerdict(...)` orchestration (gates, verify, evidence index, summary). |
| `internal/verdict/mappers.go` | verdict | ~55 | `gates.Status`→string, summary computation, evidence enumeration + sort helper. |
| `cmd/centinela/verdict.go` | main | ~70 | `verdict <feature>` cobra cmd: load cfg+wf, build Deps (GeneratedAt=time.Now().UTC().Format(RFC3339), Headless via resolver+flag, VerifyDeps via verifyRoot/NewExecRunner, Filter=nil full scan), `json.MarshalIndent` to stdout, exit code via `os.Exit(packet.Summary.ExitCode)` (or returning a coded error). `--headless` flag. |

Changed files:

| File | Change | Budget impact |
|------|--------|---------------|
| `internal/config/config.go` | Add `Headless HeadlessConfig \`toml:"headless"\`` to `Config`. | +1 line (under 100). |
| `internal/config/defaults.go` | No normalization needed (bools default false = off); no change unless a guard is wanted. | none. |
| `cmd/centinela/hook_context_review_mode.go` | `shouldRenderReviewPrompt`: add `if config.IsHeadless(cfg) { return false }` as the first check after the nil/done guard. | +3 lines. |
| `cmd/centinela/hook_plan_advisor.go` | `runHookPlanAdvisor`: add `if config.IsHeadless(cfg) { return nil }` before `loadActiveWorkflows`. | +3 lines. |

**Layer/G2 check:** `internal/verdict` is a new DOMAIN package importing
config (leaf), gitdiff (leaf), orchestration (leaf), workflow (domain), gates
(domain), verify (domain), evidence. It must NOT import `internal/ui` or `cmd/`.
`internal/verify` already imports config/evidence/orchestration/worktree, so
verdict→verify is fine. This new package must be **added to the `import_graph`
gate's layer map** in `centinela.toml` (DOMAIN tier) or it surfaces as an
unassigned-package warning. `cmd/centinela/verdict.go` stays thin: it only wires
Deps and marshals — all assembly logic lives in `internal/verdict` (G7).

**G1 file-size:** every new file is budgeted ≤100 lines; `packet.go` (structs)
and `assemble.go` (logic) are split deliberately. If `assemble.go` grows past
100, split the evidence-enumeration into `mappers.go` (already planned).

#### JSON shape notes

- `schema: "centinela.verdict/v1"` is the first field — lets consumers
  version-gate. Bump to `/v2` on any breaking shape change.
- Verify statuses are kept in their native UPPERCASE form
  (`PASS`/`WARN`/...) to match `internal/verify`; gate statuses are lowercased
  (`pass`/`fail`/...) to match how the rest of the packet reads. (Decision:
  preserve each subsystem's native vocabulary rather than force a single casing
  — documented so consumers aren't surprised.)

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|------------|
| **Headless short-circuit regresses the review/advisor hooks for non-headless users** | High — silent loss of confirmation prompts breaks the core governance UX | Low | `IsHeadless` returns false for the entire default/zero-config path (no env, `enabled=false`, `detect_ci=false`). Unit-test the resolver's full precedence truth table; add a regression test asserting `shouldRenderReviewPrompt` is unchanged when headless is off. |
| **`detect_ci` always-on would change behavior for local `CI=1` shells** | Medium | Medium (if shipped always-on) | Ship `detect_ci` opt-in; default false. Documented. |
| **Headless overrides an explicit `step_confirmation_mode=every_step` the user set** | Medium — surprising precedence | Medium | This is *intended* (headless = unattended contract) but must be documented in the brief and the `[headless]` config comment. Short-circuit lives before the per-knob resolver so the override is auditable. |
| **`AssembleVerdict` runs gates + full verify (re-runs tests via verify)** — could be slow / have side effects in CI | Medium | Medium | Reuse `verify.Deps.PriorTestRun` plumbing if a caller already ran the suite (future); for v1 the `verdict` command runs them fresh and documents that it executes read-only checks. Filter defaults to full scan (`nil`) for a complete verdict. |
| **New `internal/verdict` package not added to import_graph layer map** | Low — surfaces as warning, not failure; but muddies the gate | Low | Add verdict to the DOMAIN tier in `centinela.toml` as part of the code step; gatekeeper checks the import_graph gate. |
| **Non-deterministic JSON (map iteration / time.Now inside)** | Medium — breaks diffable/cacheable output | Low | Inject `GeneratedAt`; marshal only structs/slices (no maps); sort evidence by role. Add a golden-file test asserting byte-stable output for fixed inputs. |
| **`os.Exit` in the command bypasses cobra's error path / defers** | Low | Low | Return a sentinel error carrying the exit code and translate in `main` (matching existing patterns), or `os.Exit` only after all output is flushed. Decide in code step; prefer the sentinel-error path for testability. |

#### Rollout (slices — smallest correct first)

1. **Slice 1 — Headless umbrella (non-interactive parity), no verdict.**
   `internal/config/headless.go` (`HeadlessConfig` + `IsHeadless`), wire into
   `config.Config`, short-circuit the two hooks. Ships full Deliverable 1 alone;
   zero behavior change when off. Smallest shippable, independently valuable
   (CI/fleet can already go silent with one signal).
2. **Slice 2 — Verdict packet structs + assembler.** `internal/verdict/*` with
   `AssembleVerdict`, unit-tested with injected fakes (fake gates/verify via
   Deps + golden JSON). No command yet.
3. **Slice 3 — `centinela verdict` command.** Wire Deps, marshal, exit code,
   `--headless` flag, surface `run.headless`. Acceptance test runs the command
   against a fixture feature and asserts JSON shape + exit code.
4. **Follow-ups (out of v1):** `--json` on validate/verify; `fail_on_warning`;
   reuse `PriorTestRun` so `complete`→`verdict` doesn't double-run tests.

#### Back-compat note

Default off is byte-identical to today. `[headless]` absent + `CENTINELA_HEADLESS`
unset + `detect_ci=false` ⇒ `IsHeadless` returns false ⇒ both hooks behave
exactly as before. `internal/verify` and `internal/gates` are untouched (read
only). The `verdict` command is purely additive — no existing command changes.
Existing workflow JSON needs no migration (the packet only reads it).

#### Handoff

- **Next role:** feature-specialist.
- **Outstanding questions for the feature-specialist to resolve in the spec /
  detailed design:**
  1. Exit mechanism: sentinel-error-with-code (testable, cobra-friendly) vs
     `os.Exit` after flush. Plan leans sentinel-error.
  2. Evidence index breadth: confirm "all on-disk role evidence for the feature"
     vs "only required roles for the current step." Plan leans all-on-disk for a
     complete review trail.
  3. Whether `verdict` should accept `--changed`/`--full` like `validate` to
     scope the gate filter, or always full-scan for a complete verdict. Plan
     leans always full-scan in v1 (a verdict should be comprehensive); note the
     flags as a follow-up.
  4. Confirm the `[headless]` config comment wording that documents the
     "headless wins over explicit knobs" precedence.
