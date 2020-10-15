package octarine

import (
	"errors"
	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"os"
)

func registrySecretName() (types.NamespacedName, error) {
	imagePullSecretName, defined := os.LookupEnv("IMAGE_PULL_SECRET_NAME")
	if !defined {
		return types.NamespacedName{}, errors.New("IMAGE_PULL_SECRET_NAME env must be set")
	}

	ns, err := k8sutil.GetWatchNamespace()
	if err != nil {
		return types.NamespacedName{}, err
	}

	return types.NamespacedName{
		Name:      imagePullSecretName,
		Namespace: ns,
	}, nil
}

func guardrailsSecretName(octarine *unstructured.Unstructured) (types.NamespacedName, error) {
	ns, err := k8sutil.GetWatchNamespace()
	if err != nil {
		return types.NamespacedName{}, err
	}

	return types.NamespacedName{
		Name:      fmt.Sprintf("%s-guardrails-enforcer-tls", octarine.GetName()),
		Namespace: ns,
	}, nil
}

func guardrailsServiceName(octarine *unstructured.Unstructured) (types.NamespacedName, error) {
	ns, err := k8sutil.GetWatchNamespace()
	if err != nil {
		return types.NamespacedName{}, err
	}

	return types.NamespacedName{
		Name:      fmt.Sprintf("%s-guardrails-enforcer", octarine.GetName()),
		Namespace: ns,
	}, nil
}

func guardrailsWebhookName(octarine *unstructured.Unstructured) string {
	return fmt.Sprintf("%s-guardrails", octarine.GetName())
}
