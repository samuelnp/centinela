---
id: fd962adabbaacf20
feature: custom-gate-sdk
step: tests
type: lesson
title: - **Exit 0 → Pass**, non-zero → Fail; gate named in the report
tags: edge-cases, lesson
sourceArtifact: .workflow/custom-gate-sdk-edge-cases.md
createdAt: 2026-06-16T19:27:54Z
---

- **Exit 0 → Pass**, non-zero → Fail; gate named in the report
- **severity=fail blocks** validate (exit 1); **severity=warn** fails
- **Empty command output on failure** → generic non-empty Detail (no opaque
- **`output="lines"`** → one Detail per stdout line (blanks dropped, bounded by
- **Timeout** (`sleep` + tiny `timeout_seconds`) → Fail with timeout message,
- **Command not found / non-zero launch** → Fail via `exitCode` = -1, clear
- **Config validation** (`validateCustomGates`, indexed): empty command, empty
- **`enabled=false`** → gate skipped (command never runs); **no/empty
- **Normalize defaults** — severity→fail, output→blob, timeout 0→60, trimmed.
- **diff-aware** — `CENTINELA_CHANGED_FILES` env injected from the `*gitdiff.Set`
- **Audit-ratchet participation (AC-4, cross-feature regression guard)** —
- **Multiple gates independent**; **determinism** — same command → same Results.
- `runCustom` ~95% per-func (a launch-error branch); aggregate gate (95.1%) met.
- Shell-command tests skip on Windows (`runtime.GOOS`), but `runCustom` uses
- Non-deterministic user commands undermine ratchet fingerprint stability — by
- Trust model: custom commands run with full process permissions (config =
