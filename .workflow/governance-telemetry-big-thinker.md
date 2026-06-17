# Plan: governance-telemetry

### Big-Thinker Report: governance-telemetry

**Date:** 2026-06-13

#### Problem

Centinela emits governance signals (block, gate-failure, verify-rejection,
refused advance) as one-shot stderr renders and then forgets them. There is no
durable record of governance friction. Five downstream roadmap features
(centinela-insights, failure-ledger-plan-advisor, capability-calibration,
team-dashboard, adaptive-skill-synthesis) are blocked on a data source that does
not exist. Because five readers consume it, the event schema is a **long-lived
contract** — churn cost is multiplied by five, so v1 must get the schema right
and version it.

#### Scope (In / Out)

**In (v1):** leaf package `internal/telemetry` (versioned `Event` contract,
non-blocking `Record`, typed constructors, `Read` reader); 5 event types;
append-only `.workflow/telemetry/events.jsonl`; `[telemetry] enabled` (`*bool`,
default ON); emission wired into `hook_prewrite.go`, `validate.go`,
`complete.go`; `internal/telemetry/**` added to the `leaf` layer.

**Out (v1):** external sinks/daemon/network; aggregation/reporting surfaces
(those *are* the downstream features); a workflow attempt counter or backward
transitions (rework is derived, not stored); schema migration tooling;
redaction policy.

#### Dependencies & Assumptions

- Depends on the shipped `governed-project-memory` non-blocking capture contract
  and `*bool` opt-out config; and on `headless-governance`'s
  `schema: "centinela.<x>/v1"` string-version precedent.
- **Verified assumptions (ground truth):**
  - `.workflow/` is git-tracked (only `.worktrees/` is ignored — confirmed in
    `.gitignore`); each worktree has its own `.workflow/`, so appends merge
    cleanly (additions only, no shared mutable index).
  - Emission chokepoints all live in `cmd/`; the domain types they read
    (`hookpolicy.PrewriteDecision`, `gates.Result`, `verify.Check`) are
    side-effect-free and unmodified.
  - `internal/telemetry` will import only `internal/config` + stdlib ⇒ qualifies
    as a leaf; no domain→telemetry edge is introduced (emit from `cmd/` only).
  - No JSONL/append helper exists; `O_APPEND|O_CREATE|O_WRONLY` single-line
    writes are atomic per `write()` on local FS — no flock needed.

#### Event Schema (the contract)

Versioned, append-only JSONL — one event per line. A **flat** struct with typed
`omitempty` fields (no marshaled maps ⇒ deterministic key order, trivial for 5
readers). The only nested type is the verify-rejection check list.

```go
// Package telemetry: governance event log. See SchemaID for the contract version.
const SchemaID = "centinela.telemetry/v1"

// EventType enumerates the v1 governance events.
type EventType string

const (
    EventBlock            EventType = "block"             // out-of-step / no-workflow write block
    EventGateFailure      EventType = "gate-failure"      // a validate gate returned Fail
    EventVerifyRejection  EventType = "verify-rejection"  // claim verification hard-blocked the advance
    EventCompleteRejected EventType = "complete-rejected" // an advance was refused (gates|verify)
    EventStepAdvanced     EventType = "step-advanced"     // an advance succeeded (brackets rework windows)
)

// Block reason values.
const (
    ReasonOutOfStep = "out-of-step" // a workflow is active but the file is wrong for the step
    ReasonNeedInit  = "need-init"   // no active workflow (NeedInit) — write needs `centinela start`
)
// complete-rejected reason values: "gates" | "verify".

// Event is one governance event. The JSON shape is the long-lived contract for
// the 5 downstream readers; add fields with omitempty, never repurpose one.
type Event struct {
    Schema     string     `json:"schema"`               // always SchemaID — per-line, self-describing
    Type       EventType  `json:"type"`                 // one of the EventType constants
    Timestamp  string     `json:"timestamp"`            // RFC3339 UTC, stamped by Record
    Feature    string     `json:"feature,omitempty"`    // feature slug (absent on need-init blocks)
    Step       string     `json:"step,omitempty"`       // workflow step the event occurred in
    Reason     string     `json:"reason,omitempty"`     // block: out-of-step|need-init; complete-rejected: gates|verify
    FileType   string     `json:"fileType,omitempty"`   // block: classified file type
    TargetPath string     `json:"targetPath,omitempty"` // block: the path that was refused
    Gate       string     `json:"gate,omitempty"`       // gate-failure: gates.Result.Name
    Message    string     `json:"message,omitempty"`    // gate-failure: gates.Result.Message
    Checks     []CheckRef `json:"checks,omitempty"`     // verify-rejection: the blocking claims
}

// CheckRef is a flattened, JSON-stable copy of a failing verify.Check. We copy
// (not import verify's type) so the contract is owned by telemetry and the
// package stays a leaf (no internal/verify import).
type CheckRef struct {
    Claim  string `json:"claim"`
    Role   string `json:"role"`
    Status string `json:"status"`
    Detail string `json:"detail,omitempty"`
}
```

