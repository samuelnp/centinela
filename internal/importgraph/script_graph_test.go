package importgraph

import (
	"errors"
	"strings"
	"testing"
)

func TestScriptProvider_Load(t *testing.T) {
	js := `{"module":"s","pkgs":[{"path":"x","imports":["y"]}]}`
	g, err := scriptProvider{cmd: []string{"emit"}, run: okRunner(js)}.Load(".")
	if err != nil || g.Module != "s" || g.Pkgs[0].Path != "x" {
		t.Fatalf("%+v %v", g, err)
	}
}

func TestScriptProvider_RunError(t *testing.T) {
	_, err := scriptProvider{cmd: []string{"x"}, run: errRunner(errors.New("nonzero"))}.Load(".")
	if err == nil {
		t.Fatal("nonzero exit must surface as an error")
	}
}

func TestScriptProvider_BadJSON(t *testing.T) {
	_, err := scriptProvider{cmd: []string{"x"}, run: okRunner("nope")}.Load(".")
	if err == nil {
		t.Fatal("malformed output must error, never a silent empty graph")
	}
}

func TestScriptProvider_Name(t *testing.T) {
	if (scriptProvider{}).Name() != "script" {
		t.Fatal("name")
	}
}

func TestDecodeGraphJSON_EmptyValid(t *testing.T) {
	g, err := decodeGraphJSON([]byte(`{"module":"m"}`))
	if err != nil || g.Module != "m" || len(g.Pkgs) != 0 {
		t.Fatalf("valid empty graph: %+v %v", g, err)
	}
}

func TestToolMissingError_Message(t *testing.T) {
	if e := (&ToolMissingError{Tool: "depcruise"}); !strings.Contains(e.Error(), "depcruise") {
		t.Fatalf("message must name the tool: %q", e.Error())
	}
}
