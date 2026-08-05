package architecture_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// Required CI jobs for FND-002. Removing any of these must fail this test,
// which is the recorded evidence that a deliberately incomplete gate set is rejected.
var requiredCIJobs = []string{
	"gitleaks",
	"oss-hygiene",
	"format",
	"lint",
	"test",
	"vulncheck",
	"scrub",
}

func TestCIWorkflowDefinesRequiredJobs(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	// Fixed relative path under the repository root (not user-controlled).
	data, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci.yaml")) //nolint:gosec // G304: fixed path
	if err != nil {
		t.Fatalf("read ci.yaml: %v", err)
	}
	body := string(data)
	for _, job := range requiredCIJobs {
		// Match top-level job keys under jobs: (two-space indent).
		re := regexp.MustCompile(`(?m)^  ` + regexp.QuoteMeta(job) + `:`)
		if !re.MatchString(body) {
			t.Errorf("ci.yaml missing required job %q", job)
		}
	}
}

func TestCodeQLWorkflowExists(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	path := filepath.Join(root, ".github", "workflows", "codeql.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("CodeQL workflow required by FND-002: %v", err)
	}
}
