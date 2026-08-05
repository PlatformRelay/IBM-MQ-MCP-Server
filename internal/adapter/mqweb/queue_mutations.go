package mqweb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/coexistence"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

func (c *adminClient) DefineQueue(
	ctx context.Context,
	name string,
	req mqadmin.DefineQueueRequest,
) (mqadmin.QueueMutationResult, error) {
	name = strings.TrimSpace(name)
	body, err := encodeQueueDefineBody(name, req)
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/queue",
		queueListAPIVersion,
		url.PathEscape(c.queueManager),
	)
	respBody, code, err := c.base.requestWithBody(ctx, http.MethodPost, path, body)
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	if code != http.StatusCreated && code != http.StatusOK {
		return mqadmin.QueueMutationResult{}, mqadmin.ReasonCodeFromHTTPStatus(code)
	}
	after, err := parseQueueDetail(respBody, name)
	if err != nil {
		after = mqadmin.QueueDetail{Name: name, Type: string(req.QueueType)}
	}
	return mqadmin.QueueMutationResult{
		Operation:    mqadmin.MutationDefine,
		Profile:      c.base.name,
		QueueManager: c.queueManager,
		QueueName:    name,
		After:        queueSnapshot(after),
		CompletedAt:  time.Now().UTC(),
	}, nil
}

func (c *adminClient) AlterQueue(
	ctx context.Context,
	name string,
	req mqadmin.AlterQueueRequest,
) (mqadmin.QueueMutationResult, error) {
	name = strings.TrimSpace(name)
	beforeDetail, beforeErr := c.GetQueue(ctx, name)
	var before *mqadmin.QueueSnapshot
	if beforeErr == nil {
		before = queueSnapshot(beforeDetail)
	}
	body, err := encodeQueueAlterBody(req)
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/queue/%s",
		queueListAPIVersion,
		url.PathEscape(c.queueManager),
		url.PathEscape(name),
	)
	respBody, code, err := c.base.requestWithBody(ctx, http.MethodPut, path, body)
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	if code != http.StatusOK {
		return mqadmin.QueueMutationResult{}, mqadmin.ReasonCodeFromHTTPStatus(code)
	}
	afterDetail, parseErr := parseQueueDetail(respBody, name)
	if parseErr != nil {
		afterDetail = mqadmin.QueueDetail{Name: name}
		if before != nil {
			afterDetail.Type = before.Type
		}
		if req.MaxDepth != nil {
			afterDetail.MaxDepth = *req.MaxDepth
		}
	}
	return mqadmin.QueueMutationResult{
		Operation:    mqadmin.MutationAlter,
		Profile:      c.base.name,
		QueueManager: c.queueManager,
		QueueName:    name,
		Before:       before,
		After:        queueSnapshot(afterDetail),
		CompletedAt:  time.Now().UTC(),
	}, nil
}

func (c *adminClient) DeleteQueue(ctx context.Context, name string) (mqadmin.QueueMutationResult, error) {
	name = strings.TrimSpace(name)
	beforeDetail, beforeErr := c.GetQueue(ctx, name)
	var before *mqadmin.QueueSnapshot
	if beforeErr == nil {
		before = queueSnapshot(beforeDetail)
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/queue/%s",
		queueListAPIVersion,
		url.PathEscape(c.queueManager),
		url.PathEscape(name),
	)
	_, code, err := c.base.requestWithBody(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return mqadmin.QueueMutationResult{}, mqadmin.ReasonCodeFromHTTPStatus(code)
	}
	return mqadmin.QueueMutationResult{
		Operation:    mqadmin.MutationDelete,
		Profile:      c.base.name,
		QueueManager: c.queueManager,
		QueueName:    name,
		Before:       before,
		CompletedAt:  time.Now().UTC(),
	}, nil
}

func queueSnapshot(detail mqadmin.QueueDetail) *mqadmin.QueueSnapshot {
	return &mqadmin.QueueSnapshot{
		Name:        detail.Name,
		Type:        detail.Type,
		MaxDepth:    detail.MaxDepth,
		Description: detail.Description,
	}
}

func encodeQueueDefineBody(name string, req mqadmin.DefineQueueRequest) ([]byte, error) {
	payload := map[string]any{
		"name": name,
		"type": mqwebQueueType(req.QueueType),
	}
	general := map[string]any{}
	if desc := strings.TrimSpace(req.Description); desc != "" {
		general["description"] = desc
	}
	storage := map[string]any{}
	if req.MaxDepth != nil {
		storage["maximumDepth"] = *req.MaxDepth
	}
	if len(storage) > 0 {
		payload["storage"] = storage
	}
	if len(general) > 0 {
		payload["general"] = general
	}
	return json.Marshal(payload)
}

func encodeQueueAlterBody(req mqadmin.AlterQueueRequest) ([]byte, error) {
	payload := map[string]any{}
	general := map[string]any{}
	if req.Description != nil {
		general["description"] = *req.Description
	}
	storage := map[string]any{}
	if req.MaxDepth != nil {
		storage["maximumDepth"] = *req.MaxDepth
	}
	if len(storage) > 0 {
		payload["storage"] = storage
	}
	if len(general) > 0 {
		payload["general"] = general
	}
	if len(payload) == 0 {
		return nil, mqadmin.ErrAlterNoChanges
	}
	return json.Marshal(payload)
}

func mqwebQueueType(qtype mqadmin.QueueType) string {
	switch qtype {
	case mqadmin.QueueTypeLocal:
		return "qlocal"
	case mqadmin.QueueTypeAlias:
		return "qalias"
	case mqadmin.QueueTypeRemote:
		return "qremote"
	case mqadmin.QueueTypeModel:
		return "qmodel"
	default:
		return strings.ToLower(string(qtype))
	}
}

func (b *baseClient) requestWithBody(ctx context.Context, method, path string, body []byte) ([]byte, int, error) {
	if b.closed {
		return nil, 0, errors.New("mqweb client closed")
	}
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, b.endpoint+path, reader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if b.authType == catalog.AuthBasic && b.username != "" {
		req.SetBasicAuth(b.username, b.password)
	}
	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if reason := parseReasonCode(respBody); reason != 0 && resp.StatusCode >= 400 {
		return respBody, resp.StatusCode, mqadmin.MapReasonCode(reason)
	}
	return respBody, resp.StatusCode, nil
}

func parseMKuratorTag(description string) string {
	const prefix = coexistence.TagManagedByMKurator + "="
	if strings.HasPrefix(description, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(description, prefix))
	}
	return ""
}
