# Feature Brief: G1 Justified File Size Exceptions

## Problem

G1 currently hard-fails any source file over 100 lines. Some files are valid
exceptions (configuration-heavy or domain-atomic files) where splitting harms
cohesion and readability.

## Goal

Keep 100 lines as the default rule, but allow explicit, justified exceptions for
rare cases with strict constraints and auditability.

## Scope

- Add `gates.file_size_exceptions` config entries.
- Allow exceptions only for `configuration` or `domain_atomic` kinds.
- Enforce max exception cap at 130 lines.
- Preserve failure behavior for non-justified oversize files.
- Report justified exceptions in gate output details.
