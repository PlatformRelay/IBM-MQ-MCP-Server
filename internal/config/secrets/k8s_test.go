package secrets_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
)

type fakeK8sReader struct {
	secrets map[string]map[string][]byte
	err     error
}

func (f *fakeK8sReader) ReadSecret(_ context.Context, namespace, name, key string) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	ns, ok := f.secrets[namespace+"/"+name]
	if !ok {
		return nil, secrets.ErrK8sSecretNotFound
	}
	value, ok := ns[key]
	if !ok {
		return nil, secrets.ErrK8sSecretKeyNotFound
	}
	return value, nil
}

func TestParseK8sReference(t *testing.T) {
	ref, err := secrets.Parse("k8s:mq-system/mq-credentials#password")
	if err != nil {
		t.Fatal(err)
	}
	if ref.Provider != secrets.ProviderK8s {
		t.Fatalf("provider = %q", ref.Provider)
	}
	if ref.K8sNamespace != "mq-system" || ref.K8sSecret != "mq-credentials" || ref.K8sKey != "password" {
		t.Fatalf("k8s target = %+v", ref)
	}
	if ref.String() != "k8s:mq-system/mq-credentials#password" {
		t.Fatalf("String() = %q", ref.String())
	}
}

func TestParseK8sReferenceRejectsMalformed(t *testing.T) {
	cases := []string{
		"k8s:",
		"k8s:namespace-only",
		"k8s:namespace/secret",
		"k8s:/secret#key",
		"k8s:namespace/#key",
	}
	for _, c := range cases {
		if _, err := secrets.Parse(c); err == nil {
			t.Fatalf("Parse(%q) expected error", c)
		}
	}
}

func TestResolveK8sSecret(t *testing.T) {
	reader := &fakeK8sReader{
		secrets: map[string]map[string][]byte{
			"mq-system/mq-credentials": {"password": []byte("alice:secret")},
		},
	}
	resolver := secrets.NewResolver(secrets.WithK8sReader(reader))
	ref, err := secrets.Parse("k8s:mq-system/mq-credentials#password")
	if err != nil {
		t.Fatal(err)
	}
	value, err := resolver.Resolve(ref)
	if err != nil {
		t.Fatal(err)
	}
	if value != "alice:secret" {
		t.Fatalf("value = %q", value)
	}
}

func TestResolveK8sMissingSecretTypedError(t *testing.T) {
	reader := &fakeK8sReader{secrets: map[string]map[string][]byte{}}
	resolver := secrets.NewResolver(secrets.WithK8sReader(reader))
	ref, err := secrets.Parse("k8s:mq-system/missing#password")
	if err != nil {
		t.Fatal(err)
	}
	_, err = resolver.Resolve(ref)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, secrets.ErrK8sSecretNotFound) {
		t.Fatalf("expected ErrK8sSecretNotFound, got %v", err)
	}
	if strings.Contains(err.Error(), "alice:secret") {
		t.Fatalf("error leaked secret: %v", err)
	}
}

func TestResolveK8sUnavailableWithoutReader(t *testing.T) {
	resolver := secrets.NewResolver(secrets.WithK8sReader(nil))
	ref, err := secrets.Parse("k8s:default/mq#pass")
	if err != nil {
		t.Fatal(err)
	}
	_, err = resolver.Resolve(ref)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, secrets.ErrK8sUnavailable) {
		t.Fatalf("expected ErrK8sUnavailable, got %v", err)
	}
}

func TestResolveK8sErrorsDoNotEchoSecretValues(t *testing.T) {
	reader := &fakeK8sReader{
		secrets: map[string]map[string][]byte{
			"ns/sec": {"password": []byte("must-not-appear-in-error")},
		},
	}
	resolver := secrets.NewResolver(secrets.WithK8sReader(reader))
	ref, err := secrets.Parse("k8s:ns/sec#missing-key")
	if err != nil {
		t.Fatal(err)
	}
	_, err = resolver.Resolve(ref)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "must-not-appear-in-error") {
		t.Fatalf("error leaked secret: %v", err)
	}
}
