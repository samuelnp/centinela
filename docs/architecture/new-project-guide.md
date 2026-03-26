<!-- centinela:doc-version=1 template=docs/architecture/new-project-guide.md -->
# New Project Setup Guide

This guide explains how to bring Centinela into a new project from scratch.

---

## Step 1: Install centinela

```bash
go install github.com/samuelnp/centinela@latest
```

Verify:

```bash
centinela --help
```

---

## Step 2: Initialize the project

Run once in the project root:

```bash
centinela init
```

This creates required files and wires Claude/OpenCode integrations automatically.

Use `--local` if you prefer not to commit the hooks to a shared settings file:

```bash
centinela init --local
```

---

## Step 3: Fill in PROJECT.md

Rename `PROJECT.md.template` to `PROJECT.md` and fill in every section:

1. **Project name and elevator pitch** — what the project does
2. **Architecture Choice** — pick one archetype (Hexagonal, Rails-native, N-Tier, ECS, Modular)
3. **Tech Stack** — language, framework, test runner, i18n approach
4. **Domain** — core entities and external integrations
5. **Folder Structure** — where source, tests, and specs live
6. **Locales** — list all locales if the project uses i18n (or leave empty)

Do not leave any `<!-- comment -->` placeholders in place — the AI agent reads this file and will be confused by unfilled sections.

---

## Step 4: Configure centinela.toml

Open `centinela.toml` and add the commands that validate your stack:

```toml
[validate]
commands = [
  "npx tsc --noEmit",   # type check
  "npx vitest run",     # unit + integration tests
  "npx cucumber-js",    # acceptance tests
]
```

If your project uses i18n, enable the built-in key parity check:

```toml
[gates]
i18n = true

[i18n]
format  = "json"
dir     = "src/i18n/messages"
locales = ["en", "es"]
```

---

## Step 5: Verify the setup

```bash
# PROJECT.md exists and is complete
cat PROJECT.md

# Start a test workflow to confirm hooks are working
centinela start test-feature
centinela status test-feature

# Clean up the test workflow file
rm .workflow/test-feature.json
```

---

## Step 6: Start your first real feature

```bash
centinela start <feature-name>
```

This sets the workflow to the `plan` step. The hooks will block any file writes outside `docs/plans/` and `specs/` until you advance.

---

## What Each File Does

| File | Purpose |
|------|---------|
| `CLAUDE.md` | Framework rules: workflow, architecture, naming conventions |
| `PROJECT.md` | Project definition: domain, stack, folder structure |
| `PROJECT.md.template` | Blank template — copy and fill in manually |
| `centinela.toml` | Validate commands + built-in gate configuration |
| `docs/architecture/` | Architecture reference documentation |

---

## Verification Checklist

Before starting your first feature, confirm:

- [ ] `PROJECT.md` exists and has no `<!-- comment -->` placeholders
- [ ] `centinela.toml` has at least one validate command (or a note explaining why none are needed)
- [ ] `centinela start test-feature` runs without error
- [ ] The folder paths in `PROJECT.md → Folder Structure` match the actual layout
- [ ] All locales in `PROJECT.md → Locales` have their locale files at the expected paths

## Preserved Custom Sections

## Step 5: Verify setup

```bash
cat PROJECT.md
centinela status-all
centinela start test-feature
centinela status test-feature
rm .workflow/test-feature.json
```


## Step 6: Start real work

```bash
centinela start <feature-name>
```

Then follow:

```
plan -> code -> tests -> validate
```


## Checklist

- `PROJECT.md` has no placeholders.
- `centinela.toml` contains validate commands for your stack.
- `centinela start test-feature` works.
- Paths in `PROJECT.md` match actual project folders.

