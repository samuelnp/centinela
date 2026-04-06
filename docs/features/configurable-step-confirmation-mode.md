# Feature Brief: Configurable Step Confirmation Mode

## Problem

Centinela currently always asks for confirmation before step advancement. Teams need
different operating modes: strict review at every step, review only after planning,
or uninterrupted automation.

## Goal

Add a configurable workflow setting that controls when Centinela asks for manual
confirmation prompts during step progression.

## Scope

- Add `workflow.step_confirmation_mode` in `centinela.toml`.
- Support `every_step`, `after_plan`, and `auto` modes.
- Keep existing behavior as default (`every_step`).
- Update hook context review prompting to respect selected mode.
- Update docs and scaffold templates to describe mode behavior.
