# Gate Keepers

Automated and manual checks that MUST pass before any code is merged.

> Tool names below are examples. Use the equivalent tools for your project's language and stack (PROJECT.md → Tech Stack). Gate rules G2 and G7 vary by architecture archetype (PROJECT.md → Architecture Choice). See [architecture-overview.md](architecture-overview.md) for the rules per archetype.

## Pre-Commit Gates (automated via lint/scripts)

### G1: File Size Limit
- **Rule**: No file exceeds 100 lines.
- **Check**: `scripts/check-file-size.sh`
- **Fail action**: Block commit. Split the file.

### G2: Layer Dependency Violations
- **Rule**: No imports cross forbidden layer boundaries. What is "forbidden" depends on the archetype in PROJECT.md → Architecture Choice → G2 rule. See [architecture-overview.md](architecture-overview.md) for the rules per archetype.
- **Check**: Static analysis tool with import boundary rules configured for the project's archetype and language (e.g., dependency-cruiser, ESLint import plugin, RuboCop custom cops, custom linter).
- **Fail action**: Block commit. Fix the import direction.

### G3: Type Safety / Static Analysis
- **Rule**: No dynamic typing shortcuts. Strictest mode enabled for the project's language.
- **Check**: Project's type checker or static analyzer (e.g., `tsc --noEmit`, `bundle exec srb tc`, `mypy`).
- **Fail action**: Block commit. Add proper types.

### G4: Test Coverage
- **Rule**: All use cases have unit tests. All adapters have integration tests.
  All `.feature` files have step definitions. All acceptance tests pass.
- **Check**: `centinela validate`
- **Fail action**: Block commit. Write missing tests.

## Pre-Feature Gates (manual checklist)

### G5: Spec First
- **Rule**: Every feature has a `.feature` file in `specs/` BEFORE implementation.
- **Check**: Manual review — does the feature file exist and cover acceptance criteria?
- **Fail action**: Do not start coding. Write the spec first.

### G5.1: Gatekeeper Conflict Review (Subagent)
- **Rule**: Before implementing, the Gatekeeper subagent reviews the new spec
  against ALL existing specs for conflicts.
- **Check**: Invoke Gatekeeper subagent (see CLAUDE.md for prompt template).
- **Output**: SAFE / WARNING / BLOCKING report.
- **Fail action**:
  - SAFE → proceed to implementation
  - WARNING → document acknowledged risks in plan, proceed with caution
  - BLOCKING → resolve conflicts before writing any code

### G6: Plan Documented
- **Rule**: Every feature has a plan in `docs/plans/`.
- **Check**: Manual review.
- **Fail action**: Do not start coding. Write the plan first.

### G7: No Business Logic in the Outer Layer
- **Rule**: The outer layer must contain no business or game logic. What the "outer layer" is depends on the archetype: UI components (Hexagonal/N-Tier), views/templates (Rails), Components (ECS), or the equivalent in your custom pattern. See PROJECT.md → Architecture Choice → G7 rule.
- **Check**: Code review — no conditionals based on business rules, no data transformations, no direct data access in the outer layer.
- **Fail action**: Move logic to the appropriate inner layer per the archetype.

### G8: Single Responsibility
- **Rule**: Each file exports one thing and does one thing.
- **Check**: Code review.
- **Fail action**: Split the file.

## Post-Feature Gates

### G9: All Tests Pass
- **Rule**: Full test suite exits with 0 (unit + integration + acceptance).
- **Check**: CI pipeline running `centinela validate`.
- **Fail action**: Fix failing tests before merge.

### G10: Acceptance Regression
- **Rule**: All existing Gherkin scenarios still pass after new feature.
- **Check**: Run acceptance tests only (subset of your `centinela validate` commands).
- **Fail action**: New feature broke existing behavior. Fix before merge.

### G11: i18n Complete
- **Rule**: No hardcoded user-facing strings. All keys present in all locale files listed in PROJECT.md → Locales.
- **Check**: `scripts/check-i18n.sh`
- **Fail action**: Add missing translations.

### G12: Production Readiness (Subagent, opt-in)

- **Enabled by**: `gates.production_readiness = true` in `centinela.toml`.
- **Check**: Invoke subagent (see CLAUDE.md → Production Readiness Subagent).
- **Output**: PASS / WARNING / BLOCKING in `.workflow/<feature>-production-readiness.md`.
- **BLOCKING** → fix CRITICAL issues, re-run subagent, then `centinela complete`
- **WARNING** → complete proceeds; centinela suggests a follow-up hardening feature
- **PASS** → complete proceeds normally

## Gate Enforcement

These gates are enforced at four levels:

1. **CLAUDE.md** — AI agents read this and follow the rules.
2. **Static analysis config** — Automated type checking and import boundary rules.
3. **`centinela validate`** — Runs built-in gates plus project validate commands.
4. **CI pipeline** — Run `centinela validate` on push/PR.
