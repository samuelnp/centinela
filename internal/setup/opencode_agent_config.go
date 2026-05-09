package setup

import "encoding/json"

var centinelaOpenCodeAgents = map[string]map[string]string{
	"big-thinker": {
		"description": "Centinela planning strategist for problem framing, scope, dependencies, and rollout risks.",
		"prompt":      "You are Centinela big-thinker. Analyze problem, scope, dependencies, constraints, risks, and sequencing. Return actionable outputs and required evidence paths.",
	},
	"feature-specialist": {
		"description": "Centinela feature specialist for acceptance criteria, specs, and edge-case framing.",
		"prompt":      "You are Centinela feature-specialist. Define observable behavior, Gherkin acceptance criteria, UX states, and edge cases. Return concrete spec outputs.",
	},
	"senior-engineer": {
		"description": "Centinela implementation specialist for code changes within architecture rules.",
		"prompt":      "You are Centinela senior-engineer. Implement the smallest correct change, preserve architecture boundaries, and return concrete implementation outputs.",
	},
	"qa-senior": {
		"description": "Centinela QA specialist for tests, regressions, and edge-case reports.",
		"prompt":      "You are Centinela qa-senior. Identify edge cases, add unit/integration/acceptance coverage, and produce the required edge-case report.",
	},
	"documentation-specialist": {
		"description": "Centinela documentation specialist for docs validation and generated project docs.",
		"prompt":      "You are Centinela documentation-specialist. Update user-facing docs when needed, validate docs inputs, and regenerate project documentation outputs.",
	},
	"validation-specialist": {
		"description": "Centinela validation specialist for gatekeeper review, full validation, and readiness checks.",
		"prompt":      "You are Centinela validation-specialist. Run gatekeeper review, centinela validate, readiness checks when enabled, and report concrete validation outputs.",
	},
	"ux-ui-specialist": {
		"description": "Centinela UX/UI specialist for user-facing flows, mobile-first design, and visual states.",
		"prompt":      "You are Centinela ux-ui-specialist. Review user-facing UI for mobile-first flow, accessibility, visual hierarchy, loading, empty, and error states.",
	},
}

func mergeOpenCodeAgents(raw map[string]json.RawMessage) bool {
	agents := map[string]json.RawMessage{}
	_ = json.Unmarshal(raw["agent"], &agents)
	changed := false
	for name, cfg := range centinelaOpenCodeAgents {
		if _, ok := agents[name]; ok {
			continue
		}
		agents[name], _ = json.Marshal(map[string]any{
			"description": cfg["description"],
			"mode":        "subagent",
			"prompt":      cfg["prompt"],
		})
		changed = true
	}
	if mergeBuildTaskPermissions(agents) {
		changed = true
	}
	if !changed {
		return false
	}
	raw["agent"], _ = json.Marshal(agents)
	return true
}

func mergeBuildTaskPermissions(agents map[string]json.RawMessage) bool {
	build := map[string]json.RawMessage{}
	_ = json.Unmarshal(agents["build"], &build)
	changed := false
	if _, ok := agents["build"]; !ok {
		build["mode"], _ = json.Marshal("primary")
		changed = true
	}
	permission := map[string]json.RawMessage{}
	_ = json.Unmarshal(build["permission"], &permission)
	task := map[string]string{}
	_ = json.Unmarshal(permission["task"], &task)
	for name := range centinelaOpenCodeAgents {
		if task[name] == "allow" {
			continue
		}
		task[name] = "allow"
		changed = true
	}
	if !changed {
		return false
	}
	permission["task"], _ = json.Marshal(task)
	build["permission"], _ = json.Marshal(permission)
	agents["build"], _ = json.Marshal(build)
	return true
}
