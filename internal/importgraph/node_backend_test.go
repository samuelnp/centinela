package importgraph

import (
	"errors"
	"testing"
)

func TestNodeProvider_Depcruise(t *testing.T) {
	swapOnPath(t, func(n string) bool { return n == "depcruise" })
	g, err := nodeProvider{run: okRunner(depcruiseFixture)}.Load(".")
	if err != nil {
		t.Fatal(err)
	}
	if g.Module != "node" || len(g.Pkgs) != 2 {
		t.Fatalf("got %+v", g)
	}
}

func TestNodeProvider_MadgeFallback(t *testing.T) {
	swapOnPath(t, func(n string) bool { return n == "madge" })
	g, err := nodeProvider{run: okRunner(`{"a.js":["b.js"]}`)}.Load(".")
	if err != nil || len(g.Pkgs) != 1 {
		t.Fatalf("got %+v %v", g, err)
	}
}

func TestNodeProvider_ToolMissing(t *testing.T) {
	swapOnPath(t, func(string) bool { return false })
	_, err := nodeProvider{run: okRunner("")}.Load(".")
	var tm *ToolMissingError
	if !errors.As(err, &tm) {
		t.Fatalf("want ToolMissingError, got %v", err)
	}
}

func TestNodeProvider_RunError(t *testing.T) {
	swapOnPath(t, func(n string) bool { return n == "depcruise" })
	if _, err := (nodeProvider{run: errRunner(errors.New("boom"))}).Load("."); err == nil {
		t.Fatal("runner error must surface")
	}
	swapOnPath(t, func(n string) bool { return n == "madge" })
	if _, err := (nodeProvider{run: errRunner(errors.New("boom"))}).Load("."); err == nil {
		t.Fatal("madge runner error must surface")
	}
}

func TestNodeProvider_Name(t *testing.T) {
	if (nodeProvider{}).Name() != "node" {
		t.Fatal("name")
	}
}
