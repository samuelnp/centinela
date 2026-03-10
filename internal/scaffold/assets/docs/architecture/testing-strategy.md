# Testing Strategy

> This document is archetype-agnostic. Tool names are examples — use the equivalent for your language and stack (PROJECT.md → Tech Stack). The structure and rules apply regardless of framework.

---

## The Three Layers (universal)

Every Centinela project requires all three, regardless of archetype or language:

| Layer | What it tests | Runs against |
|-------|--------------|--------------|
| **Unit** | A single logical unit in isolation | Pure code — no I/O, no network, no DB |
| **Integration** | A boundary crossing (DB, API, file system) | Real or containerised external dependency |
| **Acceptance** | Observable behaviour described in Gherkin | The application layer, with infrastructure mocked |

---

## Unit Tests

Test the smallest meaningful unit with all dependencies mocked or replaced.

**What to test by archetype:**

| Archetype | Unit test targets |
|-----------|------------------|
| Hexagonal | Domain entities, value objects, domain services, use cases (mock ports) |
| Rails-native | Model validations, scopes, service objects |
| N-Tier | Service layer methods (mock repository) |
| ECS | Individual Systems (inject component data directly, no scene) |
| Modular | Each module's internal services (no cross-module calls) |

**Example — Hexagonal (TypeScript/Vitest):**
```typescript
describe('PlaceOrder', () => {
  it('rejects an order with zero items', async () => {
    const useCase = new PlaceOrder(new InMemoryOrderRepository());
    await expect(useCase.execute({ items: [] })).rejects.toThrow('Order must have at least one item');
  });
});
```

**Example — ECS System (GDScript/GUT):**
```gdscript
func test_damage_system_reduces_health():
    var entity = create_entity()
    entity.add_component(HealthComponent.new(100))
    entity.add_component(DamageComponent.new(25))

    DamageSystem.new().process(entity)

    assert_eq(entity.get_component(HealthComponent).value, 75)
```

**Example — N-Tier Service (Python/pytest):**
```python
def test_create_order_calculates_total():
    repo = InMemoryOrderRepository()
    service = OrderService(repo)
    order = service.create_order(items=[Item("book", 12.99), Item("pen", 1.99)])
    assert order.total == 14.98
```

---

## Integration Tests

Test that your code correctly communicates with a real external boundary.

**What to test by archetype:**

| Archetype | Integration test targets |
|-----------|-------------------------|
| Hexagonal | Infrastructure adapters: API clients, DB repositories, cache stores |
| Rails-native | Controller + model + DB roundtrip; mailer delivery |
| N-Tier | Repository queries against a test DB or container |
| ECS | Save/load system against real file I/O; network sync against a local server |
| Modular | Each module's public API contract (verified against a real DB or in-memory store) |

**Example — HTTP adapter (TypeScript/MSW):**
```typescript
it('maps a 404 response to a NotFoundError', async () => {
  server.use(rest.get('/api/orders/:id', (req, res, ctx) => res(ctx.status(404))));
  const client = new HttpOrderClient(baseUrl);
  await expect(client.getOrder('missing')).rejects.toBeInstanceOf(NotFoundError);
});
```

**Example — DB repository (Go/testcontainers):**
```go
func TestOrderRepository_Save(t *testing.T) {
    db := startTestDB(t)
    repo := NewSQLOrderRepository(db)
    order := domain.NewOrder([]domain.Item{{Name: "book", Price: 12.99}})
    require.NoError(t, repo.Save(order))
    found, err := repo.FindByID(order.ID)
    require.NoError(t, err)
    assert.Equal(t, order.Total(), found.Total())
}
```

---

## Acceptance Tests (Gherkin)

Acceptance tests verify observable behaviour described in `.feature` files. They run against the application layer with all infrastructure replaced by in-memory fakes or mocks.

**Gherkin works for every project type.** The step definitions change; the Gherkin format does not.

### Web / API

Step definitions call use cases or HTTP endpoints directly. Infrastructure is an in-memory fake.

