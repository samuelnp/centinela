### Feature-Specialist Report: deep-codebase-analysis
**Date:** 2026-06-17

#### Behavior Summary
`centinela analyze` performs a mechanical, read-only, no-LLM scan of the current
repository and emits a deterministic, machine-readable `Inventory`. It walks the
project root with a fixed skip set (`vendor/`, `node_modules/`, `.git/`,
`.workflow/`, `dist/`, `build/`, gitignored paths), counts source files by
extension to derive `languages` + `primaryLanguage`, detects manifests
(`go.mod`, `package.json`+scripts, `Gemfile`, `Cargo.toml`,
`pyproject.toml`/`requirements.txt`, `Makefile`) and extracts each one's
build/test/framework signals and declared dependency names, detects i18n
locales, builds a depth-bounded package layout, and assembles a dependency graph
(Go: real `go list -json` package edges via the extracted `internal/golist`
leaf; other ecosystems: declared manifest dep names). The typed `Inventory`
(carrying `schemaVersion`) is written deterministically — every list sorted,
`MarshalIndent` + trailing newline, byte-stable on re-run — to the well-known
`.workflow/analysis.json` (overridable via `--out`), and a concise summary
(primary language, build/test signal, locale count, package count, graph edge
count) is printed to stdout. Analyze is diagnostic, not a gate: any sub-detector
failure degrades to a best-effort/empty result with a recorded reason and the
command still exits 0. The only hard failures are an un-writable output path and
a non-existent/unreadable analysis root.

#### Gherkin Scenarios
Full spec: `specs/deep-codebase-analysis.feature` (scenario titles map 1:1 to Go
acceptance tests under `tests/acceptance/`).

- **Analyzing a Go module writes a complete inventory and prints a summary** —
  happy path (AC-1/2/4): `analysis.json` written; `schemaVersion=1`;
  `primaryLanguage=Go`; go-mod + make manifests; non-empty `go-packages` graph;
  stdout summary.
- **Re-running analyze on an unchanged repo produces a byte-identical inventory**
  — determinism (AC-3); sorted/stable lists.
- **Analyzing a Node project detects the npm manifest with build and test
  scripts and declared deps** — polyglot/non-Go manifest path; scripts → build/
  test; sorted deps.
- **A polyglot repo counts every language and picks the highest-count primary
  deterministically** — multi-language counting; alphabetical tiebreak; all
  manifests listed.
- **Analyzing a repo with locale files lists the detected locale codes** /
  **A repo with no i18n reports an empty locale list and exit 0** — i18n
  detection + empty-locale non-error.
- **The scan skips dependency and build directories so counts reflect real
  source** — skip set + gitignore + read-only guarantee (AC-5/6).
- **A repo with no recognized manifest still produces a valid inventory and
  exits 0** — best-effort on unfamiliar repos (AC-7).
- **A malformed package.json is recorded as detected-but-unparsable and the scan
  continues** — manifest parse degradation.
- **When go list fails the Go graph is recorded as best-effort empty with a note
  and the rest still emits** — Go graph degradation.
- **Running analyze with an un-writable output path fails clearly with a
  non-zero exit and writes no partial inventory** — NEGATIVE path; the canonical
  hard error; no partial artifact.
- **Running analyze against a non-existent or unreadable root fails clearly and
  writes no inventory** — NEGATIVE path; unreadable root.
- **The --out flag redirects the inventory to a custom path** — `--out`
  override.
- **An empty or docs-only repo yields a valid empty inventory and exits 0** —
  empty-repo edge.

#### UX States
`centinela analyze` is a CLI; states are described as stdout/stderr/exit-code
behaviors.

| State   | Trigger | Surface |
|---------|---------|---------|
| loading | Scan in progress on a large repo | Synchronous; the only load-bearing output is the final summary. May print a single "Analyzing <root>…" line; not asserted. |
| empty   | Empty / docs-only / no-manifest repo | Exit 0; valid inventory with empty `manifests`/`graph` and `primaryLanguage=""`; summary shows 0 locales / 0 edges. The empty state is success, not error. |
| error   | Un-writable output path or non-existent/unreadable root | Non-zero exit; clear error message on stderr naming the cause; NO partial/corrupt `.workflow/analysis.json` left behind. Sub-detector failures are NOT this state — they degrade best-effort and stay exit 0. |
| success | Normal scan completes | Exit 0; `.workflow/analysis.json` (or `--out` target) written deterministically; stdout summary: primary language, build/test signal, locale count, package count, graph edge count. |

#### Edge Cases
Enumerated in full in `.workflow/deep-codebase-analysis-edge-cases.md` (16 cases
mapped to spec scenarios). Highlights: empty/docs-only repo; polyglot with
alphabetical tiebreak; `go list` failure → best-effort empty graph + note;
malformed `package.json` → detected-but-unparsable; no-i18n → empty list;
depth-bounded huge-repo layout; byte-identical re-run; symlinks/unreadable files
skipped; skip set + gitignore; un-writable output (hard error, no partial
artifact); non-existent root (hard error); read-only guarantee; `--out`
override; `schemaVersion` stability.

#### Out-of-Scope
- **LLM inference of archetype / specs / adoption baseline** — the deliberate job
  of downstream Phase 9 features (`archetype-inference-project-synthesis`,
  `spec-reconstruction`, `adoption-baseline`); this feature is the deterministic
  substrate only. (Deliberate exclusion — not a new defer.)
- **Source-level import graphs for non-Go languages** — v1 records declared
  manifest deps only. (Already deferred by big-thinker.)
- **Broad framework fingerprinting** beyond manifest scripts. (Already deferred.)
- **Incremental / cached re-analysis** of only changed directories. (Already
  deferred.)
- **Metrics enrichment** (LOC, complexity, churn, coverage inference). (Already
  deferred.)
- Editing/normalizing source files, committing the artifact for the user, or
  acting as a blocking gate — analyze is read-only and diagnostic only.

#### Deferred Findings
- none. All out-of-scope items are either the deliberate downstream-feature
  exclusion or already deferred by the big-thinker
  (`non-go-source-import-graphs`, `brownfield-framework-fingerprinting`,
  `incremental-codebase-analysis`, `codebase-metrics-enrichment`). No new gaps
  discovered during acceptance authoring.

#### Handoff
- Next role: senior-engineer
- Open clarifications: none — the plan answered the three big-thinker open
  questions (output path `.workflow/analysis.json`, `internal/golist` extraction
  = option a, the v1 manifest set + build/test/framework signals). The spec
  pins exit-code semantics (best-effort exit 0 vs hard-error non-zero on
  un-writable output / unreadable root) for the implementation to honor.
