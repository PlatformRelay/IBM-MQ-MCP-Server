package mcpserver

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func toolSuccess(text string) *mcp.CallToolResult {
	if text == "" {
		return &mcp.CallToolResult{}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
