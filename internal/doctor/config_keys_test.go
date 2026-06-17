package doctor

import (
	"strings"
	"testing"
)

func TestUnknownConfigKeysNoFile(t *testing.T) {
	repoFixture(t)
	if got := unknownConfigKeys(); got != nil {
		t.Fatalf("missing centinela.toml must yield no findings, got %v", got)
	}
}

func TestUnknownConfigKeysKnownOnly(t *testing.T) {
	repoFixture(t)
	writeFile(t, "centinela.toml", "[verify]\nverify_timeout = 240\n")
	if got := unknownConfigKeys(); len(got) != 0 {
		t.Fatalf("only-known keys must yield nothing, got %v", got)
	}
}

func TestUnknownConfigKeysDetected(t *testing.T) {
	repoFixture(t)
	writeFile(t, "centinela.toml", "bogus_top = 1\n[verify]\nmystery = 2\n")
	got := unknownConfigKeys()
	joined := strings.Join(got, " ")
	if !strings.Contains(joined, "bogus_top") || !strings.Contains(joined, "mystery") {
		t.Fatalf("must name each unknown key, got %v", got)
	}
}

func TestUnknownConfigKeysParseErrorIgnored(t *testing.T) {
	repoFixture(t)
	writeFile(t, "centinela.toml", "[bad\n")
	if got := unknownConfigKeys(); got != nil {
		t.Fatalf("parse error path is handled elsewhere, got %v", got)
	}
}
