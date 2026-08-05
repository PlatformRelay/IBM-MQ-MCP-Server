package remotemcp_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/remotemcp"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/runtime"
)

func TestUnauthenticatedOversizedBodyLimitedBeforeAuth(t *testing.T) {
	t.Parallel()

	srv := newLimitedRemoteServer(t, remotemcp.Limits{
		MaxBodyBytes:      32,
		RequestsPerSecond: 100,
		Burst:             10,
		MaxConcurrency:    8,
	}, "gate-token")
	t.Cleanup(func() { _ = srv.Close() })

	body := strings.Repeat("x", 64)
	req, err := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	setMCPHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d (limits before auth)", resp.StatusCode, http.StatusRequestEntityTooLarge)
	}
}

func TestUnauthenticatedRateLimitedBeforeAuth(t *testing.T) {
	t.Parallel()

	srv := newLimitedRemoteServer(t, remotemcp.Limits{
		MaxBodyBytes:      1 << 20,
		RequestsPerSecond: 100,
		Burst:             1,
		MaxConcurrency:    64,
	}, "gate-token")
	t.Cleanup(func() { _ = srv.Close() })

	var lastStatus int
	for i := 0; i < 8; i++ {
		req, err := http.NewRequest(http.MethodPost, srv.URL, io.NopCloser(strings.NewReader("{}")))
		if err != nil {
			t.Fatal(err)
		}
		setMCPHeaders(req)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		lastStatus = resp.StatusCode
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}

	if lastStatus != http.StatusTooManyRequests {
		t.Fatalf("last status = %d, want %d (limits must run before bearer auth)", lastStatus, http.StatusTooManyRequests)
	}
}

func TestOversizedBodyWithValidBearerRejected(t *testing.T) {
	t.Parallel()

	srv := newLimitedRemoteServer(t, remotemcp.Limits{
		MaxBodyBytes:      1 << 20,
		RequestsPerSecond: 100,
		Burst:             10,
		MaxConcurrency:    8,
	}, "gate-token")
	t.Cleanup(func() { _ = srv.Close() })

	body := strings.Repeat("x", (1<<20)+1)
	req, err := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	setMCPHeaders(req)
	req.Header.Set("Authorization", "Bearer gate-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusRequestEntityTooLarge)
	}
}

func TestConcurrencyLimitReturns503(t *testing.T) {
	t.Parallel()

	block := make(chan struct{})
	release := make(chan struct{})
	slow := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(block)
		<-release
		w.WriteHeader(http.StatusOK)
	})
	handler := remotemcp.LimitHandler(remotemcp.Limits{
		MaxBodyBytes:      1 << 20,
		RequestsPerSecond: 1000,
		Burst:             1000,
		MaxConcurrency:    1,
	}, remotemcp.BearerAuth("gate-token", slow))

	ts := httptest.NewServer(handler)
	t.Cleanup(func() {
		close(release)
		ts.Close()
	})

	go func() {
		req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(`{}`))
		if err != nil {
			t.Error(err)
			return
		}
		setMCPHeaders(req)
		req.Header.Set("Authorization", "Bearer gate-token")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		defer func() { _ = resp.Body.Close() }()
	}()

	<-block

	req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	setMCPHeaders(req)
	req.Header.Set("Authorization", "Bearer gate-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}
}

func newLimitedRemoteServer(t *testing.T, limits remotemcp.Limits, token string) *remotemcp.TestListener {
	t.Helper()

	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	rt := runtime.New()
	rt.SetConfigValid(true)

	srv, err := remotemcp.NewTestServer(remotemcp.Config{
		AuthToken:     token,
		Limits:        limits,
		TransportName: "streamable-http",
	}, mcpServer, rt)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	return srv
}
