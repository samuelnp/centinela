# Feature-Specialist Report: governance-telemetry

**Date:** 2026-06-13
**Role:** feature-specialist (plan step, after big-thinker)
**Handoff to:** senior-engineer

## Resolved open questions (big-thinker leans confirmed)

1. **`.gitignore` targets the directory `.workflow/telemetry/` in THIS repo only — CONFIRMED.**
   Exact line to add (a trailing-slash directory entry, so the local log still
   exists for dogfooding but never enters a commit / per-step auto-commit):

   ```gitignore
   # Local governance telemetry — dogfooded, not committed in this repo
   .workflow/telemetry/
   ```

   This repo already ignores `.worktrees/` only; `.workflow/` is otherwise
   git-tracked. The feature CONTRACT for other projects is unchanged: telemetry
   is git-trackable and default-on; only this repo opts its own log out of VCS to
   keep `commitStep` diffs clean.

2. **`step-advanced` carries the JUST-COMPLETED step — CONFIRMED.**
   It is recorded at the same point as `memory.Capture(feature, current, cfg)`
   (after `saveWorkflow` succeeds), using `current` (the step that just
   advanced). This lets a reader bracket a rework window: the run of
   `complete-rejected{feature,step}` events terminates at the `step-advanced`
   carrying that SAME `step`. Recording the next step would break that pairing.

3. **`Read` is LENIENT — CONFIRMED.** Unparseable lines are skipped, valid events
   still returned; a missing file is `(nil, nil)`. Rationale: the log is
   append-only and merged across worktrees, so a partial/garbage line (crash mid
   write, merge artifact) must never poison the whole read for the 5 downstream
   readers. A strict reader variant is deferred to a reader-side feature.

4. **`gate-failure` WITHOUT a feature field is ACCEPTABLE for v1 — CONFIRMED, with
   the join contract documented.** `validate` is not feature-scoped, so the
   `gate-failure` event carries only the WHAT (`gate`, `message`). The WHEN/WHERE
   (feature, step) lives on the co-occurring `complete-rejected{reason:"gates"}`
   event emitted in the same `centinela complete` invocation.

   **Documented join contract for downstream readers
   (centinela-insights / failure-ledger-plan-advisor):** to attribute a
   `gate-failure` to a feature, join it to the nearest following
   `complete-rejected{reason:"gates"}` event by timestamp proximity within the
   same log (same worktree, same process). Both are written microseconds apart in
   one `runComplete` call, in order: gate-failure(s) first (during the validate
   sub-run), then the single complete-rejected{gates}. A reader pairs the trailing
   complete-rejected with the immediately preceding contiguous run of gate-failure
   events. A bare `validate` invocation (no `complete`) yields gate-failure events
   with no complete-rejected partner — those are correctly feature-unattributed
   and contribute to global "which gates bite most" stats only.

## Behavior Summary

`internal/telemetry` is a leaf package (imports only `internal/config` + stdlib)
that appends governance events to `.workflow/telemetry/events.jsonl`, one JSON
object per line, via `os.OpenFile(O_APPEND|O_CREATE|O_WRONLY, 0o644)` (atomic
per-line on local FS, worktree-safe, no flock). `Record(cfg, Event)` is
non-blocking and best-effort: it is a no-op when `cfg==nil || !IsEnabled()`,
stamps `Schema=SchemaID` and an RFC3339-UTC `Timestamp`, marshals the flat
`Event`, appends a line, and swallows all I/O errors to a stderr warning —
exactly mirroring `memory.Capture`. Five typed constructors
(`RecordBlock/RecordGateFailure/RecordVerifyRejection/RecordCompleteRejected/RecordStepAdvanced`)
keep `cmd/` thin. `Read(dir)` is the lenient downstream reader: skips
unparseable lines, returns `(nil,nil)` on a missing file.

Five event types are emitted from seven `cmd/` chokepoints only (domain stays
side-effect-free, G7): `block` (need-init | out-of-step, with fileType +
targetPath; out-of-step also carries feature+step), `gate-failure` (gate +
message, one per Fail, no feature), `verify-rejection` (feature + step +
failing checks as owned `CheckRef` copies), `complete-rejected`
(reason gates|verify, feature + step), `step-advanced` (feature + just-completed
step). `rework` is DERIVED, never stored: N `complete-rejected` for a
(feature,step) before the `step-advanced` for that same (feature,step).

