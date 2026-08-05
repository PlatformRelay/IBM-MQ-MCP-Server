package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const toolConsumeQueueMessages = "consume_queue_messages"

type consumeQueueMessagesInput struct {
	Profile         string `json:"profile" jsonschema:"required"`
	Queue           string `json:"queue" jsonschema:"required"`
	Count           int    `json:"count,omitempty"`
	WaitIntervalMs  int    `json:"waitIntervalMs,omitempty"`
	IncludePayload  bool   `json:"includePayload,omitempty"`
	MaxPayloadBytes int    `json:"maxPayloadBytes,omitempty"`
}

// RegisterConsumeTools wires MSG-003 consume tools when a consumer is configured.
func RegisterConsumeTools(server *mcp.Server, consumer *application.Consumer) {
	if server == nil || consumer == nil {
		return
	}
	registered := []ToolSpec{
		{
			Name:               toolConsumeQueueMessages,
			RequiredCapability: policy.Consume,
			Description:        "Destructively get bounded queue messages; payloads are opt-in.",
		},
	}
	RegisteredTools = append(RegisteredTools, registered...)

	mcp.AddTool(server, &mcp.Tool{
		Name: toolConsumeQueueMessages,
		Description: ToolDescription(
			"Destructively get bounded queue messages (mqweb DELETE per message); metadata by default, optional payloads. "+
				"No syncpoint or exactly-once delivery — see docs/messaging/mqweb-semantics.md.",
			policy.Consume,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in consumeQueueMessagesInput) (
		*mcp.CallToolResult,
		collection.Page[messaging.MessageRecord],
		error,
	) {
		page, err := consumer.ConsumeQueueMessages(ctx, in.Profile, in.Queue, messaging.ConsumeRequest{
			Count:           in.Count,
			WaitIntervalMs:  in.WaitIntervalMs,
			IncludePayload:  in.IncludePayload,
			MaxPayloadBytes: in.MaxPayloadBytes,
		})
		if err != nil {
			return toolError(err), collection.Page[messaging.MessageRecord]{}, nil
		}
		return &mcp.CallToolResult{}, page, nil
	})
}
