package config

import "testing"

func cfgDriver(id string) *Config {
	c := &Config{}
	c.Orchestration.DriverModel = id
	return c
}

// DriverModelFrom: flag > env > config; trimming; all-empty → ""; nil cfg.
func TestDriverModelFrom(t *testing.T) {
	cases := []struct {
		name, flag, env string
		cfg             *Config
		want            string
	}{
		{"flag wins over env+config", "flag-model", "env-model", cfgDriver("config-model"), "flag-model"},
		{"env wins over config", "", "env-model", cfgDriver("config-model"), "env-model"},
		{"config fallback", "", "", cfgDriver("config-model"), "config-model"},
		{"all empty", "", "", cfgDriver(""), ""},
		{"nil cfg, flag set", "flag-model", "", nil, "flag-model"},
		{"nil cfg, env set", "", "env-model", nil, "env-model"},
		{"nil cfg, all empty", "", "", nil, ""},
		{"flag trimmed", "  flag-model  ", "", cfgDriver(""), "flag-model"},
		{"env trimmed wins", "   ", "  env-model  ", cfgDriver("config-model"), "env-model"},
		{"config trimmed", "", "", cfgDriver("  config-model  "), "config-model"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CENTINELA_MODEL", tc.env)
			if got := DriverModelFrom(tc.flag, tc.cfg); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
