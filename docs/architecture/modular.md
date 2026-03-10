# Modular Monolith Architecture Guide

> This document describes the **Modular Monolith archetype**. It is one of five supported architecture archetypes in Centinela. See [architecture-overview.md](architecture-overview.md) to confirm this is the right pattern for your project before reading further. Examples use TypeScript, but the conventions apply to any language with module/package support.

---

## When to Choose Modular Monolith

Choose Modular Monolith when:

- Your application has clear business domain boundaries (billing, users, reports, notifications)
- You need monolith deployment simplicity but want to prevent spaghetti dependencies from forming as the codebase grows
- You may eventually extract modules into microservices and want the structure to make that cheap
- Your team is large enough that multiple people working on the same codebase without domain isolation causes conflicts

Do NOT choose Modular Monolith when:

- The project is small — one module is overkill, three modules is over-engineering
- You already need network-level isolation (latency, independent deploy, separate scaling) — use microservices from the start
- The domain cannot be cleanly partitioned — use Hexagonal within a single module instead

---

## Core Idea

The defining tension of a modular monolith: **monolith deployment** (one process, one database, shared infrastructure) combined with **microservice-level isolation of concerns** (each module hides its internals and exposes only a controlled public API).

The key discipline is the **import boundary**. Modules are partitioned into `public/` and `internal/`. Any code outside a module may import from `public/`. No code outside a module may import from `internal/`. This boundary is enforced by the Gatekeeper and by linting rules.

A module that respects this boundary can be extracted into a standalone service by:
1. Replacing `public/` direct calls with HTTP endpoints
2. Moving `internal/` into its own deployment unit
3. Replacing cross-module event handlers with message queue subscribers

The structure anticipates this extraction without requiring it.

---

## Module Structure

### Public API (`module/public/`)

The only interface another module may use. Everything here is a stable contract.

**What goes in `public/`:**

- Type definitions / interfaces that other modules consume
- Facade functions or classes that expose the module's capabilities
- Event type definitions that this module emits (for subscribers to import)
- DTOs for data that crosses the module boundary

```
modules/billing/public/
  index.ts                       # re-exports everything the public API exposes
  BillingFacade.ts               # methods: chargeCustomer, getInvoice, cancelSubscription
  types.ts                       # InvoiceSummary, SubscriptionStatus, ChargeResult
  events.ts                      # PaymentSucceeded, PaymentFailed, SubscriptionCancelled
```

**Rules for `public/`:**

1. `public/` is a thin facade — it calls into `internal/` services. No business logic lives here.
2. All types exposed in `public/` are DTOs or plain interfaces — never internal model classes.
3. `public/` may not import from `internal/` of any other module.
4. The `index.ts` (or `__init__.py`, or `mod.rs`) explicitly lists every exported symbol — no wildcard re-exports.
5. Breaking changes to `public/` require updating all consuming modules — treat these with the same care as a microservice API contract change.

---

### Internal Implementation (`module/internal/`)

The private implementation. Only files within the same module may import from here.

**What goes in `internal/`:**

- Services — business logic for this domain
- Repositories — data access for this domain's data
- Domain entities and value objects (if the module uses Hexagonal internally)
- Event handlers — handle events from other modules (subscribed via the event bus)
- Helpers, validators, and mappers specific to this module

```
modules/billing/internal/
  services/
    ChargeService.ts             # executes payment charges
    InvoiceService.ts            # creates and retrieves invoices
    SubscriptionService.ts       # manages subscription state
  repositories/
    InvoiceRepository.ts
    SubscriptionRepository.ts
    PaymentMethodRepository.ts
  handlers/
    UserRegisteredHandler.ts     # reacts to users/public/events.ts → UserRegistered
  entities/
    Invoice.ts
    Subscription.ts
    PaymentMethod.ts
  mappers/
    InvoiceMapper.ts
    StripeResponseMapper.ts
  validators/
    ChargeValidator.ts
```

**Rules for `internal/`:**

1. No file outside `modules/billing/` may import from `modules/billing/internal/`. This is absolute.
2. Internal services may import from `shared/` (cross-cutting utilities).
3. Internal services may import from another module's `public/` (to call its API or read its event types).
4. Internal structure within a module is unconstrained — use whatever pattern fits (Hexagonal, N-Tier, flat services). The Gatekeeper only enforces the module boundary, not the internal structure.
5. Internal event handlers subscribe to events from other modules' `public/events.ts` files — they must not import from other modules' `internal/`.

