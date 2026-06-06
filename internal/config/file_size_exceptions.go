package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

const FileSizeExceptionCap = 130

type FileSizeException struct {
	Path     string `toml:"path"`
	Kind     string `toml:"kind"`
	Reason   string `toml:"reason"`
	MaxLines int    `toml:"max_lines"`
}

func validateConfig(cfg *Config) error {
	for i, ex := range cfg.Gates.FileSizeExceptions {
		if strings.TrimSpace(ex.Path) == "" {
			return fmt.Errorf("gates.file_size_exceptions[%d].path is required", i)
		}
		normalized := filepath.ToSlash(filepath.Clean(ex.Path))
		cfg.Gates.FileSizeExceptions[i].Path = strings.ReplaceAll(normalized, "\\", "/")
		if ex.Kind != "configuration" && ex.Kind != "domain_atomic" {
			return fmt.Errorf("gates.file_size_exceptions[%d].kind must be configuration or domain_atomic", i)
		}
		if strings.TrimSpace(ex.Reason) == "" {
			return fmt.Errorf("gates.file_size_exceptions[%d].reason is required", i)
		}
		if ex.MaxLines <= 100 || ex.MaxLines > FileSizeExceptionCap {
			return fmt.Errorf("gates.file_size_exceptions[%d].max_lines must be 101..%d", i, FileSizeExceptionCap)
		}
	}
	if err := validateOrchestrationModels(cfg); err != nil {
		return err
	}
	if err := validateOrchestrationModelMap(cfg); err != nil {
		return err
	}
	if err := validateSecurityGate(cfg.Gates.Security); err != nil {
		return err
	}
	return nil
}
