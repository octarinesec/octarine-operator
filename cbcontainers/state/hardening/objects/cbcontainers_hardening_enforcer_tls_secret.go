package objects

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
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
}

func NewEnforcerTlsK8sObject(tlsSecretsValuesCreator TlsSecretsValuesCreator) *EnforcerTlsK8sObject {
	return &EnforcerTlsK8sObject{
		tlsSecretsValuesCreator: tlsSecretsValuesCreator,
	}
}

func (obj *EnforcerTlsK8sObject) EmptyK8sObject() client.Object {
	return &coreV1.Secret{}
}

func (obj *EnforcerTlsK8sObject) HardeningChildNamespacedName(cbContainersHardening *cbcontainersv1.CBContainersHardening) types.NamespacedName {
	return types.NamespacedName{Name: EnforcerTlsName, Namespace: cbContainersHardening.Namespace}
}

func (obj *EnforcerTlsK8sObject) MutateHardeningChildK8sObject(k8sObject client.Object, cbContainersHardening *cbcontainersv1.CBContainersHardening) error {
	secret, ok := k8sObject.(*coreV1.Secret)
	if !ok {
		return fmt.Errorf("expected Secret K8s object")
	}

	tlsSecretValues, err := obj.tlsSecretsValuesCreator.CreateTlsSecretsValues(types.NamespacedName{Name: EnforcerName, Namespace: cbContainersHardening.Namespace})
	if err != nil {
		return err
	}

	secret.Data = tlsSecretValues.ToDataMap()

	return nil
}
