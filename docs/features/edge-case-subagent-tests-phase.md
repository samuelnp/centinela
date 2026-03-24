# Feature Brief: Edge-Case Subagent for Tests Phase

## Problem
The current tests phase can pass with mostly happy-path tests.
This misses edge cases and hard paths that commonly fail in production.

## Goal
Make edge-case discovery and risk-based test planning mandatory in the tests phase.

## Users
- Product teams shipping features with AI-assisted implementation.
- Maintainers enforcing production-readiness quality bars.

## Acceptance Criteria
- Tests phase requires an edge-case report artifact at `.workflow/<feature>-edge-cases.md`.
- A dedicated edge-case subagent prompt exists in architecture docs.
- Context hook reminds the agent during tests step when edge-case report is missing.
- Docs and scaffold assets document the requirement.

## Non-Goals
- Replacing gatekeeper or production-readiness reports.
- Changing the workflow step order.
