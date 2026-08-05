package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/output"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const toolPutQueueMessage = "put_queue_message"

type putQueueMessageInput struct {
	Profile       string `json:"profile" jsonschema:"required"`
	Queue         string `json:"queue" jsonschema:"required"`
	ContentType   string `json:"contentType" jsonschema:"required"`
	Payload       string `json:"payload" jsonschema:"required"`
	CorrelationID string `json:"correlationId,omitempty"`
}

// RegisterProduceTools wires MSG-002 produce tools when a producer is configured.
func RegisterProduceTools(server *mcp.Server, producer *application.Producer) {
	if server == nil || producer == nil {
		return
	}
	registered := []ToolSpec{
		{
			Name:               toolPutQueueMessage,
			RequiredCapability: policy.Produce,
			Description:        "Put one validated message; returns identifiers only.",
		},
	}
	RegisteredTools = append(RegisteredTools, registered...)

	mcp.AddTool(server, &mcp.Tool{
		Name: toolPutQueueMessage,
		Description: ToolDescription(
			"Put one validated message with named content types and size limits; "+
				"returns message and correlation identifiers only.",
			policy.Produce,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in putQueueMessageInput) (
		*mcp.CallToolResult,
		messaging.PutResult,
		error,
	) {
		result, err := producer.PutQueueMessage(ctx, in.Profile, in.Queue, messaging.PutRequest{
			ContentType:   in.ContentType,
			Payload:       in.Payload,
			CorrelationID: in.CorrelationID,
		})
		if err != nil {
			return toolError(err), messaging.PutResult{}, nil
		}
		return toolSuccess(output.RenderPutResult(result)), result, nil
	})
}
