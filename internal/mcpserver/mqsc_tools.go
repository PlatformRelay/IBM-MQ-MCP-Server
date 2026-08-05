package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/output"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const toolExecuteMQSC = "execute_mqsc"

type executeMQSCInput struct {
	Profile string `json:"profile" jsonschema:"required"`
	Command string `json:"command" jsonschema:"required"`
}

// RegisterMQSCTools wires ADM-003 exceptional raw MQSC execution when server opt-in is enabled.
func RegisterMQSCTools(server *mcp.Server, executor *application.MQSCExecutor) {
	if server == nil || executor == nil {
		return
	}
	registered := []ToolSpec{
		{
			Name:               toolExecuteMQSC,
			RequiredCapability: policy.ExecuteMQSC,
			Description:        "Execute one read-only MQSC command (DISPLAY/DIS/PING only); exceptional double opt-in.",
		},
	}
	RegisteredTools = append(RegisteredTools, registered...)

	mcp.AddTool(server, &mcp.Tool{
		Name: toolExecuteMQSC,
		Description: ToolDescription(
			"Execute one read-only MQSC command. v0 allowlist: DISPLAY, DIS, and PING; "+
				"denied verbs are rejected before IBM MQ I/O.",
			policy.ExecuteMQSC,
		),
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptrBool(false), ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in executeMQSCInput) (
		*mcp.CallToolResult,
		mqadmin.RawMQSCResult,
		error,
	) {
		result, err := executor.ExecuteRawMQSC(ctx, in.Profile, in.Command)
		if err != nil {
			return toolError(err), mqadmin.RawMQSCResult{}, nil
		}
		return toolSuccess(output.RenderRawMQSCResult(result)), result, nil
	})
}
