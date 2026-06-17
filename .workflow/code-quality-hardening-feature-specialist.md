### Feature-Specialist Report: code-quality-hardening
**Date:** 2026-06-09

#### Behavior Summary
- This feature fixes four verified quality defects and adds the mechanical enforcement that keeps them fixed. The postwrite hook's evidence formatter regains the `"coverage"` key (between `mobileFirst` and `handoffTo`) so reformatting a coverage-bearing evidence file stays byte-identical to canonical `evidence.MarshalJSON` output, guarded by a new behavior-level parity test. Formatting becomes gated: a new `scripts/check-fmt.sh` exits non-zero naming any non-gofmt file under `cmd internal tests`, and the script is appended to `[validate] commands` so `centinela validate` enforces it. Config-error policy is aligned by surface: `centinela start` hard-fails on a corrupted `centinela.toml` (naming the file, creating no workflow state) like `complete` already does, while the prompt context hook degrades gracefully — exiting zero and injecting a `config warning:` line so the host session never breaks. Finally, `workflow.Load()` becomes transparent: only a genuinely missing state file reports "no workflow found"; read failures and parse failures surface as errors naming the state file path.

#### Gherkin Scenarios
All scenarios live in `specs/code-quality-hardening.feature` and map 1:1 to executable Go acceptance assertions in `tests/acceptance/`.

1. **Hook formatter preserves the canonical evidence key order** — Given a role evidence JSON document containing a coverage field; When the evidence is marshalled canonically and reformatted by the postwrite hook formatter; Then both outputs are byte-identical, And the coverage key appears between mobileFirst and handoffTo.
2. **Unformatted Go source fails the format check** — Given a Go source file that is not gofmt-formatted; When the format check script runs over that file's tree; Then it exits non-zero, And it prints the offending file path.
3. **Formatted tree passes the format check** — Given a source tree where every Go file is gofmt-formatted; When the format check script runs; Then it exits zero, And it prints nothing.
4. **Validate suite gates formatting** — Given the project centinela.toml; When the validate command list is read; Then it includes the format check script `./scripts/check-fmt.sh`. *(Added: the plan's `[validate]` wiring had no scenario.)*
5. **Starting a feature with a corrupted config fails loudly** — Given a centinela.toml that cannot be parsed; When the user runs centinela start for a new feature; Then the command exits with an error naming centinela.toml, And no workflow state file is created.
6. **Prompt hook degrades with a warning on corrupted config** — Given a centinela.toml that cannot be parsed; When the prompt context hook runs; Then the hook exits zero so the host session continues, And the injected context contains a config warning naming the failure.
7. **Loading a missing workflow reports absence** — Given no workflow state file exists for a feature; When the workflow is loaded by name; Then the error states no workflow was found for that feature.
8. **Loading a corrupted workflow reports the cause** — Given a workflow state file containing invalid JSON; When the workflow is loaded by name; Then the error names the state file path, And the error includes the underlying parse failure.
9. **Loading an unreadable workflow is not reported as absence** — Given a workflow state file that exists but cannot be read; When the workflow is loaded by name; Then the error names the state file path, And the error does not state that no workflow was found. *(Added: the masked-read-failure branch is the core of defect 4 and had no scenario; invalid-JSON only exercises the parse branch, which was already wrapped.)*

#### UX States
| State | Trigger | Surface |
|-------|---------|---------|
| Format check failure | `centinela validate` (or direct script run) on a tree with non-gofmt files | stderr/stdout: offending file paths, exit 1 |
| Format check pass | `centinela validate` on a clean tree | silent, exit 0 |
| Start blocked by corrupted config | `centinela start <feature>` with unparseable centinela.toml | CLI error naming `centinela.toml`, non-zero exit, no `.workflow` state file written |
| Hook degraded mode | prompt context hook with unparseable centinela.toml | injected context contains `config warning: <error>`; hook exits 0; defaults used |
| Workflow absent | `status`/`complete` for a feature with no state file | error: "no workflow found for %q" |
| Workflow unreadable/corrupted | state file unreadable or invalid JSON | error naming the state file path with the underlying cause wrapped |
| Visual/graphical UI | n/a | n/a (CLI + hook feature only) |

#### Out-of-Scope
- CWD-relative path architecture refactor.
- Test-suite quality overhaul / coverage-padding cleanup.
- Consolidation of the three shell-exec wrappers.
- Any new built-in gate type — the format check rides `[validate] commands`.
- Hard policy decision for `ActiveWorkflows` on corrupted state files (warn vs fail) — flagged as an open clarification, not specified here.

#### Handoff
- Next role: senior-engineer
- Open clarifications: (a) `format_evidence_order.go` doc comment — drop the (currently false) import-cycle rationale entirely, or keep it as a forward-looking guard? Recommend justifying the duplication on layer thinness and citing the new parity test. (b) `ActiveWorkflows` with a corrupted state file: stderr warning while listing, or hard-fail? Spec deliberately leaves listing behavior unpinned; `Load()` semantics (scenarios 7–9) are pinned. (c) Land the mass `gofmt -w` as its own commit and coordinate with sibling worktrees to limit merge conflicts. (d) Scenario 9 (unreadable file) should use chmod-based fixtures and skip when running as root, where permission bits don't apply.
