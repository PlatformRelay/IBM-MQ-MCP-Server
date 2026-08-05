package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const (
	toolListProfiles       = "list_profiles"
	toolQueueManagerStatus = "queue_manager_status"
	toolListQueues         = "list_queues"
	toolGetQueue           = "get_queue"
)

type listProfilesInput struct {
	Limit  int    `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}

type profileInput struct {
	Profile string `json:"profile" jsonschema:"required"`
}

type listQueuesInput struct {
	Profile    string `json:"profile" jsonschema:"required"`
	NamePrefix string `json:"namePrefix,omitempty"`
	QueueType  string `json:"queueType,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
}

type getQueueInput struct {
	Profile string `json:"profile" jsonschema:"required"`
	Queue   string `json:"queue" jsonschema:"required"`
}

// RegisterInspectionTools wires INS-001 tools when a profile pool is configured.
func RegisterInspectionTools(server *mcp.Server, inspector *application.Inspector) {
	if server == nil || inspector == nil {
		return
	}
	registered := []ToolSpec{
		{
			Name:               toolListProfiles,
			RequiredCapability: policy.Inspect,
			Description:        "List configured connection profiles with capabilities (no secrets).",
		},
		{
			Name:               toolQueueManagerStatus,
			RequiredCapability: policy.Inspect,
			Description:        "Report queue manager health with configured vs observed identity.",
		},
		{
			Name:               toolListQueues,
			RequiredCapability: policy.Inspect,
			Description:        "List queues on a profile with filters, cursor pagination, and truncation metadata.",
		},
		{
			Name:               toolGetQueue,
			RequiredCapability: policy.Inspect,
			Description:        "Get queue definition and live status including current depth.",
		},
	}
	RegisteredTools = append(RegisteredTools, registered...)

	mcp.AddTool(server, &mcp.Tool{
		Name: toolListProfiles,
		Description: ToolDescription(
			"List configured connection profiles with capabilities (no secrets).",
			policy.Inspect,
		),
	}, func(_ context.Context, _ *mcp.CallToolRequest, in listProfilesInput) (
		*mcp.CallToolResult,
		collection.Page[application.ProfileSummary],
		error,
	) {
		page, err := inspector.ListProfilesPage(in.Limit, in.Cursor)
		if err != nil {
			return toolError(err), collection.Page[application.ProfileSummary]{}, nil
		}
		return &mcp.CallToolResult{}, page, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolQueueManagerStatus,
		Description: ToolDescription(
			"Report queue manager health with configured vs observed identity.",
			policy.Inspect,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in profileInput) (
		*mcp.CallToolResult,
		mqadmin.QueueManagerStatus,
		error,
	) {
		status, err := inspector.QueueManagerStatus(ctx, in.Profile)
		if err != nil {
			return toolError(err), mqadmin.QueueManagerStatus{}, nil
		}
		return &mcp.CallToolResult{}, status, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolListQueues,
		Description: ToolDescription(
			"List queues on a profile with filters, cursor pagination, and truncation metadata.",
			policy.Inspect,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listQueuesInput) (
		*mcp.CallToolResult,
		collection.Page[mqadmin.QueueSummary],
		error,
	) {
		page, err := inspector.ListQueues(ctx, in.Profile, mqadmin.ListQueuesRequest{
			Filter: mqadmin.ListQueuesFilter{NamePrefix: in.NamePrefix, QueueType: in.QueueType},
			Limit:  in.Limit,
			Cursor: in.Cursor,
		})
		if err != nil {
			return toolError(err), collection.Page[mqadmin.QueueSummary]{}, nil
		}
		return &mcp.CallToolResult{}, page, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolGetQueue,
		Description: ToolDescription(
			"Get queue definition and live status including current depth.",
			policy.Inspect,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getQueueInput) (
		*mcp.CallToolResult,
		mqadmin.QueueDetail,
		error,
	) {
		detail, err := inspector.GetQueue(ctx, in.Profile, in.Queue)
		if err != nil {
			return toolError(err), mqadmin.QueueDetail{}, nil
		}
		return &mcp.CallToolResult{}, detail, nil
	})
}

func toolError(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
	}
}

// NewWithInspector returns an MCP server with INS-001 inspection tools registered.
func NewWithInspector(inspector *application.Inspector) *mcp.Server {
	server := New()
	if inspector != nil {
		RegisterInspectionTools(server, inspector)
	}
	return server
}

// ToolCount returns the number of registered tool specifications (tests).
func ToolCount() int {
	return len(RegisteredTools)
}

// ResetRegisteredTools clears tool specs (tests only).
func ResetRegisteredTools() {
	RegisteredTools = nil
}

// RegisteredToolNames returns registered tool names in order (tests).
func RegisteredToolNames() []string {
	names := make([]string, len(RegisteredTools))
	for i, spec := range RegisteredTools {
		names[i] = spec.Name
	}
	return names
}

// MustRegister panics when a tool name duplicates (tests).
func MustRegister(spec ToolSpec) {
	for _, existing := range RegisteredTools {
		if existing.Name == spec.Name {
			panic(fmt.Sprintf("duplicate tool %q", spec.Name))
		}
	}
	RegisteredTools = append(RegisteredTools, spec)
}
