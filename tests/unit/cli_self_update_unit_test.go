package unit_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/selfupdate"
)

// Unit tests for the selfupdate package's exported pure API.

func TestSelfUpdate_AssetNameLinux(t *testing.T) {
	got := selfupdate.AssetName("linux", "amd64", "v0.40.2")
	if got != "centinela-v0.40.2-linux-amd64" {
		t.Fatalf("AssetName = %q", got)
	}
}

func TestSelfUpdate_AssetNameDarwin(t *testing.T) {
	got := selfupdate.AssetName("darwin", "arm64", "v0.40.2")
	if got != "centinela-v0.40.2-darwin-arm64" {
		t.Fatalf("AssetName = %q", got)
	}
}

func TestSelfUpdate_AssetNameWindowsHasExe(t *testing.T) {
	got := selfupdate.AssetName("windows", "amd64", "v0.40.2")
	if !strings.HasSuffix(got, ".exe") {
		t.Fatalf("windows asset should end in .exe, got %q", got)
	}
}

func TestSelfUpdate_KindConstants(t *testing.T) {
	kinds := []selfupdate.Kind{
		selfupdate.KindNetwork,
		selfupdate.KindAPI,
		selfupdate.KindPlatform,
		selfupdate.KindChecksum,
		selfupdate.KindReplace,
		selfupdate.KindPermission,
	}
	seen := map[selfupdate.Kind]bool{}
	for _, k := range kinds {
		if seen[k] {
			t.Fatalf("duplicate Kind value: %q", k)
		}
		seen[k] = true
		if string(k) == "" {
			t.Fatal("empty Kind string")
		}
	}
}

func TestSelfUpdate_ErrorNoWrapped(t *testing.T) {
	e := &selfupdate.Error{Kind: selfupdate.KindChecksum, Msg: "bad"}
	if e.Unwrap() != nil {
		t.Fatal("expected nil Unwrap when no cause")
	}
	if !strings.Contains(e.Error(), "checksum") {
		t.Fatalf("error string missing kind: %q", e.Error())
	}
	if !strings.Contains(e.Error(), "bad") {
		t.Fatalf("error string missing msg: %q", e.Error())
	}
}

func TestSelfUpdate_ErrorWithWrappedCause(t *testing.T) {
	cause := errors.New("root")
	e := &selfupdate.Error{Kind: selfupdate.KindNetwork, Msg: "connect", Err: cause}
	if e.Unwrap() != cause {
		t.Fatalf("Unwrap = %v, want %v", e.Unwrap(), cause)
	}
	if !errors.Is(e, cause) {
		t.Fatal("errors.Is should find wrapped cause")
	}
}

func TestSelfUpdate_NewConstructor(t *testing.T) {
	u := selfupdate.New("1.2.3")
	if u.Version != "1.2.3" {
		t.Fatalf("Version = %q", u.Version)
	}
	if u.HTTP == nil {
		t.Fatal("HTTP must not be nil")
	}
	if u.Target == nil {
		t.Fatal("Target must not be nil")
	}
	if u.Now == nil {
		t.Fatal("Now must not be nil")
	}
}

func TestSelfUpdate_CheckResultFields(t *testing.T) {
	r := selfupdate.CheckResult{Current: "0.37.0", Latest: "v0.40.2", Behind: true, Message: "msg"}
	if r.Current != "0.37.0" || r.Latest != "v0.40.2" || !r.Behind {
		t.Fatalf("CheckResult fields not accessible: %+v", r)
	}
}
