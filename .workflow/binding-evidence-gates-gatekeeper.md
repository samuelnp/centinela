### Adversarial Verifier Report: binding-evidence-gates
**Date:** 2026-07-31
**Status:** WARNING

#### Inputs Read

Re-verification from a fresh context. I did **not** read any pre-existing
`.workflow/*-gatekeeper.md`, and I treated no role's `.workflow/*.md` narrative
as proof of behaviour — every claim below is re-derived from source or from a
command I ran.

- `git diff main...HEAD` (62 files) **plus** the uncommitted working tree
  (`docs/features/binding-evidence-gates.md`, `specs/binding-evidence-gates.feature`,
  `internal/workflow/validate_docs.go`, the three `tests/acceptance/binding_evidence_gates_*_test.go`,
  `ROADMAP.md`, `.workflow/roadmap.json`) and the two **untracked** implementation
  files `internal/orchestration/fill_stub.go` + `internal/orchestration/fill_marker_stub_test.go`.
- `docs/features/binding-evidence-gates.md`, `specs/binding-evidence-gates.feature`,
  `docs/plans/binding-evidence-gates.md`.
- Implementation: `internal/workflow/handoff.go`, `handoff_chain.go`,
  `validate_orchestration.go`, `validate_docs.go`, `validate_freshness.go`,
  `validate_gatekeeper.go`, `stamp.go`, `order.go`;
  `internal/gatereport/commands_schema.go`, `block.go`, `stamp.go`, `grounding.go`, `check.go`;
  `internal/orchestration/fill_marker.go`, `fill_stub.go`, `handoff_read.go`,
  `evidence.go`, `policy.go`, `feature_surface.go`, `paths.go`, `validate.go`;
  `internal/evidence/schema_init.go`, `validate.go`, `fill.go`, `artifact_changelog.go`;
  `internal/gates/file_size.go`, `file_size_scan.go`; `internal/treestate/digest.go`, `untracked.go`.
- Tests: `internal/workflow/handoff_{tolerance,migration,fixture}_test.go`,
  `internal/orchestration/fill_marker_stub_test.go`,
  `tests/acceptance/binding_evidence_gates_{helper,handoff,migration}_test.go`.
- `centinela.toml`, `.workflow/binding-evidence-gates.json`,
  `docs/architecture/evidence-contract.md` + its `internal/scaffold/assets/` mirror.
- Output of the commands under **Commands Run**, and of four throwaway probe
  packages I wrote, executed, and deleted (see Refutation Attempts). The tree
  is clean of them — `git status --short` after deletion shows only the
  feature's own files.

**Orchestrator prompt:** contained no narrative summary of the implementation.
It named the three gates and the attack surfaces, which I used as a starting
point and then went past; it asserted no behaviour I have taken on trust.

#### Analyzed Specs

`specs/binding-evidence-gates.feature` — 15 scenarios across the three defects
(handoff chain 6, stamp commands schema 4, changelog placeholder 3, migration 2).
The spec-traceability gate reports all 15 mapped to acceptance tests; I confirmed
the four scenarios added in the uncommitted diff (null exitCode, stub behind a
heading, tolerance refuses another step's role, old prefill on a user-facing
feature) now carry real `// Scenario:` anchors rather than the "proposed" comments
they replaced.

#### Refutation Attempts

**Claim attacked:** "An unfilled changelog template no longer passes the docs gate."
**How:** Wrote a throwaway `internal/orchestration` probe over `StubEntry` and a
throwaway `internal/workflow` probe driving the real `validateChangelog` against
18 + 8 hand-built changelog bodies: the verbatim scaffold, the scaffold behind a
heading, all-citation forms (`- <FILL: ...>: <FILL: ...>`), unicode-ellipsis
descriptions, CRLF, tab indent, unterminated `<FILL:`, alternate bullets
(`>`, `1.`, `+`, `•`), case variants, doubled `>`, and — the decisive family —
the *residues of erasing* the scaffold rather than filling it.
**Result:** **REFUTED in part.** The verbatim scaffold and the heading-hidden
scaffold are correctly rejected (line-numbered, actionable errors). But erasing
the two slots in place leaves `- : `, and `validateChangelog` **accepts** it —
as it accepts `- `, `-`, `.`, `TODO`, and `## Changelog` alone. `StubEntry` fires
only while a literal `<FILL:` survives, so the cheapest route past the gate is to
delete the marker, not to write an entry. See F1/F2.

