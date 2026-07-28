package evidence

import "fmt"

// gatekeeperBody renders the validate-step verifier report stub. The
// `**Status:**` line is parsed by the complete gate, so the literal
// SAFE | WARNING | CRITICAL vocabulary is preserved verbatim. "Analyzed Specs"
// stays a section of its own because it is mechanically pre-filled from
// specs/*.feature and must stay deterministic (see analyzedSpecsList);
// "Inputs Read" is the verifier's own account of what it opened.
//
// The stub ships **Status:** CRITICAL and an EMPTY commands array on purpose:
// a freshly scaffolded stub must FAIL the gate until a verifier fills it and
// runs `centinela artifact stamp`. Fail-closed is the whole point — a stub
// that passed would reintroduce the dead-subagent hole. CRITICAL rather than
// SAFE because the delivery composer reads the RAW verdict with no Assess, so
// a SAFE stub would surface there as a passing verdict.
func gatekeeperBody(feature string) []byte {
	return []byte(fmt.Sprintf(`### Adversarial Verifier Report: %s
**Date:** %s
**Status:** CRITICAL

#### Inputs Read
- %s

#### Analyzed Specs
%s

#### Refutation Attempts
- %s

#### Commands Run
- %s

#### Findings
- %s

#### Recommendation
- %s

`+"```json centinela:verification"+`
{
  "revision": "",
  "treeDigest": "",
  "commands": []
}
`+"```"+`
`, feature, today(),
		FillSlot("every path you actually read; flag any narrative summary you were handed"),
		analyzedSpecsList(),
		FillSlot("claim attacked / how / what happened"),
		FillSlot("argv, exit code, duration — then run `centinela artifact stamp "+feature+"`"),
		FillSlot("affected spec / scenario / risk / suggestion; empty if SAFE"),
		FillSlot("SAFE/WARNING/CRITICAL + one-line rationale")))
}
