//go:build e2e

package e2e

import (
	"io"
	"net/http"
	"testing"
)

func TestMQWeb_AdminREST_QueueManagerReachable(t *testing.T) {
	env := requireE2E(t)

	path := "/ibmmq/rest/v3/admin/qmgr/" + env.queueMgr
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, env.url(path), nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.SetBasicAuth(env.user, env.password)

	resp, err := env.httpClient.Do(req)
	if err != nil {
		t.Fatalf("admin REST request failed (is MQ up? see docs/development/local-mq.md): %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin REST expected 200 with valid credentials, got %d for %s", resp.StatusCode, path)
	}
}

func TestMQWeb_MessagingREST_Reachable(t *testing.T) {
	env := requireE2E(t)

	// HEAD against a well-known queue path proves the messaging REST servlet is up.
	path := "/ibmmq/rest/v3/messaging/qmgr/" + env.queueMgr + "/queue/SYSTEM.DEFAULT.LOCAL.QUEUE/message"
	req, err := http.NewRequestWithContext(t.Context(), http.MethodHead, env.url(path), nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.SetBasicAuth(env.user, env.password)

	resp, err := env.httpClient.Do(req)
	if err != nil {
		t.Fatalf("messaging REST request failed (is MQ up? see docs/development/local-mq.md): %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent, http.StatusMethodNotAllowed, http.StatusNotFound:
		// HEAD on queue message path: 200/204 when reachable; 404 if queue missing; 405 if method unsupported.
	default:
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			t.Fatalf("messaging REST auth failed (%d) — check MQ_ADMIN_PASSWORD / IBM_MQ_MCP_E2E_PASSWORD", resp.StatusCode)
		}
		t.Fatalf("messaging REST unexpected status %d for %s", resp.StatusCode, path)
	}
}
