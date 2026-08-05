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

	toolListChannels      = "list_channels"
	toolGetChannel        = "get_channel"
	toolGetChannelStatus  = "get_channel_status"
	toolListListeners     = "list_listeners"
	toolGetListener       = "get_listener"
	toolGetListenerStatus = "get_listener_status"
	toolListSubscriptions = "list_subscriptions"
	toolGetSubscription   = "get_subscription"
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

type listChannelsInput struct {
	Profile     string `json:"profile" jsonschema:"required"`
	NamePrefix  string `json:"namePrefix,omitempty"`
	ChannelType string `json:"channelType,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Cursor      string `json:"cursor,omitempty"`
}

type channelInput struct {
	Profile string `json:"profile" jsonschema:"required"`
	Channel string `json:"channel" jsonschema:"required"`
}

type listListenersInput struct {
	Profile    string `json:"profile" jsonschema:"required"`
	NamePrefix string `json:"namePrefix,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
}

type listenerInput struct {
	Profile  string `json:"profile" jsonschema:"required"`
	Listener string `json:"listener" jsonschema:"required"`
}

type listSubscriptionsInput struct {
	Profile    string `json:"profile" jsonschema:"required"`
	NamePrefix string `json:"namePrefix,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
}

type subscriptionInput struct {
	Profile      string `json:"profile" jsonschema:"required"`
	Subscription string `json:"subscription" jsonschema:"required"`
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
		{
			Name:               toolListChannels,
			RequiredCapability: policy.Inspect,
			Description:        "List channels with filters, cursor pagination, and truncation metadata.",
		},
		{
			Name:               toolGetChannel,
			RequiredCapability: policy.Inspect,
			Description:        "Get channel definition attributes without runtime status.",
		},
		{
			Name:               toolGetChannelStatus,
			RequiredCapability: policy.Inspect,
			Description:        "Get channel runtime status, distinguishing unavailable from configured definitions.",
		},
		{
			Name:               toolListListeners,
			RequiredCapability: policy.Inspect,
			Description:        "List listeners with filters, cursor pagination, and truncation metadata.",
		},
		{
			Name:               toolGetListener,
			RequiredCapability: policy.Inspect,
			Description:        "Get listener definition attributes without runtime status.",
		},
		{
			Name:               toolGetListenerStatus,
			RequiredCapability: policy.Inspect,
			Description:        "Get listener runtime status, distinguishing unavailable from configured definitions.",
		},
		{
			Name:               toolListSubscriptions,
			RequiredCapability: policy.Inspect,
			Description:        "List subscriptions with filters, cursor pagination, and truncation metadata.",
		},
		{
			Name:               toolGetSubscription,
			RequiredCapability: policy.Inspect,
			Description:        "Get subscription definition by id or name.",
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

	mcp.AddTool(server, &mcp.Tool{
		Name: toolListChannels,
		Description: ToolDescription(
			"List channels with filters, cursor pagination, and truncation metadata.",
			policy.Inspect,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listChannelsInput) (
		*mcp.CallToolResult,
		collection.Page[mqadmin.ChannelSummary],
		error,
	) {
		page, err := inspector.ListChannels(ctx, in.Profile, mqadmin.ListChannelsRequest{
			Filter: mqadmin.ListChannelsFilter{NamePrefix: in.NamePrefix, ChannelType: in.ChannelType},
			Limit:  in.Limit,
			Cursor: in.Cursor,
		})
		if err != nil {
			return toolError(err), collection.Page[mqadmin.ChannelSummary]{}, nil
		}
		return &mcp.CallToolResult{}, page, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolGetChannel,
		Description: ToolDescription(
			"Get channel definition attributes without runtime status.",
			policy.Inspect,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in channelInput) (
		*mcp.CallToolResult,
		mqadmin.ChannelDetail,
		error,
	) {
		detail, err := inspector.GetChannel(ctx, in.Profile, in.Channel)
		if err != nil {
			return toolError(err), mqadmin.ChannelDetail{}, nil
		}
		return &mcp.CallToolResult{}, detail, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolGetChannelStatus,
		Description: ToolDescription(
			"Get channel runtime status, distinguishing unavailable from configured definitions.",
			policy.Inspect,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in channelInput) (
		*mcp.CallToolResult,
		mqadmin.ChannelStatus,
		error,
	) {
		status, err := inspector.GetChannelStatus(ctx, in.Profile, in.Channel)
		if err != nil {
			return toolError(err), mqadmin.ChannelStatus{}, nil
		}
		return &mcp.CallToolResult{}, status, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolListListeners,
		Description: ToolDescription(
			"List listeners with filters, cursor pagination, and truncation metadata.",
			policy.Inspect,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listListenersInput) (
		*mcp.CallToolResult,
		collection.Page[mqadmin.ListenerSummary],
		error,
	) {
		page, err := inspector.ListListeners(ctx, in.Profile, mqadmin.ListListenersRequest{
			Filter: mqadmin.ListListenersFilter{NamePrefix: in.NamePrefix},
			Limit:  in.Limit,
			Cursor: in.Cursor,
		})
		if err != nil {
			return toolError(err), collection.Page[mqadmin.ListenerSummary]{}, nil
		}
		return &mcp.CallToolResult{}, page, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolGetListener,
		Description: ToolDescription(
			"Get listener definition attributes without runtime status.",
			policy.Inspect,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listenerInput) (
		*mcp.CallToolResult,
		mqadmin.ListenerDetail,
		error,
	) {
		detail, err := inspector.GetListener(ctx, in.Profile, in.Listener)
		if err != nil {
			return toolError(err), mqadmin.ListenerDetail{}, nil
		}
		return &mcp.CallToolResult{}, detail, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolGetListenerStatus,
		Description: ToolDescription(
			"Get listener runtime status, distinguishing unavailable from configured definitions.",
			policy.Inspect,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listenerInput) (
		*mcp.CallToolResult,
		mqadmin.ListenerStatus,
		error,
	) {
		status, err := inspector.GetListenerStatus(ctx, in.Profile, in.Listener)
		if err != nil {
			return toolError(err), mqadmin.ListenerStatus{}, nil
		}
		return &mcp.CallToolResult{}, status, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolListSubscriptions,
		Description: ToolDescription(
			"List subscriptions with filters, cursor pagination, and truncation metadata.",
			policy.Inspect,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listSubscriptionsInput) (
		*mcp.CallToolResult,
		collection.Page[mqadmin.SubscriptionSummary],
		error,
	) {
		page, err := inspector.ListSubscriptions(ctx, in.Profile, mqadmin.ListSubscriptionsRequest{
			Filter: mqadmin.ListSubscriptionsFilter{NamePrefix: in.NamePrefix},
			Limit:  in.Limit,
			Cursor: in.Cursor,
		})
		if err != nil {
			return toolError(err), collection.Page[mqadmin.SubscriptionSummary]{}, nil
		}
		return &mcp.CallToolResult{}, page, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolGetSubscription,
		Description: ToolDescription(
			"Get subscription definition by id or name.",
			policy.Inspect,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in subscriptionInput) (
		*mcp.CallToolResult,
		mqadmin.SubscriptionDetail,
		error,
	) {
		detail, err := inspector.GetSubscription(ctx, in.Profile, in.Subscription)
		if err != nil {
			return toolError(err), mqadmin.SubscriptionDetail{}, nil
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
		RegisterBrowseTools(server, application.NewBrowser(inspectorPool(inspector)))
	}
	RegisterDiagnosticsTools(server, inspector)
	return server
}

func inspectorPool(inspector *application.Inspector) *application.ProfilePool {
	if inspector == nil {
		return nil
	}
	return inspector.Pool()
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
