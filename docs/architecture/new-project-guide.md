# New Project Setup Guide

This guide explains how to bring Centinela into a new project from scratch.

## Step 1: Install centinela

```bash
go install github.com/samuelnp/centinela@latest
centinela --help
```

## Step 2: Initialize the project

Run once in the project root:

```bash
centinela init
```

Use `--local` if you prefer `.claude/settings.local.json`:

```bash
centinela init --local
```

Choose integration target explicitly when needed:

```bash
centinela init --agent claude
centinela init --agent opencode
centinela init --agent both
```

## Step 3: Fill in PROJECT.md

Rename `PROJECT.md.template` to `PROJECT.md` and fill every section:

1. Project concept and problem.
2. Architecture archetype.
3. Tech stack and testing tools.
4. Domain entities and integrations.
5. Folder structure and locales.

Do not leave `<!-- comment -->` placeholders.

## Step 4: Configure centinela.toml

Add your stack validation commands:

```toml
[validate]
commands = [
  "npx tsc --noEmit",
  "npx vitest run",
  "npx cucumber-js"
]
```

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
