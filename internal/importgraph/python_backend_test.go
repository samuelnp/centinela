package importgraph

import (
	"errors"
	"testing"
)

func TestPythonProvider_Load(t *testing.T) {
	swapOnPath(t, func(n string) bool { return n == "python3" })
	js := `{"module":"m","pkgs":[{"path":"a","imports":["b"]}]}`
	g, err := pythonProvider{run: okRunner(js)}.Load(".")
	if err != nil || g.Module != "m" || len(g.Pkgs) != 1 {
		t.Fatalf("%+v %v", g, err)
	}
}

func TestPythonProvider_ToolMissing(t *testing.T) {
	swapOnPath(t, func(string) bool { return false })
	_, err := pythonProvider{run: okRunner("")}.Load(".")
	var tm *ToolMissingError
	if !errors.As(err, &tm) || tm.Tool != "python3" {
		t.Fatalf("want python3 ToolMissingError, got %v", err)
	}
}

func TestPythonProvider_RunError(t *testing.T) {
	swapOnPath(t, func(string) bool { return true })
	if _, err := (pythonProvider{run: errRunner(errors.New("x"))}).Load("."); err == nil {
		t.Fatal("runner error must surface")
	}
}

func TestPythonProvider_Name(t *testing.T) {
	if (pythonProvider{}).Name() != "python" {
		t.Fatal("name")
	}
}