Config: `[telemetry] enabled` (`*bool`, default ON / opt-out), wired like
`[memory]`. Every line is self-describing via `schema="centinela.telemetry/v1"`
(matches the shipped `centinela.verdict/v1` convention). The only observable
change with telemetry on is the existence of the gitignored log file — no exit
code, block decision, or advance outcome changes.

## Gherkin Scenarios

Full spec: `specs/governance-telemetry.feature` (18 scenarios, stable titles for
`// Scenario:` traceability).

| Scenario | Covers |
|----------|--------|
| Out-of-step write appends a block event with full context | block / out-of-step |
| Write with no active workflow appends a need-init block event | block / need-init |
| A failing gate during validate appends a gate-failure event | gate-failure (no feature) |
| Each failing gate appends its own gate-failure event | one event per Fail |
| A failed claim verification appends a verify-rejection event with the failing checks | verify-rejection + CheckRef |
| An advance aborted by validate gates appends complete-rejected with reason gates | complete-rejected / gates |
| An advance aborted by verification appends complete-rejected with reason verify | complete-rejected / verify |
| A successful advance appends a step-advanced event carrying the just-completed step | step-advanced (just-completed) |
| Telemetry disabled is a no-op and writes no file | disabled no-op |
| Absent telemetry config defaults to enabled and records events | default-on (opt-out) |
| Every recorded event carries the schema id and an RFC3339 timestamp | schema + timestamp contract |
| Multiple events accumulate append-only in call order | append-only ordering |
| Two sequential records both land intact under append-only writes | worktree-safe append |
| An I/O error while recording does not fail the host command | non-blocking / best-effort |
| Read skips a corrupt line and returns the valid events | lenient Read |
| Read of a missing telemetry log returns no events and no error | missing file (nil,nil) |
| Rework is derivable from two complete-rejected events before a step-advanced | derived rework |

## UX States (CLI / file surfaces)

- **Loading:** n/a — synchronous single `OpenFile`+`Write`+`Close`; no spinner,
  no read on the write path.
- **Empty:** missing log file ⇒ `Read` returns `(nil, nil)`; disabled config ⇒
  no file ever created.
- **Error:** unwritable dir / marshal failure ⇒ `Record` returns nothing, warns
  `[telemetry] warning: ...` to stderr, host command exit code / block decision /
  advance outcome unchanged. Corrupt line in the log ⇒ `Read` skips it, returns
  the valid events.
- **Success:** one JSON line appended per event, self-describing via `schema`,
  ordered append-only; `Read` returns the events in file order.
- **Stream note:** events go to the file only; the human-facing governance render
  (block/gate/verify messages) is unchanged — telemetry is a silent side effect.

## Out of Scope (v1)

- External sinks / daemon / network emission.
- Aggregation / reporting / dashboard surfaces — those ARE the 5 downstream
  features (centinela-insights, team-dashboard, etc.); v1 only writes + exposes
  a reader API.
- A workflow attempt counter or backward step transitions — rework is derived,
  not stored.
- Schema migration tooling — v1 is `v1`; the `schema` field exists so a future
  reader can branch, nothing more.
- A strict (error-on-bad-line) reader variant — deferred to a reader-side feature.
- Per-event redaction / PII policy — events carry only paths, gate names, and
  claim metadata already visible in the repo.
- Importing `verify.Check` — telemetry owns a flat `CheckRef` COPY to stay a leaf.

## Handoff

- **Next role:** senior-engineer.
- **Outputs:** `specs/governance-telemetry.feature`, this report, the plan.
- **Build order (per big-thinker slices):** S1 contract + storage + config + leaf
  layer edit + `.gitignore` (ships the reader API/schema to downstream
  immediately, no call-sites); S2 gate-failure + complete-rejected + step-advanced
  (cold paths); S3 verify-rejection; S4 block in prewrite (hottest path, last,
  with a benchmark; emit BEFORE `exitPrewrite(2)`).
- **Watch-outs:** keep each `internal/telemetry/*.go` ≤100 lines (split
  constructors if needed); overridable `now` clock var for deterministic
  timestamp tests; `toCheckRefs` mapper lives in `cmd/` (not telemetry);
  `appendLine` must `os.MkdirAll(LogDir, 0o755)` first; `Record` must never
  return an error or alter control flow.
