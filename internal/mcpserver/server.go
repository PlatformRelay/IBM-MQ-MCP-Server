// Package mcpserver registers MCP protocol surfaces and maps results.
// MQ tools are intentionally absent until later stories land.
package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/runtime"
)

const (
	// Name is the MCP implementation name advertised to clients.
	Name = "ibm-mq-mcp-server"
	// Version is the pre-release implementation version.
	Version = "0.0.0"
)

// New returns a minimal MCP server with no tools registered.
func New() *mcp.Server {
	return mcp.NewServer(&mcp.Implementation{
		Name:    Name,
		Version: Version,
	}, nil)
}

// RunStdio serves the MCP session over stdin/stdout until the client disconnects.
// When rt is non-nil, transport readiness is reflected for ops probes without MQ contact.
func RunStdio(ctx context.Context, server *mcp.Server, rt *runtime.Runtime) error {
	stop := TrackTransport(rt, "stdio")
	defer stop()
	return server.Run(ctx, &mcp.StdioTransport{})
}

// TrackTransport marks transport readiness for ops probes and returns a stop callback.
func TrackTransport(rt *runtime.Runtime, name string) func() {
	if rt == nil {
		return func() {}
	}
	rt.SetTransportReady(true, name)
	return func() { rt.SetTransportReady(false, "") }
}
