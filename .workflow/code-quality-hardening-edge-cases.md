# Edge Cases: code-quality-hardening

Each hard path below lists the risk and how it is covered (or why deferred).

## Format gate
- **gofmt always exits 0** — `gofmt -l` lists offenders but returns 0, so a
  naive validate command can never fail. The wrapper `scripts/check-fmt.sh`
  forces `exit 1` when `gofmt -l` output is non-empty. Tested:
  `TestUnformattedSourceFailsFormatCheck` (non-zero exit + offender path) and
  `TestFormattedTreePassesFormatCheck` (exit 0, no output), both via the
  `FMT_DIRS` override against a throwaway temp tree so the test never depends
  on the repo being dirty.
- **Validate wiring drift** — the gate is useless unless `[validate] commands`
  actually invokes it. `TestValidateSuiteGatesFormatting` reads `centinela.toml`
  and asserts `./scripts/check-fmt.sh` is present.

## Evidence key-order invariant
- **coverage key position invariant** — the hookpolicy formatter duplicates the
  evidence `jsonKeyOrder`; dropping/misplacing `coverage` silently reorders
  coverage-bearing evidence files on the next postwrite. `TestEvidenceKeyOrderParity`
  byte-compares `evidence.MarshalJSON` output against the formatter for a doc
  carrying a coverage field and asserts coverage lands strictly between
  `mobileFirst` and `handoffTo`. The doc comment in
  `format_evidence_order.go` now names this exact test — the assertion makes
  the comment honest.

## Workflow-load transparency (three distinct outcomes)
- **missing vs unreadable vs invalid-JSON** — collapsing all three into
  "no workflow found" was the original defect. Three colocated tests pin each
  outcome separately: missing → "no workflow found" (`TestLoadMissingReportsAbsence`);
  invalid JSON → names the path + wraps the parse cause, never "no workflow found"
  (`TestLoadCorruptReportsPathAndCause`); unreadable → names the path AND does
  not say "no workflow found" (`TestLoadUnreadableIsNotAbsence`).
- **root-skip for the chmod fixture** — `chmod 000` does not deny root, so the
  unreadable test would spuriously fail under a root CI runner. It skips when
  `os.Geteuid() == 0` and restores 0644 in cleanup so the temp dir is removable.
- **silent drop in ActiveWorkflows** — a corrupt state file whose name matches
  its feature must not vanish from the active list without a trace.
  `TestActiveWorkflowsWarnsOnCorruptStateFile` captures stderr and asserts a
  `workflow warning:` line while the file is excluded from the active set.

## Config-error policy by surface
- **hook must never break the session on bad config** — the context hook degrades
  instead of erroring. `TestRunHookContextCorruptConfigWarnsAndExitsZero` asserts
  `runHookContext` returns nil AND injects a `config warning:` line into stdout.
- **start must fail loudly and leave no half-state** — corrupt TOML must stop
  `start` before any workflow JSON is written.
  `TestRunStartCorruptConfigFailsAndWritesNoState` asserts the error names
  `centinela.toml` and that no `.workflow/<feature>.json` exists afterward.
- **statusline hook can't carry a warning** — the statusline is a single-line
  protocol surface, so a `config warning:` line would corrupt it. The
  senior-engineer deliberately retains a silent default there; the context hook
  surfaces the same failure every prompt. Deferred by design — covered by the
  context-hook test rather than a statusline assertion.
