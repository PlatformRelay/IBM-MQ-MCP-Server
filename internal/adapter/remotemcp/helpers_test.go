package remotemcp_test

import (
	"os"
	"path/filepath"
	"testing"
)

func writeProfileYAML(t *testing.T, dir, mqEndpoint, capability string) string {
	t.Helper()
	path := filepath.Join(dir, "profiles.yaml")
	content := `
profiles:
  prod:
    queueManager: QM1
    endpoint: ` + mqEndpoint + `
    tls:
      insecureSkipVerify: true
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_TOOL_SECRET
    capabilities:
      - ` + capability + `
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
