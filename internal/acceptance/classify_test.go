package acceptance

import "strings"

import "testing"

// oldPredicate is a verbatim copy of the pre-move loop body in
// internal/workflow. The parity table below asserts the extracted per-command
// predicate agrees with it on every branch — the proof that this feature MOVED
// the classification rather than broadening it.
func oldPredicate(cmd string) bool {
	c := strings.ToLower(strings.TrimSpace(cmd))
	if c == "" {
		return false
	}
	if strings.Contains(c, "tests/acceptance") {
		return true
	}
	if strings.Contains(c, "go test") && strings.Contains(c, "./...") {
		return true
	}
	if strings.Contains(c, "cucumber") || strings.Contains(c, "godog") || strings.Contains(c, "behave") {
		return true
	}
	if strings.Contains(c, "acceptance") && strings.Contains(c, "test") {
		return true
	}
	return false
}

func TestIsExecutionCommand_ParityWithThePreviousClassifier(t *testing.T) {
	commands := []string{
		"", "   ", "go vet ./...", "go build ./...", "go test", "go test ./...",
		"go test ./tests/acceptance/...", "GO111MODULE=on go test ./... -race",
		"npx cucumber-js", "go run github.com/cucumber/godog/cmd/godog",
		"behave features/", "npm run test:acceptance", "npm run acceptance",
		"pytest tests/unit", "make test", "  GO TEST ./...  ",
		"./scripts/check-fmt.sh", "COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh",
		"go test ./... -coverprofile=coverage.out",
	}
	for _, cmd := range commands {
		if got, want := IsExecutionCommand(cmd), oldPredicate(cmd); got != want {
			t.Fatalf("IsExecutionCommand(%q) = %v, previous classifier said %v", cmd, got, want)
		}
	}
}

func TestAnyExecutionCommand(t *testing.T) {
	if AnyExecutionCommand(nil) {
		t.Fatal("no commands means no acceptance execution")
	}
	if AnyExecutionCommand([]string{"go vet ./...", "./scripts/check-fmt.sh"}) {
		t.Fatal("neither command is an acceptance execution")
	}
	if !AnyExecutionCommand([]string{"go vet ./...", "go test ./... -coverprofile=c.out"}) {
		t.Fatal("expected the go test ./... command to qualify")
	}
}
