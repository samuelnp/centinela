<!-- centinela:doc-version=1 template=docs/architecture/gatekeepers.md -->
# Gate Keepers

Automated and manual checks that MUST pass before any feature ships.

> Gate rules G2 and G7 are archetype-specific. See [architecture-overview.md](architecture-overview.md) for the exact rules per archetype. All other gates are universal.

---

## Pre-Commit Gates

### G1: File Size Limit
- **Rule:** Source files default to a 100-line maximum. Rare, explicit exceptions are allowed only when justified in `centinela.toml` and capped at 130 lines.
- **Check:** `scripts/check-file-size.sh` (or equivalent for your language).
- **Fail action:** Block commit. Split the file or add a valid justified exception.

### G2: Layer Dependency Violations
- **Rule:** No imports cross forbidden layer boundaries. The forbidden boundaries are defined per archetype in PROJECT.md → Architecture Choice → G2 rule.
- **Check:** Static analysis configured for the archetype (dependency-cruiser, ESLint import plugin, RuboCop, custom linter).
- **Fail action:** Block commit. Fix the import direction.

### G3: Type Safety / Static Analysis
- **Rule:** Strictest static analysis mode enabled for the project language. No dynamic typing shortcuts.
- **Check:** Language type checker (`tsc --noEmit`, `mypy --strict`, `cargo check`, `srb tc`).
- **Fail action:** Block commit. Add proper types.

### G4: Tests Present and Passing
- **Rule:** Every logical unit has a unit test. Every external boundary has an integration test. Every `.feature` file has passing step definitions.
- **What "logical unit" means per archetype:**

  | Archetype | Unit tests cover | Integration tests cover |
  |-----------|-----------------|------------------------|
  | Hexagonal | Domain entities, value objects, use cases | Infrastructure adapters (API clients, repositories) |
  | Rails-native | Models, service objects | Controller + DB roundtrip |
  | N-Tier | Service layer | Repository layer (real DB or test container) |
  | ECS | Systems in isolation | System + real component data |
  | Modular | Each module's internal services | Module public API contracts |

- **Check:** `centinela validate` (runs all commands in `centinela.toml`)
- **Fail action:** Block commit. Write the missing tests.

---

## Pre-Feature Gates

### G5: Spec First — Gherkin `.feature` File
- **Rule:** Every feature has a `.feature` file in `specs/` BEFORE any implementation starts.
- **Check:** Manual review — does the feature file exist and describe acceptance criteria in Given/When/Then form?
- **Fail action:** Do not write code. Write the spec first.

#### Gherkin applies to every project type

Gherkin describes **observable behaviour**, not HTTP endpoints. The form is the same regardless of stack:

```
Given  [some initial state / precondition]
When   [an action is taken]
Then   [the observable outcome]
```

**Web application:**
```gherkin
Scenario: User places an order
  Given a user has 3 items in their cart
  When they confirm the order
  Then an order record is created
  And a confirmation email is queued
```

**Game (ECS):**
```gherkin
Scenario: Player takes damage
  Given a player entity with 100 health
  When the DamageSystem processes a collision with a 25-damage enemy
  Then the player health component value is 75

Scenario: Player dies when health reaches zero
  Given a player entity with 10 health
  When the DamageSystem applies 15 damage
  Then the player health is 0
  And a PlayerDied event is emitted
```

**CLI tool:**
```gherkin
Scenario: Generate report from config file
  Given a config file with two data sources
  When the CLI runs with the --report flag
  Then a report file is written to the output directory
  And the exit code is 0
```

**Library / SDK:**
```gherkin
Scenario: Parse malformed input
  Given the parser is initialised
  When it receives input with a missing required field
  Then it returns a ParseError with code MISSING_FIELD
```

The step definitions for a game don't use a browser or HTTP client — they set up component data, call systems directly, and assert on component state. The Gherkin layer stays the same; only the driver changes.

### G5.1: Gatekeeper Conflict Review
- **Rule:** Before implementing, invoke the Gatekeeper subagent to review the new spec against all existing specs for conflicts.
- **Check:** Run the Gatekeeper subagent (prompt in [gatekeeper-prompt.md](gatekeeper-prompt.md)).
- **Output:** `SAFE` / `WARNING` / `BLOCKING` report saved to `.workflow/<feature>-gatekeeper.md`.
- **Fail action:**
  - `SAFE` → proceed
  - `WARNING` → document acknowledged risks in the plan, proceed with caution
  - `BLOCKING` → resolve conflicts before writing any code

