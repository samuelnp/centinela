package cost

// Status is one scope's spend measured against its budget. Budget 0 means "no
// budget for this scope" — Over is then always false (a 0 budget never warns).
type Status struct {
	Scope  string `json:"scope"`  // "feature" | "step" | "model"
	Name   string `json:"name"`   // feature slug, "feature/step", or model id
	Used   int    `json:"used"`   // total tokens
	Budget int    `json:"budget"` // configured budget (0 = unset)
	Over   bool   `json:"over"`   // Used > Budget && Budget > 0
}

// Remaining is Budget-Used, floored at 0; meaningless (0) when Budget is unset.
func (s Status) Remaining() int {
	if s.Budget <= 0 || s.Used >= s.Budget {
		return 0
	}
	return s.Budget - s.Used
}

func status(scope, name string, used, budget int) Status {
	return Status{Scope: scope, Name: name, Used: used, Budget: budget, Over: budget > 0 && used > budget}
}
