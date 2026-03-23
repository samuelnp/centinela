# Feature Brief: Harden OpenCode Plugin Compatibility

## Problem
The generated OpenCode plugin currently assumes one event/payload shape. Future OpenCode changes could break hook enforcement silently.

## Goal
Make plugin behavior more defensive and version-tolerant while preserving current functionality.

## Acceptance Criteria
- Plugin safely handles missing/alternative tool payload fields.
- Prompt append logic remains safe when output structure differs.
- Tests cover fallback paths and defensive branches.
