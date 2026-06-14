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

> The flagship promise — separation of concerns / layer dependencies — was
> a *manual* code-review gate. Convert the remaining "requested" gates into
> mechanically enforced ones, then open the gate engine to teams' own rules.

- **g2-import-graph-gate** — Mechanically enforce per-archetype layer-dependency rules by parsing the import graph (Go `go/packages`, TS madge/ts-morph, Python AST). Turns the prose G2 rule in `PROJECT.md` into a checkable allow/deny matrix.
  *Fixes: agent silently introduces a forbidden cross-layer import that no gate catches.*
- **security-gate** — Mechanical secret-scanning + dependency-vuln audit (`gitleaks` / `osv-scanner` / `govulncheck`, configurable) wired into validate.
  *Fixes: agent commits a secret or pulls a vulnerable dependency; subagent review misses it.*
- **spec-traceability-gate** — Verify every Gherkin scenario in `specs/*.feature` maps to an executed step definition in `tests/acceptance/`.
  *Fixes: spec and acceptance tests drift apart; scenarios silently go unimplemented.*

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

## Phase 12: Self-Improvement

> Centinela applies its own founding principle to itself: a recurring friction
> signal is an engineering problem to fix permanently — by generating automation.
> Comes last: it depends on mature instrumentation from Phase 7.

- **adaptive-skill-synthesis** — When telemetry shows a workflow isn't running as expected — e.g. the same feature/step requires too many manual validation rounds, repeated rework, or the user keeps hand-correcting the same thing — Centinela proposes and scaffolds a **skill** (a Claude Code / OpenCode skill, or a reusable validate command / sub-agent prompt) that automates the repetitive step further. Generated skills are reviewable artifacts the user approves before they take effect — never silently self-modifying. (depends on governance-telemetry, centinela-insights)
  *Fixes: the human becomes the slow loop, re-validating the same step over and over, with no path for the harness to absorb that toil.*

## Backlog

- **rawio-reformat-diff-churn** — First defer/promote reformats untouched phases of roadmap.json, creating spurious git diff churn *(deferred 2026-06-12T16:29:13Z · deferred-findings-roadmap-capture/senior-engineer)*
- **roadmap-import-graph-layer-mapping** — Map internal/roadmap as an import_graph layer so the gates->roadmap and ui->roadmap read-only edges are mechanically enforced, not just documented in PROJECT.md G2 *(deferred 2026-06-14T10:30:31Z · roadmap-doc-sync/gatekeeper)*
