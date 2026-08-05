//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/mqweb"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
)

func TestConnectivity_HealthyProfile(t *testing.T) {
	env := requireE2E(t)
	inspector := newE2EInspector(t, env)

	report, err := inspector.CheckProfileConnectivity(context.Background(), "e2e")
	if err != nil {
		t.Fatalf("connectivity check: %v", err)
	}
	if !report.Reachable {
		t.Fatalf("expected reachable profile: %+v", report)
	}
	if !report.IdentityMatch {
		t.Fatalf("identity mismatch: %+v", report.Identity)
	}
	if report.LatencyMs < 0 {
		t.Fatalf("latency = %d", report.LatencyMs)
	}
}

func TestConnectivity_UnreachableEndpoint(t *testing.T) {
	env := requireE2E(t)
	doc := fmt.Sprintf(`
profiles:
  bad:
    queueManager: %s
    endpoint: https://127.0.0.1:1
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_E2E_SECRET
    tls:
      insecureSkipVerify: true
    capabilities:
      - inspect
`, env.queueMgr)
	t.Setenv("IBM_MQ_MCP_E2E_SECRET", env.user+":"+env.password)
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	pool := application.NewProfilePool(
		cat,
		cat.Validate(),
		secrets.NewResolver(),
		nil,
		application.WithAdminFactory(mqweb.NewAdminClient),
	)
	t.Cleanup(func() { _ = pool.Close() })
	inspector := application.NewInspector(pool)

	report, err := inspector.CheckProfileConnectivity(context.Background(), "bad")
	if err != nil {
		t.Fatal(err)
	}
	if report.Reachable {
		t.Fatalf("expected unreachable: %+v", report)
	}
	if report.FailureCause == "" {
		t.Fatalf("expected typed failure cause: %+v", report)
	}
}

func newE2EInspector(t *testing.T, env mqEnv) *application.Inspector {
	t.Helper()
	secret := env.user + ":" + env.password
	t.Setenv("IBM_MQ_MCP_E2E_SECRET", secret)
	endpoint := profileEndpoint(env)
	insecure := envOr("IBM_MQ_MCP_MQ_INSECURE_TLS", defaultInsecureTL) == "true"
	doc := fmt.Sprintf(`
profiles:
  e2e:
    queueManager: %s
    endpoint: %s
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_E2E_SECRET
    tls:
      insecureSkipVerify: %t
    capabilities:
      - inspect
`, env.queueMgr, endpoint, insecure)
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	pool := application.NewProfilePool(
		cat,
		cat.Validate(),
		secrets.NewResolver(),
		nil,
		application.WithAdminFactory(mqweb.NewAdminClient),
	)
	t.Cleanup(func() { _ = pool.Close() })
	return application.NewInspector(pool)
}

func profileEndpoint(env mqEnv) string {
	raw := strings.TrimRight(env.endpoint.String(), "/")
	if env.host != "" && strings.Contains(raw, "127.0.0.1") {
		port := env.endpoint.Port()
		if port == "" {
			port = "30443"
		}
		return fmt.Sprintf("https://%s:%s", env.host, port)
	}
	return raw
}
