package selfupdate

import (
	"os"
	"strings"
	"testing"
)

func TestUpdateHappyPath(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2", asset: []byte("NEW"), withAsset: true, goodSum: true})
	u, bin := newUpdater(t, fs, "0.37.0")
	msg, err := u.Update()
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if msg != "0.37.0 -> 0.40.2" {
		t.Fatalf("msg = %q", msg)
	}
	if got, _ := os.ReadFile(bin); string(got) != "NEW" {
		t.Fatalf("binary not replaced: %q", got)
	}
}

func TestUpdateNoOp(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2", asset: []byte("NEW"), withAsset: true, goodSum: true})
	u, bin := newUpdater(t, fs, "0.40.2")
	msg, err := u.Update()
	if err != nil || msg != "already up to date" {
		t.Fatalf("msg=%q err=%v", msg, err)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD-BINARY" {
		t.Fatalf("binary changed on no-op: %q", got)
	}
}

func TestUpdateDevBuild(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2", asset: []byte("NEW"), withAsset: true, goodSum: true})
	u, bin := newUpdater(t, fs, "dev")
	msg, err := u.Update()
	if err != nil || !strings.Contains(msg, "development build") {
		t.Fatalf("msg=%q err=%v", msg, err)
	}
	if fs.hits != 0 {
		t.Fatalf("dev build hit network %d times", fs.hits)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD-BINARY" {
		t.Fatal("dev build modified binary")
	}
}
