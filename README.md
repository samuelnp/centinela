# Centinela

Development workflow enforcer for Claude Code projects.

Centinela keeps Claude inside a strict 4-step workflow — plan → code → tests → validate — and blocks writes that happen in the wrong step. It ships as a single Go binary with no runtime dependencies.

---

## Install

```bash
go install github.com/samuelnp/centinela@latest
```

Or download a pre-built binary from [Releases](https://github.com/samuelnp/centinela/releases) and move it to `/usr/local/bin/`.

---

## Bootstrap a project

Run once in any new project directory:

```bash
centinela init
```

Creates:
- `CLAUDE.md` — framework rules Claude must follow
- `PROJECT.md.template` — fill this in to define your project
- `docs/architecture/` — all architecture reference docs
- `specs/` `docs/plans/` `tests/` — required directory scaffolding
- `.claude/settings.json` — hooks wired automatically

Safe to re-run — existing files are never overwritten.

---

## Workflow commands

```bash
centinela start <feature>     # begin a new feature (creates .workflow/<feature>.json)
centinela status <feature>    # interactive status view
centinela status-all          # all active workflows
centinela complete <feature>  # validate current step and advance to next
```

### The 4 steps

| Step | What Claude may write | Required before advancing |
|------|-----------------------|--------------------------|
| `plan` | `docs/plans/` and `specs/` only | Plan file + `.feature` spec |
| `code` | `src/` `app/` and plan files | — |
| `tests` | `tests/` and source | Unit + acceptance test files |
| `validate` | anything (small fixes) | Gatekeeper report + `scripts/validate.sh` passes |

---

## Hooks

Three Claude Code hooks enforce the workflow automatically:

| Event | Hook | Effect |
|-------|------|--------|
| `PreToolUse` (Write/Edit) | `centinela hook prewrite` | Blocks writes in the wrong step |
| `PostToolUse` (Write/Edit) | `centinela hook postwrite` | Injects workflow tag after every write |
| `UserPromptSubmit` | `centinela hook context` | Shows active workflow on every prompt |

`centinela init` wires these into `.claude/settings.json` automatically.

---

## Build from source

```bash
git clone https://github.com/samuelnp/centinela
cd centinela
go build -o centinela ./cmd/centinela/
```

Cross-compile:

```bash
GOOS=linux  GOARCH=amd64 go build -o centinela-linux  ./cmd/centinela/
GOOS=darwin GOARCH=arm64 go build -o centinela-darwin ./cmd/centinela/
```
