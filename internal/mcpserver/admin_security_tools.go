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
	toolDefineChannel = "define_channel"
	toolAlterChannel  = "alter_channel"
	toolDeleteChannel = "delete_channel"

	toolDefineCHLAUTH = "define_chlauth"
	toolAlterCHLAUTH  = "alter_chlauth"
	toolDeleteCHLAUTH = "delete_chlauth"

	toolDefineAuthrec = "define_authrec"
	toolAlterAuthrec  = "alter_authrec"
	toolDeleteAuthrec = "delete_authrec"
)

type defineChannelInput struct {
	Profile           string `json:"profile" jsonschema:"required"`
	Channel           string `json:"channel" jsonschema:"required"`
	ChannelType       string `json:"channelType" jsonschema:"required"`
	Description       string `json:"description,omitempty"`
	ConnectionName    string `json:"connectionName,omitempty"`
	TransmissionQueue string `json:"transmissionQueue,omitempty"`
}

type alterChannelInput struct {
	Profile           string  `json:"profile" jsonschema:"required"`
	Channel           string  `json:"channel" jsonschema:"required"`
	Description       *string `json:"description,omitempty"`
	ConnectionName    *string `json:"connectionName,omitempty"`
	TransmissionQueue *string `json:"transmissionQueue,omitempty"`
}

type deleteChannelInput struct {
	Profile string `json:"profile" jsonschema:"required"`
	Channel string `json:"channel" jsonschema:"required"`
}

type chlauthTargetInput struct {
	ChannelName string `json:"channelName" jsonschema:"required"`
	RuleType    string `json:"ruleType" jsonschema:"required"`
	Address     string `json:"address,omitempty"`
	ClientUser  string `json:"clientUser,omitempty"`
	SSLPeer     string `json:"sslPeer,omitempty"`
	QMgrName    string `json:"qMgrName,omitempty"`
}

type defineCHLAUTHInput struct {
	Profile    string             `json:"profile" jsonschema:"required"`
	Target     chlauthTargetInput `json:"target" jsonschema:"required"`
	UserSource string             `json:"userSource" jsonschema:"required"`
	MCAUser    string             `json:"mcaUser,omitempty"`
}

type alterCHLAUTHInput struct {
	Profile    string             `json:"profile" jsonschema:"required"`
	Target     chlauthTargetInput `json:"target" jsonschema:"required"`
	UserSource *string            `json:"userSource,omitempty"`
	MCAUser    *string            `json:"mcaUser,omitempty"`
}

type deleteCHLAUTHInput struct {
	Profile string             `json:"profile" jsonschema:"required"`
	Target  chlauthTargetInput `json:"target" jsonschema:"required"`
}

type authrecTargetInput struct {
	Profile    string `json:"profile" jsonschema:"required"`
	ObjectType string `json:"objectType" jsonschema:"required"`
	Entity     string `json:"entity" jsonschema:"required"`
	EntityType string `json:"entityType" jsonschema:"required"`
}

type defineAuthrecInput struct {
	Profile     string             `json:"profile" jsonschema:"required"`
	Target      authrecTargetInput `json:"target" jsonschema:"required"`
	Authorities []string           `json:"authorities" jsonschema:"required"`
}

type alterAuthrecInput struct {
	Profile     string             `json:"profile" jsonschema:"required"`
	Target      authrecTargetInput `json:"target" jsonschema:"required"`
	AddAuths    []string           `json:"addAuths,omitempty"`
	RemoveAuths []string           `json:"removeAuths,omitempty"`
}

type deleteAuthrecInput struct {
	Profile string             `json:"profile" jsonschema:"required"`
	Target  authrecTargetInput `json:"target" jsonschema:"required"`
}

