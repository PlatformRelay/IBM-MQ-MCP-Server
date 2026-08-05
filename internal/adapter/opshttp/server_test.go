package opshttp_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/opshttp"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/metrics"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/runtime"
)

func TestHealthzAlwaysOK(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t, runtime.New())
	res := get(t, ts.URL, "/healthz")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET /healthz = %d, want 200", res.StatusCode)
	}
}

func TestReadyzReflectsConfigAndTransportWithoutMQ(t *testing.T) {
	t.Parallel()

	rt := runtime.New()
	ts := newTestServer(t, rt)

	res := get(t, ts.URL, "/readyz")
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("GET /readyz before ready = %d, want 503", res.StatusCode)
	}

	rt.SetConfigValid(true)
	rt.SetTransportReady(true, "stdio")

	res = get(t, ts.URL, "/readyz")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET /readyz when ready = %d, want 200", res.StatusCode)
	}

	rt.SetConfigValid(false)
	res = get(t, ts.URL, "/readyz")
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("GET /readyz with invalid config = %d, want 503", res.StatusCode)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	t.Parallel()

	rt := runtime.New()
	reg := metrics.New()
	reg.RecordRequest("_none", 0)
	srv := opshttp.NewHandler(rt, reg)

	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)

	res := get(t, ts.URL, "/metrics")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET /metrics = %d, want 200", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), "ibm_mq_mcp_requests_total") {
		t.Fatalf("metrics body missing request counter: %s", body)
	}
}

func newTestServer(t *testing.T, rt *runtime.Runtime) *httptest.Server {
	t.Helper()
	return httptest.NewServer(opshttp.NewHandler(rt, metrics.New()))
}

func get(t *testing.T, baseURL, path string) *http.Response {
	t.Helper()
	res, err := http.Get(baseURL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	return res
}
