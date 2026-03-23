package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestValidateCommands_IncludeCoverageScript(t *testing.T) {
	orig, _ := os.Getwd()
	defer os.Chdir(orig)                //nolint:errcheck
	os.Chdir(filepath.Join("..", "..")) //nolint:errcheck

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load error: %v", err)
	}
	found := false
	for _, c := range cfg.Validate.Commands {
		if c == "./scripts/check-coverage.sh" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("validate commands should include ./scripts/check-coverage.sh")
	}
}