---

### Shared Kernel (`shared/`)

Cross-cutting infrastructure utilities with no business logic. Every module may import from here.

**What goes in `shared/`:**

- Logging utilities
- Error base classes
- Pagination types and helpers
- Date/time formatting utilities
- Configuration loading
- Database connection management
- HTTP client base classes

```
shared/
  errors/
    AppError.ts                  # base error class
    NotFoundError.ts
    ForbiddenError.ts
    ValidationError.ts
  logging/
    logger.ts
  pagination/
    PaginatedResult.ts
    PaginationParams.ts
  database/
    connection.ts
    transaction.ts
  config/
    env.ts
  http/
    HttpClient.ts
```

**Rules for `shared/`:**

1. Zero business logic in `shared/`. If a helper makes a domain-specific decision, it belongs in a module's `internal/`.
2. `shared/` may not import from any module's `public/` or `internal/`. It has no knowledge of modules.
3. Every item in `shared/` must be genuinely used by at least two modules. A utility used only by `billing/` belongs in `billing/internal/`.
4. Avoid "God shared" — `shared/utils.ts` with 500 lines of mixed utilities. Group by responsibility.

---

## Realistic Module Tree

```
src/
  modules/
    users/
      public/
        index.ts                 # exports UserFacade, UserCreatedEvent, UserSummary
        UserFacade.ts            # getUser(id), createUser(data), deactivateUser(id)
        types.ts                 # UserSummary, UserRole
        events.ts                # UserCreated, UserDeactivated, EmailChanged
      internal/
        services/
          UserService.ts
          AuthService.ts
          EmailVerificationService.ts
        repositories/
          UserRepository.ts
        entities/
          User.ts
        handlers/
          (none — users module reacts to nothing from other modules)
        validators/
          CreateUserValidator.ts

    billing/
      public/
        index.ts
        BillingFacade.ts         # chargeCustomer, getInvoice, getSubscriptionStatus
        types.ts                 # InvoiceSummary, SubscriptionStatus, ChargeResult
        events.ts                # PaymentSucceeded, PaymentFailed, SubscriptionCancelled
      internal/
        services/
          ChargeService.ts
          InvoiceService.ts
          SubscriptionService.ts
        repositories/
          InvoiceRepository.ts
          SubscriptionRepository.ts
        entities/
          Invoice.ts
          Subscription.ts
        handlers/
          UserRegisteredHandler.ts   # on UserCreated → create free-tier subscription
          UserDeactivatedHandler.ts  # on UserDeactivated → cancel active subscriptions

    notifications/
      public/
        index.ts
        NotificationFacade.ts    # sendEmail, sendPush, scheduleReminder
        types.ts                 # NotificationChannel, NotificationResult
      internal/
        services/
          EmailService.ts
          PushService.ts
          ReminderService.ts
        handlers/
          PaymentSucceededHandler.ts  # on PaymentSucceeded → send receipt email
          PaymentFailedHandler.ts     # on PaymentFailed → send failure alert
          UserCreatedHandler.ts       # on UserCreated → send welcome email

    reports/
      public/
        index.ts
        ReportFacade.ts          # generateMonthlyReport, getRevenueBreakdown
        types.ts                 # ReportSummary, RevenueBreakdown
      internal/
        services/
          ReportGeneratorService.ts
          RevenueCalculatorService.ts
        repositories/
          ReportReadModel.ts         # cross-module read model (see below)

  shared/
    errors/
      AppError.ts
      NotFoundError.ts
    logging/
      logger.ts
    pagination/
      PaginatedResult.ts
    database/
      connection.ts
      transaction.ts
    config/
      env.ts
```

---

## Cross-module Communication

Three patterns. Use the right one for the situation.

### 1. Direct call to public API (synchronous, same process)

Use when you need a result immediately and the dependency is acceptable.

```typescript
// notifications/internal/handlers/UserCreatedHandler.ts
import { UserFacade } from "../../users/public";  // CORRECT — public API only

export class UserCreatedHandler {
  async handle(event: UserCreatedEvent): Promise<void> {
    const user = await UserFacade.getUser(event.userId);  // direct synchronous call
    await this.emailService.sendWelcome(user.email, user.name);
  }
}
```

