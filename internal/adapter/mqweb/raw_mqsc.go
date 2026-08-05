package mqweb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

// ExecuteRawMQSC submits a validated plain-text MQSC command via mqweb runCommand.
func (c *adminClient) ExecuteRawMQSC(ctx context.Context, command string) (mqadmin.RawMQSCResult, error) {
	body, err := encodeRawMQSCCommand(command)
	if err != nil {
		return mqadmin.RawMQSCResult{}, err
	}
	respBody, code, err := c.postMQSC(ctx, body)
	if err != nil {
		return mqadmin.RawMQSCResult{}, err
	}
	if code != http.StatusOK {
		if reason := parseReasonCode(respBody); reason != 0 {
			return mqadmin.RawMQSCResult{}, mqadmin.MapReasonCode(reason)
		}
		return mqadmin.RawMQSCResult{}, mqadmin.ReasonCodeFromHTTPStatus(code)
	}
	completion, err := parseRawMQSCCompletion(respBody)
	if err != nil {
		return mqadmin.RawMQSCResult{}, err
	}
	return mqadmin.RawMQSCResult{
		Profile:      c.base.name,
		QueueManager: c.queueManager,
		Command:      command,
		Completion:   completion,
		CompletedAt:  time.Now().UTC(),
	}, nil
}

func (c *adminClient) postMQSC(ctx context.Context, body []byte) ([]byte, int, error) {
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/action/qmgr/%s/mqsc",
		mqscActionAPIVersion,
		url.PathEscape(c.queueManager),
	)
	return c.base.requestWithBody(ctx, http.MethodPost, path, body)
}

func encodeRawMQSCCommand(command string) ([]byte, error) {
	payload := map[string]any{
		"type": "runCommand",
		"parameters": map[string]string{
			"command": command,
		},
	}
	return json.Marshal(payload)
}

func parseRawMQSCCompletion(body []byte) (mqadmin.MQSCCompletion, error) {
	var payload struct {
		OverallCompletionCode int `json:"overallCompletionCode"`
		OverallReasonCode     int `json:"overallReasonCode"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return mqadmin.MQSCCompletion{}, fmt.Errorf("parse mqsc completion: %w", err)
	}
	if payload.OverallCompletionCode != 0 || payload.OverallReasonCode != 0 {
		if payload.OverallReasonCode != 0 {
			return mqadmin.MQSCCompletion{}, mqadmin.MapReasonCode(payload.OverallReasonCode)
		}
		return mqadmin.MQSCCompletion{}, fmt.Errorf(
			"mqsc command failed with completion code %d",
			payload.OverallCompletionCode,
		)
	}
	return mqadmin.MQSCCompletion{
		OverallCompletionCode: payload.OverallCompletionCode,
		OverallReasonCode:     payload.OverallReasonCode,
	}, nil
}
