package hardening

import (
	"fmt"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	EnforcerTlsName = "cbcontainers-hardening-enforcer-tls"
)

type EnforcerTlsK8sObject struct {
	CBContainersHardeningChildK8sObject
	tlsSecretsValuesCreator tlsSecretsValuesCreator
}

func NewEnforcerTlsK8sObject(tlsSecretsValuesCreator tlsSecretsValuesCreator) *EnforcerTlsK8sObject {
	return &EnforcerTlsK8sObject{
		tlsSecretsValuesCreator: tlsSecretsValuesCreator,
	}
}

func (obj *EnforcerTlsK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: EnforcerTlsName, Namespace: obj.cbContainersHardening.Namespace}
}

func (obj *EnforcerTlsK8sObject) EmptyK8sObject() client.Object {
	return &coreV1.Secret{}
}

func (obj *EnforcerTlsK8sObject) MutateK8sObject(k8sObject client.Object) error {
	secret, ok := k8sObject.(*coreV1.Secret)
	if !ok {
		return fmt.Errorf("expected Secret K8s object")
	}

	tlsSecretValues, err := obj.tlsSecretsValuesCreator.CreateTlsSecretsValues(types.NamespacedName{Name: EnforcerName, Namespace: obj.cbContainersHardening.Namespace})
	if err != nil {
		return err
	}

	secret.Data["ca.crt"] = tlsSecretValues.CaCert
	secret.Data["ca.key"] = tlsSecretValues.CaKey
	secret.Data["signed_cert"] = tlsSecretValues.SignedCert
	secret.Data["key"] = tlsSecretValues.Key

	return nil
}
