//go:build e2e

package e2e

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestMQWeb_ConsumeDecreasesQueueDepth(t *testing.T) {
	env := requireE2E(t)
	queue := envOr("IBM_MQ_MCP_E2E_QUEUE", "SYSTEM.DEFAULT.LOCAL.QUEUE")

	depthBefore, err := queueCurrentDepth(t, env, queue)
	if err != nil {
		t.Fatalf("depth before: %v", err)
	}

	putPath := "/ibmmq/rest/v3/messaging/qmgr/" + env.queueMgr + "/queue/" + url.PathEscape(queue) + "/message"
	putReq, err := http.NewRequestWithContext(t.Context(), http.MethodPost, env.url(putPath), strings.NewReader("e2e-consume-probe"))
	if err != nil {
		t.Fatal(err)
	}
	putReq.SetBasicAuth(env.user, env.password)
	putReq.Header.Set("ibm-mq-rest-csrf-token", "1")
	putReq.Header.Set("Content-Type", "text/plain;charset=utf-8")
	putResp, err := env.httpClient.Do(putReq)
	if err != nil {
		t.Fatal(err)
	}
	_ = putResp.Body.Close()
	if putResp.StatusCode != http.StatusOK && putResp.StatusCode != http.StatusCreated && putResp.StatusCode != http.StatusNoContent {
		t.Fatalf("put status = %d", putResp.StatusCode)
	}

	depthAfterPut, err := queueCurrentDepth(t, env, queue)
	if err != nil {
		t.Fatalf("depth after put: %v", err)
	}
	if depthAfterPut <= depthBefore {
		t.Fatalf("depth did not increase after put: before=%d after=%d", depthBefore, depthAfterPut)
	}

	delPath := "/ibmmq/rest/v3/messaging/qmgr/" + env.queueMgr + "/queue/" + url.PathEscape(queue) + "/message"
	delReq, err := http.NewRequestWithContext(t.Context(), http.MethodDelete, env.url(delPath), nil)
	if err != nil {
		t.Fatal(err)
	}
	delReq.SetBasicAuth(env.user, env.password)
	delReq.Header.Set("ibm-mq-rest-csrf-token", "1")
	delReq.Header.Set("Accept", "text/plain")
	delResp, err := env.httpClient.Do(delReq)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(delResp.Body)
	_ = delResp.Body.Close()
	if delResp.StatusCode != http.StatusOK && delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d body=%s", delResp.StatusCode, body)
	}

	depthAfterConsume, err := queueCurrentDepth(t, env, queue)
	if err != nil {
		t.Fatalf("depth after consume: %v", err)
	}
	if depthAfterConsume >= depthAfterPut {
		t.Fatalf("depth did not decrease after consume: afterPut=%d afterConsume=%d", depthAfterPut, depthAfterConsume)
	}
}

func queueCurrentDepth(t *testing.T, env mqEnv, queue string) (int, error) {
	t.Helper()
	path := "/ibmmq/rest/v1/admin/qmgr/" + env.queueMgr + "/queue/" + url.PathEscape(queue)
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, env.url(path), nil)
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(env.user, env.password)
	resp, err := env.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("get queue status %d: %s", resp.StatusCode, body)
	}
	// Minimal parse — currentDepth integer in JSON.
	const key = `"currentDepth":`
	idx := strings.Index(string(body), key)
	if idx < 0 {
		return 0, fmt.Errorf("currentDepth not found in %s", body)
	}
	rest := string(body)[idx+len(key):]
	end := strings.IndexAny(rest, ",}")
	if end < 0 {
		return 0, fmt.Errorf("parse currentDepth from %s", rest)
	}
	val, err := strconv.Atoi(strings.TrimSpace(rest[:end]))
	if err != nil {
		return 0, err
	}
	return val, nil
}
