# Feature Brief: Docs Consistency Pass

## Problem
Documentation still contains stale references to legacy script commands and Claude-only wording in places where Centinela now supports both Claude and OpenCode.

## Users
- New teams onboarding Centinela.
- Existing teams migrating from Claude-only setups to mixed agent usage.

## Goals
- Standardize command references to current CLI commands.
- Align wording across root docs and scaffolded docs.
- Keep architecture and workflow guidance accurate and consistent.

## Acceptance Criteria
- No remaining references to `scripts/centinela-workflow.sh` in active docs.
- New-project guide and workflow docs reference `centinela` commands.
- Agent wording is accurate where integration is now Claude/OpenCode.

## Risks
- Missing scattered stale references in scaffold assets.
- Over-updating historical/example sections that are intentionally illustrative.

## Decomposition
- Scan docs and scaffold assets for legacy references.
- Update references to `centinela` commands.
- Run tests and validate command docs remain coherent.
