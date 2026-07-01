# Roadmap

> Centinela is a harness-governance layer for AI coding agents. Its guiding
> principle: **treat every agent failure as an engineering problem to fix
> permanently, not a prompt to retry — make correctness enforced, not requested.**
> Each item below names the agent or team failure mode it permanently fixes.
>
> **Capability-spectrum principle.** Centinela must govern *any* model — from a
> small local model running on a laptop to a frontier reasoning model. The
> weaker the model, the more scaffolding it needs (step-gating, artifact
> templates, confirmations); the stronger the model, the more enforcement
> shifts to outcomes (gates + claim verification at delivery). What never
> varies is independent mechanical verification: no model's claims are
> trusted, ever. **Governance adapts; verification is constant.**
>
> Status is derived from workflow state — run `centinela roadmap` for the live view.
> Phases are ordered by leverage and dependency; scheduling follows the
> `dependsOn` graph, not the phase number.

## Phase 0: Bootstrap

- **docs-migration-managed-docs** — managed, incrementally upgradable scaffolded docs via `centinela migrate`.

## Phase 1: Harness Capabilities

- **governed-project-memory** — a structured, git-tracked memory ledger that harvests Centinela's own artifacts (edge-case lessons, gate verdicts, decisions) at step boundaries and recalls the relevant slice into later features via the existing context-injection hook. Governed (reviewable, deterministic, no semantic-store dependency), not fuzzy recall.

## Phase 2: Configurable Model Routing

- **configurable-model-routing** — runner-keyed tier→model remapping plus per-role concrete-model overrides (`claude` | `opencode` | `codex`), so operators route reasoning/coding work to the models they actually want while every unconfigured role keeps the built-in default. The foundation the capability-adaptive phase builds on.

## Phase 3: Close the Mechanical-Verification Gap

- **g2-import-graph-gate** — Mechanically enforce per-archetype layer-dependency rules by parsing the import graph (Go `go/packages`, TS madge/ts-morph, Python AST). Turns the prose G2 rule in `PROJECT.md` into a checkable allow/deny matrix.
  *Fixes: agent silently introduces a forbidden cross-layer import that no gate catches.*
- **security-gate** — Mechanical secret-scanning + dependency-vuln audit (`gitleaks` / `osv-scanner` / `govulncheck`, configurable) wired into validate.
  *Fixes: agent commits a secret or pulls a vulnerable dependency; subagent review misses it.*
- **spec-traceability-gate** — Verify every Gherkin scenario in `specs/*.feature` maps to an executed step definition in `tests/acceptance/`.
  *Fixes: spec and acceptance tests drift apart; scenarios silently go unimplemented.*
- **g2-multi-language-import-graph**

## Phase 4: Loop Velocity

