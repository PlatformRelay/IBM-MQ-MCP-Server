package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const toolBrowseQueueMessages = "browse_queue_messages"

type browseQueueMessagesInput struct {
	Profile         string `json:"profile" jsonschema:"required"`
	Queue           string `json:"queue" jsonschema:"required"`
	Count           int    `json:"count,omitempty"`
	WaitIntervalMs  int    `json:"waitIntervalMs,omitempty"`
	IncludePayload  bool   `json:"includePayload,omitempty"`
	MaxPayloadBytes int    `json:"maxPayloadBytes,omitempty"`
}

// RegisterBrowseTools wires MSG-001 browse tools when a browser is configured.
func RegisterBrowseTools(server *mcp.Server, browser *application.Browser) {
	if server == nil || browser == nil {
		return
	}
	registered := []ToolSpec{
		{
			Name:               toolBrowseQueueMessages,
			RequiredCapability: policy.Browse,
			Description:        "Browse bounded queue messages non-destructively; payloads are opt-in.",
		},
	}
	RegisteredTools = append(RegisteredTools, registered...)

	mcp.AddTool(server, &mcp.Tool{
		Name: toolBrowseQueueMessages,
		Description: ToolDescription(
			"Browse bounded queue messages non-destructively with metadata by default; set includePayload for bodies.",
			policy.Browse,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in browseQueueMessagesInput) (
		*mcp.CallToolResult,
		collection.Page[messaging.MessageRecord],
		error,
	) {
		page, err := browser.BrowseQueueMessages(ctx, in.Profile, in.Queue, messaging.BrowseRequest{
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
