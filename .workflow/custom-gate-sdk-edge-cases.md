# Edge Cases: custom-gate-sdk

## Covered

- **Exit 0 → Pass**, non-zero → Fail; gate named in the report
  (`customResult` + acceptance).
- **severity=fail blocks** validate (exit 1); **severity=warn** fails
  non-blocking (exit 0).
- **Empty command output on failure** → generic non-empty Detail (no opaque
  blank).
- **`output="lines"`** → one Detail per stdout line (blanks dropped, bounded by
  `customLineCap`, overflow marker); **`output="blob"`** → single Detail capped
  at 4 KiB.
- **Timeout** (`sleep` + tiny `timeout_seconds`) → Fail with timeout message,
  returns fast, never hangs.
- **Command not found / non-zero launch** → Fail via `exitCode` = -1, clear
  message, no crash.
- **Config validation** (`validateCustomGates`, indexed): empty command, empty
  name, duplicate names, collision with a built-in gate (`import_graph`),
  invalid severity, invalid output mode — each errors at config load.
- **`enabled=false`** → gate skipped (command never runs); **no/empty
  `[[gates.custom]]`** → byte-identical validate (append gated by `len>0`).
- **Normalize defaults** — severity→fail, output→blob, timeout 0→60, trimmed.
- **diff-aware** — `CENTINELA_CHANGED_FILES` env injected from the `*gitdiff.Set`
  when `diff_aware=true`; unset otherwise.
- **Audit-ratchet participation (AC-4, cross-feature regression guard)** —
  custom-gate Names are folded into `participatingGates`, so `output="lines"`
  violations baseline per-line, are tolerated, and a NEW line blocks as "new".
- **Multiple gates independent**; **determinism** — same command → same Results.

## Residual Risks

- `runCustom` ~95% per-func (a launch-error branch); aggregate gate (95.1%) met.
- Shell-command tests skip on Windows (`runtime.GOOS`), but `runCustom` uses
  `cmd /C` on Windows by construction; the unix path is fully exercised.
- Non-deterministic user commands undermine ratchet fingerprint stability — by
  design the user's responsibility; documented.
- Trust model: custom commands run with full process permissions (config =
  checked-in code), same as `[validate] commands`; no sandbox in v1.
