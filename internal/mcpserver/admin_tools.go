package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/output"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const (
	toolDefineQueue = "define_queue"
	toolAlterQueue  = "alter_queue"
	toolDeleteQueue = "delete_queue"
)

type defineQueueInput struct {
	Profile     string `json:"profile" jsonschema:"required"`
	Queue       string `json:"queue" jsonschema:"required"`
	QueueType   string `json:"queueType" jsonschema:"required"`
	MaxDepth    *int   `json:"maxDepth,omitempty"`
	Description string `json:"description,omitempty"`
}

type alterQueueInput struct {
	Profile     string  `json:"profile" jsonschema:"required"`
	Queue       string  `json:"queue" jsonschema:"required"`
	MaxDepth    *int    `json:"maxDepth,omitempty"`
	Description *string `json:"description,omitempty"`
}

type deleteQueueInput struct {
	Profile string `json:"profile" jsonschema:"required"`
	Queue   string `json:"queue" jsonschema:"required"`
}

// RegisterAdminTools wires ADM-001 typed queue administration tools.
func RegisterAdminTools(server *mcp.Server, administrator *application.Administrator) {
	if server == nil || administrator == nil {
		return
	}
	registered := []ToolSpec{
		{
			Name:               toolDefineQueue,
			RequiredCapability: policy.Administer,
			Description:        "Create a queue using typed attributes; destructive create.",
		},
		{
			Name:               toolAlterQueue,
			RequiredCapability: policy.Administer,
			Description:        "Alter supported queue attributes; may change runtime behaviour.",
		},
		{
			Name:               toolDeleteQueue,
			RequiredCapability: policy.Administer,
			Description:        "Delete a queue definition; destructive and irreversible for messages on the queue.",
		},
	}
	RegisteredTools = append(RegisteredTools, registered...)

	mcp.AddTool(server, &mcp.Tool{
		Name: toolDefineQueue,
		Description: destructiveToolDescription(
			"Create a queue with validated LOCAL/ALIAS/REMOTE/MODEL types.",
			policy.Administer,
		),
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptrBool(true), IdempotentHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in defineQueueInput) (
		*mcp.CallToolResult,
		mqadmin.QueueMutationResult,
		error,
	) {
		result, err := administrator.DefineQueue(ctx, in.Profile, in.Queue, mqadmin.DefineQueueRequest{
			QueueType:   mqadmin.QueueType(in.QueueType),
			MaxDepth:    in.MaxDepth,
			Description: in.Description,
		})
		if err != nil {
			return toolError(err), mqadmin.QueueMutationResult{}, nil
		}
		return toolSuccess(output.RenderQueueMutationResult(result)), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolAlterQueue,
		Description: destructiveToolDescription("Alter maxDepth and/or description on an existing queue.", policy.Administer),
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptrBool(true), IdempotentHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in alterQueueInput) (
		*mcp.CallToolResult,
		mqadmin.QueueMutationResult,
		error,
	) {
		result, err := administrator.AlterQueue(ctx, in.Profile, in.Queue, mqadmin.AlterQueueRequest{
			MaxDepth:    in.MaxDepth,
			Description: in.Description,
		})
		if err != nil {
			return toolError(err), mqadmin.QueueMutationResult{}, nil
		}
		return toolSuccess(output.RenderQueueMutationResult(result)), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolDeleteQueue,
		Description: destructiveToolDescription("Delete a queue definition from the queue manager.", policy.Administer),
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptrBool(true), IdempotentHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in deleteQueueInput) (
		*mcp.CallToolResult,
		mqadmin.QueueMutationResult,
		error,
	) {
		result, err := administrator.DeleteQueue(ctx, in.Profile, in.Queue)
		if err != nil {
			return toolError(err), mqadmin.QueueMutationResult{}, nil
		}
		return toolSuccess(output.RenderQueueMutationResult(result)), result, nil
	})
}

func destructiveToolDescription(summary string, required policy.Capability) string {
	return ToolDescription(summary+" Dry-run is not supported for queue mutations.", required)
}

func ptrBool(v bool) *bool { return &v }
