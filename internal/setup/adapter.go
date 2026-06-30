package setup

import (
	"errors"
	"fmt"
	"strings"
)

// Capability is a governance feature a host harness adapter can provide.
type Capability string

const (
	CapBlocksWrites  Capability = "blocks-writes"
	CapPromptContext Capability = "prompt-context"
	CapRulesFile     Capability = "rules-file"
)

// HarnessAdapter describes how Centinela wires one host harness.
type HarnessAdapter interface {
	Name() string
	Capabilities() []Capability
	PlanItems() ([]SyncItem, error)
}

// ErrUnknownAgent is the sentinel wrapped when an agent name is not registered.
var ErrUnknownAgent = errors.New("unknown agent")

// orderedAgents is the registration order used for lookups and error messages.
var orderedAgents = []string{"claude", "opencode", "aider"}

var registry = map[string]HarnessAdapter{
	"claude":   claudeAdapter{},
	"opencode": openCodeAdapter{},
	"aider":    aiderAdapter{},
}

// composites maps a multi-harness selector to its component adapter names.
var composites = map[string][]string{
	"both": {"claude", "opencode"},
}

// Lookup resolves a single agent name to its adapter, or a wrapped
// ErrUnknownAgent listing the registered harnesses.
func Lookup(agent string) (HarnessAdapter, error) {
	if a, ok := registry[agent]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("%w %q (registered: %s)", ErrUnknownAgent, agent, strings.Join(orderedAgents, ", "))
}
