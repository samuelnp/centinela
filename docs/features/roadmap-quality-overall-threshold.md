# Feature Brief: Roadmap Quality Overall Threshold

## Problem

Roadmap analysis checks dependencies, but it does not score feature quality. Weakly
defined features can still pass and later cause planning churn.

## Goal

Require a roadmap quality evaluator artifact that scores each roadmap feature and
blocks progress until every feature has `overall >= 9`.

## Scope

- Add quality analysis artifacts (`.md` + `.json`) for roadmap setup.
- Validate quality coverage against `.workflow/roadmap.json`.
- Enforce quality readiness in greenfield `centinela start` guard.
- Extend `centinela roadmap validate` to include quality checks.