**Why these decisions:**

- **String schema id `centinela.telemetry/v1`, not `schemaVersion: 1` int.**
  Diverges from the orchestrator's lean. Reason: the just-shipped verdict packet
  established `schema: "centinela.verdict/v1"` (verified in
  `internal/verdict/packet.go` / `assemble.go`). Five readers plus the verdict
  reader benefit from one self-describing, greppable convention. Consistency wins
  over one saved byte per line.
- **Per-line `schema` field**, not a file header. The file is append-only and
  git-merged across worktrees; a header can't be guaranteed first or unique after
  merges. Self-describing lines are robust to concatenation.
- **Flat + typed + `omitempty`**, no `map[string]any`. Deterministic key order
  (Go marshals struct fields in declaration order) ⇒ golden-testable, trivial to
  parse in 5 places, and `go vet`-checkable field names.
- **`CheckRef` is owned, not imported.** Copying the 4 fields keeps telemetry a
  pure leaf (config + stdlib only) and freezes the contract under telemetry's
  control rather than coupling it to `verify`'s internal struct.

#### Divergence from proposed design (with reasons)

1. **Schema version style: adopt string `centinela.telemetry/v1`** (orchestrator
   leaned int `schemaVersion:1`). Reason above — matches the shipped verdict
   precedent; consistency across machine artifacts.
2. **Keep all 5 event types, including `step-advanced`.** Not YAGNI: `rework` is
   explicitly a *derived* metric and the derivation
   ("N `complete-rejected` for (feature,step) before it advances") needs a
   terminator. Without `step-advanced`, a reader can't tell "3 rejections then
   succeeded" from "3 rejections, still stuck". It is one cheap line appended
   right where `memory.Capture` already runs. Adopt as proposed.
3. **"rework" stays derived — no workflow state.** Confirmed: `internal/workflow`
   has no attempt counter and adding one would be scope creep into the domain.
   Adopt as proposed.
4. **Block emission: only on genuine blocks, before `os.Exit`.** Adopt. The
   `d.Allow` early-return already short-circuits allowed writes, so we never emit
   for them — zero added work on the allow path. We emit the two block branches
   immediately before each `exitPrewrite(2)`.
5. **THIS-repo storage decision: default-on + `.gitignore` `.workflow/telemetry/`
   in this repo only (option b).** See dedicated note below. The *feature
   contract* (default-on, git-trackable) is unchanged for other projects.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|------------|
| Hot-hook slowdown — prewrite is the most-fired hook | Perceptible latency on every blocked write | Low | One `OpenFile`+`Write`+`Close` of a sub-1KB line, only on the block path (not allow). No flush/fsync, no read. Emit before `os.Exit`. Benchmark in tests. |
| Schema churn breaks 5 dependents | A v1 field rename forces 5 coordinated changes | Medium | Freeze the contract: golden JSON test; documented append-only rule (new fields `omitempty`, never repurpose); owned `CheckRef`; `schema` version string lets readers branch. |
| Commit noise — per-step auto-commit bundles telemetry lines | Noisy diffs / merge churn in *this* repo | Medium | `.gitignore` `.workflow/telemetry/` in this repo only; the events file still exists locally for dogfooding. Other projects keep it git-tracked. |
| Emission throws / partial line on crash | Corrupt JSONL line | Low | `O_APPEND` makes each `write` atomic on local FS; reader skips unparseable lines (lenient `Read`). Errors swallowed to stderr like memory. |
| Concurrent worktrees writing | Interleaved/lost lines | Low | Per-worktree `.workflow/` ⇒ no shared file; single-line `O_APPEND` is atomic within a worktree. |
| Telemetry accidentally becomes blocking | Breaks a flow on disk-full / perms | Low | `Record` returns nothing, never errors; all failures → `warn()` stderr. Mirrors `memory.Capture` exactly; covered by a disabled-config + write-error test. |

#### Rollout (slices, smallest first)

1. **S1 — contract + storage + config (no call-sites).** `internal/telemetry`
   (`event.go`, `record.go`, `read.go`), `internal/config/telemetry.go`, wiring
   in `config.go` + `defaults.go`, leaf-layer edit in `centinela.toml`,
   `.gitignore` entry. Unit-tested in isolation. **Ships value to downstream
   readers immediately** (the reader API + schema exist even before any emitter).