**Claim attacked:** "Prose that legitimately quotes the marker still passes."
**How:** Probed both directions of the citation rule with real changelogs,
including the one shipped changelog in this repo that quotes the marker
(`.workflow/deterministic-artifact-scaffolds-changelog.md`, which carries both
`<FILL: …>` and a bare unterminated `` `<FILL:` ``).
**Result:** **NOT REFUTED.** That real historical changelog still passes, as do
the four quoting forms the unit test pins. The known cost — a changelog naming a
*concrete* slot (`- docs: describe the <FILL: one-line summary of the change> slot`)
is refused — is real, reproduced, and explicitly documented in `fill_marker.go`
as the deliberate alternative to judging surrounding prose. Not a finding.

**Claim attacked:** "handoffTo is validated against the chain the workflow's own
contract defines; an out-of-chain value fails `complete`."
**How:** Drove `validateHandoffChain` through the real fixture (temp repo, real
brief, real saved workflow) over 10 hand-picked rows spanning: same-step hops on
a user-facing code step, the terminal hop on internal *and* user-facing validate
steps, the legacy `validation-specialist` pin, the legacy `big-thinker` /
`feature-specialist` plan pair, `production-readiness` as a foreign role, and the
docs-step terminal. Cross-checked against the shipped `outOfChainValues` table
(case, whitespace, newline, 2000-char junk).
**Result:** **NOT REFUTED.** No input skipped a required same-step role and no
foreign step's role was admitted. `acceptsHandoff` short-circuits the
alternate-pin tolerance for same-step and terminal hops, and the tolerance is
scoped to `alternateContractRoles(feature, nextStep)` — one step, not the union.
`banana` is rejected with the derived successor and a runnable remedy.

**Claim attacked:** "The derivation is pinned to the workflow's own contract."
**How:** Built a user-facing feature at the `code` step whose `senior-engineer`
evidence carries the deliberately-refused old-prefill `qa-senior`, confirmed the
refusal, then edited **only** `docs/features/<feature>.md` to drop the
`surface: user-facing` line and re-ran the identical gate. Separately deleted
`.workflow/<feature>.json` and re-ran.
**Result:** **REFUTED.** The refusal flips to acceptance on the brief edit alone.
`validateContract` and `planContract` are pinned in workflow state, but the third
derivation input — `IsUserFacingFeature`, which reads `docs/features/<feature>.md`
at gate time — is an unpinned, agent-writable markdown line. See F3. Deleting
workflow state also disarms the chain check (`CheckHandoffTo` returns nil), but
that is stated in the code as deliberate and `complete` cannot run without state.

**Claim attacked:** "Legacy and archetype-subset workflows must keep completing."
**How:** Ran the old prefill's literal (`documentation-specialist`) against a
`gatekeeper` and a `validation-specialist` on an **internal** feature.
**Result:** **REFUTED as written.** Both are refused, because an internal
feature's docs step requires no role, making validate terminal and short-circuiting
the tolerance. This is a knowing, unit-pinned decision — but it is a *second*
unaccommodated class, and the brief states there is exactly one, with the spec
carrying a scenario only for the user-facing one. See F4.

**Claim attacked:** "`artifact stamp` validates the commands schema, and the read
side is exactly as strict as the write side."
**How:** Threw 19 payloads at `Stamped` (write), `ParseVerification` (read) and
`Assess` (admissibility) side by side inside one probe: missing `exitCode`,
`exitCode: null / 0.0 / true / -0 / "0"`, duplicate `exitCode` keys, empty argv,
argv as a string, entries that are arrays, `[null]`, `null`, `[]`, a bare object,
key-case variants (`ARGV`, `exitcode`), `durationMs: "fast"`, a leading failing
entry followed by a passing one, and a non-`centinela` basename. Plus a stamp
round-trip, a no-block report, and a two-block report whose *second* block is
malformed.
**Result:** **NOT REFUTED.** Write and read agreed on **all 19** payloads — one
`ValidateCommandsSchema` genuinely serves both, so a post-stamp hand-edit cannot
sneak through. `null` decodes without setting the field and is caught by the
`*int` decode. Duplicate keys are last-wins on both sides identically. `commands:
null` and `[]` stamp but read as ungrounded via `Assess`. The stamp is idempotent
and re-`Assess`es clean. A malformed *second* block is invisible to both sides
(only the first tagged fence is ever read or written), so it is not a bypass.

