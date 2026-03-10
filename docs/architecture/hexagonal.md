# Hexagonal Architecture Guide

> This document describes the **hexagonal archetype**. It is one of five supported architecture archetypes in Centinela. See [architecture-overview.md](architecture-overview.md) to confirm this is the right pattern for your project before reading further. Adapt all file paths to your project per PROJECT.md → Folder Structure.

---

## When to Choose Hexagonal

Choose Hexagonal when your project has:

- Multiple external integrations (APIs, databases, queues) that need to be independently replaceable
- Domain logic complex enough that it deserves to be tested without any infrastructure running
- A requirement to swap implementations without touching business logic (e.g., switch from Postgres to MongoDB, or from Stripe to Adyen)

Do NOT choose Hexagonal for:

- Simple CRUD apps — the abstraction cost is not worth it
- Framework-opinionated stacks (Rails, Django) — these have their own mature conventions; see [rails-native.md](rails-native.md)
- Games — use [ecs.md](ecs.md) instead

---

## Core Idea

The Dependency Inversion Principle applied end to end: **the domain owns its own interfaces (ports)**. Infrastructure implements those interfaces (adapters). The application layer orchestrates the domain. Nothing in the domain knows about HTTP, databases, or external APIs.

If you can delete every file in `infrastructure/` and replace them with in-memory fakes, the domain and application layers must still compile and pass their tests. That is the test of correct Hexagonal structure.

---

## Layer Details

### Domain

The innermost layer. Zero external dependencies — no framework imports, no database drivers, no HTTP clients. Pure logic and data.

**What it contains:**

- **Entities** — business objects with identity (two Orders with the same `id` are the same Order regardless of other fields)
- **Value objects** — immutable descriptors without identity (two `Money(100, "USD")` are equal if their values match)
- **Ports** — interfaces (abstract classes, protocols, interfaces) that the domain requires from the outside world
- **Domain services** — logic that operates across multiple entities and doesn't belong on any single one

```
domain/
  entities/
    Order.ts
    Customer.ts
    Product.ts
    LineItem.ts
  value-objects/
    Money.ts
    OrderStatus.ts
    Address.ts
    Quantity.ts
    Email.ts
  ports/
    repositories/
      OrderRepository.ts       # interface: findById, save, findByCustomer
      CustomerRepository.ts    # interface: findById, findByEmail, save
      ProductRepository.ts     # interface: findById, findBySku, listAvailable
    services/
      PaymentGateway.ts        # interface: charge(amount, method): PaymentResult
      NotificationSender.ts    # interface: send(recipient, template, data): void
      InventoryService.ts      # interface: reserve(productId, qty): boolean
  services/
    OrderPricingService.ts     # computes totals, discounts, taxes across line items
    FulfillmentEligibility.ts  # decides if an order can be fulfilled
```

**Rules:**

1. No `import` from Application, Infrastructure, or any framework. The only allowed imports are other files within `domain/` and the language standard library.
2. Entities are instantiated via factory methods or constructors with explicit parameters — never from raw database rows.
3. Value objects must be immutable. All properties are `readonly`. Mutation returns a new instance.
4. Ports are interfaces only — no implementation code, no logic.
5. Domain services contain only logic that cannot logically belong to a single entity. If the logic only touches one entity, put it on that entity.
6. No `async` in domain entities or value objects — IO is an infrastructure concern.
7. Domain validation lives in value object constructors and entity factory methods, not in application layer validators.

---

### Application

Orchestrates domain objects to fulfill use cases. This layer knows the domain but knows nothing about HTTP, SQL, or external APIs.

**What it contains:**

- **Use cases** — one class per operation, one public `execute()` method
- **DTOs** — plain data transfer objects for input and output; never expose domain entities to callers
- **Application services** — rare; shared logic across use cases that is not domain logic (e.g., computing a paginated response)

```
application/
  use-cases/
    orders/
      PlaceOrder.ts            # validates input, checks inventory, creates Order entity, saves
      CancelOrder.ts
      GetOrderDetail.ts
      ListCustomerOrders.ts
    customers/
      RegisterCustomer.ts
      UpdateShippingAddress.ts
    products/
      GetProductCatalog.ts
      CheckProductAvailability.ts
  dtos/
    OrderDetailDto.ts
    OrderSummaryDto.ts
    PlaceOrderCommand.ts       # input DTO
    CustomerProfileDto.ts
    ProductListDto.ts
    LineItemDto.ts
  services/
    PaginationService.ts       # shared: compute offset/limit, build PaginatedResult<T>
```

**Rules:**

