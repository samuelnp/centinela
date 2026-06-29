package cost

import "github.com/samuelnp/centinela/internal/telemetry"

// Usage is summed token spend.
type Usage struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

// Tokens is the total billed unit (input + output).
func (u Usage) Tokens() int { return u.Input + u.Output }

func (u *Usage) add(in, out int) { u.Input += in; u.Output += out }

// Aggregate is recorded cost-sample spend folded by scope.
type Aggregate struct {
	Feature map[string]Usage            // feature → total
	Step    map[string]map[string]Usage // feature → step → total
	Model   map[string]Usage            // model id → total
}

// Fold reduces the telemetry log to per-feature, per-step, and per-model spend.
// Only cost-sample events contribute; every other event type is ignored.
func Fold(events []telemetry.Event) Aggregate {
	a := Aggregate{
		Feature: map[string]Usage{},
		Step:    map[string]map[string]Usage{},
		Model:   map[string]Usage{},
	}
	for _, e := range events {
		if e.Type != telemetry.TypeCostSample {
			continue
		}
		f := a.Feature[e.Feature]
		f.add(e.InputTokens, e.OutputTokens)
		a.Feature[e.Feature] = f

		if a.Step[e.Feature] == nil {
			a.Step[e.Feature] = map[string]Usage{}
		}
		s := a.Step[e.Feature][e.Step]
		s.add(e.InputTokens, e.OutputTokens)
		a.Step[e.Feature][e.Step] = s

		if e.Model != "" {
			m := a.Model[e.Model]
			m.add(e.InputTokens, e.OutputTokens)
			a.Model[e.Model] = m
		}
	}
	return a
}
