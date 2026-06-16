package gates

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func cg(name, sev, out string) config.CustomGate {
	return config.CustomGate{Name: name, Severity: sev, Output: out, TimeoutSeconds: 60}
}

// TestCustomResultPass: exit 0 => Pass with no Details.
func TestCustomResultPass(t *testing.T) {
	r := customResult(cg("p", "fail", "blob"), "", 0, false)
	if r.Status != Pass || len(r.Details) != 0 {
		t.Fatalf("want Pass/no details, got %+v", r)
	}
	if r.Message != "p passed" {
		t.Fatalf("message = %q", r.Message)
	}
}

// TestCustomResultFailAndWarn: non-zero maps to Fail (severity fail) or Warn.
func TestCustomResultFailAndWarn(t *testing.T) {
	f := customResult(cg("f", "fail", "blob"), "boom", 1, false)
	if f.Status != Fail || f.Message != "f failed (exit 1)" {
		t.Fatalf("fail mapping wrong: %+v", f)
	}
	w := customResult(cg("w", "warn", "blob"), "nit", 2, false)
	if w.Status != Warn {
		t.Fatalf("warn mapping wrong: %+v", w)
	}
}

// TestCustomResultTimeout: a timed-out command Fails with a timeout message
// regardless of severity.
func TestCustomResultTimeout(t *testing.T) {
	r := customResult(cg("h", "warn", "blob"), "", -1, true)
	if r.Status != Fail {
		t.Fatalf("timeout must Fail even at warn: %+v", r)
	}
	if !strings.Contains(r.Message, "timed out after 60s") || len(r.Details) != 1 {
		t.Fatalf("timeout detail wrong: %+v", r)
	}
}

// TestCustomResultEmptyOutput: an empty-output Fail falls back to a generic detail.
func TestCustomResultEmptyOutput(t *testing.T) {
	r := customResult(cg("e", "fail", "blob"), "   ", 3, false)
	if len(r.Details) != 1 || !strings.Contains(r.Details[0], "with no output") {
		t.Fatalf("generic fallback missing: %+v", r.Details)
	}
}

// TestCustomResultLinesSplit: output=lines yields one Detail per non-empty line.
func TestCustomResultLinesSplit(t *testing.T) {
	r := customResult(cg("l", "fail", "lines"), "a.go:1\n\nb.go:2\nc.go:3", 1, false)
	if len(r.Details) != 3 {
		t.Fatalf("want 3 line details, got %d: %+v", len(r.Details), r.Details)
	}
	if r.Details[0] != "a.go:1" || r.Details[2] != "c.go:3" {
		t.Fatalf("lines order/content wrong: %+v", r.Details)
	}
}

// TestCustomResultLinesBounded: more than customLineCap lines are bounded with a
// "… (N more)" overflow marker.
func TestCustomResultLinesBounded(t *testing.T) {
	var b strings.Builder
	for i := 0; i < customLineCap+5; i++ {
		b.WriteString("v\n")
	}
	r := customResult(cg("l", "fail", "lines"), b.String(), 1, false)
	if len(r.Details) != customLineCap+1 {
		t.Fatalf("want %d details, got %d", customLineCap+1, len(r.Details))
	}
	if !strings.Contains(r.Details[customLineCap], "5 more") {
		t.Fatalf("overflow marker wrong: %q", r.Details[customLineCap])
	}
}

// TestCustomResultBlobTruncates: a >4 KiB blob is capped with the truncation marker.
func TestCustomResultBlobTruncates(t *testing.T) {
	big := strings.Repeat("x", customBlobCap+500)
	r := customResult(cg("b", "fail", "blob"), big, 1, false)
	if len(r.Details) != 1 {
		t.Fatalf("blob should be one detail, got %d", len(r.Details))
	}
	if !strings.HasSuffix(r.Details[0], customTruncMsg) {
		t.Fatalf("truncation marker missing")
	}
	if len(r.Details[0]) != customBlobCap+len(customTruncMsg) {
		t.Fatalf("blob not capped: len=%d", len(r.Details[0]))
	}
}
