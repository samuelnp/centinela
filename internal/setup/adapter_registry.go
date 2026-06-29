package setup

// RegisteredAgents lists the registered single-harness names in order.
func RegisteredAgents() []string {
	out := make([]string, len(orderedAgents))
	copy(out, orderedAgents)
	return out
}

// RegisteredAdapters returns every registered single-harness adapter in order.
func RegisteredAdapters() []HarnessAdapter {
	out := make([]HarnessAdapter, 0, len(orderedAgents))
	for _, n := range orderedAgents {
		out = append(out, registry[n])
	}
	return out
}

// IsValidAgent reports whether agent names a registered harness or composite.
func IsValidAgent(agent string) bool {
	if _, ok := registry[agent]; ok {
		return true
	}
	_, ok := composites[agent]
	return ok
}

// AgentsFor resolves a selector to its ordered single-harness names.
func AgentsFor(agent string) ([]string, error) {
	adapters, err := adaptersFor(agent)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(adapters))
	for i, a := range adapters {
		names[i] = a.Name()
	}
	return names, nil
}

// adaptersFor resolves an agent selector to its ordered adapters.
func adaptersFor(agent string) ([]HarnessAdapter, error) {
	if names, ok := composites[agent]; ok {
		out := make([]HarnessAdapter, 0, len(names))
		for _, n := range names {
			out = append(out, registry[n])
		}
		return out, nil
	}
	a, err := Lookup(agent)
	if err != nil {
		return nil, err
	}
	return []HarnessAdapter{a}, nil
}
