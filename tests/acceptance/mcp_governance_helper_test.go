package acceptance_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var mcpBinOnce sync.Once
var mcpBin string

func buildMcpBin(t *testing.T) string {
	t.Helper()
	mcpBinOnce.Do(func() {
		dir, _ := os.MkdirTemp("", "cent-mcp-bin")
		mcpBin = filepath.Join(dir, "centinela")
		c := exec.Command("go", "build", "-o", mcpBin, "./cmd/centinela")
		c.Dir = repoRoot(t)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("build: %v\n%s", err, out)
		}
	})
	return mcpBin
}

// mcpRepo builds a minimal repo: file_size-only gates and one active workflow.
// When block, an oversized file under internal/ trips G1 (a whole-repo scan).
func mcpRepo(t *testing.T, block bool) string {
	t.Helper()
	dir := t.TempDir()
	w := func(rel, body string) {
		p := filepath.Join(dir, rel)
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w("centinela.toml", "[gates]\nfile_size = true\n")
	w(".workflow/demo.json", `{"feature":"demo","currentStep":"plan","stepOrder":["plan"],"steps":{}}`)
	if block {
		w("internal/big/big.go", "package big\nfunc F() int {\n"+strings.Repeat("\t_ = 0\n", 130)+"\treturn 0\n}\n")
	}
	return dir
}

func connectMcp(t *testing.T, bin, dir string) *sdk.ClientSession {
	t.Helper()
	cmd := exec.Command(bin, "mcp", "serve")
	cmd.Dir = dir
	client := sdk.NewClient(&sdk.Implementation{Name: "harness", Version: "v1"}, nil)
	sess, err := client.Connect(context.Background(), &sdk.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatalf("connect mcp: %v", err)
	}
	return sess
}

// toolText returns the JSON text payload of a tool call with the given args.
func toolText(t *testing.T, sess *sdk.ClientSession, tool string, args map[string]any) string {
	t.Helper()
	res, err := sess.CallTool(context.Background(), &sdk.CallToolParams{
		Name: tool, Arguments: args,
	})
	if err != nil {
		t.Fatalf("call %s: %v", tool, err)
	}
	for _, c := range res.Content {
		if tc, ok := c.(*sdk.TextContent); ok {
			return tc.Text
		}
	}
	t.Fatalf("%s: no text content", tool)
	return ""
}
