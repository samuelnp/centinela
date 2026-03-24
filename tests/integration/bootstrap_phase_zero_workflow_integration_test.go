package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartGuardEnforcesBootstrapAndExistingBypass(t *testing.T) {
	path := filepath.Join("..", "..", "cmd", "centinela", "start_guard.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read guard file: %v", err)
	}
	content := string(data)
	checks := []string{"projectstage.Existing", "roadmap.HasBootstrapPhase", "roadmap.BootstrapComplete", "workflow.BootstrapStepOrder"}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("start guard missing %q", c)
		}
	}
}
