package synthesize

// layerSlot maps an abstract layer name to a substring matched against package
// paths to derive its concrete path in the target project.
type layerSlot struct{ name, keyword string }

// profile carries the archetype-specific content the synthesizer fills into the
// PROJECT.md Architecture Choice / G2 / G7 / Layer Mapping sections. The text is
// distilled from docs/architecture/architecture-overview.md.
type profile struct {
	pattern, reference, g2, g7 string
	layers                     []layerSlot
}

var profiles = map[Archetype]profile{
	NTier: {
		pattern:   "N-Tier / Layered",
		reference: "docs/architecture/n-tier.md",
		g2:        "Handler/Controller may call Service; Service may call Repository; Repository may not call up. No skipping layers.",
		g7:        "HTTP handlers/controllers are the outer layer — request parsing and response shaping only, no business logic.",
		layers:    []layerSlot{{"Handler / Controller", "handler"}, {"Service", "service"}, {"Repository", "repository"}, {"App / Routing", "cmd"}},
	},
	Hexagonal: {
		pattern:   "Hexagonal (Ports and Adapters)",
		reference: "docs/architecture/hexagonal.md",
		g2:        "Domain imports nothing outward; Application depends on Domain + Ports; Infrastructure adapters implement Ports. Dependencies point inward.",
		g7:        "UI components and infrastructure adapters are the outer layer — they wire ports, never hold domain logic.",
		layers:    []layerSlot{{"Domain", "domain"}, {"Application / Use Cases", "application"}, {"Ports", "ports"}, {"Infrastructure / Adapters", "infrastructure"}},
	},
	RailsNative: {
		pattern:   "Rails-native (MVC + Fat Model)",
		reference: "docs/architecture/rails-native.md",
		g2:        "Controllers orchestrate; Models hold business logic (Active Record couples logic to persistence by design); Views render only.",
		g7:        "Views and route handlers are the outer layer — presentation only, no business logic.",
		layers:    []layerSlot{{"Model", "app/models"}, {"Controller", "app/controllers"}, {"View", "app/views"}, {"Service", "app/services"}},
	},
	ECS: {
		pattern:   "ECS (Entity-Component-System)",
		reference: "docs/architecture/ecs.md",
		g2:        "Components are pure data; Systems hold all logic and operate over components; Entities compose components. Components must not contain logic.",
		g7:        "Scene nodes / components are the outer layer — pure data, no behavior.",
		layers:    []layerSlot{{"Entities", "entities"}, {"Components", "components"}, {"Systems", "systems"}, {"Autoloads", "autoload"}},
	},
	Modular: {
		pattern:   "Modular Monolith",
		reference: "docs/architecture/modular.md",
		g2:        "A module's internals are private; modules communicate only through published public APIs. No reaching into another module's internal/.",
		g7:        "A module's internal files are the outer layer relative to its public API — not exposed across module boundaries.",
		layers:    []layerSlot{{"Module Public API", "/public"}, {"Module Internal", "/internal"}},
	},
	Custom: {
		pattern:   "Custom (confirm and define manually)",
		reference: "docs/architecture/architecture-overview.md",
		g2:        "TODO: define the layer-boundary rule for this project.",
		g7:        "TODO: define the outer layer for this project.",
		layers:    []layerSlot{{"Outer layer", "cmd"}, {"Core logic", "internal"}},
	},
}