1. Use cases receive ports via constructor injection — never instantiate infrastructure classes directly.
2. Use cases return DTOs, never domain entities. Callers must not depend on domain internals.
3. Each use case has exactly one public method named `execute()` (or `run()` — pick one and be consistent).
4. Use cases must not call other use cases. Extract shared logic to a domain service or application service.
5. Application layer may import Domain. It must not import Infrastructure.
6. Input DTOs are validated at the application boundary — reject bad input before touching domain logic.
7. Use case files have one class each. If a file exceeds 100 lines, the use case is doing too much — split it.

---

### Infrastructure

Concrete implementations of the ports defined in the domain. This is where framework code, database drivers, and HTTP clients live.

**What it contains:**

- **Repository adapters** — implement domain repository ports against a real database or external API
- **Service adapters** — implement domain service ports (payment gateway, email sender, etc.)
- **Mappers** — translate between external data formats (API responses, DB rows) and domain entities
- **DI container** — the single location where ports are bound to concrete implementations

```
infrastructure/
  persistence/
    postgres/
      PostgresOrderRepository.ts    # implements OrderRepository port
      PostgresCustomerRepository.ts
      PostgresProductRepository.ts
      mappers/
        OrderMapper.ts              # DB row → Order entity / Order entity → DB row
        CustomerMapper.ts
        ProductMapper.ts
    redis/
      RedisCacheStore.ts
  payments/
    stripe/
      StripePaymentGateway.ts       # implements PaymentGateway port
      mappers/
        StripeResponseMapper.ts
  notifications/
    sendgrid/
      SendGridNotificationSender.ts # implements NotificationSender port
  inventory/
    WarehouseApiClient.ts           # raw HTTP client
    WarehouseInventoryService.ts    # implements InventoryService port
    mappers/
      WarehouseResponseMapper.ts
  di/
    container.ts                    # binds all ports to their implementations
```

**Rules:**

1. Infrastructure adapters implement exactly one domain port each. One class, one port.
2. Infrastructure never exports domain types to callers — it only satisfies port interfaces.
3. All external data (API responses, DB rows) is mapped to domain entities via dedicated mapper files. Mappers handle missing or malformed fields defensively.
4. `container.ts` is the only place where infrastructure classes are instantiated and bound to ports. No `new ConcreteRepository()` anywhere else.
5. No business logic in mappers or adapters. A mapper that branches on business conditions is a violation — that logic belongs in the domain.
6. HTTP clients (raw API wrappers) are separate from repository/service adapters. The adapter uses the client; they are not the same class.

---

### Presentation / UI

The outermost layer. Accepts data via props or hooks and renders it. Contains no business logic.

**What it contains:**

- **UI components** — purely presentational; receive all data via props
- **Hooks / controllers** — bridge between the UI and the application layer; manage UI state (loading, error, data); call use cases
- **Providers** — make the DI container available to the component tree

```
ui/
  components/
    orders/
      OrderList.tsx
      OrderCard.tsx
      OrderStatusBadge.tsx
      LineItemRow.tsx
    checkout/
      CheckoutForm.tsx
      PaymentMethodSelector.tsx
      AddressInput.tsx
    products/
      ProductGrid.tsx
      ProductCard.tsx
    common/
      LoadingSpinner.tsx
      ErrorBanner.tsx
      Pagination.tsx
  hooks/
    useOrderList.ts           # calls ListCustomerOrders use case, exposes loading/error/data
    useOrderDetail.ts
    useCheckout.ts            # calls PlaceOrder use case, handles form state
    useProductCatalog.ts
  providers/
    AppProvider.tsx           # injects DI container into React context
  pages/                      # or app/ for Next.js — thin routing only
    OrdersPage.tsx            # renders <OrderList /> with useOrderList
    OrderDetailPage.tsx
    CheckoutPage.tsx
```

**Rules:**

1. Components receive all data and callbacks via props. No direct use case calls, no `fetch`, no imports from Infrastructure.
2. Hooks import from Application (use cases) only — never from Infrastructure or Domain directly.
3. Hooks manage UI-level state: loading booleans, error messages, pagination cursor. Business state lives in DTOs returned by use cases.
4. Pages/route handlers are thin: instantiate the relevant hook and render one component from `ui/`.
5. A component that contains an `if (order.status === 'fulfilled') { applyDiscount() }` branch is a violation — that logic belongs in a domain service.
6. Components must be unit-testable with props alone — no context, no DI container required for isolated tests.

---

## Dependency Direction

```
Domain ← Application ← Infrastructure
              ↑
         Presentation
         (via hooks/bridge)
```

Arrows indicate the direction of allowed imports. Domain is imported by Application. Application is imported by Infrastructure (to implement ports) and by Presentation (via hooks). Domain knows nothing above it.

Presentation does not import Infrastructure. The DI container (Infrastructure) wires things together at startup; components consume use cases through a context provider.

---

## Forbidden Imports (G2)