```typescript
// FORBIDDEN — importing from another module's internal
import { UserService } from "../../users/internal/services/UserService";  // VIOLATION
```

**When to use:** Module A needs data from Module B to complete its own operation, and the operation is synchronous.

---

### 2. Domain events / event bus (decoupled, async-friendly)

Use when Module A should not depend on Module B's availability, or when multiple modules react to the same event.

```typescript
// billing/internal/services/ChargeService.ts — emits an event
import { EventBus } from "../../../shared/events/EventBus";
import { PaymentSucceeded } from "../public/events";

async processPayment(customerId: string, amount: Money): Promise<ChargeResult> {
  const result = await this.stripeClient.charge(amount);
  if (result.ok) {
    await EventBus.emit(new PaymentSucceeded({ customerId, amount, invoiceId: result.invoiceId }));
  }
  return result;
}

// notifications/internal/handlers/PaymentSucceededHandler.ts — reacts to the event
import { PaymentSucceeded } from "../../billing/public/events";  // imports event TYPE from public

EventBus.subscribe(PaymentSucceeded, async (event) => {
  await emailService.sendReceipt(event.customerId, event.invoiceId);
});
```

**When to use:** Multiple modules react to the same occurrence. The emitting module should not know who consumes its events. The reaction can be asynchronous.

---

### 3. Shared read model (cross-module queries)

Use when a module (typically `reports`) needs to query data that spans multiple modules' domains without coupling to each module's internals.

```typescript
// reports/internal/repositories/ReportReadModel.ts
// This repository has direct DB access to produce cross-module aggregated views.
// It is read-only. It never writes to tables owned by other modules.

export class ReportReadModel {
  async getMonthlyRevenueByPlan(month: Date): Promise<RevenueBreakdown[]> {
    // Direct DB query joining billing.invoices + users.subscriptions
    // This is acceptable ONLY for read models used in reporting.
    // Reports module owns this query; it does not import billing/internal or users/internal.
    return this.db.query(`
      SELECT s.plan_id, SUM(i.amount) as total
      FROM billing.invoices i
      JOIN users.subscriptions s ON i.subscription_id = s.id
      WHERE i.created_at >= $1 AND i.created_at < $2
      GROUP BY s.plan_id
    `, [startOfMonth(month), startOfMonth(addMonths(month, 1))]);
  }
}
```

**When to use:** Reporting and analytics modules that need aggregated views across domains. The read model is always read-only. It never modifies data owned by another module. Treat cross-module DB queries in read models as a deliberate exception — document them.

---

## Dependency Direction

```
shared/ ← modules/X/internal/ ← modules/X/public/
                 ↑
     modules/Y/public/ (Y calls X via public API)
                 ↑
     modules/Y/internal/handlers/ (Y reacts to X's events)
```

Valid imports:
- `internal/` → `shared/`
- `internal/` → same module's `public/`
- `internal/` → another module's `public/`
- `public/` → same module's `internal/`

Forbidden imports:
- `modules/X/internal/` → `modules/Y/internal/` (cross-module internal access)
- `shared/` → any module
- Circular: X's `public/` → Y's `public/` → X's `public/`

---

## Forbidden Patterns (G2)

| Pattern | Why it is forbidden |
|---|---|
| `import` from `modules/billing/internal/` in `modules/users/` | Violates the module boundary — only `billing/public/` is accessible to other modules |
| A module that every other module imports | This is a God module — break it apart or move the shared logic to `shared/` |
| `modules/A/public/` imports `modules/B/public/` imports `modules/A/public/` | Circular dependency — redesign the boundary or introduce an event |
| Business logic in `shared/` | `shared/` is infrastructure utilities only; business logic belongs in a module |
| `public/` exposing internal entity classes | `public/` exposes DTOs and interfaces, not internal domain objects |
| Direct DB access in another module's tables from a non-read-model service | Cross-module writes must go through the module's public API or event handlers |

---

## What "No Business Logic in Outer Layer" Means (G7)

In Modular Monolith, the "outer layer" is each **module's `public/` API**. It must be a thin facade — a dispatcher to `internal/` services. Business logic that lives in `public/` is a violation.

**Violation:**

