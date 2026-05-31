# Roadmap

> Centinela is a harness-governance layer for AI coding agents. Its guiding
> principle: **treat every agent failure as an engineering problem to fix
> permanently, not a prompt to retry — make correctness enforced, not requested.**
> Each item below names the agent or team failure mode it permanently fixes.
>
> Status is derived from workflow state — run `centinela roadmap` for the live view.
> Phases are ordered by leverage and dependency; scheduling follows the
> `dependsOn` graph, not the phase number.

## ✅ Phase 0: Bootstrap

- **docs-migration-managed-docs** — managed, incrementally upgradable scaffolded
  docs via `centinela migrate`.

## ✅ Phase 1: Harness Capabilities

- **governed-project-memory** — a structured, git-tracked memory ledger that
  harvests Centinela's own artifacts (edge-case lessons, gate verdicts,
  decisions) at step boundaries and recalls the relevant slice into later
  features via the existing context-injection hook. Governed (reviewable,
  deterministic, no semantic-store dependency), not fuzzy recall.

## Phase 2: Close the Mechanical-Verification Gap

> The flagship promise — separation of concerns / layer dependencies — is still
> a *manual* code-review gate. Convert the remaining "requested" gates into
> mechanically enforced ones.

- **g2-import-graph-gate** — Mechanically enforce per-archetype layer-dependency
  rules by parsing the import graph (Go `go/packages`, TS madge/ts-morph, Python
  AST). Turns the prose G2 rule in `PROJECT.md` into a checkable allow/deny matrix.
  *Fixes: agent silently introduces a forbidden cross-layer import that no gate catches.*
- **security-gate** — Mechanical secret-scanning + dependency-vuln audit
  (`gitleaks` / `osv-scanner` / `govulncheck`, configurable) wired into validate.
  *Fixes: agent commits a secret or pulls a vulnerable dependency; subagent review misses it.*
- **spec-traceability-gate** — Verify every Gherkin scenario in `specs/*.feature`
  maps to an executed step definition in `tests/acceptance/`.
  *Fixes: spec and acceptance tests drift apart; scenarios silently go unimplemented.*
- **custom-gate-sdk** — Let teams define their own mechanical gates (project-specific
  rules) via config/plugin and run them inside `centinela validate` — without
  forking Centinela. Generalizes the built-in gate interface so the three Phase 2
  gates become reference implementations, not the ceiling.
  *Fixes: teams with bespoke rules have no enforced path short of forking, so those rules stay "requested, not enforced."* (depends on g2-import-graph-gate)

## Phase 3: Operability & DX

> Reduce the friction of running and recovering Centinela itself — keep its own
> artifacts honest and self-healing. Cheap, dependency-free wins that smooth
> every later phase, so they come early.

- **roadmap-doc-sync** — Treat `.workflow/roadmap.json` as the source of truth and
  generate the human-readable `ROADMAP.md` from it, with a drift check that fails
  if the two disagree. *(Dogfood note: this roadmap was hand-synced across both
  files repeatedly while drafting it — exactly the toil this removes.)*
  *Fixes: the machine roadmap and the human roadmap drift apart because they're maintained by hand.*
- **centinela-doctor** — `centinela doctor`: one command that diagnoses (and, where
  safe, repairs) broken hook wiring, stale or orphaned `.workflow` state, abandoned
  worktrees, config drift, and roadmap drift — extending the existing `evidence
  repair` into a holistic health check.
  *Fixes: a broken or drifted install fails opaquely with no guided path back to a healthy state — the graceful-recovery gap.*

## Phase 4: Instrument the Loop

> You can't "fix failures permanently" if you can't see them. Make Centinela
> observe itself.

- **governance-telemetry** — Local, git-tracked append-only event log of every
  block, gate failure, verify rejection, and rework cycle (no external service,
  consistent with the governed-memory design).
- **centinela-insights** — Reads telemetry and reports most-triggered blocks,
  most-failed gates, features with the most rework, and mean steps-to-green.
  *Fixes: roadmap prioritization by anecdote instead of evidence.* (depends on governance-telemetry)
