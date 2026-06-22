package reconstruct

import "testing"

func TestTemplateFor(t *testing.T) {
	if templateFor(RoleCommand).name != "the command performs its primary behavior" {
		t.Fatal("command template wrong")
	}
	if templateFor(RoleEndpoint).name != "the endpoint handles a request" {
		t.Fatal("endpoint template wrong")
	}
	// unknown/empty roles fall back to the module template.
	if templateFor("").name != templateFor(RoleModule).name {
		t.Fatal("empty role must fall back to module template")
	}
	if templateFor("nonsense").name != templateFor(RoleModule).name {
		t.Fatal("unknown role must fall back to module template")
	}
}
