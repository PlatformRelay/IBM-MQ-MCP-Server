package secrets

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type clientGoK8sReader struct {
	client kubernetes.Interface
}

func (c *clientGoK8sReader) ReadSecret(
	ctx context.Context,
	namespace, name, key string,
) ([]byte, error) {
	secret, err := c.client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrK8sSecretNotFound
		}
		return nil, fmt.Errorf("get secret: %w", err)
	}
	value, ok := secret.Data[key]
	if !ok {
		if _, stringOK := secret.StringData[key]; stringOK {
			return nil, ErrK8sSecretKeyNotFound
		}
		return nil, ErrK8sSecretKeyNotFound
	}
	return value, nil
}

func newDefaultK8sReader() (K8sSecretReader, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			loadingRules,
			&clientcmd.ConfigOverrides{},
		).ClientConfig()
		if err != nil {
			return nil, err
		}
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &clientGoK8sReader{client: client}, nil
}

// compile-time check
var _ K8sSecretReader = (*clientGoK8sReader)(nil)