**Claim attacked:** "The verifier's own evidence can satisfy the new rules."
**How:** Ran `artifact new --force` then `evidence init` and inspected what the
CLI seeded, before writing anything.
**Result:** **NOT REFUTED**, and a genuine dogfood: `evidence init` derived
`handoffTo: "complete"` for this internal feature from the workflow's own
contract — the pre-derivation static chain would have written
`documentation-specialist`, which this feature's own gate now refuses. The
CLI's prefill no longer fails the CLI's own gate.

**Claim attacked:** "No gate was weakened, and the 100-line cap holds repo-wide."
**How:** Full-scan (not diff-aware) line count over every `.go` file under
`internal/` and `cmd/` — the gate's own source roots, with no
`[[gates.file_size_exceptions]]` configured in `centinela.toml`. Also diffed
`docs/architecture/evidence-contract.md` against its `internal/scaffold/assets/`
mirror, and read the per-function coverage from the profile the validate run wrote.
**Result:** **NOT REFUTED.** Zero files over 100 lines repo-wide. Mirror is
byte-identical. Aggregate coverage 97.3%; every new function is 100% except
`hasSubstanceOutsideSlots` at 85.7% (its unterminated-`<FILL:` early return is
unexercised — the input class that produces F1-adjacent acceptances).

#### Commands Run

All from `/Users/samuelnp/projects/personal/centinela/.worktrees/binding-evidence-gates`
at revision `e8c9133ac79bc0ad9a60ae2608a6fd2da6ca8dd3`.

| # | argv | exit | duration |
|---|------|------|----------|
| 1 | `go build -o /tmp/centinela-v5 ./cmd/centinela` | 0 | ~11s |
| 2 | `/tmp/centinela-v5 validate` | **0** | **410s** (wall clock, `date +%s` delta) |
| 3 | `/tmp/centinela-v5 audit` | 0 | <1s — prints `no baseline`; no baseline committed in this worktree, so it produced no ratchet signal |
| 4 | `find internal cmd -type f -name "*.go" -exec sh -c 'n=$(wc -l < "$1"); [ "$n" -gt 100 ] && echo "$n $1"' _ {} \; \| sort -rn` | 0 | ~1s — **full scan**, empty output |
| 5 | `diff docs/architecture/evidence-contract.md internal/scaffold/assets/docs/architecture/evidence-contract.md` | 0 | <1s — identical |
| 6 | `go tool cover -func=coverage.out` | 0 | ~2s — total 97.3% |
| 7 | `/tmp/centinela-v5 artifact new binding-evidence-gates gatekeeper --force` | 0 | <1s |
| 8 | `/tmp/centinela-v5 evidence init binding-evidence-gates gatekeeper` | 0 | <1s |

Probe runs (throwaway `zzprobe_test.go` files, since deleted — all reported `ok`;
their exit codes were not captured verbatim, so they are **not** recorded in the
machine block below):
`go test ./internal/orchestration/ -run TestZZProbeStub -v`,
`go test ./internal/workflow/ -run TestZZProbe -v`,
`go test ./internal/workflow/ -run TestZZProbeBlanked -v`,
`go test ./internal/gatereport/ -run TestZZProbe -v`.

**The suite was run exactly once**, as command 1 of the single `centinela validate`
run: `centinela.toml` `[validate] commands` begins with
`go test ./... -coverprofile=coverage.out`, which is the full suite including
`tests/acceptance`. Per the mandate I did not re-run it. Full validate output:

