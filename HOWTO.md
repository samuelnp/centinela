# HOWTO: Build a Landing Page MVP with Centinela

This guide shows how to collaborate with Claude Code or OpenCode while Centinela enforces the full workflow. The example is a small landing page MVP with a hero, benefit cards, social proof, and a signup call to action.

## 1. Install and Initialize

Run this once in the project root:

```bash
centinela init --agent both
```

If `PROJECT.md` is missing, open your coding agent and let Centinela guide the setup interview. The agent should create `PROJECT.md`, then define roadmap artifacts before feature work starts.

Use this prompt:

```text
Use Centinela to set up this project. Interview me for any missing PROJECT.md details, then create the roadmap artifacts Centinela requires before implementation.
```

Verify setup before starting the MVP:

```bash
centinela roadmap validate
```

## 2. Start the Feature

Start a named feature:

```bash
centinela start landing-page-mvp
```

You can also ask the agent:

```text
Start a Centinela feature named landing-page-mvp for a small marketing landing page MVP.
```

## 3. Plan Step

Ask for the smallest useful plan first:

```text
Plan a landing page MVP with a hero, three benefits, social proof, and a signup call to action. Follow Centinela's plan step: create the feature brief, docs/plans artifact, specs Gherkin file, and required specialist evidence. Ask only missing high-value questions before writing artifacts.
```

Expected outputs include:

- `docs/features/landing-page-mvp.md`
- `docs/plans/landing-page-mvp.md`
- `specs/landing-page-mvp.feature`
- `.workflow/landing-page-mvp-big-thinker.*`
- `.workflow/landing-page-mvp-feature-specialist.*`

When the agent asks `Step plan complete — shall I advance to code?`, review the plan and approve only if it matches the MVP scope.

## 4. Code Step

After approval, Centinela advances with:

```bash
centinela complete landing-page-mvp
```

Then ask the agent:

```text
Implement the landing page MVP from the approved plan. Keep UI components thin, keep business logic outside the outer layer, use existing design patterns, and include required senior-engineer evidence. If this is user-facing, include ux-ui-specialist evidence with mobileFirst true and real UI outputs.
```

Do not ask the agent to add tests during `code` unless the workflow has advanced to `tests`. If Centinela blocks a write, ask the agent to explain which step or artifact is missing.

## 5. Tests Step

Advance after code review, then ask:

```text
Add Centinela-required tests for landing-page-mvp: unit tests, integration tests, executable acceptance tests, and the edge-case report. Ensure acceptance tests are real assertions or executable steps, not placeholders.
```

Expected outputs include:

- files in `tests/unit/` or the project unit-test location
- files in `tests/integration/`
- executable files in `tests/acceptance/`
- `.workflow/landing-page-mvp-edge-cases.md`
- `.workflow/landing-page-mvp-qa-senior.*`

Make sure `centinela.toml` runs acceptance tests during validation:

```toml
[validate]
commands = [
  "npm test",
  "npm run test:acceptance"
]
```

Use your project's real commands, such as `go test ./...`, `pytest`, `npx vitest run`, or `bundle exec rspec`.

## 6. Validate Step

Advance after tests are in place, then run:

```bash
centinela validate
```

The agent should also produce the required gatekeeper report:

```text
Run the Centinela validate step for landing-page-mvp. Create the required gatekeeper report, fix any failures, and rerun validation until it passes.
```

Expected output:

- `.workflow/landing-page-mvp-gatekeeper.md`
- passing `centinela validate`
- passing lint, type checks, unit tests, integration tests, and acceptance tests from `[validate].commands`

If G1 file-size checks fail, split files by responsibility. Use `gates.file_size_exceptions` only for rare configuration-heavy or domain-atomic files, and keep exceptions at or below 130 lines.

## 7. Docs Step

Advance after validation passes, then ask:

```text
Complete the Centinela docs step for landing-page-mvp. Create documentation-specialist evidence and regenerate project docs.
```

Run:

```bash
centinela docs validate
centinela docs generate --out docs/project-docs/index.html --title "Project Documentation"
```

Expected outputs include:

- `.workflow/landing-page-mvp-documentation-specialist.md`
- `.workflow/landing-page-mvp-documentation-specialist.json`
- `docs/project-docs/index.html`

Finish the workflow only after docs validation and generated docs succeed.

## Recovery Tips

- If setup stalls, run `centinela roadmap validate` and create the missing roadmap artifact it names.
- If a file write is blocked, check `centinela status landing-page-mvp` and work in the current step only.
- If the agent tries to skip tests or docs, remind it that Centinela has no skip command.
- If managed docs or integration files are stale, preview changes with `centinela migrate` and apply them only after review with `centinela migrate --apply`.
