// Package mcpserver registers MCP protocol surfaces and maps results.
package mcpserver

import (
	"fmt"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

// ToolSpec describes one MCP tool and its single required capability (ADR-0003).
type ToolSpec struct {
	Name               string
	RequiredCapability policy.Capability
	Description        string
}

// RegisteredTools lists globally registered tools. Empty until INS-* stories land.
var RegisteredTools []ToolSpec

// ToolDescription prefixes a tool summary with capability requirements.
func ToolDescription(summary string, required policy.Capability) string {
	return fmt.Sprintf(
		"%s Requires profile capability %q; the active profile must grant it or the call is denied before IBM MQ I/O.",
		summary,
		required,
	)
}
