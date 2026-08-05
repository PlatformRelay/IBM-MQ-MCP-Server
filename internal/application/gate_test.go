package application_test

import (
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin/fake"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

type spyRecorder struct {
	denials []string
}

func (s *spyRecorder) RecordRequest(string, float64) {}

func (s *spyRecorder) RecordPolicyDenial(profile string) {
	s.denials = append(s.denials, profile)
}

func TestProfilePoolDeniesAdminBeforeSecretResolve(t *testing.T) {
	doc := `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_NEVER_SET_SECRET
    tls:
      insecureSkipVerify: true
    capabilities:
      - inspect
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	factory := func(profile catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
		return fake.New(profile.Name), nil
	}
	gate := application.NewPolicyGate()
	pool := newPool(t, cat, gate, factory)

	_, err = pool.Admin("prod", policy.Administer)
	if err == nil {
		t.Fatal("expected policy denial")
	}
	if !strings.Contains(err.Error(), "administer") {
		t.Fatalf("unexpected error: %v", err)
	}
	decisions := gate.Decisions()
	if len(decisions) != 1 || decisions[0].Granted {
		t.Fatalf("decisions = %+v", decisions)
	}
}

func TestProfilePoolDenialIncrementsRecorder(t *testing.T) {
	doc := `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_NEVER_SET_SECRET
    tls:
      insecureSkipVerify: true
    capabilities:
      - inspect
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	rec := &spyRecorder{}
	gate := application.NewPolicyGate(application.WithRecorder(rec))
	pool := application.NewProfilePool(cat, cat.Validate(), secrets.NewResolver(), gate)
	t.Cleanup(func() { _ = pool.Close() })

	_, err = pool.Messaging("prod", policy.Produce)
	if err == nil {
		t.Fatal("expected policy denial")
	}
	if len(rec.denials) != 1 || rec.denials[0] != "prod" {
		t.Fatalf("denials = %+v", rec.denials)
	}
}

func TestProfilePoolAllowsAdminWithGrant(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_POOL_SECRET", "user:pass")
	doc := `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_POOL_SECRET
    tls:
      insecureSkipVerify: true
    capabilities:
      - administer
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	factory := func(profile catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
		return fake.New(profile.Name), nil
	}
	pool := newPool(t, cat, nil, factory)

	client, err := pool.Admin("prod", policy.Administer)
	if err != nil {
		t.Fatal(err)
	}
	if client.ProfileName() != "prod" {
		t.Fatalf("profile = %q", client.ProfileName())
	}
}

func TestProfilePoolAuthorizeWithoutClient(t *testing.T) {
	doc := `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_POOL_SECRET
    capabilities:
      - inspect
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	pool := application.NewProfilePool(cat, cat.Validate(), secrets.NewResolver(), nil)
	t.Cleanup(func() { _ = pool.Close() })

	if err := pool.Authorize("prod", policy.Produce, "future-tool"); err == nil {
		t.Fatal("expected denial")
	}
}