// registerSecurityAdminTools wires ADM-002 channel, CHLAUTH, and authrec administration tools.
func registerSecurityAdminTools(server *mcp.Server, administrator *application.Administrator) {
	registered := []ToolSpec{
		{
			Name: toolDefineChannel, RequiredCapability: policy.Administer,
			Description: "Create a channel with validated types; destructive.",
		},
		{
			Name: toolAlterChannel, RequiredCapability: policy.Administer,
			Description: "Alter supported channel attributes; destructive.",
		},
		{
			Name: toolDeleteChannel, RequiredCapability: policy.Administer,
			Description: "Delete a channel definition; destructive.",
		},
		{
			Name: toolDefineCHLAUTH, RequiredCapability: policy.Administer,
			Description: "Create a channel authentication rule with exact target identity; security-sensitive.",
		},
		{
			Name: toolAlterCHLAUTH, RequiredCapability: policy.Administer,
			Description: "Alter a channel authentication rule; security-sensitive.",
		},
		{
			Name: toolDeleteCHLAUTH, RequiredCapability: policy.Administer,
			Description: "Delete a channel authentication rule; security-sensitive and irreversible.",
		},
		{
			Name: toolDefineAuthrec, RequiredCapability: policy.Administer,
			Description: "Grant authority on an object profile; security-sensitive.",
		},
		{
			Name: toolAlterAuthrec, RequiredCapability: policy.Administer,
			Description: "Add or remove authority grants on an object profile; security-sensitive.",
		},
		{
			Name: toolDeleteAuthrec, RequiredCapability: policy.Administer,
			Description: "Delete an authority record; security-sensitive and irreversible.",
		},
	}
	RegisteredTools = append(RegisteredTools, registered...)

	mcp.AddTool(server, &mcp.Tool{
		Name: toolDefineChannel,
		Description: destructiveToolDescription(
			"Create a channel with validated SDR/SVR/RCVR/RQSTR/CLNTCONN/SVRCONN/CLUSSDR/CLUSRCVR types.",
		),
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptrBool(true), IdempotentHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in defineChannelInput) (
		*mcp.CallToolResult, mqadmin.ChannelMutationResult, error,
	) {
		result, err := administrator.DefineChannel(ctx, in.Profile, in.Channel, mqadmin.DefineChannelRequest{
			ChannelType:       mqadmin.ChannelType(in.ChannelType),
			Description:       in.Description,
			ConnectionName:    in.ConnectionName,
			TransmissionQueue: in.TransmissionQueue,
		})
		if err != nil {
			return toolError(err), mqadmin.ChannelMutationResult{}, nil
		}
		return toolSuccess(output.RenderChannelMutationResult(result)), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolAlterChannel,
		Description: destructiveToolDescription(
			"Alter description, connectionName, and/or transmissionQueue on a channel.",
		),
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptrBool(true), IdempotentHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in alterChannelInput) (
		*mcp.CallToolResult, mqadmin.ChannelMutationResult, error,
	) {
		result, err := administrator.AlterChannel(ctx, in.Profile, in.Channel, mqadmin.AlterChannelRequest{
			Description:       in.Description,
			ConnectionName:    in.ConnectionName,
			TransmissionQueue: in.TransmissionQueue,
		})
		if err != nil {
			return toolError(err), mqadmin.ChannelMutationResult{}, nil
		}
		return toolSuccess(output.RenderChannelMutationResult(result)), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolDeleteChannel,
		Description: destructiveToolDescription("Delete a channel definition from the queue manager."),
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptrBool(true), IdempotentHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in deleteChannelInput) (
		*mcp.CallToolResult, mqadmin.ChannelMutationResult, error,
	) {
		result, err := administrator.DeleteChannel(ctx, in.Profile, in.Channel)
		if err != nil {
			return toolError(err), mqadmin.ChannelMutationResult{}, nil
		}
		return toolSuccess(output.RenderChannelMutationResult(result)), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolDefineCHLAUTH,
		Description: securitySensitiveDescription(
			"Create a channel authentication (CHLAUTH) rule with exact target identity.",
		),
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptrBool(true), IdempotentHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in defineCHLAUTHInput) (
		*mcp.CallToolResult, mqadmin.CHLAUTHMutationResult, error,
	) {
		result, err := administrator.DefineCHLAUTH(ctx, in.Profile, mqadmin.DefineCHLAUTHRequest{
			Target:     toCHLAUTHTarget(in.Target),
			UserSource: mqadmin.CHLAUTHUserSource(in.UserSource),
			MCAUser:    in.MCAUser,
		})
		if err != nil {
			return toolError(err), mqadmin.CHLAUTHMutationResult{}, nil
		}
		return toolSuccess(output.RenderCHLAUTHMutationResult(result)), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolAlterCHLAUTH,
		Description: securitySensitiveDescription(
			"Alter USERSRC and/or MCAUSER on an existing CHLAUTH rule.",
		),
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptrBool(true), IdempotentHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in alterCHLAUTHInput) (
		*mcp.CallToolResult, mqadmin.CHLAUTHMutationResult, error,
	) {
		req := mqadmin.AlterCHLAUTHRequest{Target: toCHLAUTHTarget(in.Target)}
		if in.UserSource != nil {
			src := mqadmin.CHLAUTHUserSource(*in.UserSource)
			req.UserSource = &src
		}
		req.MCAUser = in.MCAUser
		result, err := administrator.AlterCHLAUTH(ctx, in.Profile, req)
		if err != nil {
			return toolError(err), mqadmin.CHLAUTHMutationResult{}, nil
		}
		return toolSuccess(output.RenderCHLAUTHMutationResult(result)), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolDeleteCHLAUTH,
		Description: securitySensitiveDescription(
			"Delete a CHLAUTH rule. Target identity must exactly match channelName, ruleType, and type-specific keys.",
		),
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptrBool(true), IdempotentHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in deleteCHLAUTHInput) (
		*mcp.CallToolResult, mqadmin.CHLAUTHMutationResult, error,
	) {
		result, err := administrator.DeleteCHLAUTH(ctx, in.Profile, toCHLAUTHTarget(in.Target))
		if err != nil {
			return toolError(err), mqadmin.CHLAUTHMutationResult{}, nil
		}
		return toolSuccess(output.RenderCHLAUTHMutationResult(result)), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolDefineAuthrec,
		Description: securitySensitiveDescription(
			"Create an authority record (AUTHREC) with typed grants on an object profile.",
		),
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptrBool(true), IdempotentHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in defineAuthrecInput) (
		*mcp.CallToolResult, mqadmin.AuthrecMutationResult, error,
	) {
		result, err := administrator.DefineAuthrec(ctx, in.Profile, mqadmin.DefineAuthrecRequest{
			Target:      toAuthrecTarget(in.Target),
			Authorities: toAuthrecAuthorities(in.Authorities),
		})
		if err != nil {
			return toolError(err), mqadmin.AuthrecMutationResult{}, nil
		}
		return toolSuccess(output.RenderAuthrecMutationResult(result)), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolAlterAuthrec,
		Description: securitySensitiveDescription(
			"Add or remove authority grants on an existing AUTHREC.",
		),
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptrBool(true), IdempotentHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in alterAuthrecInput) (
		*mcp.CallToolResult, mqadmin.AuthrecMutationResult, error,
	) {
		result, err := administrator.AlterAuthrec(ctx, in.Profile, mqadmin.AlterAuthrecRequest{
			Target:      toAuthrecTarget(in.Target),
			AddAuths:    toAuthrecAuthorities(in.AddAuths),
			RemoveAuths: toAuthrecAuthorities(in.RemoveAuths),
		})
		if err != nil {
			return toolError(err), mqadmin.AuthrecMutationResult{}, nil
		}
		return toolSuccess(output.RenderAuthrecMutationResult(result)), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: toolDeleteAuthrec,
		Description: securitySensitiveDescription(
			"Delete an authority record with exact profile, objectType, entity, and entityType.",
		),
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptrBool(true), IdempotentHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in deleteAuthrecInput) (
		*mcp.CallToolResult, mqadmin.AuthrecMutationResult, error,
	) {
		result, err := administrator.DeleteAuthrec(ctx, in.Profile, toAuthrecTarget(in.Target))
		if err != nil {
			return toolError(err), mqadmin.AuthrecMutationResult{}, nil
		}
		return toolSuccess(output.RenderAuthrecMutationResult(result)), result, nil
	})
}

