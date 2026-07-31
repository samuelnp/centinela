package acceptance

import "testing"

// The scope split must not change WHICH commands are acceptance executions —
// only how much of their output is attributed to the acceptance tier.
func TestScopeOf_ClassifiesWithoutChangingTheExecutionPredicate(t *testing.T) {
	cases := map[string]Scope{
		"":                                    ScopeNone,
		"   ":                                 ScopeNone,
		"go vet ./...":                        ScopeNone,
		"./scripts/check-fmt.sh":              ScopeNone,
		"go test ./internal/...":              ScopeNone,
		"go test ./...":                       ScopeMixed,
		"go test -v ./...":                    ScopeMixed,
		"go test -json ./... -coverprofile=c": ScopeMixed,
		"go test ./tests/acceptance/...":      ScopeAcceptance,
		"go test -v ./... # tests/acceptance": ScopeAcceptance,
		"npx cucumber-js":                     ScopeAcceptance,
		"go run .../godog/cmd/godog":          ScopeAcceptance,
		"behave features/":                    ScopeAcceptance,
		"npm run test:acceptance":             ScopeAcceptance,
	}
	for cmd, want := range cases {
		if got := ScopeOf(cmd); got != want {
			t.Fatalf("ScopeOf(%q) = %v, want %v", cmd, got, want)
		}
		if got, want := IsExecutionCommand(cmd), want != ScopeNone; got != want {
			t.Fatalf("IsExecutionCommand(%q) = %v, want %v", cmd, got, want)
		}
	}
}

func TestIsAcceptancePackage(t *testing.T) {
	yes := []string{
		"tests/acceptance",
		"tests/acceptance/sub",
		"github.com/x/y/tests/acceptance",
		"github.com/x/y/tests/acceptance/deep",
		"x\\y\\tests\\acceptance",
	}
	for _, p := range yes {
		if !isAcceptancePackage(p) {
			t.Fatalf("%q must be an acceptance package", p)
		}
	}
	no := []string{"", "  ", "github.com/x/y/internal/config", "tests/acceptancex", "x/tests"}
	for _, p := range no {
		if isAcceptancePackage(p) {
			t.Fatalf("%q must NOT be an acceptance package", p)
		}
	}
}

// ScopeMixed is the only scope that filters; the others count everything.
func TestScope_Counts(t *testing.T) {
	if !ScopeAcceptance.counts("github.com/x/y/unitpkg") {
		t.Fatal("an acceptance-scoped command's own packages all count")
	}
	if ScopeMixed.counts("github.com/x/y/unitpkg") {
		t.Fatal("a whole-repo command must not count a unit package")
	}
	if !ScopeMixed.counts("github.com/x/y/tests/acceptance") {
		t.Fatal("a whole-repo command must count the acceptance package")
	}
	if ScopeMixed.counts("") {
		t.Fatal("an unattributable result must never be blamed on the acceptance tier")
	}
}
