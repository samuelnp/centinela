---
feature: roadmap-checkpoint-prompt
summary: Once your roadmap is fully defined, Centinela asks you once whether to keep refining the roadmap or start building the first feature, instead of leaving Claude to guess.
audience: end-user
status: done
---

## What it does
After project bootstrap has produced every roadmap-definition artifact — `PROJECT.md`, `ROADMAP.md`, a valid `roadmap.json`, and the senior-PM analysis and quality reports — Centinela used to go quiet, and Claude would silently guess whether to keep working on the roadmap or jump into the first feature. Now Centinela emits one clear checkpoint and shows you a panel with two choices: keep iterating on the roadmap, or start the first incomplete Phase 0 bootstrap feature it names for you. You make the call. Once you pick "keep iterating," the prompt stays out of your way until you actually change a roadmap file again; once you start the first feature, it goes quiet on its own.

## When you'd use it
You'll see this the moment your roadmap is finished and you submit your next message. It's the explicit hand-off between "defining the plan" and "building the first thing" — the point where you decide whether the roadmap is good enough to act on, or whether you want to refine it further. There's no time pressure: you can keep iterating as long as you like, and the prompt only comes back if a roadmap file changes after you've told Centinela you're still iterating.

## How it behaves
- When the roadmap is complete but you haven't made a choice yet, Centinela shows the checkpoint directive and a panel naming the first incomplete Phase 0 feature to start.
- After you choose to keep iterating, the prompt stays silent on later messages as long as none of your roadmap files have changed since you made that choice.
- If you edit `ROADMAP.md` after choosing to keep iterating, the checkpoint comes back on your next message — your earlier choice is treated as out of date.
- The same re-prompt happens if you edit any other roadmap-definition file (such as the roadmap analysis report), not just `ROADMAP.md`.
- Once every Phase 0 bootstrap feature is finished, the checkpoint never appears again — there's nothing left to start.
- If your roadmap defines no Phase 0 bootstrap features at all, the checkpoint stays silent, because there's no first feature to point you to.
- The moment you start the first feature (its workflow file exists), the checkpoint goes quiet on its own — no extra step needed to dismiss it.
- If your roadmap isn't finished — for example `ROADMAP.md` is still missing — Centinela's existing setup guidance fires first and the checkpoint waits its turn.
- The same is true when `roadmap.json` is broken: you get the directive about fixing the roadmap file, not the checkpoint.
- When your first Phase 0 feature is already finished and a later one isn't, the panel names that next unfinished feature instead.
- If the file that remembers your "keep iterating" choice gets corrupted, Centinela doesn't crash — it just shows the checkpoint again so a damaged file can never silently hide the prompt forever.
- The same safe behavior applies if that file has a garbled timestamp: Centinela treats it as out of date and shows the checkpoint again.

## Examples
When the roadmap is complete and you haven't chosen yet, your next message surfaces the checkpoint, naming the first feature to build:

    CENTINELA DIRECTIVE: roadmap checkpoint
    ┌─ Roadmap definition iteration complete ──────────────────────────┐
    │ Continue iterating on the roadmap, or start the first incomplete │
    │ Phase 0 bootstrap feature "phase-0-feature-a"?                    │
    └──────────────────────────────────────────────────────────────────┘

To keep refining the roadmap, persist that choice — the checkpoint then stays quiet until you change a roadmap file again:

    centinela roadmap iterate

To begin building instead, start the named feature — the checkpoint goes silent once its workflow file exists:

    centinela start phase-0-feature-a
