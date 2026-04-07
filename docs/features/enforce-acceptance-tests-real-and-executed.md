# Feature Brief: Enforce Real and Executed Acceptance Tests

## Problem

The tests step currently accepts any non-hidden file in `tests/acceptance`, which
allows placeholder artifacts without executable test logic. Also, acceptance
execution during `validate` depends only on user convention.

## Goal

Require acceptance artifacts to contain executable test code and enforce that
acceptance tests are executed as part of validation.

## Scope

- Reject acceptance files that are comment-only, whitespace-only, or placeholder/no-op.
- Keep allowing explanatory comments alongside real executable assertions/steps.
- Require at least one validate command that executes acceptance tests.
- Add unit and acceptance coverage for new enforcement paths.
