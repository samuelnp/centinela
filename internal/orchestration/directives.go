package orchestration

// validateDelegationContract is the paths-only rule for the validate step. The
// verifier's whole value is that it has not seen the implementation
// conversation; summarizing the work into its prompt destroys that and turns
// refutation back into confirmation.
const validateDelegationContract = "gatekeeper is the ADVERSARIAL VERIFIER — spawn it in a FRESH context and pass ONLY the feature slug and file paths; " +
	"do NOT summarize the implementation into the verifier's prompt. It must run `centinela validate` and the test suite itself, " +
	"then `centinela artifact stamp <feature>` as its LAST action."

// DelegationContract returns the extra delegation rule a step imposes on the
// orchestrator, or "" when the step has none.
func DelegationContract(step string) string {
	if step == "validate" {
		return validateDelegationContract
	}
	return ""
}
