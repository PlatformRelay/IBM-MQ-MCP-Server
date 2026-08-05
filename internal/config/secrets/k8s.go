package secrets

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors for Kubernetes secret resolution (safe to compare with errors.Is).
var (
	ErrK8sUnavailable       = errors.New("kubernetes secret provider unavailable")
	ErrK8sSecretNotFound    = errors.New("kubernetes secret not found")
	ErrK8sSecretKeyNotFound = errors.New("kubernetes secret key not found")
)

// K8sSecretReader fetches a key from a Kubernetes Secret.
type K8sSecretReader interface {
	ReadSecret(ctx context.Context, namespace, name, key string) ([]byte, error)
}

func parseK8sReference(ref string) (Reference, error) {
	rest := strings.TrimSpace(strings.TrimPrefix(ref, "k8s:"))
	if rest == "" {
		return Reference{}, errors.New("k8s reference requires namespace/secret#key")
	}
	hash := strings.LastIndex(rest, "#")
	if hash <= 0 || hash == len(rest)-1 {
		return Reference{}, errors.New("k8s reference requires namespace/secret#key")
	}
	key := strings.TrimSpace(rest[hash+1:])
	if key == "" {
		return Reference{}, errors.New("k8s reference requires a data key")
	}
	nsName := strings.TrimSpace(rest[:hash])
	slash := strings.Index(nsName, "/")
	if slash <= 0 || slash == len(nsName)-1 {
		return Reference{}, errors.New("k8s reference requires namespace/secret#key")
	}
	namespace := strings.TrimSpace(nsName[:slash])
	secret := strings.TrimSpace(nsName[slash+1:])
	if namespace == "" || secret == "" {
		return Reference{}, errors.New("k8s reference requires namespace/secret#key")
	}
	return Reference{
		Provider:     ProviderK8s,
		K8sNamespace: namespace,
		K8sSecret:    secret,
		K8sKey:       key,
	}, nil
}

func (r *Resolver) resolveK8s(ref Reference) (string, error) {
	reader, err := r.k8sReader()
	if err != nil {
		return "", err
	}
	data, err := reader.ReadSecret(context.Background(), ref.K8sNamespace, ref.K8sSecret, ref.K8sKey)
	if err != nil {
		return "", mapK8sError(ref, err)
	}
	value := strings.TrimRight(string(data), "\r\n")
	if value == "" {
		return "", fmt.Errorf("kubernetes secret %q key %q is empty", ref.String(), ref.K8sKey)
	}
	return value, nil
}

func mapK8sError(ref Reference, err error) error {
	switch {
	case errors.Is(err, ErrK8sSecretNotFound):
		return fmt.Errorf("kubernetes secret %q not found: %w", ref.String(), err)
	case errors.Is(err, ErrK8sSecretKeyNotFound):
		return fmt.Errorf("kubernetes secret %q key %q not found: %w", ref.String(), ref.K8sKey, err)
	default:
		return fmt.Errorf("read kubernetes secret %q: %w", ref.String(), err)
	}
}

func (r *Resolver) k8sReader() (K8sSecretReader, error) {
	if r.k8sConfigured {
		if r.k8s == nil {
			return nil, ErrK8sUnavailable
		}
		return r.k8s, nil
	}
	r.lazyOnce.Do(func() {
		r.lazyReader, r.lazyErr = newDefaultK8sReader()
	})
	if r.lazyErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrK8sUnavailable, r.lazyErr)
	}
	return r.lazyReader, nil
}