- **failure-ledger-plan-advisor** — Feed recurring gate failures from the ledger
  into the plan advisor so the next feature is pre-warned about the failure modes
  that actually bite this repo. (depends on governance-telemetry, governed-project-memory)

## Phase 5: Continuous Governance

> Governance currently evaporates at merge. Extend it beyond build-time.

- **audit-baseline-ratchet** — Whole-repo gate scan that records a baseline and
  *ratchets*: never lets new violations in, lets teams pay down old ones over time.
  Enables adoption on legacy codebases without a big-bang cleanup.
  *Fixes: the "we'll never reach zero so we disable the gate" failure.* (depends on g2-import-graph-gate)
- **precommit-and-pr-gate** — Run the mechanical gates as a fast pre-commit hook
  and post gate verdicts as PR review comments.
  *Fixes: violations discovered only at the end of the loop, maximizing rework cost.* (depends on audit-baseline-ratchet)

## Phase 6: Brownfield Onboarding

> Centinela excels on greenfield projects, where it interviews the user and
> generates a roadmap from a blank slate. On an existing codebase the truth
> already lives in the code and must be *reverse-engineered*, not interviewed.
> This phase makes Centinela adoptable on a mature repo without a big-bang rewrite.

- **deep-codebase-analysis** — `centinela analyze`: scan the existing repo and
  produce a machine-readable inventory — language(s), framework, build/test setup,
  i18n locales, module/package layout, and the current import/dependency graph.
  The foundation every other brownfield feature reads from.
- **archetype-inference-project-synthesis** — From the analysis, infer the best-fit
  architecture archetype and draft a complete `PROJECT.md` (layer mapping, folder
  structure, naming conventions, gatekeeper paths) reflecting the code *as it is*,
  for the user to confirm or correct — instead of interviewing from scratch.
  *Fixes: forcing a brownfield user to hand-author PROJECT.md and guess an archetype that doesn't match reality.* (depends on deep-codebase-analysis)
- **spec-reconstruction** — Have the LLM read existing modules, endpoints, and
  flows and generate `specs/*.feature` Gherkin scenarios plus `docs/features/*.md`
  briefs that document the behavior the system *already* exhibits, as confirmable,
  editable acceptance criteria.
  *Fixes: an existing system with zero specs, so Centinela's spec-first gates have nothing to anchor to.* (depends on deep-codebase-analysis)
- **brownfield-roadmap-generation** — Generate a roadmap that distinguishes
  already-built capability (recorded as baseline / done) from the net-new work and
  gaps (TODOs, incomplete areas, user-stated goals), so existing functionality
  isn't re-planned and the team can immediately `centinela start` the next real gap.
  *Fixes: a greenfield roadmap generator that assumes everything is still ahead of you.* (depends on archetype-inference-project-synthesis, spec-reconstruction)
- **adoption-baseline** — On brownfield init, record current gate violations
  (pre-existing G1 file-size, G2 layer, etc.) as an accepted baseline so `validate`
  isn't drowned by thousands of legacy findings; only new work is gated strictly
  while legacy debt is ratcheted down over time.
  *Fixes: a mature repo where day-one `validate` reports thousands of pre-existing violations, making the gates unusable on adoption.* (depends on deep-codebase-analysis, audit-baseline-ratchet)

## Phase 7: Workflow Flexibility & Delivery

> The 5-step is feature-shaped. Real teams also do fixes, refactors, and spikes —
> and need help getting completed work delivered.

- **completion-delivery-prompt** — When a feature reaches completion (final step
  done), Centinela asks how to deliver it and acts on the choice: **(a)** commit,
  push, and open a PR via `gh` when a git `origin` remote is configured, or
  **(b)** merge the branch into `main` locally using the merge-steward
  (`centinela merge <feature>`, with `--continue` recovery on conflicts). Detects
  remote presence to offer only valid options; never pushes or merges without explicit confirmation.
  *Fixes: completed work stalls in its worktree because the agent doesn't know the team's delivery convention, or merges/pushes without asking.*