```
Built-in Gates (diff-aware: 65 files changed since main)
✓ G1: File Size    ✓ G-Build: Cross-Compile    ⚠ import_graph (unmapped packages, non-failing)
✓ spec-traceability-gate  All 15 scenarios have acceptance coverage.
✓ roadmap_drift  ROADMAP.md is in sync.
✓ go test ./... -coverprofile=coverage.out
✓ COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh
✓ ./scripts/check-fmt.sh
🛡️👁️  CLI  All gates passed.
```

#### Findings

**F1 — Erasing the changelog scaffold still passes the gate it was written to fail.**
- **Affected spec:** `specs/binding-evidence-gates.feature`
- **Affected scenario:** "An unfilled changelog template fails the docs gate"
- **Risk:** `StubEntry` only fires while a literal `<FILL:` survives the line.
  Deleting the scaffold's two slots in place leaves `- : `, which
  `validateChangelog` **accepts** (probe-confirmed, together with `- `, `-`, `.`,
  and `TODO`). `internal/orchestration/fill_stub.go`'s own doc comment states the
  rule exists because "blanking the scaffold's two descriptions to ellipses …
  is erasing the evidence that one was never written" — blanking them to *nothing*
  is the same erasure with fewer characters, and is the cheaper move. This is the
  **third** bypass of this one rule, and the fix for the previous bypass does not
  generalise over it. The docs step therefore still completes on a changelog nobody
  wrote; the delivery composer (`cmd/centinela/deliver_artifacts.go`) will then
  lift `- : ` into `CHANGELOG.md`.
- **Suggestion:** Judge the *entry* rather than the marker: after stripping list
  bullets, separators and whitespace, require the entry to retain some minimum
  substance. `hasSubstanceOutsideSlots` already computes exactly that residue —
  apply it to every entry, not only to slot-bearing ones. Erasing the scaffold
  then fails for the same reason blanking it to ellipses does.

**F2 — A changelog with a heading and no entry passes.**
- **Affected spec:** `specs/binding-evidence-gates.feature`
- **Affected scenario:** "A changelog stub behind a heading fails the docs gate"
- **Risk:** `validateChangelog` increments `entries` for *any* non-blank line, so
  `## Changelog\n` alone satisfies the "at least one entry" rule (probe-confirmed).
  The newly added heading scenario proves the stub-behind-a-heading case, but the
  sibling case it invites — heading kept, stub line **deleted** — is the natural
  way an agent clears the newly failing gate, and it passes. Same root cause as F1:
  the rule counts lines, not entries.
- **Suggestion:** Exclude ATX/setext headings (and HTML comments) from the entry
  count so a document that is all scaffolding and no entry reports the existing
  `changelog entry is empty` error instead of passing.

**F3 — The handoff derivation's third input is unpinned and agent-writable.**
- **Affected spec:** `specs/binding-evidence-gates.feature`
- **Affected scenario:** "A valid same-step handoff when UX is required" /
  "Evidence seeded by the old prefill on a user-facing feature"
- **Risk:** `RequiredEvidenceRoles` → `orchestration.RequiredRolesForFeature` →
  `IsUserFacingFeature`, which re-reads `docs/features/<feature>.md` **at gate
  time**. Probe: a user-facing `code` step refuses `senior-engineer → qa-senior`
  with the documented remedy; delete the `surface: user-facing` line from the
  brief and the *identical* gate call returns nil. One markdown line — writable by
  every step's agent — converts the feature's flagship "same-step hops are exact by
  design" refusal into a pass, and drops the `ux-ui-specialist` evidence
  requirement with it. The brief and `docs/architecture/evidence-contract.md` both
  say the successor is derived from "this workflow's own contract"; two of the
  three inputs (`validateContract`, `planContract`) are indeed pinned in
  `.workflow/<feature>.json`, the third is not. The exposure pre-dates this feature
  for the *role-requirement* gate, but this feature is what makes an exactness gate
  depend on it.
- **Suggestion:** Pin the surface at `centinela start` into
  `.workflow/<feature>.json` (e.g. `userFacing: true|false`) and derive from the
  pinned value, falling back to the brief only when the field is absent — the same
  shape the contract pins already use. At minimum, have `centinela doctor` report a
  brief whose surface disagrees with the roles the feature's evidence contains.