### G6: Plan Documented
- **Rule:** Every feature has a written plan in `docs/plans/` before implementation starts.
- **Check:** Manual review.
- **Fail action:** Write the plan first.

### G7: No Business Logic in the Outer Layer
- **Rule:** The outer layer contains no business or game logic. What counts as the "outer layer" is archetype-specific:

  | Archetype | Outer layer | What is forbidden there |
  |-----------|-------------|------------------------|
  | Hexagonal | UI components, infrastructure adapters | Conditionals based on domain rules, data transformation |
  | Rails-native | Views, templates, route handlers | DB queries, business conditionals |
  | N-Tier | HTTP handlers / controllers | DB queries, business rules |
  | ECS | Scene nodes, visual nodes (Node2D, MonoBehaviour) | Game logic, state mutation |
  | Modular | A module's `public/` API | Implementation logic (it must be a thin facade) |

- **Check:** Code review.
- **Fail action:** Move logic to the appropriate inner layer.

### G8: Single Responsibility
- **Rule:** Each file exports one thing and does one thing.
- **Check:** Code review.
- **Fail action:** Split the file.

---

## Post-Feature Gates

### G9: Full Test Suite Passes
- **Rule:** All tests exit with 0 — unit, integration, and acceptance.
- **Check:** `centinela validate` (runs all commands in `centinela.toml`). Also run in CI pipeline.
- **Fail action:** Fix failing tests before merge.

### G10: Acceptance Regression
- **Rule:** All existing Gherkin scenarios still pass after the new feature.
- **Check:** Run your acceptance test command (configured in `centinela.toml → [validate] commands`).
- **Fail action:** New feature broke existing behaviour. Fix before merge.

### G11: i18n Complete *(only if project uses i18n)*
- **Rule:** No hardcoded user-facing strings. Every locale listed in PROJECT.md → Locales is complete — no missing keys or entries.
- **Applies to:** Any project with locales defined in PROJECT.md → Locales. **Skip this gate entirely** if PROJECT.md → Locales is empty or absent.
- **Check:** Configure `centinela.toml → [gates] i18n = true` and set `[i18n]` section. For formats not natively supported, add a custom command to `[validate] commands`.
- **Fail action:** Add the missing translations in the appropriate format for your stack.

**Translation formats — `centinela.toml` configuration:**

| Stack / Engine | Format | `centinela.toml` setting |
|---------------|--------|--------------------------|
| Web (next-intl, i18next, vue-i18n) | JSON locale files (`en.json`, `es.json`) | `format = "json"` — built-in key parity check |
| Godot | Gettext `.po` / `.pot` files | `format = "gettext"` — built-in untranslated entry check |
| Unity | Localization Tables (CSV or JSON asset) | `format = "none"` + custom command in `[validate] commands` |
| Android | `res/values-<locale>/strings.xml` | `format = "none"` + custom command |
| iOS / macOS | `<locale>.lproj/Localizable.strings` | `format = "none"` + custom command |
| Game (custom CSV) | String table CSV | `format = "none"` + custom command |

**Example `centinela.toml` for a JSON web project:**
```toml
[gates]
i18n = true

[i18n]
format  = "json"
dir     = "src/i18n/messages"
locales = ["en", "es"]
```

**Example `centinela.toml` for Godot (gettext):**
```toml
[gates]
i18n = true

[i18n]
format  = "gettext"
dir     = "i18n"
locales = ["en", "es"]
```

---

## Gate Enforcement

Gates are enforced at four levels:

1. **CLAUDE.md** — the AI agent reads and follows all gates as hard rules.
2. **`centinela hook prewrite`** — blocks file writes in the wrong workflow step (via agent integrations).
3. **`centinela validate`** — runs built-in gates (G1, G11) + all user commands from `centinela.toml`. G1 and G11 honor `[validate] diff_mode` and walk only files changed since `diff_base` (default `main`) when the resolved mode is diff-aware. Use `--changed` / `--full` to override per invocation. CI (detected via `CI=true`) defaults to a full scan so the ship gate stays strict; local runs default to diff-aware for a faster inner loop. User `[validate] commands` are not scoped by the diff.
4. **CI pipeline** — all gates run on every push (once CI is configured).
