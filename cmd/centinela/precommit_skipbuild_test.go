package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestPrecommitCfg_SkipBuildDisablesBuildGateWithoutMutating(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gates.Build.Enabled = true
	cfg.Precommit.SkipBuild = true

	got := precommitCfg(cfg)
	if got.Gates.Build.Enabled {
		t.Fatal("skip_build must disable the build gate in the returned copy")
	}
	if !cfg.Gates.Build.Enabled {
		t.Fatal("the original cfg must not be mutated by precommitCfg")
	}
	if got == cfg {
		t.Fatal("precommitCfg must return a copy, not the same pointer, when skipping build")
	}
}

func TestPrecommitCfg_KeepsBuildWhenSkipBuildOff(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gates.Build.Enabled = true
	cfg.Precommit.SkipBuild = false

	got := precommitCfg(cfg)
	if !got.Gates.Build.Enabled {
		t.Fatal("the build gate must stay enabled when skip_build is false")
	}
	if got != cfg {
		t.Fatal("precommitCfg must return the original cfg unchanged when not skipping build")
	}
}