> **Pulled forward (2026-06-11).** The per-feature loop charges full ceremony —
> seven sequential specialist roles, five manual confirmations, full docs
> regeneration — regardless of feature size or surface. The 5-step is
> feature-shaped and frontier-shaped; real work also includes fixes, refactors,
> and spikes. These features cut the round-trips while keeping every gate and
> claim verification intact, and they pay back on every remaining roadmap item —
> so they ship before the long tail. Verification stays constant; only
> *requested* process shrinks.
>
> **Dogfood note (2026-06-10).** The `code-quality-hardening` fix (PR #23) was
> built through the full five-step workflow with seven specialist subagents and
> five manual confirmations — ~370K subagent tokens — for a ~250-line semantic
> change across ten source files. The discipline paid off (subagents corrected
> two reviewer claims, caught two missing spec scenarios, and independently
> re-verified the highest-risk change), but the ceremony-to-change ratio on an
> internal fix is exactly the cost `enforcement-profiles` (an `outcome` profile)
> and `right-size-docs-step` exist to cut: same gates and claim verification, a
> fraction of the round-trips. First-party evidence that this phase earns its
> place.

- **enforcement-profiles** — Named governance strictness presets — `strict` (the back-compat default: full step-gating, per-step confirmation, mandatory subagent evidence — exactly today's behavior), `guided` (a lighter opt-in: step-gating on, but no mandatory subagent ceremony and review only after planning), `outcome` (work in any order, prompts suppressed; `complete`/merge still requires all gates + claim verification green) — selectable in `centinela.toml` per project and overridable per feature with `centinela start --profile`. Decouples *how much process is enforced* from *whether outcomes are verified*; the latter is constant across all profiles.
  *Fixes: one-size enforcement either burdens a strong model with ceremony or under-scaffolds a weak one — both end in governance being switched off.*
- **workflow-archetypes** — First-class lightweight tracks beside the 5-step: `hotfix` (reproduce → fix → test → ship), `refactor` (characterize → change → prove-equivalent), `spike` (timeboxed, no ship gate).
  *Fixes: forcing a diagnosis/bugfix through a plan→docs pipeline it doesn't fit.*
- **right-size-docs-step** — Make the `docs` step surface-aware, mirroring how the `code` step requires `ux-ui-specialist` only for `surface: user-facing` features. A user-facing feature still writes the plain-language KB guide (the one genuinely valuable, reader- and memory-useful doc artifact); an internal refactor / bugfix / chore instead emits a one-line changelog entry (via `delivery-artifact-generation`, stubbed as a plain entry until that feature ships) and skips the KB guide, the per-feature HTML portal regeneration, and the documentation-specialist evidence ceremony. The 108 KB `index.html` portal moves to merge/release-time regeneration rather than per-feature.
  *Fixes: mandatory full docs ceremony — KB guide, HTML regen, evidence — runs on every feature including internal ones with no end-user story, burning tokens for zero reader value.*

## Phase 5: Operability & DX

> Reduce the friction of running and recovering Centinela itself — keep its own
> artifacts honest and self-healing. Cheap, dependency-free wins that smooth
> every later phase, so they come early.

- **roadmap-doc-sync** — Treat `.workflow/roadmap.json` as the source of truth and generate the human-readable `ROADMAP.md` from it, with a drift check that fails if the two disagree. *(Dogfood note: this roadmap has now drifted twice while being maintained by hand — exactly the toil this removes.)*
  *Fixes: the machine roadmap and the human roadmap drift apart because they're maintained by hand.*
- **centinela-doctor** — `centinela doctor`: one command that diagnoses (and, where safe, repairs) broken hook wiring, stale or orphaned `.workflow` state, abandoned worktrees, config drift, and roadmap drift — extending the existing `evidence repair` into a holistic health check.
  *Fixes: a broken or drifted install fails opaquely with no guided path back to a healthy state — the graceful-recovery gap.*
- **deferred-findings-roadmap-capture** — When plan-step agents (big-thinker, feature-specialist) detect something outside the current feature's scope, or code/tests-step agents (senior-engineer, qa-senior) surface findings they won't fix immediately, route that information into the roadmap instead of losing it: a `centinela roadmap` capture path that appends the finding to the roadmap together with the analysis/quality entries `roadmap validate` demands, plus contract updates in the four role prompts (and their scaffold mirrors) making the capture mandatory whenever such a finding exists. Exact capture surface (dedicated backlog phase vs deferred-findings ledger promoted at triage) is decided at plan.
  *Fixes: out-of-scope discoveries and deferred fixes live only in per-feature prose artifacts (Out-of-Scope, Residual Risks, Outstanding TODOs) and evaporate — they never reach the roadmap, the single planning source of truth.*
- **cli-self-update** — Add `centinela update`: a self-update command plus a passive startup version notice. The command resolves the latest GitHub Release, downloads the prebuilt asset matching the host OS/arch (`centinela-v<tag>-<goos>-<goarch>[.exe]`), verifies it against the release `SHA256SUMS`, and atomically replaces the running binary (write a temp file in the SAME directory as the target, fsync, copy the existing mode bits, then `os.Rename`). A `--check` flag reports availability without installing. The startup notice is a throttled, non-blocking check wired into the existing SessionStart hook, comparing the running version to the latest release tag. The version-check cache is a single JSON file at an XDG-rooted path: `${XDG_CACHE_HOME:-~/.cache}/centinela/update-check.json` holding the last-seen latest tag and a unix timestamp; a check within the TTL (default 24h) reads the cache and performs NO network call. Lives in a new leaf package `internal/selfupdate` (net/http + crypto/sha256 + os + encoding/json only); `cmd/centinela` wires the `update` command and the notice into the session hook. All HTTP (GitHub releases API, asset download, SHA256SUMS) is injected behind an interface so tests drive an `httptest.Server`. Acceptance: (1) on an outdated binary, `centinela update` downloads the matching asset, the SHA256 matches `SHA256SUMS`, the binary is replaced atomically, and it prints old->new versions; on the latest it is a no-op printing `already up to date`. (2) `centinela update --check` performs the version check (honoring the TTL cache), prints the availability verdict, exits non-zero when a newer version exists and zero when current, and makes ZERO writes to the binary or any temp file. (3) A `SHA256SUMS` mismatch aborts WITHOUT touching the installed binary (fail-safe) and removes the temp file. (4) An unsupported platform / missing matching asset returns a clear typed error with no partial write. (5) When the target binary's directory is not writable (permission denied), `centinela update` returns a clear typed error, leaves the installed binary untouched, and cleans up any temp file created before the failure. (6) The startup notice is cache-throttled (a second start within the TTL performs no network call — asserted against the httptest call count), fails silent when offline/rate-limited (no error, no block), and appears only when running < latest. (7) GitHub API + asset download + SHA256SUMS + the cache file are exercised against an `httptest.Server` and a temp HOME/XDG dir so tests are deterministic and offline.
  *Fixes: users running an installed release binary have no in-tool upgrade path — they must manually rebuild or re-download, so they silently keep running stale governance.*

## Phase 6: Capability-Adaptive Governance

> One-size enforcement fails at both ends of the model spectrum: it taxes a
> frontier model with ceremony it no longer needs, and it under-scaffolds a
> small local model that needs *more* rails, not fewer. Both failure modes end
> the same way — the user disables governance. Building on the
> `enforcement-profiles` presets shipped in Phase 4, this phase makes the
> amount of process a function of model capability, while verification stays
> constant for everyone.

- **model-capability-profiles** — A registry mapping each configured model — cloud or local — to a declared capability profile (instruction-following reliability, tool-use reliability, context budget) that selects its default enforcement profile and routing tier. Declarable in `centinela.toml` so any Ollama / llama.cpp / OpenAI-compatible local model is as first-class as a frontier model. (depends on enforcement-profiles, configurable-model-routing)
  *Fixes: model assumptions are hardcoded for frontier cloud models; a local model has no place to declare what it can and cannot reliably do.*
- **deterministic-artifact-scaffolds** — Pre-generated artifact skeletons with explicit fill-in slots (plan, spec, edge-case analysis) plus mechanical generation wherever content is derivable from existing state — extending the proven docs CLI-fallback pattern. Under the `strict` profile, weak models fill constrained templates instead of inventing structure. (depends on enforcement-profiles)
  *Fixes: a low-capability model fails artifact contracts on shape rather than substance, burning retries; the dumbest models need rails that are physical, not prose instructions.*
- **headless-governance** — Full non-interactive parity: every confirmation prompt gets a config/flag equivalent, and every run can emit a machine-readable end-of-run verdict packet (gate results, verify results, evidence index) as the reviewable output. The foundation for CI, Capataz fleets, and reviewing agent work by evidence instead of by transcript.
  *Fixes: prompts assume a human in a chat session; unattended runs stall on questions or silently bypass them.*

## Phase 7: Instrument the Loop

> You can't "fix failures permanently" if you can't see them. Make Centinela
> observe itself — including how each *model* actually performs under
> governance.

- **governance-telemetry** — Local, git-tracked append-only event log of every block, gate failure, verify rejection, and rework cycle (no external service, consistent with the governed-memory design).
- **centinela-insights** — Reads telemetry and reports most-triggered blocks, most-failed gates, features with the most rework, and mean steps-to-green. (depends on governance-telemetry)
  *Fixes: roadmap prioritization by anecdote instead of evidence.*
- **failure-ledger-plan-advisor** — Feed recurring gate failures from the ledger into the plan advisor so the next feature is pre-warned about the failure modes that actually bite this repo. (depends on governance-telemetry, governed-project-memory)
- **capability-calibration** — Read per-model telemetry (block rate, gate failures, verify rejections, rework cycles) and report whether each model is over- or under-governed, recommending an enforcement-profile change backed by evidence. Answers "how much scaffolding does *this* model need on *this* repo" with measurement instead of vibes. (depends on governance-telemetry, model-capability-profiles)
  *Fixes: enforcement-profile assignment by intuition; a model that quietly needs tighter (or looser) governance is never recalibrated.*

## Phase 8: Continuous Governance

> Governance currently evaporates at merge. Extend it beyond build-time —
> and open the gate engine so teams bring their own rules.

- **custom-gate-sdk** — Let teams define their own mechanical gates (project-specific rules) via config/plugin and run them inside `centinela validate` — without forking Centinela. Generalizes the built-in gate interface so the built-in gates become reference implementations, not the ceiling. This is the pivot from "opinionated workflow tool" to "policy engine": Centinela's own rules become the default profile, not the architecture. (depends on g2-import-graph-gate)
  *Fixes: teams with bespoke rules have no enforced path short of forking, so those rules stay "requested, not enforced."*
- **audit-baseline-ratchet** — Whole-repo gate scan that records a baseline and *ratchets*: never lets new violations in, lets teams pay down old ones over time. Enables adoption on legacy codebases without a big-bang cleanup. (depends on g2-import-graph-gate)
  *Fixes: the "we'll never reach zero so we disable the gate" failure.*
- **precommit-and-pr-gate** — Run the mechanical gates as a fast pre-commit hook and post gate verdicts as PR review comments. (depends on audit-baseline-ratchet)
  *Fixes: violations discovered only at the end of the loop, maximizing rework cost.*

## Phase 9: Brownfield Onboarding

> Centinela excels on greenfield projects, where it interviews the user and
> generates a roadmap from a blank slate. On an existing codebase the truth
> already lives in the code and must be *reverse-engineered*, not interviewed.
> This phase makes Centinela adoptable on a mature repo without a big-bang rewrite.

- **deep-codebase-analysis** — `centinela analyze`: scan the existing repo and produce a machine-readable inventory — language(s), framework, build/test setup, i18n locales, module/package layout, and the current import/dependency graph. The foundation every other brownfield feature reads from.
- **archetype-inference-project-synthesis** — From the analysis, infer the best-fit architecture archetype and draft a complete `PROJECT.md` (layer mapping, folder structure, naming conventions, gatekeeper paths) reflecting the code *as it is*, for the user to confirm or correct — instead of interviewing from scratch. (depends on deep-codebase-analysis)
  *Fixes: forcing a brownfield user to hand-author PROJECT.md and guess an archetype that doesn't match reality.*
- **spec-reconstruction** — Have the LLM read existing modules, endpoints, and flows and generate `specs/*.feature` Gherkin scenarios plus `docs/features/*.md` briefs that document the behavior the system *already* exhibits, as confirmable, editable acceptance criteria. (depends on deep-codebase-analysis)
  *Fixes: an existing system with zero specs, so Centinela's spec-first gates have nothing to anchor to.*
- **brownfield-roadmap-generation** — Generate a roadmap that distinguishes already-built capability (recorded as baseline / done) from the net-new work and gaps (TODOs, incomplete areas, user-stated goals), so existing functionality isn't re-planned and the team can immediately `centinela start` the next real gap. (depends on archetype-inference-project-synthesis, spec-reconstruction)
  *Fixes: a greenfield roadmap generator that assumes everything is still ahead of you.*
- **adoption-baseline** — On brownfield init, record current gate violations (pre-existing G1 file-size, G2 layer, etc.) as an accepted baseline so `validate` isn't drowned by thousands of legacy findings; only new work is gated strictly while legacy debt is ratcheted down over time. (depends on deep-codebase-analysis, audit-baseline-ratchet)
  *Fixes: a mature repo where day-one `validate` reports thousands of pre-existing violations, making the gates unusable on adoption.*

## Phase 10: Delivery

> Completed work should ship with its evidence — and the humans reviewing
> agent output at scale review evidence, not transcripts.

- **completion-delivery-prompt** — When a feature reaches completion (final step done), Centinela asks how to deliver it and acts on the choice: **(a)** commit, push, and open a PR via `gh` when a git `origin` remote is configured, or **(b)** merge the branch into `main` locally using the merge-steward (`centinela merge <feature>`, with `--continue` recovery on conflicts). Detects remote presence to offer only valid options; never pushes or merges without explicit confirmation.
  *Fixes: completed work stalls in its worktree because the agent doesn't know the team's delivery convention, or merges/pushes without asking.*
- **delivery-artifact-generation** — Compose the PR description and a `CHANGELOG` / release-notes entry automatically from evidence Centinela already holds — the feature brief, plan, gatekeeper report, and verify results — so delivery output is consistent and traceable instead of hand-written each time. (depends on completion-delivery-prompt)
  *Fixes: rich step evidence exists but the PR body and changelog are still written from scratch (or skipped).*
- **team-dashboard** — Aggregate view across worktrees and contributors: who owns what feature/step, aggregate gate health, roadmap burn-down. (depends on governance-telemetry)
  *Fixes: multi-feature, multi-contributor state is invisible without polling each worktree by hand.*
- **cost-governance** — Per-feature/per-step token and model-tier budgets surfaced from the host harness, with a soft gate. For local models the budget unit is wall-clock/compute rather than spend, but the runaway-step problem is identical.
  *Fixes: a runaway agent burning budget on one step with no visibility.*

## Phase 11: Ecosystem

> Bring-your-own-harness is the moat: a vendor-neutral governance contract
> across every harness and every model class, cloud or local.

- **host-harness-adapters** — Extend the hook/parity pattern (today Claude + OpenCode) to Cursor, Aider, Windsurf, and Copilot Workspace. Best done after a shared integration core, so adding a third target isn't a maintenance tax.
- **codex-support** — First-class support for OpenAI **Codex** as a host harness alongside Claude Code and OpenCode: wire Codex's config/integration surface to Centinela's prewrite enforcement, postwrite status tags, and prompt-context injection, generate the Codex-native rules file (e.g. `AGENTS.md`), and extend the parity tests so the three harnesses stay in lockstep through `init` and `migrate`. Codex has a large and growing user base, so dual-support with Claude Code materially widens who can adopt Centinela. (depends on host-harness-adapters)
  *Fixes: Codex users locked out of Centinela's governance because enforcement is wired only for Claude Code / OpenCode.*
- **local-harness-support** — First-class local-model targets: OpenCode backed by Ollama, plus generic OpenAI-compatible endpoints (llama.cpp, vLLM, LM Studio). Acceptance bar: a small local model completes a governed feature end-to-end using the `strict` profile and deterministic scaffolds, with all gates and claim verification passing. (depends on host-harness-adapters, model-capability-profiles)
  *Fixes: local-model users — the fastest-growing segment that needs governance the most — can't adopt Centinela because integration and enforcement assume frontier cloud harnesses.*
- **cross-project-memory** — Promote governed-memory lessons from per-repo to an org-shared ledger, so a failure fixed in one repo warns every repo. (depends on governed-project-memory)
- **mcp-governance-server** — Expose Centinela's rules, gate engine, and claim-verification as a versioned MCP server (v1 tools: read_rules, run_gates, verify_claims, workflow_state) reusing the headless-governance verdict packet as the wire payload. The MCP server is advisory-by-protocol: it returns a structured verdict (allow/warn/block) — it cannot itself stop a write — so enforcement stays harness-side via a thin shim that maps a block verdict onto the harness's existing pre-write deny (the same contract the native hook uses). The tool surface and verdict schema are explicitly versioned so harnesses pin a compatibility level. Acceptance: a harness with zero Centinela-specific code obtains a verdict purely through MCP tool calls; the shim test asserts a block verdict aborts the write while allow proceeds; a parity test asserts the MCP verdict and the native-hook verdict are identical for the same diff and workflow state. (depends on headless-governance)
  *Fixes: every new host harness needs a bespoke hook/parity adapter, so governance can only be consumed by harnesses Centinela has explicitly integrated.*
- **agents-md-canonical-surface** — Make AGENTS.md the canonical rules surface via a managed region fenced by `<!-- centinela:managed:start -->` / `<!-- centinela:managed:end -->`: init/migrate emit Centinela's hard rules and active profile into that region, and Centinela re-ingests a constrained, documented directive schema (rule toggles, profile, locales, layer dependency overrides) authored outside the managed block. Free-form prose outside the managed block and outside the known schema is preserved verbatim and ignored by the rule engine — never silently honored — and any directive that conflicts with a hard rule surfaces as a lint WARNING rather than an override. Acceptance: init produces a valid AGENTS.md with the fenced managed region; round-trip (emit -> hand-edit outside the region -> re-emit) preserves user content byte-for-byte; a known-schema directive is ingested into the rule set; a conflicting directive produces a lint warning; an unrecognized free-form line is preserved and ignored.
  *Fixes: Centinela's rules live in a Centinela-specific surface, so the 60K+ repos and AGENTS.md-reading harnesses cannot read or contribute to Centinela governance.*
- **sdd-spec-kit-interop** — Position Centinela as the enforcement layer for Spec-Driven Development with a pinned mapping: GitHub Spec Kit `spec.md` requirements -> Gherkin Scenarios in specs/*.feature, `plan.md` -> docs/plans, `tasks.md` -> step checklist; AWS Kiro requirements map the same way. Each imported requirement receives a stable traceability ID linked to its Scenario; 'links intact' means spec-traceability-gate tracks that ID and FAILS when no code/test covers it. Export regenerates a Spec-Kit-compatible `spec.md` from Centinela's Gherkin + plan artifacts, round-tripping lossily-mapped external fields (e.g. Kiro design rationale) through a preserved sidecar block. External formats are pinned to committed fixtures (Spec Kit v0.8.x and a Kiro schema snapshot under testdata). Acceptance: a fixture spec.md imports with traceability IDs attached; the gate fails when a tracked requirement loses coverage; export round-trips the fixture without dropping sidecar fields. (depends on spec-traceability-gate)
  *Fixes: teams adopting Spec-Driven Development (Spec Kit, Kiro) generate specs that nothing enforces, so spec and shipped code silently diverge.*
- **agent-action-provenance** — Produce a hash-chained, signed provenance trail of every gate verdict, claim-verification result, and step transition, exported as an in-toto Statement carrying a SLSA v1 provenance predicate. Signing uses a repo-scoped key created by `centinela attest init` (private key in durable .workflow state or a git-config-referenced path; public key committed for verification); org deployments may swap in a KMS/sigstore backend behind the same signer interface (interface only in v1). Each record names the model/harness that made the claim and the gate that cleared it. Acceptance: each completed step appends a hash-chained record; `centinela attest <feature>` emits a verifiable in-toto/SLSA bundle; verification passes on an untampered bundle and fails when any prior record is altered; the bundle attributes each claim to its model/harness. (depends on governance-telemetry)
  *Fixes: gate verdicts and claim-verification results aren't tamper-evident or exportable, so an enterprise can't audit or attest to what an agent was permitted to do.*
- **agent-threat-security-gate** — Extend the security gate with per-class agent-native threat detection: (a) secret/credential exfiltration — deterministic regex+entropy signature scan over added diff lines, plus flagging new outbound-network calls introduced adjacent to a secret read; (b) prompt injection — a heuristic signature set over agent-fetched/MCP-sourced content, explicitly best-effort (WARNING, not a completeness claim); (c) unexpected capability — a diff introducing a new network/exec/credential API absent from the pre-change tree. Acceptance: runs against a committed corpus (testdata/agent-threats/ with planted-positive and clean cases); a planted exfiltration diff BLOCKs under the strict profile; an injected-instruction fixture WARNs; zero false positives on the existing security-gate clean corpus (the numeric bar = 0 FP on that fixed corpus). (depends on security-gate)
  *Fixes: the security gate scans for conventional vulnerabilities but not agent-native threats (prompt injection from fetched/MCP content, secret exfiltration in diffs), the attack surface that exists only because an autonomous agent wrote the code.*

## Phase 12: Self-Improvement

> Centinela applies its own founding principle to itself: a recurring friction
> signal is an engineering problem to fix permanently — by generating automation.
> Comes last: it depends on mature instrumentation from Phase 7.

- **adaptive-skill-synthesis** — When telemetry shows a workflow isn't running as expected — e.g. the same feature/step requires too many manual validation rounds, repeated rework, or the user keeps hand-correcting the same thing — Centinela proposes and scaffolds a **skill** (a Claude Code / OpenCode skill, or a reusable validate command / sub-agent prompt) that automates the repetitive step further. Generated skills are reviewable artifacts the user approves before they take effect — never silently self-modifying. (depends on governance-telemetry, centinela-insights)
  *Fixes: the human becomes the slow loop, re-validating the same step over and over, with no path for the harness to absorb that toil.*
- **cost-observability-ledger** — A historical, queryable ledger layered on cost-governance: records per-feature/per-step token and model-tier spend over time, supports showback/chargeback rollups, and flags anomalies (e.g. a step whose cost regressed 4x). Complements the preventive budget gate with explanatory analytics, mirroring the proven telemetry/insights append-and-rollup pattern. Acceptance: each governed step appends a spend record; `centinela cost report` rolls up by feature/step/model; an injected cost spike triggers an anomaly flag; the ledger survives across features and is git-tracked durable state. (depends on cost-governance, governance-telemetry)
  *Fixes: per-step budgets prevent runaway spend but leave no history, so teams cannot attribute, trend, or anomaly-detect token cost across features and steps.*
- **governance-roi-eval-harness** — Deterministic A/B evaluation over a committed fixture: a recorded multi-step agent transcript whose steps introduce planted, oracle-known bugs. The intervention model is veto-and-drop (no live agent, no corrective branching): the governed arm replays each recorded step and runs gates + claim-verification against that step's diff; a step whose diff fails a gate or makes a false claim is REJECTED and its diff is not applied to the accumulating tree, so the governed final diff omits the vetoed buggy steps. The ungoverned arm applies every recorded step unconditionally. Both arms are fully deterministic because the transcript is fixed and the gates are deterministic. Metrics: surviving-bug count (planted bugs present in each arm's final diff per the oracle test suite), plus vetoed-step count, claim-verification rejections, and token cost; output is a reviewable delta report. Acceptance: `centinela eval` yields identical results on repeated runs of the fixture; the governed arm's surviving-bug count is strictly lower than the ungoverned arm's because vetoed steps don't land; a deliberately weakened profile (fewer active gates) vetoes fewer steps and measurably raises the governed arm's surviving-bug count. (depends on governance-telemetry, centinela-insights)
  *Fixes: Centinela asserts it improves correctness but can't measure it, so there's no governed-vs-ungoverned baseline to prove ROI or catch governance regressions over time.*

## Backlog

- **rawio-reformat-diff-churn** — First defer/promote reformats untouched phases of roadmap.json, creating spurious git diff churn *(deferred 2026-06-12T16:29:13Z · deferred-findings-roadmap-capture/senior-engineer)*
- **roadmap-import-graph-layer-mapping** — Map internal/roadmap as an import_graph layer so the gates->roadmap and ui->roadmap read-only edges are mechanically enforced, not just documented in PROJECT.md G2 *(deferred 2026-06-14T10:30:31Z · roadmap-doc-sync/gatekeeper)*
- **non-go-source-import-graphs** — Parse source-level import graphs for non-Go languages (JS/TS, Ruby, Rust, Python); analyze v1 records declared manifest deps only *(deferred 2026-06-17T21:17:33Z · deep-codebase-analysis/big-thinker)*
- **brownfield-framework-fingerprinting** — Detect frameworks (Rails/Next/Django/etc.) via directory+dependency heuristics beyond manifest scripts in centinela analyze *(deferred 2026-06-17T21:17:33Z · deep-codebase-analysis/big-thinker)*
- **incremental-codebase-analysis** — Incremental/cached re-analysis that only re-scans changed directories in centinela analyze *(deferred 2026-06-17T21:17:33Z · deep-codebase-analysis/big-thinker)*
- **codebase-metrics-enrichment** — Enrich analyze inventory with LOC, complexity, churn, and test-coverage inference *(deferred 2026-06-17T21:17:33Z · deep-codebase-analysis/big-thinker)*
- **brownfield-route-flow-extraction** — Framework-specific HTTP route and call-flow extraction for reconstructed specs; spec-reconstruction v1 ships package/manifest-derived targets only *(deferred 2026-06-22T07:52:17Z · spec-reconstruction/big-thinker)*
- **centinela-changelog-subcommand** — Standalone centinela changelog subcommand for merge-path changelog parity (independent of PR delivery) *(deferred 2026-06-25T16:36:23Z · delivery-artifact-generation/big-thinker)*
- **gitignore-durable-state-guard** — Add a test asserting durable .workflow state (roadmap{,-analysis,-quality}.json, audit-baseline.json) stays tracked/not-gitignored, so the evidence-footprint allow-list can't silently drop bootstrap state again (regression fixed in 420ab3c) *(deferred 2026-06-29T15:08:01Z · lean-evidence-footprint/validation-specialist)*
- **selfupdate-notice-http-timeout** — centinela update + startup notice use http.DefaultClient (no timeout); a stalled GitHub connection can block session start indefinitely — give the production Updater a bounded client timeout (~5s) so the cold-cache notice fails fast and silent *(deferred 2026-06-30T09:45:19Z · cli-self-update/gatekeeper)*
- **brownfield-onboarding-docs** — Document the brownfield onboarding path (analyze, synthesize, enrich, confirm) in new-project-guide.md *(deferred 2026-06-29T17:42:52Z · brownfield-setup-detection/big-thinker)*
- **brownfield-manifest-breadth** — Extend HasSource manifest list to pom.xml, composer.json, build.gradle for broader ecosystem coverage *(deferred 2026-06-29T17:46:35Z · brownfield-setup-detection/feature-specialist)*
- **unit-test-mcp-server-in-memory-transport** — Cover runMcpServe/mcpConnectSelf via an in-memory MCP transport *(deferred 2026-06-30T08:12:27Z · coverage-hardening/big-thinker)*
- **fault-inject-atomic-write-error-paths** — Cover WriteBytesAtomic and low-level I/O error branches via fault injection *(deferred 2026-06-30T08:12:27Z · coverage-hardening/big-thinker)*
- **unit-test-vuln-tool-external-seam** — Cover runVulnTool by stubbing the external vulnerability-scanner binary behind a test seam *(deferred 2026-06-30T08:12:27Z · coverage-hardening/big-thinker)*
- **aider-local-provider-wiring** — Point the Aider/Claude harness at a local endpoint; local block currently wires only OpenCode's provider surface *(deferred 2026-06-30T14:45:14Z · local-harness-support/big-thinker)*
- **workflow-save-atomic-write** — workflow.Save uses plain os.WriteFile; a crash mid-write can truncate .workflow/<feature>.json — make it write-temp-then-rename *(deferred 2026-06-30T17:52:51Z · workflow-revise-loop/validation-specialist)*
- **roadmap-crud-add-remove** — roadmap add/remove/rm + per-feature Draft flag (validate-exempt until scored) + generalized promote (in-place draft finalize); establishes generalized raw-feature helpers. Depends on roadmap-json-contract (done). *(deferred 2026-07-01T13:32:50Z · roadmap-editing-suite-design/big-thinker)*
- **roadmap-edit-move** — roadmap edit/update, move, reorder; dependent dependsOn rewrite on rename; cycle re-validation across renames/moves. Depends on roadmap-crud-add-remove. *(deferred 2026-07-01T13:32:50Z · roadmap-editing-suite-design/big-thinker)*
- **roadmap-phase-ops** — roadmap phase add/rename/remove with dirty-map reindex in the raw layer; refuse non-empty phase remove unless --force. Highest raw-layer complexity, land last. *(deferred 2026-07-01T13:32:50Z · roadmap-editing-suite-design/big-thinker)*
