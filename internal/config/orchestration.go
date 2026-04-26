package config

import (
	"path/filepath"
	"strings"
)

type OrchestrationConfig struct {
	UIPaths []string `toml:"ui_paths"`
}

var defaultUIPaths = []string{"src/ui", "src/components", "app/views", "web", "styles", "internal/ui"}

func UIPaths(cfg *Config) []string {
	paths := defaultUIPaths
	if cfg != nil && len(cfg.Orchestration.UIPaths) > 0 {
		paths = cfg.Orchestration.UIPaths
	}
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		clean := strings.TrimSpace(strings.TrimPrefix(path, "./"))
		if clean != "" {
			out = append(out, filepath.ToSlash(filepath.Clean(clean)))
		}
	}
	return out
}
