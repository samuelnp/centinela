<!-- centinela:doc-version=1 template=docs/architecture/ecs.md -->
# ECS (Entity-Component-System) Architecture Guide

> This document describes the **ECS archetype**. It is one of five supported architecture archetypes in Centinela. See [architecture-overview.md](architecture-overview.md) to confirm this is the right pattern for your project before reading further. Examples use GDScript (Godot) and C# (Unity) where engine-specific syntax matters.

---

## When to Choose ECS

Choose ECS when:

- You are building a game, simulation, or real-time interactive application
- Behaviour must emerge from combining independent capabilities rather than inheritance chains
- You need the same "thing" (e.g., a character) to behave completely differently based on which data components are attached to it
- Performance matters and you need cache-friendly data layout (especially in pure ECS implementations)

Do NOT choose ECS when:

- You are building a traditional web application — the pattern is designed for entirely different problems
- Your project uses Rails, Django, or Laravel — use Rails-native
- Your domain has clear service boundaries but no real-time composition requirement — use Hexagonal or Modular

---

## Core Idea

ECS is a fundamental departure from Object-Oriented layered architecture. In OOP, a `Player` class has health, position, and behaviour methods all bundled together. In ECS:

- An **Entity** is only an ID — a handle to associate components together
- A **Component** is only data — a struct with no methods, no logic
- A **System** is only logic — it queries all entities that have a specific set of components, and processes them

A player character is not a `Player` class. It is an entity with a `Position` component, a `Health` component, an `InputController` component, and a `Sprite` component. The `MovementSystem` processes every entity that has `Position` + `InputController`. The `HealthSystem` processes every entity with `Health`. Neither system knows about "player" as a concept.

This data-driven composition replaces inheritance. Adding a new capability means adding a new component and writing a new system — not modifying existing classes.

---

## The Three Primitives

### Entities

An entity is an integer ID and nothing else. It has no data fields. It has no methods. Its only purpose is to serve as a unique key that groups components together.

```gdscript
# Godot — entities are often Node instances used as ID containers, or pure integers
# Entity ID 42 might have: Position, Health, Velocity, Inventory

# Pure ECS (e.g., a custom implementation)
var entity_id: int = world.create_entity()   # returns an integer
world.add_component(entity_id, Position.new(100, 200))
world.add_component(entity_id, Health.new(100))
world.add_component(entity_id, Velocity.new())
```

```csharp
// Unity DOTS — entities are structs
Entity player = entityManager.CreateEntity();
entityManager.AddComponentData(player, new Position { Value = new float3(0, 0, 0) });
entityManager.AddComponentData(player, new Health { Current = 100, Max = 100 });
```

**Rules for entities:**

1. Entities carry no data fields beyond an ID.
2. Entities have no methods.
3. "Creating a player" means creating an entity and attaching the correct components — there is no `Player` class to instantiate.
4. Entities are destroyed by removing them from the world (which also removes all their components).

---

### Components

Pure data containers. No methods (except simple constructors). No game logic.

**What goes in a component:** any piece of state that a system needs to read or write about an entity.

```
components/
  physics/
    Position.gd         # x: float, y: float
    Velocity.gd         # dx: float, dy: float
    Collider.gd         # shape: Shape2D, layer: int, mask: int
    RigidBody.gd        # mass: float, drag: float, is_grounded: bool
  gameplay/
    Health.gd           # current: int, max: int, is_dead: bool
    Inventory.gd        # items: Array[ItemData], capacity: int
    Equipment.gd        # weapon_slot: ItemData, armor_slot: ItemData
    Experience.gd       # level: int, xp: int, xp_to_next: int
    Status.gd           # effects: Array[StatusEffect]
  ai/
    AiBrain.gd          # state: AiState, target_entity: int, patrol_points: Array[Vector2]
    Aggression.gd       # detection_radius: float, aggression_level: float
  rendering/
    Sprite.gd           # texture: Texture2D, animation: StringName, flip_h: bool
    Billboard.gd        # text: String, color: Color, visible: bool
  input/
    PlayerInput.gd      # move_vector: Vector2, jump_pressed: bool, attack_pressed: bool
    AiInput.gd          # computed_move: Vector2, intent: AiIntent
```

