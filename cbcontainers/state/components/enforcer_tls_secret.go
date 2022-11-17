package components

import (
	"fmt"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	EnforcerTlsName = "cbcontainers-hardening-enforcer-tls"
)

type TlsSecretsValuesCreator interface {
	CreateTlsSecretsValues(resourceNamespacedName types.NamespacedName) (models.TlsSecretValues, error)
}

type EnforcerTlsK8sObject struct {
	tlsSecretsValuesCreator TlsSecretsValuesCreator

	// Namespace is the Namespace in which the Enforcer TLS secret will be created.
	Namespace string
}

func NewEnforcerTlsK8sObject(tlsSecretsValuesCreator TlsSecretsValuesCreator) *EnforcerTlsK8sObject {
	return &EnforcerTlsK8sObject{
		tlsSecretsValuesCreator: tlsSecretsValuesCreator,
		Namespace:               commonState.DataPlaneNamespaceName,
	}
}

func (obj *EnforcerTlsK8sObject) EmptyK8sObject() client.Object {
	return &coreV1.Secret{}
}

func (obj *EnforcerTlsK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: EnforcerTlsName, Namespace: obj.Namespace}
}

func (obj *EnforcerTlsK8sObject) MutateK8sObject(k8sObject client.Object, spec *cbcontainersv1.CBContainersAgentSpec) error {
	secret, ok := k8sObject.(*coreV1.Secret)
	if !ok {
		return fmt.Errorf("expected Secret K8s object")
	}

	secret.Namespace = spec.Namespace
	tlsSecretValues, err := obj.tlsSecretsValuesCreator.CreateTlsSecretsValues(types.NamespacedName{Name: EnforcerName, Namespace: obj.Namespace})
	if err != nil {
		return err
	}

	secret.Data = tlsSecretValues.ToDataMap()

	return nil
}
