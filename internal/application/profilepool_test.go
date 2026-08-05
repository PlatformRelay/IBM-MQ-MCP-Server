package application_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
)

func TestProfilePoolLazyResolveMissingSecret(t *testing.T) {
	doc := `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_MISSING_POOL_SECRET
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	validation := cat.Validate()
	pool := application.NewProfilePool(cat, validation, secrets.NewResolver())
	t.Cleanup(func() { _ = pool.Close() })

	_, err = pool.Admin("prod")
	if err == nil {
		t.Fatal("expected missing secret error on first use")
	}
	if !strings.Contains(err.Error(), "IBM_MQ_MCP_MISSING_POOL_SECRET") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProfilePoolRejectsInvalidProfile(t *testing.T) {
	doc := `
profiles:
  bad:
    queueManager: QM1
    endpoint: http://bad:9443
    authentication:
      type: basic
      secretRef: env:MQ_PASS
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	validation := cat.Validate()
	pool := application.NewProfilePool(cat, validation, secrets.NewResolver())
	t.Cleanup(func() { _ = pool.Close() })

	if _, err := pool.Admin("bad"); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestProfilePoolResolvesBasicSecret(t *testing.T) {
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
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	validation := cat.Validate()
	pool := application.NewProfilePool(cat, validation, secrets.NewResolver())
	t.Cleanup(func() { _ = pool.Close() })

	client, err := pool.Admin("prod")
	if err != nil {
		t.Fatal(err)
	}
	if client.ProfileName() != "prod" {
		t.Fatalf("profile = %q", client.ProfileName())
	}
}

func TestConfigReadyFailOpen(t *testing.T) {
	doc := `
profiles:
  good:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:MQ_GOOD
  bad:
    queueManager: QM2
    endpoint: http://bad
    authentication:
      type: basic
      secretRef: env:MQ_BAD
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	validation := cat.Validate()
	if !application.ConfigReady(cat, validation) {
		t.Fatal("expected ready when at least one profile valid")
	}
}

func TestLoadCatalogFromFileYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")
	if err := os.WriteFile(path, []byte(`
profiles:
  local:
    queueManager: QM1
    endpoint: https://localhost:9443
    authentication:
      type: basic
      secretRef: env:MQ_PASS
`), 0o600); err != nil {
		t.Fatal(err)
	}
	cat, err := application.LoadCatalogFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cat.Validate().IsValid("local") {
		t.Fatal("expected valid profile from file")
	}
}
