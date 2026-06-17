---
feature: workflow-archetypes
summary: Named lightweight workflow tracks let you pick a right-sized path for work that isn't a full feature — a hotfix, a refactor, or a throwaway spike — instead of always running the full plan-to-docs workflow, while verification still applies wherever the chosen track includes the ship step.
audience: end-user
status: done
---

## What it does

Workflow archetypes let you pick a lightweight track for a piece of work that isn't a full net-new feature. Instead of forcing an urgent bugfix, an internal refactor, or a quick experiment through the full five-step plan-to-docs workflow, you choose a track that runs only the steps that work actually needs. The tracks are not a separate workflow engine — they reuse the same canonical steps, just a smaller subset in a sensible order. Because the steps are the same, verification still applies wherever the track includes the ship step, so a lighter track never means a weaker check on the steps it keeps.

## When you'd use it

Reach for an archetype when the shape of the work doesn't match a full feature. An urgent production bug fits the hotfix track: you fix it, prove it with tests, and ship it through the gate — no upfront design document. A restructuring that changes how code is organized without changing behavior fits the refactor track: you plan the change, make it, prove behavior is unchanged, and ship it — no user-facing documentation, since nothing user-visible changed. A timeboxed experiment you might throw away fits the spike track: a light plan and then code, with no ship gate, so you can explore freely and either discard the work or promote it later. When the work really is a full feature, you simply use the default and nothing changes.

## How it behaves

- Four tracks are available. Canonical (the default) runs plan, code, tests, validate, docs. Hotfix runs code, tests, validate. Refactor runs plan, code, tests, validate. Spike runs plan, code.
- Canonical is the default, so nothing changes unless you deliberately choose a track — existing work behaves exactly as it does today.
- You choose a track when you start work, either with `centinela start --archetype <name>` or by setting the track on the feature's roadmap entry; if you give both, the flag wins.
- The active track is shown in `centinela status`, so you can always see which path a piece of work is on; a spike is marked as having no ship gate.
- Any track that includes the validate step is ship-gated exactly like a normal feature — the gates and claim verification run before the work can advance, with no exception made for the track.
- A spike has no validate step, so it is never ship-gated. This is not a verification hole: the gate is attached to the validate step, not to a track name, so a spike simply never reaches it, and there is no way to relabel work to dodge a check. If you later promote spike work, it is validated when it merges.
- An unknown track name is rejected with an error that names the offending value, so a typo can't quietly put work on an unexpected track.
- A track is independent of the strictness profile: a track sets which steps run, a profile sets how strictly each step is enforced, so any track can run under any profile.

## Examples

Start an urgent fix on the hotfix track:

```bash
centinela start fix-login --archetype hotfix
```

Start a throwaway experiment on the spike track:

```bash
centinela start probe-idea --archetype spike
```

`centinela status` shows the active track for a feature — for a spike it also notes that there is no ship gate — so you can always see which path the work is on.
