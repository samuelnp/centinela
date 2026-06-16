# Implementation Plan — audit-baseline-ratchet

> Feature brief: `docs/features/audit-baseline-ratchet.md`.
> Spec: `specs/audit-baseline-ratchet.feature`.

Phase 8 makes Centinela's all-or-nothing mechanical gates adoptable on a legacy
codebase. We add a **baseline + ratchet** layer *on top of* the existing gate
results without touching any gate's logic: a new `centinela audit` command runs
every participating gate in **full-repo scan** (`gates.RunWithFilter(cfg, nil)`),
**fingerprints** each `Result.Details` line to a stable identity, and compares
against a committed `.workflow/audit-baseline.json`. New violations block;
baselined ones are tolerated; resolved ones are pruned on the next record. The
whole comparison lives in a new `internal/audit/` package; `gates` is untouched
and `cmd/` only wires. A `[gates.audit_baseline]` config block (mirroring
`RoadmapDriftConfig`) lets `validate` optionally enforce the ratchet, defaulting
safe (warn) until a baseline is recorded.

## Decisions (DECIDED)

1. **Parse `Details`, don't refactor gates (v1).** We fingerprint the existing
   `[]string` `Result.Details` rather than adding a structured `Finding` type to
   every gate's `Result`. The structured-findings refactor (Risk: scope creep)
   is **out of v1 scope** — it touches all 7 gates and their tests for no
   ratchet behavior gain. The per-gate identity extractor (Decision #2) absorbs
   the brittleness of string parsing in one place.

2. **Per-gate identity extractor with a generic fallback (the central call —
   see Fingerprint CALL-OUT).** Each participating gate's `Result.Name` selects
   a reducer that maps a raw `Detail` to a **stable key**; unknown gates fall
   back to a generic normalizer (strip the trailing parenthetical + digits).

3. **Command surface = `audit` group, no mandatory validate coupling (smallest
   thing that unblocks `precommit-and-pr-gate`).**
   - `centinela audit baseline` — record/update: full-scan, fingerprint, write
     `.workflow/audit-baseline.json` (deterministic). Never loosens (Decision
     #5). Exit 0.
   - `centinela audit` (default `RunE`) — ratchet check: full-scan, partition
     into new / baselined / resolved, render, exit **1 iff any new** (else 0).
     `--json` emits the machine-readable verdict (`{new, baselined, resolved}`
     counts + the new fingerprints) so `precommit-and-pr-gate` reads a verdict
     without scraping text.
   - **Plus** an optional `audit_baseline` gate wired into `validate` via
     `[gates.audit_baseline]`, defaulting `enabled=false`. When enabled it runs
     the same ratchet diff and maps `new>0` → Fail/Warn per severity. This is
     additive: `precommit-and-pr-gate` can call `centinela audit --json`
     directly (fast, no full `validate`), OR a team can fold the ratchet into
     `validate`. We ship both because the gate is ~30 lines once the diff logic
     lives in `internal/audit`.

4. **Ratchet semantics + missing-baseline default.**
   - **new** (current ∉ baseline) → blocking (exit 1 / gate Fail).
   - **baselined** (current ∈ baseline, still present) → tolerated, reported.
   - **resolved** (baseline ∉ current) → reported as prunable; `audit baseline`
     drops them on next record (ratchet only tightens).
   - **Missing baseline file** → non-blocking: print `no baseline — run
     centinela audit baseline`, exit 0, gate → Skip. Treating "no baseline" as
     "nothing is new" is the safe adoption default (matches `roadmap_drift`
     defaulting to `warn`); it never blocks a repo that hasn't opted in.
   - **`audit baseline` never re-adds resolved fingerprints** — it records the
     current full-scan set verbatim, so a fixed violation can never be
     re-tolerated. (The on-disk file is fully replaced, not merged.)
   - **Severity default = `warn`** in `[gates.audit_baseline]` (safe adoption);
     teams ratchet to `fail` once green. The standalone `centinela audit`
     command is **always blocking on new** regardless of severity — severity
     only governs the optional `validate` gate.

5. **Participating gates = those that emit per-violation `Details`, filtered by
   an optional `target_gates` allowlist.** A gate whose `Result` is pass/fail
   with no `Details` (or whose only Detail is the gate's own summary message,
   e.g. `roadmap_drift`) is **excluded** — it can't be baselined per-violation.
   Default participation: `G1: File Size`, `import_graph`, `spec-traceability-gate`,
   `G-Secrets: Secret Scan`, `G11: i18n`. Empty `target_gates` ⇒ all
   detail-emitting gates participate; a non-empty list restricts to those Names.

6. **Layering: all logic in `internal/audit/`** (record, load, fingerprint,
   ratchet-diff). `cmd/centinela/audit*.go` only wires flags → calls → render.
   `internal/gates` is **not modified** and must **not** import `internal/audit`
   (would cycle / break the matrix — see Import-graph CALL-OUT).

## Fingerprint & stability (CALL-OUT — the make-or-break correctness risk)

A violation's **identity** is its location + rule, never its volatile payload
(line count, drift line number). The naive `sha256(gate + rawDetail)` is wrong:
`src/a.go (150 lines)` and `src/a.go (151 lines)` would hash differently, so a
baselined oversized file looks "new" after any edit (AC-5 fails).

**Design: a per-gate identity extractor keyed by `Result.Name`.** It lives in
`internal/audit/fingerprint.go`:

```go
// Fingerprint is a stable per-violation identity used by the ratchet.
type Fingerprint struct {
    Gate string `json:"gate"` // Result.Name
    Key  string `json:"key"`  // normalized stable identity (e.g. path, edge)
    Hash string `json:"hash"` // sha256(scheme + "\x00" + Gate + "\x00" + Key), hex
    Raw  string `json:"raw"`  // last-seen raw Detail, for human PR review only
}

// fingerprintScheme versions the normalization. Bump on any extractor change.
const fingerprintScheme = "v1"

// identityKey reduces one raw Detail to its stable key for a given gate Name.
func identityKey(gate, detail string) string
```

`identityKey` switches on `gate`:

| Gate (`Result.Name`)        | Raw Detail example                                  | Stable key                       |
|-----------------------------|-----------------------------------------------------|----------------------------------|
| `G1: File Size`             | `internal/x.go (150 lines)` / `… exceeds justified max 130` | `internal/x.go` (path before first ` (`) |
| `import_graph`              | `internal/ui → internal/orchestration (forbidden)`  | `internal/ui → internal/orchestration` (text before ` (`) |
| `spec-traceability-gate`    | `specs/x.feature: "scenario name"`                  | `specs/x.feature: "scenario name"` (already stable — pass through) |
| `G-Secrets: Secret Scan`    | `path/to/file: rule aws-key`                        | `path/to/file: rule aws-key` (path+rule, already stable) |
| `G11: i18n`                 | `src/a.ts` or `key.path`                             | trimmed Detail (already a path/key) |
| *(any other / future gate)* | arbitrary                                            | **generic fallback** |

**Generic fallback** (`genericKey`): strip a trailing ` (…)` parenthetical and
any trailing run of digits/whitespace, then `strings.TrimSpace`. This makes
"file (N lines)"-shaped details from gates we don't special-case stable by
default, and is a no-op on already-stable keys.

```go
// genericKey strips a trailing "(…)" group and trailing digits so volatile
// counts (line numbers, sizes) don't change a violation's identity.
func genericKey(detail string) string

// Compute builds the deduplicated fingerprint set for one gate's details.
func Compute(gate string, details []string) []Fingerprint
```

- **Hash** = `sha256(fingerprintScheme + "\x00" + gate + "\x00" + key)` rendered
  hex. We store **both** `Hash` (the compare key — compact, scheme-versioned)
  **and** `Key`/`Raw` (human-readable, for clean PR diffs of the baseline).
  Comparison is `Hash`-only, so a `Raw` churn (line-count change) doesn't move
  the identity (AC-5). `Key` round-trips for review.
- **Versioning**: `fingerprintScheme` is folded into every Hash *and* written as
  a top-level `scheme` field in the baseline file. If a gate's Detail format
  changes between Centinela versions, bump the scheme; a mismatched scheme on
  load ⇒ treat baseline as stale and report "scheme changed — re-run audit
  baseline" (non-blocking), so an old baseline never silently mis-matches.
- **Dedup**: identical fingerprints within one gate collapse to one entry; we
  keep a stable set (sorted by Hash) — duplicate identical details are not
  double-counted (edge case).

## Import-graph / layer decision (CALL-OUT)

**Verdict: `internal/audit` joins the `aggregator` layer. No new failing edge,
no cycle.**

- `internal/audit` is a **read-only analytics aggregator** over gate results +
  config + an on-disk artifact, imported **only by `cmd/`** — structurally
  identical to `internal/insights` and `internal/calibration`, which already sit
  in the `aggregator` layer (`centinela.toml` `[[gates.import_graph.layers]]`
  name=`aggregator`, paths `internal/doctor/**`, `internal/insights/**`,
  `internal/calibration/**`, `allow = ["domain","leaf"]`).
- Its edges are `internal/audit → internal/gates` (**domain**) and
  `internal/audit → internal/config` (**leaf**). The aggregator layer **allows
  both `domain` and `leaf`**, so adding `audit` to that layer's `paths`
  (`internal/audit/**`) makes both edges pass cleanly. This is *strictly better*
  than leaving it unmapped (which would only Warn): mapping it asserts the
  allowed dependency direction explicitly.
- **No cycle:** `gates` (domain) `allow = ["leaf"]` only — so if `gates`
  imported `audit` (aggregator) the gate would **Fail**. The plan forbids that
  edge (Decision #6); `audit` imports `gates`, never the reverse. `config` is a
  leaf and imports neither. Graph `cmd → audit → {gates → config, config}` is
  acyclic.
- **Required toml change:** add `internal/audit/**` to the `aggregator` layer
  `paths` (one line), with a comment mirroring the insights/calibration note.
  **Mirror the change in `internal/scaffold/assets`** if the import-graph matrix
  is part of the scaffolded `centinela.toml` template (verify during code; the
  scaffold parity test only covers 8 arch docs, so a toml drift won't be caught
  automatically — check by hand).

## v1 scope

**In:** `internal/audit` package (fingerprint + extractors, baseline schema,
load/save, ratchet diff); `centinela audit` (ratchet, exit 1 on new, `--json`)
and `centinela audit baseline` (record/replace, deterministic); optional
`[gates.audit_baseline]` gate in `validate` (default off, severity warn);
full-scan always (`RunWithFilter(cfg, nil)`, ignoring `diff_mode`); per-gate
identity extractors for the 5 detail-emitting gates + generic fallback;
scheme-versioned baseline; deterministic sorted output.

**Out (deferred):** a structured `Finding` refactor of every gate's `Result`
(big-bang, no behavior gain in v1 — Decision #1); auto-recording a baseline on
first run (explicit `audit baseline` only); time-window / per-author
attribution of new violations; baselining gates with no per-violation `Details`
(`roadmap_drift`, `build` summary-only); a `--fix`/auto-prune-and-commit flow
(prune happens only on explicit `audit baseline`).

## Step 2 — code

New / edited source files (each ≤100 lines):

| File | Change | Budget |
|------|--------|--------|
| `internal/audit/fingerprint.go` | NEW. `Fingerprint` struct, `fingerprintScheme`, `Compute(gate, details) []Fingerprint`, `identityKey(gate, detail)`, `genericKey(detail)`, hash helper | ~95 |
| `internal/audit/baseline.go` | NEW. `Baseline` + `GateEntry` structs (Decision: schema below); `Save(path, Baseline)` (deterministic JSON, sorted, trailing `\n`); `Load(path) (Baseline, bool, error)` (bool = exists) | ~90 |
| `internal/audit/record.go` | NEW. `Record(cfg) Baseline` — runs `gates.RunWithFilter(cfg, nil)`, filters to participating gates (Decision #5), `Compute`s each, assembles sorted `Baseline{Scheme: fingerprintScheme}` | ~70 |
| `internal/audit/ratchet.go` | NEW. `Diff` struct (`New, Baselined, Resolved []Fingerprint`); `Ratchet(cfg, baseline) Diff` — current set vs baseline by `Hash`; `(Diff).HasNew() bool` | ~85 |
| `internal/audit/participation.go` | NEW. `participatingGates(cfg) map[string]bool` over the default set ∩ `target_gates`; `isParticipating(name, cfg)` | ~45 |
| `internal/audit/gate.go` | NEW. `Check(cfg) gates.Result` — the `audit_baseline` gate body: load baseline (missing ⇒ Skip), `Ratchet`, map `len(New)>0` → Fail/Warn per severity, Details = new fingerprints' `Raw` | ~70 |
| `internal/config/audit_baseline.go` | NEW. `AuditBaselineConfig{Enabled bool; Severity string; BaselinePath string; TargetGates []string}`; `NormalizeAuditBaseline` (default severity `warn`, default path `.workflow/audit-baseline.json`); `validateAuditBaseline` (severity ∈ {fail,warn}) — mirrors `roadmap_drift.go` | ~55 |
| `internal/config/config.go` | add `AuditBaseline AuditBaselineConfig \`toml:"audit_baseline"\`` to `GatesConfig` | +1 line |
| `internal/config/defaults.go` | `cfg.Gates.AuditBaseline = NormalizeAuditBaseline(cfg.Gates.AuditBaseline)` in `applyDefaults` | +1 line |
| `internal/config/file_size_exceptions.go` | call `validateAuditBaseline(cfg.Gates.AuditBaseline)` in `validateConfig` | +3 lines |
| `internal/gates/gates.go` | in `RunWithFilter`, append `auditCheck(cfg)` **only** when `cfg.Gates.AuditBaseline.Enabled` — but to avoid `gates → audit` cycle, the gate is invoked from `cmd/` after `RunWithFilter`, NOT inside `gates`. See note. | 0 (no edit) |
| `cmd/centinela/audit.go` | NEW. `auditCmd` group + default `RunE` = ratchet check; `--json` flag; exit code via `HasNew()` | ~90 |
| `cmd/centinela/audit_baseline.go` | NEW. `audit baseline` subcommand → `audit.Record` → `audit.Save`; confirmation line | ~50 |
| `cmd/centinela/audit_render.go` | NEW. render `Diff` (new/baselined/resolved counts + lists) reusing `ui` styles; `--json` marshaling | ~85 |
| `cmd/centinela/validate.go` | when `cfg.Gates.AuditBaseline.Enabled`, append `audit.Check(cfg)` to `results` before `AllPassed` (cmd wires the aggregator gate; keeps `gates` clean) | +4 lines |
| `internal/ui/render_audit.go` *(optional)* | if `audit_render.go` exceeds budget, move pure formatting here | ~60 |

**Cycle-avoidance note (load-bearing):** the `audit_baseline` gate result is
produced by `internal/audit/gate.go` `Check`, but **wired into `validate` from
`cmd/centinela/validate.go`**, not from inside `gates.RunWithFilter` — because
`gates` (domain) may not import `audit` (aggregator). `cmd` (allows
domain+leaf+aggregator) is the correct seam. So `RunWithFilter` is *not* edited;
`validate.go` appends `audit.Check(cfg)` to the `[]gates.Result` it already
collects. `audit.Check` returns a plain `gates.Result`, so `ui.RenderGateResult`
and `gates.AllPassed` handle it with no new rendering path.

### Baseline file schema (`.workflow/audit-baseline.json`)

```go
// Baseline is the committed, deterministic ratchet snapshot.
type Baseline struct {
    Scheme  string      `json:"scheme"`  // == fingerprintScheme; stale if mismatched
    Version int         `json:"version"` // file-format version, currently 1
    Gates   []GateEntry `json:"gates"`   // sorted by Gate name
}

type GateEntry struct {
    Gate         string        `json:"gate"`         // Result.Name
    Fingerprints []Fingerprint `json:"fingerprints"` // sorted by Hash
}
```

- **Determinism (AC-7):** gates sorted by `Gate`, fingerprints sorted by `Hash`;
  `Save` uses `json.MarshalIndent(_, "", "  ")` + trailing newline. No maps in
  the serialized form. Re-recording an unchanged repo yields a byte-identical
  file ⇒ clean git diffs.
- **Reviewability:** each `Fingerprint` carries `Key` + `Raw`, so a PR reviewer
  reads human text, not just hashes.

## Step 3 — tests

Colocated per-package `_test.go` (95% per-package coverage gate is NOT moved by
`tests/` tier files — add coverage next to the code). Each ≤100 lines (G1
applies to `_test.go` too):

- `internal/audit/fingerprint_test.go` — **the AC-5 guard.** `identityKey` for
  each gate Name maps the documented raw Detail to the expected key; **`src/a.go
  (150 lines)` and `src/a.go (151 lines)` produce the SAME `Hash`** (stability);
  `genericKey` strips trailing `(…)`/digits and is a no-op on stable keys;
  `Compute` dedups identical details; scheme is folded into the hash (changing
  `fingerprintScheme` changes the hash).
- `internal/audit/ratchet_test.go` — `Ratchet` partitions correctly: a current
  fingerprint absent from baseline ⇒ `New` (AC-3); present-in-both ⇒ `Baselined`
  (AC-2); baseline-only ⇒ `Resolved` (AC-4); `HasNew()` true iff `New` non-empty.
- `internal/audit/baseline_test.go` — `Save`→`Load` round-trips; output is
  sorted + byte-stable across two records of the same set (AC-7); `Load` of a
  missing path returns `exists=false` (edge case); a scheme-mismatched file is
  surfaced as stale (edge case).
- `internal/audit/record_test.go` — `Record` over a fake `*config.Config`
  includes only participating gates (Decision #5); `participation_test.go`
  covers `target_gates` allowlist filtering + the empty-list-means-all default.
- `internal/audit/gate_test.go` — `Check`: missing baseline ⇒ `Skip`
  (non-blocking, AC edge); `New>0` ⇒ Fail at severity `fail`, Warn at `warn`;
  all baselined ⇒ Pass.
- `internal/config/audit_baseline_test.go` — `NormalizeAuditBaseline` defaults
  severity→`warn` + path→`.workflow/audit-baseline.json`; `validateAuditBaseline`
  rejects an unknown severity; no-op when disabled.

**Integration:** `tests/integration/audit_test.go` — in a `t.TempDir()` mini-repo
with one oversized file: (a) `audit baseline` writes a baseline naming that file;
(b) `audit` with no change reports it `baselined`, exit 0 (AC-2); (c) add a
*second* oversized file ⇒ `audit` exits 1, names only the new one, keeps the
first baselined (AC-3); (d) delete the original ⇒ it shows `resolved`, exit 0,
and the next `audit baseline` prunes it (AC-4); (e) grow the baselined file by
lines ⇒ still `baselined`, not new (AC-5). Drive via the package APIs or the
built binary; assert `--json` verdict counts match.

**Acceptance:** `tests/acceptance/audit_*` (executable, one per Gherkin
scenario) — run the real `centinela audit` / `centinela audit baseline` binary
against a fixture repo and assert exit codes + summary lines for: record,
no-change-tolerated, new-blocks, fix-prunes, growth-stable, missing-baseline-
non-blocking. Register the acceptance runner in `validate.commands` in
`centinela.toml`.

`.workflow/audit-baseline-ratchet-edge-cases.md` — map every brief edge case (no
baseline, empty repo, newly-enabled gate, no-Details gate excluded, scheme
change, deleted/renamed file resolves, diff-mode-on-still-full-scan, duplicate
details) to the test covering it.

Note: `go test ./...` runs ~75s; `[verify] verify_timeout = 240` gives margin.

## Step 4 — validate

Gatekeeper report `.workflow/audit-baseline-ratchet-gatekeeper.md`; `centinela
validate` green (lint + types + full suite). **Confirm the G2 import-graph gate
output: the new `internal/audit/**` mapping must produce zero new *failing*
edges** (audit→gates is domain, audit→config is leaf, both allowed by the
aggregator layer); `gates` must NOT import `audit`. Confirm every touched source
file ≤100 lines (including `_test.go`). Dogfood the new `centinela audit`
subcommands from a `/tmp` binary built from `./cmd/centinela` before relying on
the installed binary. Production-readiness subagent if the gate is enabled.

## Step 5 — docs

Documentation-specialist `.md` + `.json`; regenerate
`docs/project-docs/index.html`; changelog artifact
`.workflow/audit-baseline-ratchet-changelog.md` (create early via `evidence
artifact new` so completion doesn't fail). Document: the `centinela audit` /
`audit baseline` commands + exit codes + `--json`; the fingerprint scheme and
its per-gate identity keys (with the stability guarantee); the
`[gates.audit_baseline]` knobs (enabled, severity, baseline_path, target_gates)
and safe defaults; the baseline file schema for PR reviewers; the ratchet
semantics (new blocks / baselined tolerated / resolved prunes, only-tightens).
Add a PROJECT.md G2 one-line note that `internal/audit` is an aggregator
importing `internal/gates` (domain) + `internal/config` (leaf) read-only, and
mirror the `centinela.toml` import-graph change into `internal/scaffold/assets`
if the matrix is scaffolded.
