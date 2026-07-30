package config

// The model-id -> capability-class table and the class -> profile defaults live
// here so capability.go keeps its resolution logic within the file-size budget.

// builtinModelCapability maps known model ids to a class, so opting in is a
// single driver_model line under either runner. Keys are exact (model ids are
// case-sensitive opaque strings).
//
// This table is a strict SUPERSET and must only ever GROW. It carries three
// key shapes: the bare family aliases that are today's orchestration tierModels
// defaults, the anthropic/... opencode forms, and every previously-pinned id.
// Deleting a retired pin is silent damage: CapabilityClassFor would return
// ok=false, DefaultProfileForModel would disengage the capability tier, and an
// operator whose driver_model still names that pin would have their default
// enforcement profile changed with no error and no warning.
var builtinModelCapability = map[string]string{
	// Family aliases — the built-in claude-runner tier defaults.
	"opus":   CapabilityFrontier,
	"sonnet": CapabilityCapable,
	"haiku":  CapabilityLimited,
	// Provider-qualified opencode ids.
	"anthropic/claude-opus-4-7":   CapabilityFrontier,
	"anthropic/claude-sonnet-4-6": CapabilityCapable,
	"anthropic/claude-haiku-4-5":  CapabilityLimited,
	// Retired pins — retained forever so no operator config loses its class.
	"claude-opus-4-7":           CapabilityFrontier,
	"claude-sonnet-4-6":         CapabilityCapable,
	"claude-haiku-4-5-20251001": CapabilityLimited,
	"claude-haiku-4-5":          CapabilityLimited,
}

// defaultProfileForClass maps a class to its default enforcement profile:
// frontier→outcome, capable→guided, limited→strict.
func defaultProfileForClass(class string) string {
	switch class {
	case CapabilityFrontier:
		return ProfileOutcome
	case CapabilityCapable:
		return ProfileGuided
	default:
		return ProfileStrict
	}
}