```csharp
// Unity DOTS — components are structs implementing IComponentData
public struct Position : IComponentData {
    public float3 Value;
}

public struct Health : IComponentData {
    public int Current;
    public int Max;
}

public struct Inventory : IComponentData {
    public FixedList512Bytes<ItemData> Items;
    public int Capacity;
}
```

**Rules for components:**

1. Components are data only. No methods that contain conditional logic or modify other components.
2. A component may have a simple constructor and a `reset()` initializer — nothing more.
3. No cross-component references inside a component (e.g., `Health` must not hold a reference to `Inventory`). Systems make those connections.
4. Components are value types or simple classes with only public data fields.
5. If a component has a method that branches on its own data (`if current <= 0: mark_dead()`), that logic belongs in a System.
6. Component files are named for the data they represent, not the entity that uses them. `Health.gd`, not `PlayerHealth.gd` — unless the data genuinely only applies to one entity type.

---

### Systems

Where all game logic lives. A system queries the world for all entities with a specific component composition and processes them.

**What a system does:**

- Queries entities by component combination (e.g., "all entities with Position + Velocity")
- Reads component data, computes results, writes results back to components
- Emits events to communicate state changes without coupling to other systems

```
systems/
  physics/
    MovementSystem.gd       # reads Velocity, writes Position
    CollisionSystem.gd      # reads Collider + Position, writes Velocity + RigidBody
    GravitySystem.gd        # reads RigidBody, writes Velocity
  gameplay/
    HealthSystem.gd         # reads Health, emits EntityDied event when current <= 0
    CombatSystem.gd         # reads Equipment + Status, writes Health on hit
    ExperienceSystem.gd     # reads Experience, writes Experience on xp gain, emits LevelUp
    StatusEffectSystem.gd   # reads Status, applies ticks, removes expired effects
  ai/
    AiBrainSystem.gd        # reads AiBrain + Aggression + Position, writes AiInput
    PatrolSystem.gd         # reads AiBrain, writes AiBrain.patrol state
  input/
    PlayerInputSystem.gd    # reads hardware input, writes PlayerInput component
  rendering/
    SpriteSystem.gd         # reads Sprite + Position, updates visual nodes
    BillboardSystem.gd      # reads Billboard, updates UI overlay
```

```csharp
// Unity DOTS — systems extend SystemBase
public partial class MovementSystem : SystemBase {
    protected override void OnUpdate() {
        float dt = SystemAPI.Time.DeltaTime;
        foreach (var (velocity, position) in
                 SystemAPI.Query<RefRO<Velocity>, RefRW<Position>>()) {
            position.ValueRW.Value += velocity.ValueRO.Value * dt;
        }
    }
}
```

**Rules for systems:**

1. Systems query components — they do not hold references to specific entities.
2. Systems do not call other systems directly. Communicate via events (an event bus Autoload or a signal) — never `get_node("/root/OtherSystem").some_method()`.
3. A system is responsible for exactly one behaviour domain. `MovementSystem` moves things. `CombatSystem` handles damage. They do not overlap.
4. Systems do not contain persistent state beyond their operational configuration. Entity state belongs in components.
5. Systems read from components they need and write only to components they own. `MovementSystem` writes `Position` — it does not write `Health`.
6. No rendering or UI code in logic systems. A system that moves entities must not also update a sprite — that is the rendering system's job.
7. System execution order is defined in one place (a system registry, Godot's `_process` ordering, or Unity's `SystemGroup`) — not scattered across system files.

---

### Autoloads / Global Services

Singletons that provide services to all systems without being entities themselves. These are fundamentally different from Systems: they manage global application state, not per-entity game logic.

