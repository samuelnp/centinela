# Configuration Guide

> Copy a ready-made `centinela.toml` for your situation, then tune it. Every key is documented in the [Configuration reference](configuration-reference.md).

Centinela is deliberately opinionated, but it flexes from *minimal ceremony* (solo prototyping) to *maximum assurance* (regulated code). Pick the recipe closest to you, drop it into `centinela.toml` in your project root, and adjust. Recipes are **complete and copy-pasteable** — run [`centinela doctor`](configuration-reference.md#validating-your-config) after pasting to confirm it loads.

## How config resolves (read this first)

Three things people expect to be in `centinela.toml` but aren't:

- **Which harness runs** (Claude Code / OpenCode / Codex) is chosen by `centinela init --agent` and the generated `.claude/settings.json` / `opencode.json`. `centinela.toml` decides *what is enforced*, not *which agent runs*.
- **Your architecture archetype** (Hexagonal, Rails, N-Tier, ECS, Modular) is inferred from your codebase and confirmed in `PROJECT.md`.
- **The enforcement profile** (`strict` / `guided` / `outcome`) scales how strictly the five steps are enforced. It does not swap config files — the same `centinela.toml` applies.

Jump to a recipe:

1. [Solo dev / rapid prototyping](#1-solo-dev--rapid-prototyping)
2. [Disciplined team with CI](#2-disciplined-team-with-ci)
3. [Local / offline models](#3-local--offline-models)
4. [Layered / multi-language codebase](#4-layered--multi-language-codebase)
5. [CI / fleet automation (headless)](#5-ci--fleet-automation-headless)
6. [Maximum governance / regulated](#6-maximum-governance--regulated)
7. [Custom model routing](#7-custom-model-routing)

---

## 1. Solo dev / rapid prototyping

**Who it's for:** one person, moving fast, still figuring out what to build.
**Goal:** keep the workflow's guardrails without the full ceremony slowing you down.

```toml
[workflow]
enforcement_profile    = "guided"   # relax step strictness; still validates at merge
step_confirmation_mode = "auto"     # advance steps without pausing for review
plan_advisor_mode      = "off"      # no planning questions

[gates]
file_size = false                    # don't sweat the 100-line limit while prototyping

[validate]
commands = ["go test ./..."]         # your one must-pass command
```

**What it enables:** the pipeline stays on (plan → code → tests → validate → docs) but nothing pauses to ask for approval, and the file-size gate is off. You still get a written plan, tests, and a validate step.
**Trade-offs:** less enforced structure — appropriate for exploration, not for code you'll hand to a team. Turn the gates back on once the idea is validated.

---

## 2. Disciplined team with CI

**Who it's for:** a team shipping production code through pull requests.
**Goal:** enforce structure locally and hard-fail in CI, with the same rules for humans and agents.

```toml
[workflow]
enforcement_profile    = "strict"
step_confirmation_mode = "after_plan"   # pause once (after plan), then flow

[validate]
commands  = ["npx tsc --noEmit", "npx vitest run", "npx cucumber-js"]
diff_mode = "auto"                       # diff-aware locally, full scan in CI
diff_base = "main"

[gates]
file_size = true

[gates.spec_traceability]
enabled  = true
severity = "fail"                        # every Gherkin scenario must have an acceptance test

[gates.security]
enabled = true                           # gitleaks secret scan (hard-fail) + dependency audit

[gates.build]
enabled = true
command = "go build ./cmd/myapp"
targets = [
  { goos = "linux",  goarch = "amd64" },
  { goos = "darwin", goarch = "arm64" },
]

[headless]
detect_ci = true                         # auto non-interactive when CI=true

[pr_gate]
enabled         = true
fail_on_warning = true                   # warnings block the PR gate in CI
```

**What it enables:** strict five-step order, one review pause after planning, secret + spec-traceability + cross-compile gates, diff-aware locally so you only see *your* violations, and a strict full-scan PR gate.
**Trade-offs:** more friction per feature. That's the point — it's the tax that keeps agent output looking like a disciplined team's.

---

## 3. Local / offline models

**Who it's for:** privacy-sensitive work, air-gapped environments, or avoiding API cost.
**Goal:** run the specialist subagents against a local model server.

```toml
[orchestration.local]
provider = "ollama"                       # or "openai-compatible"
endpoint = "http://localhost:11434/v1"
model    = "qwen2.5-coder"

[validate]
commands = ["go test ./..."]
```

For any other OpenAI-compatible server (llama.cpp, vLLM, LM Studio) that needs an auth header:

```toml
[orchestration.local]
provider    = "openai-compatible"
endpoint    = "http://localhost:8080/v1"
model       = "my-local-model"
api_key_env = "LOCAL_API_KEY"             # runner reads the value of this env var
```

**What it enables:** `centinela init` / `migrate` wire the OpenCode provider for you. A local model with no explicit capability class defaults to the `limited` class → `strict` profile, so you get maximum scaffolding to compensate for a smaller model.
**Trade-offs:** Centinela validates only the *shape* of the block — it never connects to your server. Availability is the runner's job. Smaller models benefit from the stricter profile, so don't loosen it. See [Getting Started → local model](getting-started.md#point-centinela-at-a-local-model).

---

## 4. Layered / multi-language codebase

**Who it's for:** a codebase with real architectural layers you want mechanically enforced.
**Goal:** fail the build when a package imports across a forbidden layer boundary.

```toml
[gates]
file_size = true

[gates.import_graph]
enabled  = true                           # G2: Layer Boundaries
# provider = "go"                         # auto-detected; also node | python | script

[[gates.import_graph.layers]]
name  = "leaf"
paths = ["internal/config/**"]
allow = []                                # a leaf imports no other layer

[[gates.import_graph.layers]]
name  = "domain"
paths = ["internal/service/**"]
allow = ["leaf"]

[[gates.import_graph.layers]]
name  = "cmd"
paths = ["cmd/**"]
allow = ["domain", "leaf"]

[validate]
commands = ["go vet ./...", "go test ./..."]
```

For a non-Go stack, set `provider = "script"` and a `script_command` that emits the import-graph JSON:

```toml
[gates.import_graph]
enabled        = true
provider       = "script"
script_command = ["./scripts/import-graph.sh"]
```

**What it enables:** a forbidden edge is reported as `internal/config -> internal/ui (leaf may not import domain)`. Packages matching no layer warn (not fail), so you can adopt the matrix incrementally.
**Trade-offs:** you have to describe your layers once. Standard-library and third-party imports are ignored. See [Quality gates → layer-boundary](gates.md#layer-boundary-import-graph-gate).

---

## 5. CI / fleet automation (headless)

**Who it's for:** running Centinela unattended across many repos or in a build fleet.
**Goal:** never block on a prompt, track token spend, and keep an audit log.

```toml
[headless]
enabled = true                            # force non-interactive (or use detect_ci = true)

[telemetry]
enabled = true                            # append-only JSONL governance event log

[cost]
enabled              = true               # soft gate — warns, never blocks complete
step_token_budget    = 150000
feature_token_budget = 800000

[gates.audit_baseline]
enabled  = true                           # ratchet: only NEW violations fail
severity = "fail"

[validate]
commands = ["go test ./..."]
```

**What it enables:** no interactive prompts, per-step/per-feature token budgets surfaced as warnings, a governance event log, and a baseline ratchet so a legacy codebase can adopt gates without a big-bang cleanup — only regressions fail.
**Trade-offs:** the cost gate is advisory by design (it never blocks `complete`); wire budgets into your own alerting if you need hard stops. You must commit the baseline snapshot (`.workflow/audit-baseline.json`).

---

## 6. Maximum governance / regulated

**Who it's for:** high-assurance or regulated code where correctness must be provable.
**Goal:** turn on every gate and tighten verification.

```toml
[workflow]
enforcement_profile    = "strict"
step_confirmation_mode = "every_step"     # review every step transition

[validate]
commands  = ["golangci-lint run", "go test ./..."]
diff_mode = "off"                         # always full scan

[gates]
file_size            = true
i18n                 = true
production_readiness = true               # G12 production-readiness subagent

[i18n]
format  = "json"
dir     = "src/i18n/messages"
locales = ["en", "es", "fr"]

[gates.security]
enabled = true

[gates.spec_traceability]
enabled  = true
severity = "fail"

[gates.roadmap_drift]
enabled  = true
severity = "fail"

[gates.audit_baseline]
enabled  = true
severity = "fail"

[verify]
verify_timeout     = 120
coverage_tolerance = 0.0005               # tighter than the 0.1% default
```

**What it enables:** every built-in gate on and hard-failing, always-full scans, i18n completeness across three locales, production-readiness review, roadmap-drift enforcement, and near-zero coverage-claim slack.
**Trade-offs:** the slowest, strictest setup. Expect every feature to take real time in `validate`. Aim to clear the coverage floor by ~2% so parallel merges don't tip `main` red.

---

## 7. Custom model routing

**Who it's for:** teams that want specific models per subagent role or per runner.
**Goal:** override the built-in tier defaults without touching harness config.

```toml
# String form: set a tier for a role across all runners.
[orchestration.models]
big-thinker    = "reasoning"
qa-senior      = "balanced"
edge-case-tester = "fast"

# Table form: pin exact models per runner for one role.
[orchestration.models.senior-engineer]
claude   = "claude-opus-4-7"
opencode = "anthropic/claude-opus-4-7"

# Remap what a tier resolves to, per runner.
[orchestration.model_map.balanced]
claude   = "claude-sonnet-4-6"
opencode = "anthropic/claude-sonnet-4-6"
```

**What it enables:** precise control over cost/quality per role — a `reasoning` model for planning and engineering, cheaper `fast` models for edge-case generation.
**Trade-offs:** roles and runners are validated against fixed sets (see the [reference](configuration-reference.md#orchestrationmodelsrole)); a typo fails config load. Leave it unset to use the sensible built-in tiers.

---

## Mixing recipes

These recipes are starting points, not modes — every block is independent. A common real-world config is **Disciplined team** + a couple of blocks from **Maximum governance** (say, `i18n` and `production_readiness`), or **Local models** + **CI fleet**. Combine freely, then run `centinela doctor` to confirm the merged file is valid.

For the exhaustive key-by-key list, see the [Configuration reference](configuration-reference.md).

---

← Back to the [documentation index](README.md) · [Getting started](getting-started.md) · [Configuration reference](configuration-reference.md)
