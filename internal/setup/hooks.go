package setup

const (
	cmdPrewrite  = "centinela hook prewrite"
	cmdPostwrite = "centinela hook postwrite"
	cmdAutostart = "centinela hook autostart"
	cmdOrch      = "centinela hook orchestration"
	cmdPlan      = "centinela hook plan-advisor"
	cmdContext   = "centinela hook context"
	cmdSetup     = "centinela hook setup"
	cmdMigrate   = "centinela hook migrate"
	cmdMerge     = "centinela hook merge"
)

func mergeHooks(pre, post, prompt *[]HookGroup) bool {
	c := ensureGroup(pre, "Write", cmdPrewrite, "Validating workflow step...")
	c = ensureGroup(pre, "Edit", cmdPrewrite, "Validating workflow step...") || c
	c = ensureGroup(post, "Write", cmdPostwrite, "") || c
	c = ensureGroup(post, "Edit", cmdPostwrite, "") || c
	c = ensurePrompt(prompt, cmdAutostart, "Detecting new feature intent...") || c
	c = ensurePrompt(prompt, cmdOrch, "Enforcing subagent orchestration...") || c
	c = ensurePrompt(prompt, cmdPlan, "Advising during plan step...") || c
	c = ensurePrompt(prompt, cmdContext, "Checking workflow status...") || c
	c = ensurePrompt(prompt, cmdSetup, "Checking project setup...") || c
	c = ensurePrompt(prompt, cmdMigrate, "Checking managed migrations...") || c
	c = ensurePrompt(prompt, cmdMerge, "Checking pending merges...") || c
	return c
}

// ensureGroup adds a matcher-scoped hook entry if it is not already present.
func ensureGroup(groups *[]HookGroup, matcher, command, statusMsg string) bool {
	if groupHasCommand(*groups, matcher, command) {
		return false
	}
	cmd := HookCmd{Type: "command", Command: command, StatusMessage: statusMsg}
	*groups = append(*groups, HookGroup{Matcher: matcher, Hooks: []HookCmd{cmd}})
	return true
}

// ensurePrompt adds an unmatched (UserPromptSubmit) hook if not already present.
func ensurePrompt(groups *[]HookGroup, command, statusMsg string) bool {
	for _, g := range *groups {
		for _, c := range g.Hooks {
			if c.Command == command {
				return false
			}
		}
	}
	cmd := HookCmd{Type: "command", Command: command, StatusMessage: statusMsg}
	*groups = append(*groups, HookGroup{Hooks: []HookCmd{cmd}})
	return true
}

func groupHasCommand(groups []HookGroup, matcher, command string) bool {
	for _, g := range groups {
		if g.Matcher != matcher {
			continue
		}
		for _, c := range g.Hooks {
			if c.Command == command {
				return true
			}
		}
	}
	return false
}
