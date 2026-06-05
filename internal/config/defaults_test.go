package config

import "testing"

// applyDefaults is exercised broadly by config.Load tests; these cases lock the
// two branches most relevant to the import-graph feature: the FileSize default
// flip and the import-graph layer normalization wiring.

func TestApplyDefaults_EnablesFileSizeWhenBothGatesOff(t *testing.T) {
	cfg := &Config{}
	cfg.Gates.FileSizeEnabled = false
	cfg.Gates.I18nEnabled = false
	applyDefaults(cfg)
	if !cfg.Gates.FileSizeEnabled {
		t.Fatal("expected FileSize gate to default on when both gates are off")
	}
}

func TestApplyDefaults_LeavesFileSizeOffWhenI18nOn(t *testing.T) {
	cfg := &Config{}
	cfg.Gates.FileSizeEnabled = false
	cfg.Gates.I18nEnabled = true
	applyDefaults(cfg)
	if cfg.Gates.FileSizeEnabled {
		t.Fatal("FileSize must stay off when I18n is explicitly enabled")
	}
}

func TestApplyDefaults_NormalizesImportGraphLayers(t *testing.T) {
	cfg := &Config{}
	cfg.Gates.ImportGraph.Layers = []Layer{
		{Name: " leaf ", Paths: []string{" internal/config/** "}, Allow: []string{" "}},
	}
	applyDefaults(cfg)
	l := cfg.Gates.ImportGraph.Layers[0]
	if l.Name != "leaf" || l.Paths[0] != "internal/config/**" || len(l.Allow) != 0 {
		t.Fatalf("import-graph layers not normalized through applyDefaults: %+v", l)
	}
}