| Layer | May import | May NOT import |
|---|---|---|
| Domain | Domain only (+ stdlib) | Application, Infrastructure, Presentation, any framework |
| Application | Domain | Infrastructure, Presentation, HTTP libraries, DB drivers |
| Infrastructure | Domain, Application | Presentation |
| Presentation (hooks) | Application | Infrastructure, Domain (except types/DTOs re-exported from Application) |
| Presentation (components) | Nothing application-logic-related | Application, Infrastructure, Domain |

The most common violation to guard against: a hook that imports a concrete repository class from Infrastructure. Hooks must consume use cases through the DI container, not instantiate infrastructure directly.

---

## What "No Business Logic in Outer Layer" Means (G7)

**Infrastructure violations — NOT allowed:**

```typescript
// BAD: business rule inside a mapper
class OrderMapper {
  toDomain(row: OrderRow): Order {
    // This is a business rule: what constitutes a "late" order
    const isLate = row.created_at < Date.now() - 7 * 24 * 60 * 60 * 1000;
    return Order.create({ ...row, flaggedAsLate: isLate });
  }
}

// BAD: conditional logic in a repository adapter
class PostgresOrderRepository implements OrderRepository {
  async save(order: Order): Promise<void> {
    if (order.totalAmount > 1000) {
      await this.notifyFraudTeam(order); // fraud logic does not belong here
    }
    await this.db.query(...);
  }
}
```

**Infrastructure — allowed:**

```typescript
// GOOD: mapper just translates shape
class OrderMapper {
  toDomain(row: OrderRow): Order {
    return Order.reconstitute({
      id: row.id,
      status: OrderStatus.fromString(row.status), // value object handles validation
      total: Money.of(row.total_cents, row.currency),
      createdAt: new Date(row.created_at),
    });
  }
}
```

**Presentation violations — NOT allowed:**

```typescript
// BAD: discount logic inside a component
function OrderCard({ order }: { order: OrderDetailDto }) {
  const discountedPrice = order.subtotal * (order.customer.isVip ? 0.9 : 1.0);
  return <div>{discountedPrice}</div>;
}

// BAD: hook making a second API call based on business condition
function useOrderDetail(id: string) {
  const order = useFetchOrder(id);
  if (order?.status === 'pending') {
    fetchInventoryCheck(order.lineItems); // this decision belongs in a use case
  }
}
```

**Presentation — allowed:**

```typescript
// GOOD: component renders what it receives
function OrderCard({ order }: { order: OrderDetailDto }) {
  return (
    <div>
      <span>{order.formattedTotal}</span>   {/* formatting computed in DTO/mapper */}
      <OrderStatusBadge status={order.status} />
    </div>
  );
}

// GOOD: hook calls use case, exposes UI state
function useOrderDetail(id: string) {
  const [state, setState] = useState<{ loading: boolean; order: OrderDetailDto | null }>(
    { loading: false, order: null }
  );
  const { getOrderDetail } = useAppContext(); // use case from DI container
  useEffect(() => {
    setState(s => ({ ...s, loading: true }));
    getOrderDetail.execute({ id }).then(order => setState({ loading: false, order }));
  }, [id]);
  return state;
}
```

---

## Testing Strategy in Hexagonal

**Unit tests — Domain and Application layers only:**

- Test entities with plain constructor calls. No database, no HTTP.
- Test use cases by injecting in-memory fakes that implement the domain ports.
- In-memory fakes live in `tests/fakes/` and implement ports directly — they are not mocks.

**Integration tests — Infrastructure adapters in isolation:**

- Test each repository against a real database (test container or embedded DB) or against a mock HTTP server (e.g., MSW for HTTP APIs).
- One integration test file per adapter. Tests confirm the adapter correctly maps to/from domain entities.
- These tests do not exercise use cases or domain logic.

**Acceptance tests — Full stack via Gherkin:**

- Feature files in `specs/` describe user-visible behaviour in domain language.
- Step definitions wire Gherkin steps to the full application (use cases + real or stub infrastructure).
- Acceptance tests do not test presentation internals — they test that a use case, given certain inputs, produces the correct output.

```
tests/
  unit/
    domain/
      entities/Order.test.ts
      value-objects/Money.test.ts
      services/OrderPricingService.test.ts
    application/
      use-cases/PlaceOrder.test.ts          # uses in-memory fakes
      use-cases/CancelOrder.test.ts
  fakes/
    InMemoryOrderRepository.ts              # implements OrderRepository port
    InMemoryPaymentGateway.ts               # implements PaymentGateway port
  integration/
    persistence/PostgresOrderRepository.test.ts
    payments/StripePaymentGateway.test.ts
  acceptance/
    place-order.steps.ts
    cancel-order.steps.ts
specs/
  place-order.feature
  cancel-order.feature
```