```
autoloads/
  EventBus.gd          # global signal dispatcher; systems emit and subscribe
  SaveManager.gd       # serialize/deserialize world state to disk
  AudioManager.gd      # play sounds; decoupled from entity/system lifecycle
  SceneManager.gd      # load/unload scenes; manages transitions
  InputManager.gd      # translates raw hardware input into game input events
  WorldRegistry.gd     # entity/component storage in custom ECS; or wraps engine ECS
```

**When to use an Autoload vs a System:**

| Situation | Use |
|---|---|
| Logic that processes entities each frame | System |
| Global event dispatch (fire-and-forget signals) | Autoload (EventBus) |
| Saving/loading game state | Autoload (SaveManager) |
| Managing audio sources | Autoload (AudioManager) |
| Logic that only runs when a specific event occurs | System (listening to events) |
| Application-level scene transitions | Autoload (SceneManager) |

**Rules for Autoloads:**

1. Autoloads provide services — they do not process entities. If an Autoload has a `for entity in entities` loop, it should be a System.
2. Autoloads are stateless where possible. Stateful Autoloads (SaveManager, AudioManager) manage infrastructure, not game logic.
3. Systems call Autoloads (to emit events, play sounds). Autoloads must not call Systems.
4. Every Autoload has a single, named responsibility. No `GameManager` that does everything.

---

## Dependency Direction

```
Components ← Systems (Systems read/write Components)

Systems → EventBus (Autoload) → Systems (listening)
```

Systems never reference each other directly. The only permitted cross-system communication is through events emitted on the EventBus. A System reads component data, does its work, and optionally emits an event. Another System listens for that event and reacts.

```gdscript
# CORRECT: HealthSystem emits an event when an entity dies
# HealthSystem.gd
func _process(delta: float) -> void:
    for entity in world.query([Health]):
        var health: Health = world.get_component(entity, Health)
        if health.current <= 0 and not health.is_dead:
            health.is_dead = true
            EventBus.emit_signal("entity_died", entity)

# CORRECT: LootSystem listens for the event and responds
# LootSystem.gd
func _ready() -> void:
    EventBus.connect("entity_died", _on_entity_died)

func _on_entity_died(entity: int) -> void:
    if world.has_component(entity, Inventory):
        _spawn_loot(world.get_component(entity, Inventory))
```

---

## Forbidden Patterns (G2)

| Pattern | Why it is forbidden |
|---|---|
| Logic inside a Component | Components are data. A `Health` component that calls `die()` when `current <= 0` violates this |
| System calling another System directly | Creates tight coupling; use events instead |
| Scene node (Node2D / MonoBehaviour) owning game state | Scene nodes are rendering/presentation — they must not hold canonical game state |
| Scene node making game decisions | `if health <= 20: flee()` inside a Node2D script is game logic — it belongs in a System |
| Autoload processing entities in a loop | That is a System's job |
| Components referencing other components | Components are independent data bags — Systems make connections |
| `get_parent()` / `get_node()` inside a System | Systems query the component world, not the scene tree |

---

## What "No Business Logic in Outer Layer" Means (G7)

In ECS, the "outer layer" is **scene nodes** — `Node2D` scripts in Godot, `MonoBehaviour` scripts in Unity (classic). Scene nodes are the presentation layer. Their job is to read component state and update visuals.

**Violations — logic in scene nodes:**

```gdscript
# BAD: Node2D script making a game decision
extends Node2D

func _process(delta: float) -> void:
    var health = $HealthBar.value
    if health <= 20:                          # game logic in scene node
        get_tree().change_scene_to_file("res://scenes/game_over.tscn")  # state change
    if Input.is_action_pressed("attack"):     # input handling in scene node
        $AnimationPlayer.play("attack")
        target.health -= 10                   # mutating another entity's component directly
```

```gdscript
# BAD: scene node holding canonical state
extends Node2D

var coins: int = 0                    # game state on a scene node
var inventory: Array = []             # should be a component on the player entity

func collect_coin():
    coins += 1                        # mutating game state in the scene layer
```

**Correct — scene node as pure renderer:**

