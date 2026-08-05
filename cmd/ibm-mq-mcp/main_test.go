package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveConfigPathPrefersFlag(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_CONFIG", "/env/config.yaml")
	if got := resolveConfigPath("/flag/config.yaml"); got != "/flag/config.yaml" {
		t.Fatalf("resolveConfigPath = %q", got)
	}
}

func TestResolveConfigPathFromEnv(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_CONFIG", "/env/config.yaml")
	if got := resolveConfigPath(""); got != "/env/config.yaml" {
		t.Fatalf("resolveConfigPath = %q", got)
	}
}

func TestLoadProfilesStrictStartup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte(`
profiles:
  bad:
    queueManager: QM1
    endpoint: http://insecure:9443
    authentication:
      type: basic
      secretRef: env:MQ_PASS
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadProfiles(path, true); err == nil {
		t.Fatal("expected strict startup failure")
	}
}

func TestLoadProfilesFailOpen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mixed.yaml")
	if err := os.WriteFile(path, []byte(`
profiles:
  good:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:MQ_GOOD
    capabilities:
      - inspect
  bad:
    queueManager: QM2
    endpoint: http://insecure:9443
    authentication:
      type: basic
      secretRef: env:MQ_BAD
`), 0o600); err != nil {
		t.Fatal(err)
	}
	pool, ready, err := loadProfiles(path, false)
	if err != nil {
		t.Fatal(err)
	}
	if pool == nil || !ready {
		t.Fatal("expected fail-open pool with ready config")
	}
}
