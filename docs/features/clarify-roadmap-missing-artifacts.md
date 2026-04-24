# Feature: Clarify Missing Roadmap Artifacts

## Problem

In greenfield setup, agents can write `ROADMAP.md` but miss `.workflow/roadmap.json`.
Centinela then fails with a generic roadmap error that does not tell the agent which
 artifact is missing or what file shapes it must create next, which can trap setup in
 a loop.

## Outcome

Make roadmap setup failures explicit and document the expected setup and per-feature
 artifact templates so agents can recover deterministically.

## Scope

- Surface `.workflow/roadmap.json` directly in start and roadmap command failures.
- Add a setup-hook branch and UI guidance for missing or invalid roadmap JSON.
- Document setup and per-feature workflow artifact templates in scaffolded docs.
- Update setup guidance so new projects validate roadmap artifacts before starting.
