package doctor

// checks is the ordered, deterministic registry. Output order follows this
// slice exactly regardless of finding severity (spec: fixed check order).
func checks() []Check {
	return []Check{
		hooksCheck{},
		roadmapCheck{},
		worktreesCheck{},
		workflowStateCheck{},
		evidenceCheck{},
		configCheck{},
		versionCheck{},
	}
}

// Run diagnoses every check in registry order. It is pure: no file is mutated.
func Run(ctx Context) []Diagnosis {
	all := checks()
	out := make([]Diagnosis, 0, len(all))
	for _, c := range all {
		out = append(out, c.Run(ctx))
	}
	return out
}

// Fix applies every safe+idempotent repair (in registry order), then re-runs
// the diagnoses and returns the post-fix report. Every safe repair is attempted
// even if an earlier one fails; a failing Apply marks that check Error with the
// error in Details and the rest still run. Destructive remediations (Apply nil)
// are never executed — they remain reported with their Command.
func Fix(ctx Context) []Diagnosis {
	all := checks()
	failed := map[string]error{}
	for _, c := range all {
		d := c.Run(ctx)
		if d.Repair == nil || !d.Repair.Safe || d.Repair.Apply == nil {
			continue
		}
		if err := d.Repair.Apply(); err != nil {
			failed[c.Name()] = err
		}
	}
	post := Run(ctx)
	for i := range post {
		if err, ok := failed[post[i].Name]; ok {
			post[i].Status = Error
			post[i].Message = "repair failed: " + err.Error()
			post[i].Details = append(post[i].Details, err.Error())
		}
	}
	return post
}

// ExitError reports whether any diagnosis is Error (drives exit code 1). WARN
// never fails the command.
func ExitError(diags []Diagnosis) bool {
	for _, d := range diags {
		if d.Status == Error {
			return true
		}
	}
	return false
}

// Counts tallies diagnoses by status for the summary line.
func Counts(diags []Diagnosis) (ok, warn, err int) {
	for _, d := range diags {
		switch d.Status {
		case OK:
			ok++
		case Warn:
			warn++
		case Error:
			err++
		}
	}
	return
}
