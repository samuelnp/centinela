# Feature Brief: Roadmap Senior PM Analysis

## Problem

Roadmaps can miss cross-feature dependencies and product-flow sequencing, which
causes invalid implementation order and poor delivery planning.

## Goal

Require a senior-product-manager roadmap analysis artifact that validates feature
dependencies and sequencing before feature workflows can start.

## Scope

- Add roadmap analysis artifacts (`.md` + `.json`) for setup/roadmap phase.
- Validate analysis content against `.workflow/roadmap.json`.
- Enforce analysis readiness in greenfield `centinela start` guard.
- Add `centinela roadmap validate` command for explicit checks.