2. **S2 — gate-failure + complete-rejected + step-advanced** in `validate.go` /
   `complete.go` (cold paths, richest context, lowest risk).
3. **S3 — verify-rejection** in `complete.go` (`runClaimVerification`).
4. **S4 — block** in `hook_prewrite.go` (hottest path, last — most care, with a
   benchmark).

#### Handoff

**Next role:** feature-specialist.

**Outstanding questions for the specialist:**

- Confirm `.gitignore` line targets `.workflow/telemetry/` (directory) so the
  reader still finds a local file but commits stay clean. (Recommended: yes.)
- Decide whether `step-advanced` carries the *just-completed* step or the *next*
  step. Recommendation: the just-completed step (matches `memory.Capture(feature,
  current, cfg)` which is called at the same point with `current`), so a reader
  pairs it cleanly with `complete-rejected` events for that same step.
- Confirm the reader `Read` is lenient (skip unparseable lines) vs strict.
  Recommendation: lenient — robust to merge artifacts; a strict variant can come
  with a reader-side feature later.

---

## Implementation Plan

### Event contract

As specified in **Event Schema** above. `SchemaID = "centinela.telemetry/v1"`.
`Record` stamps `Schema` and `Timestamp`; call-sites never set them.

### `internal/telemetry` API + file split (each ≤100 lines)

**`event.go`** — the contract: `SchemaID`, `EventType` + constants, reason
constants, `Event`, `CheckRef`. No logic.

**`record.go`** — emission:

```go
// now is overridable for deterministic tests.
var now = func() time.Time { return time.Now().UTC() }

// Record appends one event to the telemetry log. It never errors and never
// blocks: nil/disabled config is a no-op; all I/O failures warn to stderr.
// Mirrors memory.Capture's non-blocking contract.
func Record(cfg *config.Config, e Event) {
    if cfg == nil || !cfg.Telemetry.IsEnabled() {
        return
    }
    e.Schema = SchemaID
    e.Timestamp = now().Format(time.RFC3339)
    line, err := json.Marshal(e)
    if err != nil { warn("marshal: %v", err); return }
    if err := appendLine(line); err != nil { warn("append: %v", err) }
}

func appendLine(line []byte) error // OpenFile(LogPath, O_APPEND|O_CREATE|O_WRONLY, 0o644), write line+"\n", Close
func warn(format string, args ...any) // fmt.Fprintf(os.Stderr, "[telemetry] warning: "+format+"\n", ...)
```

**`constructors.go`** — one-liner call-site helpers so `cmd/` stays thin (G7):

```go
func RecordBlock(cfg *config.Config, reason, fileType, feature, step, targetPath string)
func RecordGateFailure(cfg *config.Config, feature, gate, message string)
func RecordVerifyRejection(cfg *config.Config, feature, step string, checks []CheckRef)
func RecordCompleteRejected(cfg *config.Config, feature, step, reason string)
func RecordStepAdvanced(cfg *config.Config, feature, step string)
```

Each builds an `Event` and calls `Record`. (If this file approaches 100 lines,
split into `constructors_block.go` / `constructors_complete.go`.)

**`read.go`** — the downstream reader contract:

```go
const LogDir = ".workflow/telemetry"
const LogFile = "events.jsonl"
func LogPath() string // filepath.Join(LogDir, LogFile)

// Read parses all events from dir's log. Lenient: unparseable lines are skipped.
// Missing file ⇒ (nil, nil). Used by tests and the 5 downstream readers.
func Read(dir string) ([]Event, error)
```

### Config

**`internal/config/telemetry.go`** (mirror `memory.go`):

```go
type TelemetryConfig struct {
    Enabled *bool `toml:"enabled"` // default true (opt-out)
}
func (t TelemetryConfig) IsEnabled() bool { return t.Enabled == nil || *t.Enabled }
```

- `config.go`: add `Telemetry TelemetryConfig \`toml:"telemetry"\`` after
  `Headless`.
- `defaults.go`: no normalization needed (the `*bool` default-on is handled by
  `IsEnabled`); no `applyTelemetryDefaults` required unless future caps are
  added.

### centinela.toml leaf-layer edit

```toml
[[gates.import_graph.layers]]
name  = "leaf"
paths = ["internal/config/**", "internal/gitdiff/**", "internal/orchestration/**", "internal/telemetry/**"]
allow = []
```

### Storage