```typescript
// billing/public/BillingFacade.ts
export const BillingFacade = {
  async chargeCustomer(customerId: string, amount: number): Promise<ChargeResult> {
    // VIOLATION: business logic in public API
    if (amount <= 0) throw new Error("Amount must be positive");
    const customer = await db.query("SELECT * FROM customers WHERE id = $1", [customerId]);
    if (!customer.emailVerified) throw new Error("Customer must verify email");
    const result = await stripeClient.charge(customer.stripeId, amount);
    if (result.declined) await db.query("UPDATE customers SET flagged = true WHERE id = $1", [customerId]);
    return result;
  }
};
```

**Correct:**

```typescript
// billing/public/BillingFacade.ts — thin facade
import { ChargeService } from "../internal/services/ChargeService";

export const BillingFacade = {
  async chargeCustomer(customerId: string, amount: number): Promise<ChargeResult> {
    return ChargeService.charge(customerId, amount);   // delegates entirely to internal
  },
  async getInvoice(invoiceId: string): Promise<InvoiceSummary> {
    return InvoiceService.getById(invoiceId);
  }
};

// billing/internal/services/ChargeService.ts — owns the business logic
export const ChargeService = {
  async charge(customerId: string, amount: number): Promise<ChargeResult> {
    if (amount <= 0) throw new ValidationError("Amount must be positive");
    const customer = await customerRepo.findById(customerId);
    if (!customer) throw new NotFoundError(`Customer ${customerId} not found`);
    if (!customer.emailVerified) throw new ForbiddenError("Customer must verify email");
    const result = await stripeGateway.charge(customer.stripeId, Money.of(amount, "USD"));
    if (!result.ok) await customerRepo.flagForReview(customerId);
    return { success: result.ok, invoiceId: result.invoiceId };
  }
};
```

---

## Preparing for Service Extraction

A module is ready to become a microservice when:

1. Its `public/` API is complete — there are no direct `internal/` imports from outside the module
2. All cross-module communication goes through either the public API or events — no shared DB writes
3. The module's `internal/` repositories only query their own tables (or read-only cross-module views)
4. The module can be stood up in a test environment without the rest of the application

**How the modular structure makes extraction cheap:**

| What you have | What changes during extraction |
|---|---|
| `modules/billing/public/BillingFacade.ts` | Replace direct function calls with HTTP client calls to new billing service |
| `billing/public/events.ts` | Replace in-process event bus emissions with message queue publishes |
| `billing/internal/handlers/` | Replace event bus subscriptions with message queue consumers |
| `billing/internal/repositories/` | Point to billing service's own database; remove shared DB access |

The module boundary was always there — extraction formalises it at the network layer.

**Signs a module is NOT ready for extraction:**

- Other modules import from `internal/` — fix the boundary first
- The module writes to tables owned by another module — those writes must go through the owning module's API
- Circular event dependencies (A emits → B handles → B emits → A handles in the same request cycle)

---

## Testing Strategy

**Unit tests — internal services of each module:**

- Test each `internal/service` in isolation with mock repositories.
- No database, no cross-module calls.
- Confirm business logic is correct.

**Integration tests — module boundary (public API contract tests):**

- Test the full module (public API → internal services → real test database) in isolation.
- Confirm that calling `BillingFacade.chargeCustomer(...)` produces the expected database state and events.
- These tests verify the contract that other modules depend on.

**Acceptance tests — cross-module Gherkin scenarios:**

- Feature files describe user-visible behaviour that spans modules.
- Step definitions exercise the full application (all modules together) with a test database.
- These tests confirm that cross-module communication (direct calls and events) produces correct end-to-end outcomes.

```
tests/
  unit/
    modules/
      billing/
        ChargeService.test.ts
        InvoiceService.test.ts
      users/
        UserService.test.ts
        AuthService.test.ts
      notifications/
        EmailService.test.ts
  integration/
    modules/
      billing/
        BillingFacade.integration.test.ts    # full module, real DB
        UserRegisteredHandler.test.ts        # handler + billing internal
      users/
        UserFacade.integration.test.ts
  acceptance/
    user-registers-and-gets-welcome-email.steps.ts
    payment-succeeds-and-receipt-is-sent.steps.ts
    subscription-cancellation.steps.ts

specs/
  user-registration.feature
  payment-processing.feature
  subscription-management.feature
```
