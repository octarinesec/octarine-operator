package operator

import (
	"context"
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretAccessTokenProvider struct {
	k8sClient client.Client
}

func NewSecretAccessTokenProvider(k8sClient client.Client) *SecretAccessTokenProvider {
	return &SecretAccessTokenProvider{k8sClient: k8sClient}
}

// GetCBAccessToken will attempt to read the access token value from a secret in the deployed namespace
// The secret should be defined in the provided Custom resource
func (provider *SecretAccessTokenProvider) GetCBAccessToken(
	ctx context.Context,
	cbContainersCluster *cbcontainersv1.CBContainersAgent,
	deployedNamespace string,
) (string, error) {
	accessTokenSecretNamespacedName := types.NamespacedName{
		Name:      cbContainersCluster.Spec.AccessTokenSecretName,
		Namespace: deployedNamespace,
	}
	accessTokenSecret := &corev1.Secret{}
	if err := provider.k8sClient.Get(ctx, accessTokenSecretNamespacedName, accessTokenSecret); err != nil {
		return "", fmt.Errorf("couldn't find access token secret k8s object: %v", err)
	}

	accessToken := string(accessTokenSecret.Data[commonState.AccessTokenSecretKeyName])
	if accessToken == "" {
		return "", fmt.Errorf("the k8s secret %v is missing the key %v", accessTokenSecretNamespacedName, commonState.AccessTokenSecretKeyName)
	}

	return accessToken, nil
}
