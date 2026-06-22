package reconstruct

// todoMarker is the canonical honest-gap token. It is counted toward TodoCount
// and signals that the scan could not infer concrete behavior — never replaced
// by a fabricated assertion.
const todoMarker = "# TODO: confirm"

// scenarioTemplate is the role-keyed Gherkin scenario skeleton. Each Given/When/
// Then is a TODO marker so no concrete behavior is ever fabricated.
type scenarioTemplate struct {
	name  string // scenario title
	given string // subject of the Given step (after "Given ")
	when  string // subject of the When step (after "When ")
	then  string // subject of the Then step (after "Then ")
}

// scenarioTemplates maps a Role to its scenario skeleton. The zero/unknown role
// falls back to moduleScenario so every target still yields a parseable feature.
var scenarioTemplates = map[Role]scenarioTemplate{
	RoleCommand: {
		name:  "the command performs its primary behavior",
		given: "a precondition for invoking the command",
		when:  "the operator invokes the command",
		then:  "the expected outcome and exit status",
	},
	RoleEndpoint: {
		name:  "the endpoint handles a request",
		given: "a request precondition and authorization state",
		when:  "a client sends the request",
		then:  "the expected response status and payload",
	},
	RoleModule: {
		name:  "the module exposes its primary behavior",
		given: "the relevant input state",
		when:  "the behavior is exercised",
		then:  "the expected result",
	},
}

// templateFor returns the scenario template for a role, defaulting to the module
// template for an unknown/empty role so no target is ever skeleton-less.
func templateFor(role Role) scenarioTemplate {
	if t, ok := scenarioTemplates[role]; ok {
		return t
	}
	return scenarioTemplates[RoleModule]
}

// narrativeFor returns the one-line "As a / I want / So that" subject for a role,
// used in the Feature: narrative.
func narrativeFor(role Role) string {
	switch role {
	case RoleCommand:
		return "exercise the " + string(role) + " surface and confirm its behavior"
	case RoleEndpoint:
		return "exercise the " + string(role) + " request/response contract"
	default:
		return "exercise the module's behavior"
	}
}
