package architecture_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestDockerfileUsesDistrolessNonroot(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	data, err := os.ReadFile(filepath.Join(root, "Dockerfile")) //nolint:gosec // G304: fixed path
	if err != nil {
		t.Fatalf("read Dockerfile: %v", err)
	}
	body := string(data)

	required := []string{
		"gcr.io/distroless/static:nonroot",
		"USER 65532:65532",
		"CGO_ENABLED=0",
	}
	for _, fragment := range required {
		if !strings.Contains(body, fragment) {
			t.Errorf("Dockerfile missing required fragment %q", fragment)
		}
	}
}

func TestReleaseWorkflowDefinesSupplyChainSteps(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	path := filepath.Join(root, ".github", "workflows", "release.yaml")
	data, err := os.ReadFile(path) //nolint:gosec // G304: fixed path
	if err != nil {
		t.Fatalf("read release.yaml: %v", err)
	}
	body := string(data)

	requiredFragments := []string{
		"sigstore/cosign-installer",
		"cosign sign --yes",
		"actions/attest",
		"anchore/sbom-action",
		"aquasecurity/trivy-action",
		"provenance: mode=max",
		"sbom: true",
		"id-token: write",
		"environment: release",
		"ghcr.io",
		"ibm-mq-mcp",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(body, fragment) {
			t.Errorf("release.yaml missing required fragment %q", fragment)
		}
	}

	trivySeverity := regexp.MustCompile(`(?m)severity:\s*CRITICAL,HIGH`)
	trivyExitCode := regexp.MustCompile(`(?m)exit-code:\s*1`)
	if !trivySeverity.MatchString(body) && !trivyExitCode.MatchString(body) {
		t.Error("release.yaml Trivy step must set severity CRITICAL,HIGH and/or exit-code 1")
	}

	concurrency := regexp.MustCompile(`(?m)^concurrency:`)
	if !concurrency.MatchString(body) {
		t.Error("release.yaml missing concurrency group for tag/ref")
	}

	// FND-003 / DQ22: no Helm or Kustomize in v0 release pipeline.
	forbidden := regexp.MustCompile(`(?i)(helm push|setup-helm|install-crds\.yaml|install\.yaml)`)
	if forbidden.MatchString(body) {
		t.Errorf("release.yaml must not include Helm/Kustomize artifacts (DQ22)")
	}
}

func TestReleaseAssetsChecksumGlobMatchesTarballNames(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	path := filepath.Join(root, "hack", "release-assets.sh")
	data, err := os.ReadFile(path) //nolint:gosec // G304: fixed path
	if err != nil {
		t.Fatalf("read release-assets.sh: %v", err)
	}
	body := string(data)

	tarballPattern := regexp.MustCompile(`\$\{DIST\}/\$\{BINARY\}_\$\{VERSION\}_\$\{GOOS\}_\$\{GOARCH\}\.tar\.gz`)
	if !tarballPattern.MatchString(body) {
		t.Fatal("release-assets.sh must build tarballs as ${BINARY}_${VERSION}_${GOOS}_${GOARCH}.tar.gz")
	}

	checksumGlob := regexp.MustCompile(`files=\("\$\{BINARY\}"_"\$\{VERSION\}"_\*\.tar\.gz`)
	if !checksumGlob.MatchString(body) {
		t.Fatal("release-assets.sh checksum glob must be ${BINARY}_${VERSION}_*.tar.gz to match tarball names")
	}

	// Guard against the v0.1.0 regression: version-before-os/arch tarballs do not match
	// a middle-version glob like ibm-mq-mcp_*_0.1.0_*.tar.gz.
	sample := "ibm-mq-mcp_0.1.0_linux_amd64.tar.gz"
	ok, err := filepath.Match("ibm-mq-mcp_0.1.0_*.tar.gz", sample)
	if err != nil {
		t.Fatalf("match checksum glob: %v", err)
	}
	if !ok {
		t.Errorf("checksum glob ibm-mq-mcp_0.1.0_*.tar.gz does not match sample tarball %q", sample)
	}
	ok, err = filepath.Match("ibm-mq-mcp_*_0.1.0_*.tar.gz", sample)
	if err != nil {
		t.Fatalf("match legacy glob: %v", err)
	}
	if ok {
		t.Errorf("legacy checksum glob ibm-mq-mcp_*_0.1.0_*.tar.gz must not match tarball %q", sample)
	}
}

func TestCIWorkflowDefinesDockerBuildJob(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	data, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci.yaml")) //nolint:gosec // G304: fixed path
	if err != nil {
		t.Fatalf("read ci.yaml: %v", err)
	}
	body := string(data)
	re := regexp.MustCompile(`(?m)^  docker-build:`)
	if !re.MatchString(body) {
		t.Error("ci.yaml missing docker-build job")
	}
}