- Path: `.workflow/telemetry/events.jsonl` (one event per line).
- Mechanism: `os.OpenFile(LogPath(), O_APPEND|O_CREATE|O_WRONLY, 0o644)`, write
  `line + "\n"`, close. No flock, no temp+rename, no fsync. `appendLine` must
  `os.MkdirAll(LogDir, 0o755)` first.

### THIS-repo commit decision

**Option (b): default-on + `.gitignore` `.workflow/telemetry/` in this repo
only.** Rationale: default-on matches `[memory]` and the feature's
git-trackable intent and feeds local dogfooding; but per-step auto-commit
(`commitStep`) would otherwise bundle telemetry lines into every feature commit,
adding diff/merge noise. Gitignoring the directory *in this repo* keeps commits
clean while the log still exists locally. The **feature contract for other
projects is unchanged** — telemetry is git-trackable and default-on; only this
repo opts its own log out of version control. Add to `.gitignore`:

```
# Local governance telemetry — dogfooded, not committed in this repo
.workflow/telemetry/
```

### Emission call-site table

| # | File | Location / anchor | Event | Payload |
|---|------|-------------------|-------|---------|
| 1 | `cmd/centinela/hook_prewrite.go` | `runHookPrewrite`, in the `d.NeedInit` branch, immediately before `exitPrewrite(2)` (line ~68) | `block` | `RecordBlock(cfg, ReasonNeedInit, string(d.FileType), "", "", filePath)` |
| 2 | `cmd/centinela/hook_prewrite.go` | `runHookPrewrite`, the out-of-step branch, immediately before the final `exitPrewrite(2)` (line ~71) | `block` | `RecordBlock(cfg, ReasonOutOfStep, string(d.FileType), d.Feature, d.Step, filePath)` |
| 3 | `cmd/centinela/validate.go` | `executeValidationWithFlag`, in the `for _, r := range results` loop, when `r.Status == gates.Fail` | `gate-failure` | `RecordGateFailure(cfg, "", r.Name, r.Message)` (feature unknown at validate; left empty — see note) |
| 4 | `cmd/centinela/complete.go` | `runClaimVerification`, in the `if res.HasFailures()` branch before the `return` | `verify-rejection` | `RecordVerifyRejection(cfg, feature, step, toCheckRefs(res.Failed()))` |
| 5 | `cmd/centinela/complete.go` | `runComplete`, the `current=="validate"` block: when `executeValidation()` returns err | `complete-rejected` | `RecordCompleteRejected(cfg, feature, current, "gates")` |
| 6 | `cmd/centinela/complete.go` | `runComplete`, when `runClaimVerification(...)` returns err | `complete-rejected` | `RecordCompleteRejected(cfg, feature, current, "verify")` |
| 7 | `cmd/centinela/complete.go` | `runComplete`, after `saveWorkflow(wf)` succeeds, alongside `memory.Capture(feature, current, cfg)` (line ~69) | `step-advanced` | `RecordStepAdvanced(cfg, feature, current)` |

**Notes on the table:**

- `validate.go` loads `cfg` already; `gate-failure` `feature` is empty because
  `validate` is not feature-scoped. The richer feature-scoped duplication is the
  `complete-rejected{reason:"gates"}` event (#5), so the gate-failure events
  carry the *what* (gate, message) and complete-rejected carries the *when/where*
  (feature, step). Downstream joins on timestamp proximity if needed.
- `prewrite.go` already loads `cfg` (defaulting to `&config.Config{}` on error);
  `Record(nil-safe)` no-ops if telemetry can't resolve — non-fatal preserved.
- `toCheckRefs` is a tiny `cmd/`-local mapper from `[]verify.Check` to
  `[]telemetry.CheckRef` (keeps telemetry leaf; cmd already imports verify).
- All seven sites are best-effort: `Record` returns nothing; in prewrite it runs
  *before* `os.Exit` so the process image is still intact.

### Back-compat note

- **Default-on must not break existing flows.** `Record` is a no-op when disabled
  and never returns an error or changes control flow. The only observable change
  with telemetry on is the existence of `.workflow/telemetry/events.jsonl`
  (gitignored in this repo). No exit code, block decision, or advance outcome
  changes — identical to the `memory.Capture` guarantee.
- **Hooks stay non-fatal.** Block events are written before `exitPrewrite(2)`;
  an I/O failure warns to stderr and the block still proceeds. The hot path only
  does work when it was already going to block.
- **Downstream reader contract.** `telemetry.Read(dir)` + the frozen `Event` /
  `CheckRef` JSON shape + `SchemaID` are the stable surface the 5 features build
  on. The append-only rule (add `omitempty` fields, never repurpose) and the
  per-line `schema` string let future readers branch without breaking v1.
