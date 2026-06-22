package reconstruct

// maxTargets bounds the reconstructed corpus so the review set stays reviewable.
const maxTargets = 50

// excludeRule drops a package from selection when match holds for its lowercased
// path. Exclusion always wins over promotion. It is data, not control flow.
type excludeRule struct {
	reason string
	match  func(pkg string) bool
}

// excludeRules drops test-only, generated/vendored, and config-leaf packages.
var excludeRules = []excludeRule{
	{"test-only package", func(p string) bool {
		return contains(p, "_test") || contains(p, "/test") || contains(p, "tests/") || contains(p, "spec/")
	}},
	{"generated or vendored package", func(p string) bool {
		return contains(p, ".pb.go") || contains(p, "node_modules") || contains(p, "vendor/") ||
			contains(p, "dist/") || contains(p, "/gen/") || contains(p, "generated")
	}},
	{"config leaf", func(p string) bool {
		return contains(p, "config") || contains(p, "mocks") || contains(p, "fixtures")
	}},
}

// promoteRule promotes a package to a Target when match holds, assigning a Role
// hint and a human-readable reason. Rules are evaluated in order; the first hit
// wins, so more specific surfaces are listed before the generic module fallback.
type promoteRule struct {
	role   Role
	reason string
	match  func(pkg string, s signals) bool
}

// promoteRules is the ordered promotion table. A package owns behavior when it is
// a command/endpoint surface or a consumed (graph in-edge) package; the final
// rule promotes any remaining non-leaf package as a generic module.
var promoteRules = []promoteRule{
	{RoleCommand, "command surface (cmd/ package or CLI framework)", func(p string, s signals) bool {
		return contains(p, "cmd/") || contains(p, "command") || s.hasFramework("cobra") || s.hasDep("cobra")
	}},
	{RoleEndpoint, "endpoint surface (handler/controller/route or HTTP framework)", func(p string, s signals) bool {
		return contains(p, "handler") || contains(p, "controller") || contains(p, "route") ||
			contains(p, "endpoint") || contains(p, "api/") ||
			s.hasFramework("express") || s.hasFramework("fastify") || s.hasFramework("gin")
	}},
	{RoleModule, "consumed surface (a dependency edge points into it)", func(p string, s signals) bool {
		return s.hasIncoming(p)
	}},
	{RoleModule, "behavioral package (service/domain/use-case/core)", func(p string, _ signals) bool {
		return contains(p, "service") || contains(p, "domain") || contains(p, "usecase") ||
			contains(p, "use_case") || contains(p, "core") || contains(p, "internal/")
	}},
}