**F4 — A second retroactively-unaccommodated class is undocumented.**
- **Affected spec:** `specs/binding-evidence-gates.feature`
- **Affected scenario:** "A valid terminal handoff on an internal feature"
- **Risk:** `docs/features/binding-evidence-gates.md` states "**One** class is
  deliberately NOT accommodated" and names the user-facing same-step case; the spec
  carries a scenario for exactly that one. In fact
  `gatekeeper|validation-specialist → documentation-specialist` on an **internal**
  feature is refused too (probe-confirmed): validate becomes terminal, and
  `acceptsHandoff` returns false for a terminal `want` before the tolerance is
  consulted. `TestHandoffToleranceDisabledForSameStepAndTerminal` pins it, so the
  behaviour is intended — but the brief understates the migration cost. Since the
  pre-derivation prefill wrote `documentation-specialist` for every gatekeeper,
  *every* Centinela-governed repo with an in-flight internal feature sitting at the
  validate step will hit this on its next `complete`. Actionable (the error prints
  the one-command fix) but unannounced.
- **Suggestion:** Amend the brief's constraint to "two classes", add the
  internal-feature terminal scenario to the spec (the acceptance test already
  exists to bind it), and name both in the changelog entry so downstream adopters
  are warned.

#### Deferred Findings

Not filed via `centinela roadmap defer` **deliberately**: that command rewrites
`ROADMAP.md`, a tracked file, and mutating the tree after the single verified
`centinela validate` run would leave the stamp describing a tree the gates never
saw. Slugs recorded here for the orchestrator to file after this verdict lands:

- `changelog-erasure-passes-gate` — deleting the `<FILL: …>` slots (leaving
  `- : `, `-`, or `TODO`) clears the docs gate without writing a changelog (F1/F2).
- `feature-surface-unpinned-at-gate-time` — `IsUserFacingFeature` re-reads the
  agent-writable brief at gate time, so one line edit disarms the same-step
  handoff exactness and the ux-ui-specialist requirement (F3).

#### Recommendation

**WARNING — ship, with F1/F2 and F3 filed.**

All three defects the brief set out to close are genuinely closed, and I could not
break the two I attacked hardest:

- `handoffTo` is derived, not hardcoded; `banana`, every foreign step's role, and
  every case/whitespace variant are refused, and the same-step and terminal hops
  correctly refuse the alternate-pin tolerance.
- The commands schema is one rule serving both sides — write and read agreed on
  all 19 adversarial payloads, including the `null`-decodes-to-zero case the spec
  calls out, so a post-stamp hand-edit cannot sneak through.
- The literal scaffold, and the scaffold behind a heading, both fail.

What keeps this off SAFE is that the changelog rule is still marker-scoped rather
than entry-scoped (F1/F2) — the cheapest way past it remains *erasing* the
evidence that nothing was written, which is the same move the last two fixes were
written to refuse — and that the handoff derivation the brief calls
contract-pinned is, for one of its three inputs, a markdown line any agent can
edit mid-workflow (F3). Neither falsifies a shipped scenario; both are the next
bypass in a rule whose history is bypasses. F4 is a documentation correction.

No finding blocks: nothing here is a false green on a claim the feature makes, and
no gate was weakened to make another pass.

```json centinela:verification
{
  "revision": "e8c9133ac79bc0ad9a60ae2608a6fd2da6ca8dd3",
  "treeDigest": "sha256:dc730948110be07e127719694d5cad63ffb8f43c2bee6af7f128709e2a4397b8",
  "commands": [
    {"argv": ["go", "build", "-o", "/tmp/centinela-v5", "./cmd/centinela"], "exitCode": 0, "durationMs": 11000},
    {"argv": ["/tmp/centinela-v5", "validate"], "exitCode": 0, "durationMs": 410000},
    {"argv": ["/tmp/centinela-v5", "audit"], "exitCode": 0, "durationMs": 500},
    {"argv": ["/tmp/centinela-v5", "artifact", "new", "binding-evidence-gates", "gatekeeper", "--force"], "exitCode": 0, "durationMs": 200},
    {"argv": ["/tmp/centinela-v5", "evidence", "init", "binding-evidence-gates", "gatekeeper"], "exitCode": 0, "durationMs": 200}
  ]
}
```
