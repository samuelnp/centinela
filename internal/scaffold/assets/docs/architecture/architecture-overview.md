# Architecture Overview

**Centinela** enforces **how you build** (workflow, testing, separation of concerns).
It does not mandate **which architecture pattern** you use. The pattern is set once,
in PROJECT.md → Architecture Choice, by the `/centinela-setup` wizard.

---

## The Universal Principle

Every architecture, no matter how different, shares one rule:

> **Business or game logic must not leak into the outer layer.**

What "outer layer" means depends on the pattern:
- Hexagonal → UI components and infrastructure adapters
- Rails-native → views and route handlers
- N-Tier → HTTP handlers / controllers
- ECS → Components (they must be pure data, no logic)
- Modular → a module's internal files (not its public API)

The framework enforces this principle. The archetype defines it.

---

## Supported Archetypes

### 1. Hexagonal (Ports and Adapters)

**When to use:** Complex domain with multiple external integrations (APIs, message queues, multiple databases) that need to be independently swappable and testable.

**When NOT to use:** Simple CRUD apps, framework-opinionated stacks (Rails), games.

**Examples:** Sauron (Jira + GitHub dashboard), enterprise backends, data pipelines with multiple sources.

**Layer structure:**
```
Domain          ← pure business logic, zero dependencies
Application     ← use cases, orchestrates domain objects
Infrastructure  ← implements domain ports (API clients, DBs, etc.)
Presentation    ← UI components, accepts data via props/hooks
```

**G2 rule:** Domain cannot import Application or Infrastructure. Application cannot import Infrastructure. Presentation cannot import Application directly (must go through a bridge layer).

**G7 rule:** No business logic in Infrastructure adapters or Presentation components.

**Reference:** [hexagonal.md](hexagonal.md)

---

### 2. Rails-native (MVC + Fat Model)

**When to use:** Any Rails, Django, or Laravel project. These frameworks have their own opinionated architecture — adding hexagonal layers on top adds complexity without benefit and fights the framework's conventions.

**When NOT to use:** Apps with complex domain logic that doesn't map to CRUD operations (use Hexagonal instead, possibly alongside Rails).

**Examples:** Internal tools, content platforms, admin dashboards, CRUD-heavy web apps.

**Layer structure:**
```
Views / Templates   ← display only, no logic (ERB, Jinja, Blade)
Controllers         ← thin: parse request, call model/service, render
Models              ← business logic + database (Active Record)
Services / POROs    ← complex operations that don't belong on a model
Jobs / Workers      ← async operations
```

**G2 rule:** Controllers must be thin — no business logic, no direct DB queries beyond what a model provides. Views must not call models directly. Services are plain objects, not framework-coupled.

**G7 rule:** No logic in views or templates. No DB queries in controllers (delegate to model/service).

**Key insight:** Active Record intentionally couples business logic to the database layer. This is a Rails design decision, not a violation. Do not fight it with unnecessary abstraction layers.

**Reference:** [rails-native.md](rails-native.md)

---

### 3. N-Tier / Layered

**When to use:** REST APIs, microservices, backend services, GraphQL servers. Straightforward request-response flow without complex domain logic.

**When NOT to use:** Projects where the domain complexity would benefit from explicit domain modelling (use Hexagonal instead).

**Examples:** Express API, FastAPI service, Go HTTP server, Fastify microservice.

**Layer structure:**
```
Handler / Controller  ← parse request, validate input, call service, format response
Service / Business    ← business logic, orchestration, transformation
Repository / Data     ← data access (DB queries, external API calls)
```

**G2 rule:** Handler may only call Service. Service may only call Repository. Repository may not call Service. No layer skips a level.

**G7 rule:** No DB queries in Handlers. No business logic in Repositories. No HTTP/request concepts in Services.

**Reference:** [n-tier.md](n-tier.md)

---

### 4. ECS (Entity-Component-System)

**When to use:** Games, simulations, real-time applications. Any project where behaviour emerges from composition of data rather than inheritance hierarchies.

**When NOT to use:** Traditional web apps — the pattern solves a different problem.

**Examples:** Godot games, Unity projects, game servers, physics simulations.

**Layer structure:**
```
Entities    ← identity only (an ID, a container); no logic, no data
Components  ← pure data bags (position, health, velocity); no methods/logic
Systems     ← all logic; query entities by component composition and process them
Autoloads   ← global services (event bus, save manager, audio manager)
```

**G2 rule:** Components have no methods (only exported data fields). Systems do not call other Systems directly — communicate via events or a message bus. Entities contain no logic.

**G7 rule:** No game logic in visual/scene nodes (e.g., Godot's Node2D subclasses). Visuals react to component state; they do not modify it.

**Key insight:** In Godot, `Node` scripts attached to scene nodes are the "presentation layer". Game logic belongs in Systems or Autoloads. Scene nodes read component state and render; they don't own logic.

**Reference:** [ecs.md](ecs.md)

---

### 5. Modular Monolith

**When to use:** Larger applications organized by business domain that maintain monolith deployment simplicity but need module isolation to prevent spaghetti dependencies.

**When NOT to use:** Small apps (over-engineering), or when you need microservice isolation at the network level.

**Examples:** Growing SaaS products, feature-rich internal platforms, apps preparing for eventual service extraction.

**Layer structure:**
```
modules/
  orders/
    public/       ← the module's public API (imports from here are OK)
    internal/     ← implementation details (no external imports allowed)
  payments/
    public/
    internal/
  shared/         ← shared utilities with no business logic
```

**G2 rule:** Module A may only import from `modules/B/public/`. Importing from `modules/B/internal/` from outside module B is forbidden. `shared/` may be imported by anyone but must contain no business logic.

**G7 rule:** Each module's `public/` API is intentionally minimal — expose only what other modules need, hide the rest.

**Reference:** [modular.md](modular.md)

---

### 6. Custom

**When to use:** Your project doesn't fit any of the above patterns.

**How to define it:**
In PROJECT.md → Architecture Choice, document:
1. The layer names and what each contains
2. The dependency direction (which layers may import from which)
3. What "business logic in the wrong layer" means for your pattern
4. The paths the Gatekeeper should scan

The `/centinela-setup` wizard will ask you to describe these explicitly.

---

## What Changes Per Archetype

| Gate | Universal rule | Archetype-specific definition |
|------|---------------|-------------------------------|
| G1 — File size | Max 100 lines | Same for all archetypes |
| G2 — Layer boundaries | No forbidden cross-layer imports | Defined per archetype above |
| G3 — Type safety | Strictest mode for the language | Same for all archetypes |
| G4 — Tests | Unit + integration + acceptance | Same for all archetypes |
| G5 — Spec first | .feature file before implementation | Same for all archetypes |
| G7 — Wrong layer | No business logic in outer layer | "Outer layer" defined per archetype |
| G11 — i18n | All locales complete | Only applies if project has i18n |

## What Never Changes

Regardless of archetype:

- **4-step workflow** — plan → code → tests → validate
- **Spec-first** — Gherkin before implementation
- **Tests mandatory** — unit + acceptance cannot be skipped
- **Gatekeeper AI** — conflict check before every implementation
- **`scripts/validate.sh`** — full validation before shipping
- **Max 100 lines per file** — no exceptions
