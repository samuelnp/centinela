# Feature: coverage-hardening

## Problem

Centinela's `scripts/check-coverage.sh` gate enforces a total statement
coverage floor of **95.0%**. The last three features each landed *exactly*
on ~95.0%, leaving effectively zero headroom above the gate. Because
`centinela validate` is **not** a required CI check, a red coverage result
on `main` does not block a merge — so the moment two PRs land in parallel,
the second merge can tip `main`'s coverage below 95% and it auto-merges
red, unnoticed.

This week a near-miss confirmed the fragility: a parallel merge dropped
total coverage to the floor with no margin to absorb it.

## Who is hurting

- **Every future PR author**: builds on a tree that is one statement away
  from a red gate, so unrelated changes get blamed for coverage drops.
- **`main` itself**: silently drifts red because validate is advisory, not
  required — the gate that should protect the trunk cannot.
- **Reviewers**: cannot trust a green local run to stay green after a
  sibling PR merges.

## Why now

Three consecutive features at the 95.0% floor plus a real near-miss this
week. The user's standing policy is explicit: **do not sit on a gate
threshold — exceed it by ~2%.** We raise total coverage from 95.0% to
**>= 97%** with real tests, buying a durable safety margin so parallel
merges no longer tip the trunk red.

## Outcome

Total statement coverage measured by `scripts/check-coverage.sh` is
**>= 97%** (target ~97%+ with a small safety margin), achieved entirely
with real, passing tests that exercise real logic — never by lowering the
gate threshold or gaming the measurement.