func securitySensitiveDescription(summary string) string {
	return ToolDescription(
		summary+" Dry-run is not supported. Security-sensitive: verify exact target identity before calling.",
		policy.Administer,
	)
}

func toCHLAUTHTarget(in chlauthTargetInput) mqadmin.CHLAUTHTarget {
	return mqadmin.CHLAUTHTarget{
		ChannelName: in.ChannelName,
		RuleType:    mqadmin.CHLAUTHType(in.RuleType),
		Address:     in.Address,
		ClientUser:  in.ClientUser,
		SSLPeer:     in.SSLPeer,
		QMgrName:    in.QMgrName,
	}
}

func toAuthrecTarget(in authrecTargetInput) mqadmin.AuthrecTarget {
	return mqadmin.AuthrecTarget{
		Profile:    in.Profile,
		ObjectType: mqadmin.AuthrecObjectType(in.ObjectType),
		Entity:     in.Entity,
		EntityType: mqadmin.AuthrecEntityType(in.EntityType),
	}
}

func toAuthrecAuthorities(values []string) []mqadmin.AuthrecAuthority {
	out := make([]mqadmin.AuthrecAuthority, 0, len(values))
	for _, value := range values {
		out = append(out, mqadmin.AuthrecAuthority(value))
	}
	return out
}
