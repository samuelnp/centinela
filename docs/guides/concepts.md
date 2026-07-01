# Concepts: Centinela & Harness Engineering

> Why Centinela exists and where it fits relative to your AI coding agent.

## Harness engineering

"Harness engineering" is the discipline of building the infrastructure around an
LLM that turns it into a reliable agent — the verification loops, guardrails,
context management, and environment control. Its guiding principle:

> Treat every agent failure as an engineering problem to fix permanently, not a
> prompt to retry. Make correctness **enforced**, not **requested**.

Centinela is **not an agent harness** — Claude Code and OpenCode are. Centinela
is the *governance layer* that sits on top of them and enforces how the harness
is used across a team. It owns the parts of harness engineering that decide
whether shipped code is trustworthy, and stays out of the parts the host agent
already does well:

| Harness subsystem            | Owned by Centinela | How                                                                 |
|------------------------------|:------------------:|---------------------------------------------------------------------|
| Verification & guardrails    |        ★★★         | PreToolUse blocks out-of-step writes; validate gates (file size, i18n, your test suite); gatekeeper + production-readiness subagents |
| Context engineering          |        ★★          | UserPromptSubmit injects the active feature, step, and required evidence; the plan advisor reads roadmap deps and prior edge-case lessons |
| Environment control          |        ★★          | `centinela init` wires hooks and scaffolds the rules; `migrate` updates them incrementally to prevent known failure modes |
| Tool integration layer       |         —          | delegated to Claude Code / OpenCode                                 |
| Memory & state management    |         ★          | `.workflow/*.json` tracks per-feature step state                    |
| The agent loop itself        |         —          | delegated to the host harness                                       |

The three principles of harness engineering map directly onto what Centinela
already does:

- **Environment control** → CLAUDE.md hard-rules, scaffolded docs, and `migrate`
  let you encode rules that prevent known failure modes — and keep them current.
- **Mechanical verification** → required artifacts and gates make correctness
  *checkable*: no plan file means no code, no tests means no validate.
- **Graceful recovery** → the merge-steward, missing-artifact recovery, and the
  plan advisor are designed for non-deterministic agent behavior.

In short: bring your own harness; Centinela makes sure it's used with discipline.

## When *not* to use Centinela

Centinela trades flexibility for discipline. Skip it if any of these apply:

- **Throwaway scripts / one-off experiments.** The 5-step ceremony is overhead you'll regret.
- **Solo prototyping in the first 48 hours of an idea.** Plans, specs, and gate suites are useful *after* you've validated the idea — not while you're still figuring out what to build.
- **You don't use an AI coding agent.** Centinela's strongest leverage is forcing structure on agent-generated code; humans typing every keystroke already have plenty of friction.
- **Your team has a different workflow you actually follow.** Centinela is opinionated. If your team already ships clean specs, tests, and docs without enforcement, the hooks will feel like a tax.

Centinela is for *production code* you intend to maintain, where an AI agent is doing meaningful work and you want the agent's output to look like it came from a disciplined human team.

---

← Back to the [documentation index](README.md) · [Getting started](getting-started.md)
