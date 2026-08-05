package catalog_test

import (
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
)

const validBasicYAML = `
profiles:
  production:
    queueManager: PROD1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:MQ_PROD_CREDENTIALS
    tls:
      caRef: file:/etc/mq/ca.pem
    capabilities:
      - inspect
`

func TestLoadYAMLValidCatalog(t *testing.T) {
	cat, err := catalog.LoadYAML([]byte(validBasicYAML))
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Profiles) != 1 {
		t.Fatalf("profiles = %d", len(cat.Profiles))
	}
	result := cat.Validate()
	if !result.AllValid() {
		t.Fatalf("validation failed: %+v", result.Statuses)
	}
}

func TestLoadYAMLDuplicateNames(t *testing.T) {
	doc := `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://a.example.test:9443
    authentication:
      type: basic
      secretRef: env:X
  prod:
    queueManager: QM2
    endpoint: https://b.example.test:9443
    authentication:
      type: basic
      secretRef: env:Y
`
	_, err := catalog.LoadYAML([]byte(doc))
	if err == nil {
		t.Fatal("expected duplicate name error")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateMissingSecretRef(t *testing.T) {
	doc := `
profiles:
  bad:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	result := cat.Validate()
	if result.AllValid() {
		t.Fatal("expected validation failure")
	}
	if result.IsValid("bad") {
		t.Fatal("bad profile should be invalid")
	}
}

func TestValidateInvalidTLSRef(t *testing.T) {
	doc := `
profiles:
  bad:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:MQ_PASS
    capabilities:
      - inspect
    tls:
      caRef: env:NOT_ALLOWED
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	result := cat.Validate()
	if result.IsValid("bad") {
		t.Fatal("expected invalid TLS")
	}
}

func TestValidateRejectsInlineSecrets(t *testing.T) {
	doc := `
profiles:
  bad:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      username: admin
      password: secret
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	result := cat.Validate()
	if result.IsValid("bad") {
		t.Fatal("expected inline secret rejection")
	}
}

func TestValidateFailOpenPartialValidity(t *testing.T) {
	doc := `
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
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	result := cat.Validate()
	if !result.AnyValid() || result.AllValid() {
		t.Fatalf("expected partial validity, got %+v", result.Statuses)
	}
}

func TestValidateRejectsInvalidTimeout(t *testing.T) {
	doc := `
profiles:
  bad:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    timeout: not-a-duration
    authentication:
      type: basic
      secretRef: env:MQ_PASS
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	result := cat.Validate()
	if result.IsValid("bad") {
		t.Fatal("expected invalid timeout to mark profile invalid")
	}
	if result.Statuses[0].Err == nil || !strings.Contains(result.Statuses[0].Err.Error(), "timeout") {
		t.Fatalf("expected timeout validation error, got %+v", result.Statuses)
	}
}

func TestValidateRejectsUnknownCapability(t *testing.T) {
	doc := `
profiles:
  bad:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:MQ_PASS
    capabilities:
      - read-only
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	result := cat.Validate()
	if result.IsValid("bad") {
		t.Fatal("expected unknown capability rejection")
	}
	if result.Statuses[0].Err == nil || !strings.Contains(result.Statuses[0].Err.Error(), "unknown capability") {
		t.Fatalf("expected unknown capability error, got %+v", result.Statuses)
	}
}

func TestValidateRejectsEmptyCapabilities(t *testing.T) {
	doc := `
profiles:
  bad:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:MQ_PASS
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	result := cat.Validate()
	if result.IsValid("bad") {
		t.Fatal("expected empty capabilities rejection")
	}
}

func TestValidateRejectsDuplicateCapability(t *testing.T) {
	doc := `
profiles:
  bad:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:MQ_PASS
    capabilities:
      - inspect
      - inspect
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	result := cat.Validate()
	if result.IsValid("bad") {
		t.Fatal("expected duplicate capability rejection")
	}
}

func TestValidateMTLSProfile(t *testing.T) {
	doc := `
profiles:
  mtls:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: mtls
      certificateRef: file:/run/secrets/client.pem
      privateKeyRef: file:/run/secrets/client-key.pem
    capabilities:
      - inspect
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if !cat.Validate().IsValid("mtls") {
		t.Fatal("expected valid mtls profile")
	}
}

func TestValidateK8sSecretRef(t *testing.T) {
	doc := `
profiles:
  k8s-profile:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: k8s:mq-system/mq-credentials#password
    capabilities:
      - inspect
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if !cat.Validate().IsValid("k8s-profile") {
		t.Fatal("expected valid k8s secretRef profile")
	}
}

func TestValidateRejectsVaultSecretRef(t *testing.T) {
	doc := `
profiles:
  bad:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: vault:secret/data/mq
    capabilities:
      - inspect
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	result := cat.Validate()
	if result.IsValid("bad") {
		t.Fatal("expected vault secretRef rejection")
	}
	var vaultErr error
	for _, s := range result.Statuses {
		if s.Name == "bad" {
			vaultErr = s.Err
			break
		}
	}
	if vaultErr == nil || !strings.Contains(vaultErr.Error(), "unsupported secret reference") {
		t.Fatalf("err = %v", vaultErr)
	}
}
