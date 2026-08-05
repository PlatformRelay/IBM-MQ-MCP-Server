package mqweb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

const mqscActionAPIVersion = "v2"

func (c *adminClient) DefineCHLAUTH(
	ctx context.Context,
	req mqadmin.DefineCHLAUTHRequest,
) (mqadmin.CHLAUTHMutationResult, error) {
	return c.runCHLAUTHMutation(ctx, mqadmin.MutationDefine, req.Target, func(params map[string]string) {
		params["USERSRC"] = string(req.UserSource)
		if mca := strings.TrimSpace(req.MCAUser); mca != "" {
			params["MCAUSER"] = mca
		}
		params["ACTION"] = "ADD"
	})
}

func (c *adminClient) AlterCHLAUTH(
	ctx context.Context,
	req mqadmin.AlterCHLAUTHRequest,
) (mqadmin.CHLAUTHMutationResult, error) {
	return c.runCHLAUTHMutation(ctx, mqadmin.MutationAlter, req.Target, func(params map[string]string) {
		if req.UserSource != nil {
			params["USERSRC"] = string(*req.UserSource)
		}
		if req.MCAUser != nil {
			params["MCAUSER"] = *req.MCAUser
		}
		params["ACTION"] = "REPLACE"
	})
}

func (c *adminClient) DeleteCHLAUTH(
	ctx context.Context,
	target mqadmin.CHLAUTHTarget,
) (mqadmin.CHLAUTHMutationResult, error) {
	body, err := encodeMQSCCommand("delete", "chlauth", target.ChannelName, chlauthIdentityParams(target))
	if err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	if err := c.executeMQSC(ctx, body); err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	snapshot := mqadmin.CHLAUTHSnapshotFromTarget(target)
	return mqadmin.CHLAUTHMutationResult{
		Operation:    mqadmin.MutationDelete,
		Profile:      c.base.name,
		QueueManager: c.queueManager,
		Target:       snapshot,
		Before:       &snapshot,
		CompletedAt:  time.Now().UTC(),
	}, nil
}

func (c *adminClient) DefineAuthrec(
	ctx context.Context,
	req mqadmin.DefineAuthrecRequest,
) (mqadmin.AuthrecMutationResult, error) {
	params := authrecIdentityParams(req.Target)
	params["AUTHADD"] = joinAuthrecAuthorities(req.Authorities)
	return c.runAuthrecMutation(ctx, mqadmin.MutationDefine, req.Target, "set", params, nil)
}

func (c *adminClient) AlterAuthrec(
	ctx context.Context,
	req mqadmin.AlterAuthrecRequest,
) (mqadmin.AuthrecMutationResult, error) {
	params := authrecIdentityParams(req.Target)
	if len(req.AddAuths) > 0 {
		params["AUTHADD"] = joinAuthrecAuthorities(req.AddAuths)
	}
	if len(req.RemoveAuths) > 0 {
		params["AUTHRMV"] = joinAuthrecAuthorities(req.RemoveAuths)
	}
	return c.runAuthrecMutation(ctx, mqadmin.MutationAlter, req.Target, "set", params, nil)
}

func (c *adminClient) DeleteAuthrec(
	ctx context.Context,
	target mqadmin.AuthrecTarget,
) (mqadmin.AuthrecMutationResult, error) {
	params := authrecIdentityParams(target)
	body, err := encodeMQSCCommand("delete", "authrec", "", params)
	if err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	if err := c.executeMQSC(ctx, body); err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	snapshot := mqadmin.AuthrecSnapshotFromTarget(target)
	return mqadmin.AuthrecMutationResult{
		Operation:    mqadmin.MutationDelete,
		Profile:      c.base.name,
		QueueManager: c.queueManager,
		Target:       snapshot,
		Before:       &snapshot,
		CompletedAt:  time.Now().UTC(),
	}, nil
}

func (c *adminClient) runCHLAUTHMutation(
	ctx context.Context,
	op mqadmin.MutationOperation,
	target mqadmin.CHLAUTHTarget,
	extend func(map[string]string),
) (mqadmin.CHLAUTHMutationResult, error) {
	params := chlauthIdentityParams(target)
	extend(params)
	body, err := encodeMQSCCommand("set", "chlauth", target.ChannelName, params)
	if err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	if err := c.executeMQSC(ctx, body); err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	snapshot := mqadmin.CHLAUTHSnapshotFromTarget(target)
	if op == mqadmin.MutationDefine {
		snapshot.UserSource = params["USERSRC"]
		snapshot.MCAUser = params["MCAUSER"]
	}
	return mqadmin.CHLAUTHMutationResult{
		Operation:    op,
		Profile:      c.base.name,
		QueueManager: c.queueManager,
		Target:       snapshot,
		After:        &snapshot,
		CompletedAt:  time.Now().UTC(),
	}, nil
}

