package analyze

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWalk_SkipSetExcludesDepsAndBuild(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "main.go"), "package main")
	for _, d := range []string{"vendor", "node_modules", ".git", ".workflow", "dist", "build"} {
		mkFile(t, filepath.Join(root, d, "junk.go"), "package x")
	}
	res, err := walk(root)
	if err != nil {
		t.Fatal(err)
	}
	if res.extCounts[".go"] != 1 {
		t.Fatalf("skip set must exclude dependency/build dirs, got %d .go files", res.extCounts[".go"])
	}
	for _, p := range res.packages {
		if p == "vendor" || p == "node_modules" || p == "build" {
			t.Fatalf("skip-set dir leaked into packages: %v", res.packages)
		}
	}
}

func TestWalk_DepthBoundedLayout(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "a", "b", "c", "deep.go"), "package c")
	res, _ := walk(root)
	for _, p := range res.packages {
		if p == "a/b/c" {
			t.Fatalf("layout must be depth-bounded (maxLayoutDepth=2), got %v", res.packages)
		}
	}
	if len(res.packages) == 0 {
		t.Fatal("shallow dirs must still be listed")
	}
}

func TestWalk_GitignoredPathExcluded(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "keep.go"), "package main")
	mkFile(t, filepath.Join(root, "generated", "gen.go"), "package gen")
	mkFile(t, filepath.Join(root, ".gitignore"), "generated/\n")
	res, _ := walk(root)
	if res.extCounts[".go"] != 1 {
		t.Fatalf("gitignored path must be excluded, got %d", res.extCounts[".go"])
	}
}

func TestWalk_SymlinkFileSkipped(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks unreliable on windows")
	}
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "real.go"), "package main")
	if err := os.Symlink(filepath.Join(root, "real.go"), filepath.Join(root, "link.go")); err != nil {
		t.Skip("symlink unsupported: " + err.Error())
	}
	res, _ := walk(root)
	if res.extCounts[".go"] != 1 {
		t.Fatalf("symlinked file must be skipped, counted %d", res.extCounts[".go"])
	}
}

func TestWalk_UnreadableRootIsHardError(t *testing.T) {
	if _, err := walk(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Fatal("unreadable root must be the sole hard error")
	}
}