- **delivery-artifact-generation** — Compose the PR description and a
  `CHANGELOG` / release-notes entry automatically from evidence Centinela already
  holds — the feature brief, plan, gatekeeper report, and verify results — so
  delivery output is consistent and traceable instead of hand-written each time.
  *Fixes: rich step evidence exists but the PR body and changelog are still written from scratch (or skipped).* (depends on completion-delivery-prompt)
- **workflow-archetypes** — First-class lightweight tracks beside the 5-step:
  `hotfix` (reproduce → fix → test → ship), `refactor` (characterize →
  change → prove-equivalent), `spike` (timeboxed, no ship gate).
  *Fixes: forcing a diagnosis/bugfix through a plan→docs pipeline it doesn't fit.*
- **right-size-docs-step** — Make the `docs` step surface-aware, mirroring how the
  `code` step requires `ux-ui-specialist` only for `surface: user-facing` features.
  A user-facing feature still writes the plain-language KB guide (the one genuinely
  valuable, reader- and memory-useful doc artifact); an internal refactor / bugfix /
  chore instead emits a one-line changelog entry (via `delivery-artifact-generation`)
  and skips the KB guide, the per-feature HTML portal regeneration, and the
  documentation-specialist evidence ceremony. The 108 KB `index.html` portal moves
  to merge/release-time regeneration rather than per-feature.
  *Fixes: mandatory full docs ceremony — KB guide, HTML regen, evidence — runs on every feature including internal ones with no end-user story, burning tokens for zero reader value.* Aggregate view across worktrees and contributors: who owns
  what feature/step, aggregate gate health, roadmap burn-down. (depends on governance-telemetry)
- **cost-governance** — Per-feature/per-step token and model-tier budgets surfaced
  from the host harness, with a soft gate.
  *Fixes: a runaway agent burning budget on one step with no visibility.*

## Phase 8: Ecosystem

- **host-harness-adapters** — Extend the hook/parity pattern (today Claude +
  OpenCode) to Cursor, Aider, Windsurf, and Copilot Workspace. Best done after a
  shared integration core, so adding a third target isn't a maintenance tax.
- **codex-support** — First-class support for OpenAI **Codex** as a host harness
  alongside Claude Code and OpenCode: wire Codex's config/integration surface to
  Centinela's prewrite enforcement, postwrite status tags, and prompt-context
  injection, generate the Codex-native rules file (e.g. `AGENTS.md`), and extend
  the parity tests so the three harnesses stay in lockstep through `init` and
  `migrate`. Codex has a large and growing user base, so dual-support with Claude
  Code materially widens who can adopt Centinela.
  *Fixes: Codex users locked out of Centinela's governance because enforcement is wired only for Claude Code / OpenCode.* (depends on host-harness-adapters)
- **cross-project-memory** — Promote governed-memory lessons from per-repo to an
  org-shared ledger, so a failure fixed in one repo warns every repo. (depends on governed-project-memory)

## Phase 9: Self-Improvement

> Centinela applies its own founding principle to itself: a recurring friction
> signal is an engineering problem to fix permanently — by generating automation.
> Comes last: it depends on mature instrumentation from Phase 4.

- **adaptive-skill-synthesis** — When telemetry shows a workflow isn't running as
  expected — e.g. the same feature/step requires too many manual validation
  rounds, repeated rework, or the user keeps hand-correcting the same thing —
  Centinela proposes and scaffolds a **skill** (a Claude Code / OpenCode skill, or
  a reusable validate command / sub-agent prompt) that automates the repetitive
  step further. Generated skills are reviewable artifacts the user approves before
  they take effect — never silently self-modifying.
  *Fixes: the human becomes the slow loop, re-validating the same step over and over, with no path for the harness to absorb that toil.* (depends on governance-telemetry, centinela-insights)
