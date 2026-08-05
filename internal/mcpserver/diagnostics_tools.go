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
	toolExplainMQReasonCode      = "explain_mq_reason_code"
	toolCheckProfileConnectivity = "check_profile_connectivity"
)

type explainMQReasonCodeInput struct {
	ReasonCode int `json:"reasonCode" jsonschema:"required"`
}

// RegisterDiagnosticsTools wires INS-003 offline and connectivity tools.
func RegisterDiagnosticsTools(server *mcp.Server, inspector *application.Inspector) {
	if server == nil {
		return
	}
	registered := []ToolSpec{
		{
			Name:        toolExplainMQReasonCode,
			Description: "Explain an IBM MQ reason code from bundled offline reference data (no MQ I/O).",
		},
		{
			Name:               toolCheckProfileConnectivity,
			RequiredCapability: policy.Inspect,
			Description:        "Verify profile mqweb reachability, identity match, and latency without mutation.",
		},
	}
	RegisteredTools = append(RegisteredTools, registered...)

	mcp.AddTool(server, &mcp.Tool{
		Name: toolExplainMQReasonCode,
		Description: "Explain an IBM MQ reason code (MQRC) using bundled offline reference data. " +
			"Unknown codes return a documented generic answer with an IBM documentation link; no network fetch.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in explainMQReasonCodeInput) (
		*mcp.CallToolResult,
		mqadmin.ReasonExplanation,
		error,
	) {
		explanation := mqadmin.ExplainReasonCode(in.ReasonCode)
		return toolSuccess(output.RenderReasonExplanation(explanation)), explanation, nil
	})

	if inspector == nil {
		return
	}

	mcp.AddTool(server, &mcp.Tool{
		Name: toolCheckProfileConnectivity,
		Description: ToolDescription(
			"Verify mqweb reachability, queue manager identity match, and round-trip latency for a profile.",
			policy.Inspect,
		),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in profileInput) (
		*mcp.CallToolResult,
		mqadmin.ConnectivityReport,
		error,
	) {
		report, err := inspector.CheckProfileConnectivity(ctx, in.Profile)
		if err != nil {
			return toolError(err), mqadmin.ConnectivityReport{}, nil
		}
		return toolSuccess(output.RenderConnectivityReport(report)), report, nil
	})
}