```gherkin
# specs/order-placement.feature
Feature: Order placement

  Scenario: Successful order with valid items
    Given a customer with a verified account
    When they place an order for 2 units of "notebook"
    Then the order status is "confirmed"
    And the inventory for "notebook" decreases by 2
```

```typescript
// tests/acceptance/steps/order-placement.steps.ts
Given('a customer with a verified account', function () {
  this.customer = createCustomer({ verified: true });
  this.orderRepo = new InMemoryOrderRepository();
  this.useCase = new PlaceOrder(this.orderRepo, new InMemoryInventory());
});

When('they place an order for {int} units of {string}', async function (qty, item) {
  this.result = await this.useCase.execute({ customer: this.customer, item, qty });
});

Then('the order status is {string}', function (status) {
  expect(this.result.status).toBe(status);
});
```

### Game (ECS)

Step definitions set up component state, run systems for one or more ticks, and assert on resulting component values or emitted events. No scene, no renderer, no engine loop.

```gherkin
# specs/combat.feature
Feature: Combat system

  Scenario: Player takes lethal damage
    Given a player with 10 health
    When the DamageSystem applies 15 damage
    Then the player health is 0
    And a PlayerDied event is emitted

  Scenario: Shield absorbs overflow damage
    Given a player with 5 health and a shield with 20 durability
    When the DamageSystem applies 30 damage
    Then the shield durability is 0
    And the player health is 0
    And a PlayerDied event is emitted
```

```gdscript
# tests/acceptance/steps/combat_steps.gd
func step_given_player_with_health(context, health):
    context.entity = EntityFactory.create_player(health)
    context.damage_system = DamageSystem.new()

func step_when_damage_applied(context, amount):
    context.entity.get_component(DamageComponent).pending = amount
    context.damage_system.process([context.entity])

func step_then_health_is(context, expected):
    assert_eq(context.entity.get_component(HealthComponent).value, expected)

func step_then_event_emitted(context, event_name):
    assert_true(context.event_bus.was_emitted(event_name))
```

### CLI Tool

Step definitions invoke the binary as a subprocess and assert on stdout, stderr, exit code, and output files.

```gherkin
# specs/report-generation.feature
Feature: Report generation

  Scenario: Generate report from valid config
    Given a config file with 2 data sources
    When I run "centinela report --output ./out"
    Then the exit code is 0
    And a file "out/report.md" is created
```

```typescript
// tests/acceptance/steps/report.steps.ts
When('I run {string}', async function (cmd) {
  this.result = await exec(cmd, { cwd: this.tmpDir });
});

Then('the exit code is {int}', function (code) {
  expect(this.result.exitCode).toBe(code);
});
```

---

## Folder Structure

Adapt paths to your project. The logical separation matters more than the exact names.

```
specs/
  <feature-name>.feature      ← one file per feature or domain area

tests/
  unit/
    <mirrors source structure>
  integration/
    <mirrors infrastructure or boundary structure>
  acceptance/
    steps/
      <feature-name>.steps.ts  ← matches specs/<feature-name>.feature
    support/
      world.ts                 ← shared test context / state between steps
      hooks.ts                 ← before/after scenario setup
  fixtures/
    <entity>.factory.ts        ← factory functions; avoid static JSON fixtures
```

---

## Test Factories over Static Fixtures

Prefer factory functions to JSON fixture files. They stay in sync with your domain and are composable.

```typescript
// tests/fixtures/order.factory.ts
export function createOrder(overrides: Partial<OrderProps> = {}): Order {
  return Order.create({
    id: 'order-001',
    items: [{ name: 'book', price: 12.99 }],
    status: 'pending',
    ...overrides,
  });
}
```

---

## Running Tests

Configure your test commands in `centinela.toml`:

```toml
[validate]
commands = [
  "npx tsc --noEmit",     # type check
  "npx vitest run",       # unit + integration
  "npx cucumber-js",      # acceptance (Gherkin)
]
```

Adapt to your stack (pytest, go test, cargo test, bundle exec rspec, etc.).
`centinela validate` and `centinela complete <feature>` (at the validate step) run all commands in sequence. All three test layers must pass before a workflow can complete.
