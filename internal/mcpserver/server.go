// Package mcpserver registers MCP protocol surfaces and maps results.
// MQ tools are intentionally absent until later stories land.
package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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
func RunStdio(ctx context.Context, server *mcp.Server) error {
	return server.Run(ctx, &mcp.StdioTransport{})
}
