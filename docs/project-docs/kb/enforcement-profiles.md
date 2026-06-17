---
feature: enforcement-profiles
summary: A named strictness preset that lets you dial how much process Centinela enforces for a model or team, while every gate and all claim verification keep running unchanged no matter which preset you pick.
audience: end-user
status: done
---

## What it does

Enforcement profiles let you choose how much process Centinela puts in front of your work. You pick one of three presets — strict, guided, or outcome — and that preset decides how heavily the workflow is scaffolded: whether writes are gated to the current step, whether Centinela stops and asks before each step, and whether the full set of subagent evidence files is mandatory. What a profile never changes is the outcome side: every validate gate and all claim verification still run, exactly the same, under every preset. A profile relaxes process, never proof.

## When you'd use it

Match the profile to the model or team doing the work. A small or local model that needs maximum rails to stay on track is best served by strict — full step-gating, a confirmation before every step, and the complete subagent evidence trail. A capable model that can drive the steps itself fits guided — the rails stay on, but the heavy evidence ceremony is dropped and it only pauses for review after planning. A strong model working fast and out of order fits outcome — it can write files in any order with no inter-step prompts, and is judged purely on whether the final gates pass. You set a default profile once per project, and you can override it for a single feature when that one piece of work needs more or fewer rails than the rest.

## How it behaves

- Three presets are available: strict, guided, and outcome, each setting a different amount of enforced process.
- Strict is the default. A project that configures nothing gets exactly today's behavior — step-gated writes, a confirmation before every step, and required subagent evidence — so upgrading changes nothing.
- Under the outcome profile you can write files in any order; the workflow no longer blocks a write just because it belongs to a later step.
- The outcome profile also suppresses the between-step "shall I advance?" review prompt, so work flows without stopping to confirm each step.
- If you have explicitly set a confirmation mode in your configuration, that explicit setting always wins over whatever the profile would have chosen — your direct choice is never silently overridden by a preset.
- The strict profile requires the full subagent evidence trail to be present; the guided and outcome profiles do not require it.
- You choose the profile per project in `centinela.toml`, or per feature when you start it with `centinela start --profile <name>`; the per-feature choice takes precedence over the project default.
- An unrecognized profile name is rejected when the configuration loads, with an error that names the offending setting, so a typo can't quietly leave you on an unexpected profile.
- The key guarantee: every gate and all claim verification run identically under strict, guided, and outcome. Completion is still blocked by a failing gate or a failed claim check no matter which profile is active — a profile only ever relaxes process, never verification.

## Examples

Set a project-wide default in `centinela.toml`:

```toml
[workflow]
enforcement_profile = "guided"
```

Override the profile for a single feature when you start it:

```bash
centinela start my-feature --profile outcome
```

`centinela status` shows the active profile for a feature, so you can always see which rails are currently in effect.
