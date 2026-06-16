# Implementation Plan — custom-gate-sdk

> Feature brief: `docs/features/custom-gate-sdk.md`.
> Spec: `specs/custom-gate-sdk.feature`.

Phase 8 turns Centinela from a closed set of opinionated gates into a **policy
engine**. Today every gate is a hardcoded `if cfg.Gates.X.Enabled { checkX(…) }`
branch in `gates.RunWithFilter`; a team's project-specific rule ("no
`console.log` in `src/`") has nowhere to live except `[validate] commands`, where
it runs as an opaque pass/fail shell line with no severity, no structured
violations, no telemetry, and no participation in the audit baseline/ratchet.

This feature adds a `[[gates.custom]]` config surface: command-backed gates that
produce the **exact same** `gates.Result{Name, Status, Message, Details}` as the
built-ins and therefore flow, with zero new rendering/telemetry code, through
`ui.RenderGateResult`, `emitGateFailures` (`gate-failure` events), `warn`/`fail`
severity, and — crucially — the `audit-baseline-ratchet` (their `Details`
fingerprint like any gate's). The "SDK" is the documented **config schema + the
`Result` contract**, not a Go-plugin API. The built-ins are declared the
reference implementations of that contract; none are rewritten.

## Decisions (DECIDED)

1. **Additive runner, NOT a `Gate` interface/registry refactor (the scope call —
   see Interface CALL-OUT).** Built-ins keep their hardcoded chain verbatim. One
   new runner, `customGates(cfg, filter) []Result`, is appended at the tail of
   `RunWithFilter`. The shared contract is already `Result`; rewriting 7 built-ins
   onto an interface buys zero v1 behavior, multiplies risk, and fights the
   100-line-file rule. **No built-in file is touched.**

2. **Shell exec (`sh -c` / `cmd /C`), not argv (the exec call — see Exec
   CALL-OUT).** Custom rules realistically need pipes, globs, and `&&`
   (`grep -rn console.log src/ | grep -v '//'`). We reuse the **exact**
   `validate.commands` model (`cmd/centinela/validate_runner.go runCommand`),
   adding a per-gate timeout via `exec.CommandContext` (default 60s, configurable
   via `timeout_seconds`). The build gate's `strings.Fields` argv model is
   deliberately rejected — it cannot express the rules teams actually have.

3. **Pass/fail + structured violations.** Exit 0 ⇒ Pass. Non-zero ⇒ Fail (when
   `severity=fail`) or Warn (when `severity=warn`). Timeout ⇒ Fail with a timeout
   message (a hung rule must not false-Pass). Combined stdout+stderr is captured
   into `Details` per the `output` mode (Decision #4).

4. **Two `output` modes (default `blob`).**
   - `output="blob"` (default): the whole combined output becomes **one**
     `Details` entry, trimmed and **truncated to a byte cap** (e.g. 4 KiB) with a
     `… (truncated)` marker so the report stays readable (brief edge: huge
     output bounded).
   - `output="lines"`: each non-empty stdout line becomes **one** `Details`
     entry, bounded to a **max line count** (e.g. 200) with a `… (N more)` final
     entry. This is what makes a custom gate baseline-able **per violation** by
     the ratchet: the audit fingerprinter `Compute`s one `Fingerprint` per
     `Detail`, so `output="lines"` lets `centinela audit baseline` tolerate the
     existing violations individually and block only genuinely new lines, exactly
     like `G1: File Size`. With `blob`, the whole gate is a single fingerprint
     (all-or-nothing baseline).
   - **Empty output on failure** ⇒ a single generic `Details` fallback
     (`"<name> failed (exit N) with no output"`), so the ratchet and report never
     see an empty Fail (brief edge).

5. **Default severity = `fail` (NOT `warn`).** Built-ins default `warn` for safe
   adoption of a rule the team did **not** ask for. A `[[gates.custom]]` entry is
   the opposite: the team *explicitly authored* this rule, so the intent is to
   enforce it. Defaulting `warn` would silently no-op the very rule they just
   added. Teams that want soft rollout set `severity="warn"` explicitly. (This is
   a deliberate divergence from `roadmap_drift`'s `warn` default, justified by the
   opt-in-authorship difference.)

6. **diff-aware via opt-in env var (kept, minimal).** A `[[gates.custom]]` entry
   with `diff_aware=true` receives the changed-file set from the `*gitdiff.Set`
   filter as a newline-separated `CENTINELA_CHANGED_FILES` env var on the child
   process; the command decides how to use it (e.g. `grep $CENTINELA_CHANGED_FILES`).
   When `filter` is nil (full-scan validate) or `diff_aware=false`, the env var is
   **unset** and the command full-scans. This is ~5 lines of env plumbing reusing
   the filter `RunWithFilter` already receives, so it stays in v1. (Centinela does
   **not** filter for the gate — it only hands over the list; mechanical filtering
   inside a shell command is the rule author's job.)

7. **Wiring seam = `gates.RunWithFilter` (domain), no cross-layer edge.** The
   custom runner imports only `internal/config` (leaf — already imported by
   `gates`) and stdlib `os/exec`/`context`. `gates` is the **domain** layer
   (`allow = ["leaf"]`), so this introduces **no new import edge** and **no
   `centinela.toml` `[gates.import_graph]` change**. This is *cleaner than the
   audit gate*, which had to be wired from `cmd/` because `internal/audit` is an
   aggregator that imports `gates`; the custom runner lives **inside** `gates` and
   reads config like every other built-in, so it appends directly in
   `RunWithFilter` (verified against the `domain → leaf` rule in
   `centinela.toml:64-67`).

## Exec & security model (CALL-OUT)

**Model: shell, bounded, trusted-config.** The runner mirrors
`cmd/centinela/validate_runner.go runCommand` (`sh -c` on unix, `cmd /C` on
windows) but wraps it with a timeout (mirroring
`internal/gates/security_exec.go runScanner`'s `exec.CommandContext` +
`DeadlineExceeded` handling):

```go
// runCustom executes one custom gate's command under a shell with a timeout,
// returning the combined stdout+stderr, the exit code, and whether it timed out.
// changed is the diff-aware file list (empty ⇒ env var unset).
func runCustom(command string, timeout time.Duration, changed []string) (output string, code int, timedOut bool)
```

- **Shell:** `exec.CommandContext(ctx, "sh", "-c", command)` (unix) /
  `"cmd", "/C", command` (windows), `ctx` from
  `context.WithTimeout(_, timeout)`. Combined `Stdout`+`Stderr` into one
  `bytes.Buffer` (custom rules print violations to either stream).
- **Timeout:** `ctx.Err() == context.DeadlineExceeded` ⇒ `timedOut=true` ⇒ the
  gate Fails with `"<name> timed out after <T>s"` regardless of severity-on-
  nonzero (a wedged rule is a Fail, never a Pass). Default `timeout_seconds=60`,
  configurable; brief edge "hung command fails the gate" is this branch.
- **Exit code:** extract via the `*exec.ExitError` pattern (mirror
  `security_exec.go exitCode`). `code != 0 && !timedOut` ⇒ Fail/Warn per severity.
  A launch failure (command-not-found, non-executable) surfaces through the shell
  as a non-zero exit with a shell error message in the captured output ⇒ a clean
  Fail with a clear `Details` line, not a panic (brief edge).
- **Diff-aware env:** when `diff_aware && len(changed)>0`, set
  `cmd.Env = append(os.Environ(), "CENTINELA_CHANGED_FILES="+strings.Join(changed, "\n"))`.

**Trust model (for the Risks section):** the command string lives in
`centinela.toml`, which is **checked-in code reviewed like any source**. Custom
gates run arbitrary shell **by design** — identical to `[validate] commands`
today, which already does `sh -c` on user strings. There is **no allowlist, no
sandbox, no privilege change**; we add only a timeout (DoS bound) the existing
`validate.commands` path lacks. Anyone who can edit `centinela.toml` can already
run code via `validate.commands`, so this opens no new privilege surface. This is
documented explicitly in Step 5.

## Interface generalization / scope (CALL-OUT)

**Verdict: additive runner. No built-in is rewritten. No `Gate` interface or
registry is introduced.**

The roadmap framing — "built-ins become reference implementations" — is satisfied
**by documentation and a shared data contract**, not by a code refactor. The
contract that custom gates and built-ins share is **already `gates.Result`**;
both are `func(cfg, filter) []Result`-shaped. A registry/interface refactor would:
(a) touch all 7 built-in gate files + their tests for **zero** v1 behavior gain;
(b) push files over the 100-line cap (the brief lists this as a top risk);
(c) risk regressing the byte-stable validate output. The additive runner gives
teams the full first-class experience (severity, telemetry, ratchet) with a
single new ~3-line `if`-append in `RunWithFilter` and one new runner file.

**`Result` IS the SDK contract.** Step 5 docs declare: a gate is anything
producing `Result{Name, Status, Message, Details}`; the built-ins
(`internal/gates/*.go`) are the canonical reference implementations; a
`[[gates.custom]]` command is the no-code path to the same contract. A future Go
plugin API is **out of scope** (see below).

## v1 scope

**In:**
- `config.CustomGate` struct + `CustomGates []CustomGate \`toml:"custom"\`` on
  `GatesConfig`; `NormalizeCustomGates` (default severity `fail`, output `blob`,
  timeout 60); `validateCustomGates` (indexed errors).
- `internal/gates/custom_command.go` (`customGates`) + `custom_command_exec.go`
  (`runCustom`), appended in `RunWithFilter`.
- `output` modes `blob` (truncated, default) + `lines` (bounded, ratchet-ready).
- Per-gate timeout (Fail on timeout).
- Opt-in `diff_aware` via `CENTINELA_CHANGED_FILES`.
- Telemetry, render, audit-baseline participation: **free** (results are plain
  `Result`s; no new code).

**Out (deferred / explicit non-goals):**
- A **Go-plugin / shared-object gate API** (`plugin` package, `.so` loading) —
  the SDK is config + the `Result` contract only.
- A full **`Gate` interface + registry refactor** of the built-in chain
  (Decision #1 / Interface CALL-OUT) — built-ins stay hardcoded.
- **Deprecating or redirecting `[validate] commands`** — both paths coexist:
  custom gates are the structured/governed path, `validate.commands` stays the
  quick opaque path. No migration is forced in v1.
- **Centinela-side mechanical filtering** for diff-aware gates — we hand over the
  changed-file list via env; filtering is the command's responsibility.
- **An allowlist / sandbox** for commands — trust model is checked-in-config
  (Exec CALL-OUT).

### Audit-ratchet coherence (AC-4, AC-7 — confirmed against ground truth)

`internal/audit` fingerprints each gate's `Result.Details` Fail violations via a
per-gate identity extractor keyed by `Result.Name`, with a **generic fallback**
(`genericKey`: strip a trailing ` (…)` parenthetical + trailing digits) for any
gate it doesn't special-case. A custom gate's `Name` is never special-cased, so
its Details hit the generic fallback — **coherent and intended**:
- `output="lines"` ⇒ one Detail per violation ⇒ one fingerprint per violation ⇒
  `audit baseline` tolerates each existing violation and blocks only new lines
  (AC-7). The generic normalizer keeps a line like `src/a.ts:42 (3 hits)` stable
  if a count changes.
- `output="blob"` ⇒ the whole gate is one fingerprint (whole-gate baseline).
- Custom gates are **detail-emitting**, so they are eligible for the audit
  default participation set; **confirm during code** whether the new gate Name(s)
  must be added to `internal/audit/participation.go`'s default set or whether the
  "empty `target_gates` ⇒ all detail-emitting gates" path already includes them
  (read `participation.go`; if it uses a hardcoded allowlist, custom Names must be
  folded in — note this in the gatekeeper report). This is the one integration
  seam to verify, called out as an explicit AC-4 check.

## Step 2 — code

New / edited source files (each ≤100 lines; **G1 applies to `_test.go` too**):

| File | Change | Budget |
|------|--------|--------|
| `internal/config/custom_gate.go` | NEW. `CustomGate{Enabled bool; Name, Command, Severity, Output string; TimeoutSeconds int; DiffAware bool}` w/ toml tags; `NormalizeCustomGates([]CustomGate) []CustomGate` (trim; default severity `fail`, output `blob`, timeout 60 when ≤0); `validateCustomGates([]CustomGate) error` (indexed) | ~95 |
| `internal/config/config.go` | add `CustomGates []CustomGate \`toml:"custom"\`` to `GatesConfig` | +1 |
| `internal/config/defaults.go` | `cfg.Gates.CustomGates = NormalizeCustomGates(cfg.Gates.CustomGates)` in `applyDefaults` | +1 |
| `internal/config/file_size_exceptions.go` | `if err := validateCustomGates(cfg.Gates.CustomGates); err != nil { return err }` in `validateConfig` | +3 |
| `internal/gates/custom_command.go` | NEW. `customGates(cfg *config.Config, filter *gitdiff.Set) []Result` — loop `cfg.Gates.CustomGates`, skip `!Enabled`, call `runCustom`, map exit/timeout → `Result` via `customResult(g, output, code, timedOut)`; `output`→`Details` mapper (`blobDetails`/`lineDetails`) | ~95 |
| `internal/gates/custom_command_exec.go` | NEW. `runCustom(command string, timeout time.Duration, changed []string) (output string, code int, timedOut bool)` — shell exec w/ `exec.CommandContext`, combined buffer, env injection, `DeadlineExceeded` handling; `customExitCode(err)` helper | ~80 |
| `internal/gates/gates.go` | in `RunWithFilter`, append `results = append(results, customGates(cfg, filter)...)` at the tail (after `roadmap_drift`) — gated by `len(cfg.Gates.CustomGates) > 0` for a byte-identical no-op when none configured | +3 |

**Key signatures / types:**

```go
// internal/config/custom_gate.go
type CustomGate struct {
    Enabled        bool   `toml:"enabled"`
    Name           string `toml:"name"`
    Command        string `toml:"command"`
    Severity       string `toml:"severity"`        // fail | warn  (default fail)
    Output         string `toml:"output"`          // blob | lines (default blob)
    TimeoutSeconds int    `toml:"timeout_seconds"` // default 60
    DiffAware      bool   `toml:"diff_aware"`
}

func NormalizeCustomGates(gs []CustomGate) []CustomGate
func validateCustomGates(gs []CustomGate) error

// internal/gates/custom_command.go
func customGates(cfg *config.Config, filter *gitdiff.Set) []Result
func customResult(g config.CustomGate, output string, code int, timedOut bool) Result

// internal/gates/custom_command_exec.go
func runCustom(command string, timeout time.Duration, changed []string) (output string, code int, timedOut bool)
```

**`validateCustomGates` rules (indexed errors, mirroring
`file_size_exceptions.go`), no-op when an entry is disabled:**
1. `name` non-empty (after `TrimSpace`) — else
   `gates.custom[%d].name is required`.
2. `name` unique across all custom gates — else
   `gates.custom[%d].name %q duplicates gates.custom[%d]`.
3. `name` does **not** collide with a built-in gate Name (case-sensitive exact),
   rejecting against this list (verified from `internal/gates/*.go`):
   `G1: File Size`, `G11: i18n`, `G-Build: Cross-Compile`, `import_graph`,
   `G-Secrets: Secret Scan`, `G-Vuln: Dependency Audit`, `spec-traceability-gate`,
   `roadmap_drift`, `audit_baseline` — else
   `gates.custom[%d].name %q collides with built-in gate`.
4. `command` non-empty after `TrimSpace` — else
   `gates.custom[%d].command is required` (brief edge:
   empty/whitespace command ⇒ validation error, not runtime panic).
5. `severity ∈ {fail, warn}` — else
   `gates.custom[%d].severity must be fail or warn, got %q`.
6. `output ∈ {blob, lines}` — else
   `gates.custom[%d].output must be blob or lines, got %q`.

`Normalize` runs **before** `validate` (as `roadmap_drift` does) so defaults are
applied first; collision/duplicate checks run on the normalized (trimmed) names.

**`customResult` mapping (DECIDED):**
- `timedOut` ⇒ `Status=Fail`, `Message="<name> timed out after <T>s"`,
  `Details=[that message]`.
- `code == 0` ⇒ `Status=Pass`, `Message="<name> passed"`, no Details.
- `code != 0` ⇒ `Status = Fail if severity=="fail" else Warn`;
  `Message="<name> failed (exit <code>)"`; `Details` from the output mapper.
- output mapper: `blob` ⇒ `[]string{truncate(trimmed, 4096)}` (or the generic
  empty-output fallback); `lines` ⇒ `boundedLines(stdout, 200)` (each non-empty
  line one entry, `… (N more)` overflow marker, generic fallback when zero lines).

## Step 3 — tests

Colocated per-package `_test.go` (95% **per-package** coverage gate is NOT moved
by `tests/` tier files — coverage must sit next to the code). Each ≤100 lines.

**Unit — `internal/config/custom_gate_test.go`:**
- `NormalizeCustomGates` defaults severity→`fail`, output→`blob`, timeout 0→60;
  trims whitespace; leaves explicit values untouched.
- `validateCustomGates` rejects: empty name, duplicate names, each built-in-name
  collision, empty/whitespace command, bad severity, bad output — asserting the
  **indexed** error message and the offending index.
- no-op when `Enabled=false` (a disabled malformed entry does not error — matches
  the disabled-gate convention).

**Unit — `internal/gates/custom_command_test.go`:**
- `customResult`: exit 0 ⇒ Pass/no Details; exit≠0 + `fail` ⇒ Fail; exit≠0 +
  `warn` ⇒ Warn (AC-2, brief edge "non-zero but warn ⇒ non-blocking"); timeout ⇒
  Fail with timeout message (AC-8).
- `blob` truncates a >4 KiB payload + marker (brief edge: huge output bounded);
  empty-output Fail ⇒ generic fallback Detail (brief edge).
- `lines` splits stdout into one Detail per non-empty line; bounds at the max +
  `… (N more)` (AC-7 + edge "thousands of lines bounded").

**Unit — `internal/gates/custom_command_exec_test.go`** (real trivial commands,
no fixture binaries; skip on windows where `sh` is absent if needed):
- `runCustom("true", …)` ⇒ `code==0`; `runCustom("false", …)` ⇒ `code!=0`;
  `runCustom("printf 'a\\nb'", …)` ⇒ output captured; `runCustom("sleep 5", 50ms)`
  ⇒ `timedOut==true` (AC-8); `diff_aware` path sets `CENTINELA_CHANGED_FILES`
  (assert via `printf "$CENTINELA_CHANGED_FILES"`).

**Integration — `tests/integration/custom_gate_test.go`** (drive
`gates.RunWithFilter` over a `t.TempDir()` config, real shell commands):
- `[[gates.custom]]` with `command="true"` ⇒ one `Result{Status:Pass}` named by
  `name` (AC-1).
- `command="false", severity="fail"` ⇒ Fail; `severity="warn"` ⇒ Warn (AC-2).
- Two custom gates, one `true` one `false` ⇒ both appear; the failing one does
  not suppress the passing one (AC-5 independence).
- No `[[gates.custom]]` / all `enabled=false` ⇒ `RunWithFilter` output identical
  to the built-in-only baseline (brief edge: byte-identical no-op).
- `output="lines"` over a multi-line-emitting command ⇒ N Details (AC-7), then
  feed those Results to `internal/audit Compute` and assert per-line fingerprints
  (AC-4 ratchet coherence).
- `diff_aware=true` with a non-nil filter ⇒ command sees `CENTINELA_CHANGED_FILES`
  (AC: diff-aware env).

**Acceptance — `tests/acceptance/custom_gate_*`** (executable, one per Gherkin
scenario; run the **built binary**, not package APIs): in a fixture repo with a
`centinela.toml` declaring a custom gate, run `centinela validate` and assert exit
code + the gate name + severity behavior for: pass, fail-blocks, warn-non-blocking,
telemetry `gate-failure` recorded (read the telemetry log), multi-gate
independence. Register the acceptance runner in `centinela.toml`
`[validate] commands` (tests-step gate requires acceptance execution wired there).

`.workflow/custom-gate-sdk-edge-cases.md` — map every brief edge case
(empty/whitespace command, command-not-found, timeout, name collision, empty
output→generic Detail, disabled/empty list no-op, huge output bounded, warn
non-blocking, `lines` with thousands bounded) to the covering test.

Note: `go test ./...` ~75s; `[verify] verify_timeout` gives margin (timeout
tests use sub-second durations to stay fast).

## Step 4 — validate

Gatekeeper report `.workflow/custom-gate-sdk-gatekeeper.md`; `centinela validate`
green (lint + types + full suite). Confirm: (1) every new source file (incl.
`_test.go`) ≤100 lines; (2) the G2 import-graph gate reports **zero new failing
edges** — `customGates` imports only `internal/config` (leaf) + stdlib, and
`gates` (domain) already imports `config`, so **no `centinela.toml`
`[gates.import_graph]` change and no `internal/scaffold/assets` mirror is needed**
(state this explicitly in the report); (3) AC-4 audit coherence — verify against
`internal/audit/participation.go` whether custom-gate Names participate by default
or need folding into the default set (Audit-ratchet coherence subsection). Dogfood
via a `/tmp` binary built from `./cmd/centinela` with a throwaway
`[[gates.custom]]` config before relying on the installed binary. Production-
readiness subagent if the gate is enabled.

## Step 5 — docs

Documentation-specialist `.md` + `.json`; regenerate
`docs/project-docs/index.html`; changelog artifact
`.workflow/custom-gate-sdk-changelog.md` (create early via `evidence artifact new`
so completion doesn't fail). Document:
- The `[[gates.custom]]` schema (every field + default), with a worked example
  (`no-console-log` over `src/`).
- The two `output` modes and **which to pick** (`lines` for baseline-able
  per-violation gates, `blob` for simple pass/fail).
- `severity` semantics + the **`fail` default rationale** (opt-in authorship vs
  built-in `warn`).
- The **exec/trust model**: shell, checked-in-config = trusted, per-gate timeout,
  no sandbox — and how it differs from `[validate] commands` (structured/governed
  vs quick/opaque; both coexist).
- `diff_aware` + `CENTINELA_CHANGED_FILES` contract (newline-separated; command
  does its own filtering).
- The **SDK framing**: `Result` is the contract, built-ins are the reference
  implementations, custom gates are first-class (telemetry + audit baseline).
- Determinism expectation for ratchet stability (non-deterministic command output
  destabilizes fingerprints — the rule author's responsibility).
- Mirror any documented schema/example into `internal/scaffold/assets` **only if**
  the scaffolded `centinela.toml` template gains a `[[gates.custom]]` example
  (optional; verify — no import-graph matrix change is required either way).
