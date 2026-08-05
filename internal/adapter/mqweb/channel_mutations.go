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

func (c *adminClient) DefineChannel(
	ctx context.Context,
	name string,
	req mqadmin.DefineChannelRequest,
) (mqadmin.ChannelMutationResult, error) {
	name = strings.TrimSpace(name)
	body, err := encodeChannelDefineBody(name, req)
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/channel",
		objectAPIVersion,
		url.PathEscape(c.queueManager),
	)
	respBody, code, err := c.base.requestWithBody(ctx, http.MethodPost, path, body)
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	if code != http.StatusCreated && code != http.StatusOK {
		return mqadmin.ChannelMutationResult{}, mqadmin.ReasonCodeFromHTTPStatus(code)
	}
	after, err := parseChannelDetail(respBody, name)
	if err != nil {
		after = mqadmin.ChannelDetail{Name: name, Type: string(req.ChannelType)}
	}
	return mqadmin.ChannelMutationResult{
		Operation:    mqadmin.MutationDefine,
		Profile:      c.base.name,
		QueueManager: c.queueManager,
		ChannelName:  name,
		After:        channelSnapshot(after),
		CompletedAt:  time.Now().UTC(),
	}, nil
}

func (c *adminClient) AlterChannel(
	ctx context.Context,
	name string,
	req mqadmin.AlterChannelRequest,
) (mqadmin.ChannelMutationResult, error) {
	name = strings.TrimSpace(name)
	beforeDetail, beforeErr := c.GetChannel(ctx, name)
	var before *mqadmin.ChannelSnapshot
	if beforeErr == nil {
		before = channelSnapshot(beforeDetail)
	}
	body, err := encodeChannelAlterBody(req)
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/channel/%s",
		objectAPIVersion,
		url.PathEscape(c.queueManager),
		url.PathEscape(name),
	)
	respBody, code, err := c.base.requestWithBody(ctx, http.MethodPut, path, body)
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	if code != http.StatusOK {
		return mqadmin.ChannelMutationResult{}, mqadmin.ReasonCodeFromHTTPStatus(code)
	}
	afterDetail, parseErr := parseChannelDetail(respBody, name)
	if parseErr != nil {
		afterDetail = mqadmin.ChannelDetail{Name: name}
		if before != nil {
			afterDetail.Type = before.Type
		}
		if req.Description != nil {
			afterDetail.Description = *req.Description
		}
		if req.ConnectionName != nil {
			afterDetail.ConnectionName = *req.ConnectionName
		}
		if req.TransmissionQueue != nil {
			afterDetail.TransmissionQueue = *req.TransmissionQueue
		}
	}
	return mqadmin.ChannelMutationResult{
		Operation:    mqadmin.MutationAlter,
		Profile:      c.base.name,
		QueueManager: c.queueManager,
		ChannelName:  name,
		Before:       before,
		After:        channelSnapshot(afterDetail),
		CompletedAt:  time.Now().UTC(),
	}, nil
}

func (c *adminClient) DeleteChannel(ctx context.Context, name string) (mqadmin.ChannelMutationResult, error) {
	name = strings.TrimSpace(name)
	beforeDetail, beforeErr := c.GetChannel(ctx, name)
	var before *mqadmin.ChannelSnapshot
	if beforeErr == nil {
		before = channelSnapshot(beforeDetail)
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/channel/%s",
		objectAPIVersion,
		url.PathEscape(c.queueManager),
		url.PathEscape(name),
	)
	_, code, err := c.base.requestWithBody(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return mqadmin.ChannelMutationResult{}, mqadmin.ReasonCodeFromHTTPStatus(code)
	}
	return mqadmin.ChannelMutationResult{
		Operation:    mqadmin.MutationDelete,
		Profile:      c.base.name,
		QueueManager: c.queueManager,
		ChannelName:  name,
		Before:       before,
		CompletedAt:  time.Now().UTC(),
	}, nil
}

func channelSnapshot(detail mqadmin.ChannelDetail) *mqadmin.ChannelSnapshot {
	return &mqadmin.ChannelSnapshot{
		Name:              detail.Name,
		Type:              detail.Type,
		Description:       detail.Description,
		ConnectionName:    detail.ConnectionName,
		TransmissionQueue: detail.TransmissionQueue,
	}
}

func encodeChannelDefineBody(name string, req mqadmin.DefineChannelRequest) ([]byte, error) {
	payload := map[string]any{
		"name": name,
		"type": mqwebChannelType(req.ChannelType),
	}
	general := map[string]any{}
	if desc := strings.TrimSpace(req.Description); desc != "" {
		general["description"] = desc
	}
	if conn := strings.TrimSpace(req.ConnectionName); conn != "" {
		general["connectionName"] = conn
	}
	if xmitq := strings.TrimSpace(req.TransmissionQueue); xmitq != "" {
		general["transmissionQueue"] = xmitq
	}
	if len(general) > 0 {
		payload["general"] = general
	}
	return json.Marshal(payload)
}

func encodeChannelAlterBody(req mqadmin.AlterChannelRequest) ([]byte, error) {
	payload := map[string]any{}
	general := map[string]any{}
	if req.Description != nil {
		general["description"] = *req.Description
	}
	if req.ConnectionName != nil {
		general["connectionName"] = *req.ConnectionName
	}
	if req.TransmissionQueue != nil {
		general["transmissionQueue"] = *req.TransmissionQueue
	}
	if len(general) > 0 {
		payload["general"] = general
	}
	if len(payload) == 0 {
		return nil, mqadmin.ErrAlterNoChanges
	}
	return json.Marshal(payload)
}

func mqwebChannelType(chType mqadmin.ChannelType) string {
	switch chType {
	case mqadmin.ChannelTypeSender:
		return "sender"
	case mqadmin.ChannelTypeServer:
		return "server"
	case mqadmin.ChannelTypeReceiver:
		return "receiver"
	case mqadmin.ChannelTypeRequester:
		return "requester"
	case mqadmin.ChannelTypeClientConnection:
		return "clientConnection"
	case mqadmin.ChannelTypeServerConnection:
		return "serverConnection"
	case mqadmin.ChannelTypeClusterSender:
		return "clusterSender"
	case mqadmin.ChannelTypeClusterReceiver:
		return "clusterReceiver"
	default:
		return strings.ToLower(string(chType))
	}
}
