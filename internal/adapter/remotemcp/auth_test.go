package remotemcp_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/remotemcp"
)

func TestBearerAuthRejectsMissingToken(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := remotemcp.BearerAuth("expected-token", next)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestBearerAuthRejectsWrongToken(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := remotemcp.BearerAuth("expected-token", next)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestBearerAuthAllowsValidToken(t *testing.T) {
	t.Parallel()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	handler := remotemcp.BearerAuth("expected-token", next)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer expected-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !called {
		t.Fatal("expected downstream handler to run")
	}
}

func TestBearerAuthDoesNotForwardClientAuthorization(t *testing.T) {
	t.Parallel()

	var seenAuth string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	})
	handler := remotemcp.BearerAuth("mcp-gate-token", next)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer mcp-gate-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if seenAuth != "" {
		t.Fatalf("Authorization forwarded to inner handler = %q, want empty", seenAuth)
	}
}

func TestValidateConfigRequiresAuthWhenRemoteEnabled(t *testing.T) {
	t.Parallel()

	err := remotemcp.Validate(remotemcp.Config{Addr: ":8080"})
	if err == nil {
		t.Fatal("expected error when remote addr set without auth token ref")
	}
}

func TestResolveAuthTokenFromEnvRef(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_REMOTE_TOKEN", "gate-secret")
	token, err := remotemcp.ResolveAuthToken("env:IBM_MQ_MCP_REMOTE_TOKEN")
	if err != nil {
		t.Fatal(err)
	}
	if token != "gate-secret" {
		t.Fatalf("token = %q", token)
	}
}

func TestOversizedRequestRejected(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := remotemcp.LimitHandler(remotemcp.Limits{MaxBodyBytes: 32}, next)

	body := strings.Repeat("x", 64)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestRateLimitEventuallyRejects(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	limits := remotemcp.Limits{
		MaxBodyBytes:      1 << 20,
		RequestsPerSecond: 100,
		Burst:             1,
		MaxConcurrency:    64,
	}
	handler := remotemcp.LimitHandler(limits, next)

	var lastStatus int
	for i := 0; i < 8; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader("{}")))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		lastStatus = rec.Code
	}
	if lastStatus != http.StatusTooManyRequests {
		t.Fatalf("last status = %d, want %d", lastStatus, http.StatusTooManyRequests)
	}
}
