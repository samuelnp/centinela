package doctor

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestConfigCheckParseErrorIsError(t *testing.T) {
	repoFixture(t)
	d := configCheck{}.Run(Context{CfgErr: errStub})
	if d.Status != Error || !strings.Contains(d.Message, "could not be parsed") {
		t.Fatalf("parse error must Error, got %v %q", d.Status, d.Message)
	}
}

func TestConfigCheckCleanOK(t *testing.T) {
	repoFixture(t)
	cfg := &config.Config{}
	cfg.Verify.TimeoutSeconds = 240
	d := configCheck{}.Run(Context{Config: cfg})
	if d.Status != OK {
		t.Fatalf("clean config must be OK, got %v %q", d.Status, d.Message)
	}
}

func TestConfigCheckLowTimeoutWarn(t *testing.T) {
	repoFixture(t)
	cfg := &config.Config{}
	cfg.Verify.TimeoutSeconds = 60
	d := configCheck{}.Run(Context{Config: cfg})
	if d.Status != Warn {
		t.Fatalf("low timeout must Warn, got %v", d.Status)
	}
	if !strings.Contains(strings.Join(d.Details, " "), "verify_timeout") {
		t.Fatalf("details must mention verify_timeout: %v", d.Details)
	}
}

func TestConfigCheckMissingGateDir(t *testing.T) {
	repoFixture(t)
	cfg := &config.Config{}
	cfg.Verify.TimeoutSeconds = 240
	cfg.I18n.Dir = "does/not/exist"
	d := configCheck{}.Run(Context{Config: cfg})
	if d.Status != Warn || !strings.Contains(strings.Join(d.Details, " "), "missing directory") {
		t.Fatalf("missing dir must Warn with detail, got %v %v", d.Status, d.Details)
	}
}

func TestMissingGateDirsExistingDirOK(t *testing.T) {
	repoFixture(t)
	cfg := &config.Config{}
	cfg.I18n.Dir = ".workflow" // exists
	if got := missingGateDirs(cfg); len(got) != 0 {
		t.Fatalf("existing dir must not be flagged, got %v", got)
	}
}
