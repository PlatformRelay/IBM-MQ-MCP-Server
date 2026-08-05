package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// DOC-001: required documentation paths must exist so MkDocs strict nav cannot
// silently drop pages without failing CI.
func TestRequiredDocumentationPathsExist(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	required := []string{
		"docs/quickstart.md",
		"docs/configuration.md",
		"docs/authentication.md",
		"docs/policy.md",
		"docs/tools/index.md",
		"docs/examples/README.md",
		"docs/examples/profile-read-only.yaml",
		"docs/examples/profile-mixed-grants.yaml",
		"docs/deployment.md",
		"docs/observability.md",
		"docs/troubleshooting.md",
		"docs/upgrade.md",
		"docs/security/threat-model.md",
		"docs/support/version-matrix.md",
		"docs/support/mkurator-coexistence.md",
		"docs/development/local-mq.md",
		"docs/examples/profile-kind-local.yaml",
		"docs/RELEASE.md",
		"docs/adr/README.md",
		"mkdocs.yml",
	}
	for _, rel := range required {
		path := filepath.Join(root, rel)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("required documentation path missing: %s (%v)", rel, err)
		}
	}
}

func TestMkdocsNavReferencesRequiredPages(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	data, err := os.ReadFile(filepath.Join(root, "mkdocs.yml")) //nolint:gosec // G304: fixed path
	if err != nil {
		t.Fatalf("read mkdocs.yml: %v", err)
	}
	body := string(data)

	fragments := []string{
		"quickstart.md",
		"configuration.md",
		"authentication.md",
		"policy.md",
		"tools/index.md",
		"security/threat-model.md",
		"support/version-matrix.md",
		"observability.md",
		"deployment.md",
		"troubleshooting.md",
		"upgrade.md",
		"development/local-mq.md",
	}
	for _, fragment := range fragments {
		if !strings.Contains(body, fragment) {
			t.Errorf("mkdocs.yml nav missing reference to %q", fragment)
		}
	}
}