```gdscript
# GOOD: scene node reads component state, updates visuals only
extends Node2D

var entity_id: int = -1

func _process(_delta: float) -> void:
    if entity_id < 0:
        return
    var pos: Position = World.get_component(entity_id, Position)
    var sprite: Sprite = World.get_component(entity_id, SpriteComponent)
    if pos:
        global_position = Vector2(pos.x, pos.y)
    if sprite:
        $AnimatedSprite2D.animation = sprite.animation
        $AnimatedSprite2D.flip_h = sprite.flip_h
```

**The rule stated precisely:** A scene node may read component data and update its own visual properties. It must not write to any component, make game decisions, or respond to input (input handling belongs in `PlayerInputSystem` which writes to the `PlayerInput` component).

---

## Engine Examples

**Godot (GDScript):**

- Use `Node` or `RefCounted` for component data classes
- Systems are `Node` children of a `SystemsRoot` in the scene tree, processed in order
- `Autoload` scripts are global singletons — use for `EventBus`, `AudioManager`, `SaveManager`
- Scene nodes (Node2D, Control, etc.) are the outer/presentation layer — they must not contain Systems or own game state
- Component storage: use a `WorldRegistry` Autoload that maps `entity_id → Dictionary[ComponentType → Component]`

```gdscript
# EventBus.gd (Autoload)
signal entity_died(entity_id: int)
signal level_up(entity_id: int, new_level: int)
signal item_collected(entity_id: int, item: ItemData)
```

**Unity (C# — two approaches):**

- **Classic Unity (MonoBehaviour):** MonoBehaviours are scene nodes — they are the outer layer. Use plain C# classes (no MonoBehaviour) for Components and Systems. MonoBehaviours only read system output and update transforms/renderers.
- **Unity DOTS:** Purpose-built ECS — `IComponentData` structs for components, `SystemBase` classes for systems, no MonoBehaviour in hot paths.

```csharp
// Classic Unity — System as a plain C# class (not MonoBehaviour)
public class HealthSystem {
    private readonly List<(int entityId, Health health)> _query;
    private readonly EventBus _eventBus;

    public void Update() {
        foreach (var (entityId, health) in _query) {
            if (health.Current <= 0 && !health.IsDead) {
                health.IsDead = true;
                _eventBus.Emit(new EntityDiedEvent(entityId));
            }
        }
    }
}
```

---

## Testing Strategy

**Unit tests — Systems in isolation:**

- Construct a System with a minimal in-memory world containing only the components it needs.
- Add entities with specific component states. Call the System's update method. Assert component values changed as expected.
- No engine, no scene tree, no rendering.

**Integration tests — Systems with real component data:**

- Compose multiple systems that interact (e.g., `PlayerInputSystem` + `MovementSystem`) in a test world.
- Feed synthetic input. Assert final component state is correct.
- Test events: assert that emitting an `entity_died` event causes `LootSystem` to spawn loot components.

**Acceptance tests — Gherkin describing game behaviours:**

- Feature files describe observable game outcomes, not internal state.
- Step definitions construct a world, run systems for N frames or until a condition, and assert on observable outcomes.

```
tests/
  unit/
    systems/
      MovementSystem.test.gd      # or .cs
      HealthSystem.test.gd
      CombatSystem.test.gd
      ExperienceSystem.test.gd
  integration/
    MovementAndCollision.test.gd  # MovementSystem + CollisionSystem together
    CombatAndLoot.test.gd         # CombatSystem + HealthSystem + LootSystem
  acceptance/
    combat.steps.gd
    levelup.steps.gd

specs/
  combat.feature
  leveling.feature
```

**Example unit test (GDScript):**

```gdscript
# tests/unit/systems/HealthSystem.test.gd
func test_entity_dies_when_health_reaches_zero() -> void:
    var world = InMemoryWorld.new()
    var event_bus = MockEventBus.new()
    var system = HealthSystem.new(world, event_bus)

    var entity = world.create_entity()
    world.add_component(entity, Health.new(1, 100))

    world.get_component(entity, Health).current = 0
    system.update(0.016)

    assert_true(world.get_component(entity, Health).is_dead)
    assert_signal_emitted(event_bus, "entity_died", [entity])
```
