package main

import (
	"context"
	"errors"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestCov2McpShimSurfacesConnectError drives the connect-failure branch: when
// the MCP session cannot be established, runMcpShim returns the error and never
// reaches the decide/exit path.
func TestCov2McpShimSurfacesConnectError(t *testing.T) {
	old := mcpConnect
	mcpConnect = func(context.Context) (*sdk.ClientSession, error) {
		return nil, errors.New("connect refused")
	}
	t.Cleanup(func() { mcpConnect = old })
	if err := runMcpShim(nil, nil); err == nil || err.Error() != "connect refused" {
		t.Fatalf("expected the connect error to propagate, got %v", err)
	}
}