func (c *adminClient) runAuthrecMutation(
	ctx context.Context,
	op mqadmin.MutationOperation,
	target mqadmin.AuthrecTarget,
	command string,
	params map[string]string,
	before *mqadmin.AuthrecSnapshot,
) (mqadmin.AuthrecMutationResult, error) {
	body, err := encodeMQSCCommand(command, "authrec", "", params)
	if err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	if err := c.executeMQSC(ctx, body); err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	snapshot := mqadmin.AuthrecSnapshotFromTarget(target)
	if add := params["AUTHADD"]; add != "" {
		snapshot.Authorities = strings.Split(add, ",")
	}
	return mqadmin.AuthrecMutationResult{
		Operation:    op,
		Profile:      c.base.name,
		QueueManager: c.queueManager,
		Target:       snapshot,
		Before:       before,
		After:        &snapshot,
		CompletedAt:  time.Now().UTC(),
	}, nil
}

func (c *adminClient) executeMQSC(ctx context.Context, body []byte) error {
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/action/qmgr/%s/mqsc",
		mqscActionAPIVersion,
		url.PathEscape(c.queueManager),
	)
	respBody, code, err := c.base.requestWithBody(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	if code == http.StatusOK {
		return parseMQSCCompletion(respBody)
	}
	if reason := parseReasonCode(respBody); reason != 0 {
		return mqadmin.MapReasonCode(reason)
	}
	return mqadmin.ReasonCodeFromHTTPStatus(code)
}

func parseMQSCCompletion(body []byte) error {
	var payload struct {
		OverallCompletionCode int `json:"overallCompletionCode"`
		OverallReasonCode     int `json:"overallReasonCode"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	if payload.OverallCompletionCode == 0 && payload.OverallReasonCode == 0 {
		return nil
	}
	if payload.OverallReasonCode != 0 {
		return mqadmin.MapReasonCode(payload.OverallReasonCode)
	}
	return fmt.Errorf("mqsc command failed with completion code %d", payload.OverallCompletionCode)
}

func encodeMQSCCommand(command, qualifier, name string, params map[string]string) ([]byte, error) {
	payload := map[string]any{
		"type":    "runCommandJSON",
		"command": command,
	}
	if qualifier != "" {
		payload["qualifier"] = qualifier
	}
	if name != "" {
		payload["name"] = name
	}
	if len(params) > 0 {
		parameters := make(map[string]string, len(params))
		for key, value := range params {
			if strings.TrimSpace(value) != "" {
				parameters[strings.ToUpper(key)] = value
			}
		}
		payload["parameters"] = parameters
	}
	return json.Marshal(payload)
}

func chlauthIdentityParams(target mqadmin.CHLAUTHTarget) map[string]string {
	params := map[string]string{"TYPE": string(target.RuleType)}
	if addr := strings.TrimSpace(target.Address); addr != "" {
		params["ADDRESS"] = addr
	}
	if user := strings.TrimSpace(target.ClientUser); user != "" {
		params["CLNTUSER"] = user
	}
	if peer := strings.TrimSpace(target.SSLPeer); peer != "" {
		params["SSLPEER"] = peer
	}
	if qmgr := strings.TrimSpace(target.QMgrName); qmgr != "" {
		params["QMNAME"] = qmgr
	}
	return params
}

func authrecIdentityParams(target mqadmin.AuthrecTarget) map[string]string {
	params := map[string]string{
		"PROFILE": strings.TrimSpace(target.Profile),
		"OBJTYPE": string(target.ObjectType),
	}
	switch target.EntityType {
	case mqadmin.AuthrecEntityGroup:
		params["GROUP"] = strings.TrimSpace(target.Entity)
	default:
		params["PRINCIPAL"] = strings.TrimSpace(target.Entity)
	}
	return params
}

func joinAuthrecAuthorities(auths []mqadmin.AuthrecAuthority) string {
	parts := make([]string, 0, len(auths))
	for _, auth := range auths {
		parts = append(parts, string(auth))
	}
	return strings.Join(parts, ",")
}
