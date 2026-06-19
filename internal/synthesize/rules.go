package synthesize

// rule is one deterministic heuristic: when match holds for a project's signals,
// weight is added to arch's score and reason is recorded. Adding a signal is a
// table edit. Framework/manifest hits weigh more than folder-name hits.
type rule struct {
	arch   Archetype
	weight int
	reason string
	match  func(signals) bool
}

// rules is the ordered scoring table. It is data, not control flow, so the
// inference stays auditable and extensible.
var rules = []rule{
	// Rails-native — framework-opinionated MVC stacks are decisive.
	{RailsNative, 4, "Gemfile detected (Ruby/Rails ecosystem)", func(s signals) bool { return s.hasKind("gem") }},
	{RailsNative, 3, "depends on rails", func(s signals) bool { return s.hasDep("rails") }},
	{RailsNative, 3, "Django/Flask framework", func(s signals) bool { return s.hasDep("django") || s.hasFramework("django") || s.hasDep("flask") }},
	{RailsNative, 2, "app/models + app/controllers layout", func(s signals) bool { return s.hasPkg("app/models") && s.hasPkg("app/controllers") }},
	{RailsNative, 1, "app/views present", func(s signals) bool { return s.hasPkg("app/views") }},

	// ECS — game engines and entity/component/system folders.
	{ECS, 3, "systems/ + components/ folders", func(s signals) bool { return s.hasPkg("systems") && s.hasPkg("components") }},
	{ECS, 2, "entities/ folder", func(s signals) bool { return s.hasPkg("entities") }},
	{ECS, 3, "ECS game-engine dependency", func(s signals) bool {
		return s.hasDep("bevy") || s.hasDep("hecs") || s.hasDep("specs") || s.hasDep("ggez")
	}},

	// Hexagonal — explicit ports/adapters stratification.
	{Hexagonal, 2, "domain/ package", func(s signals) bool { return s.hasPkg("domain") }},
	{Hexagonal, 2, "application/ package", func(s signals) bool { return s.hasPkg("application") }},
	{Hexagonal, 2, "ports or adapters package", func(s signals) bool { return s.hasPkg("ports") || s.hasPkg("adapters") }},
	{Hexagonal, 2, "infrastructure/ package", func(s signals) bool { return s.hasPkg("infrastructure") }},

	// Modular monolith — modules/<x>/public + internal.
	{Modular, 4, "modules/*/public + modules/*/internal layout", func(s signals) bool {
		return s.hasPkg("modules/") && s.hasPkg("/public") && s.hasPkg("/internal")
	}},

	// N-Tier — layered handler/service/repository (also the general default).
	{NTier, 2, "handler or controller layer", func(s signals) bool { return s.hasPkg("handler") || s.hasPkg("controller") }},
	{NTier, 2, "service layer", func(s signals) bool { return s.hasPkg("service") }},
	{NTier, 2, "repository or store layer", func(s signals) bool { return s.hasPkg("repository") || s.hasPkg("store") }},
	{NTier, 2, "Express/Fastify HTTP framework", func(s signals) bool { return s.hasFramework("express") || s.hasFramework("fastify") }},
}
