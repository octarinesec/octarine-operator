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

type RegistrySecretK8sObject struct {
	registrySecretValues *models.RegistrySecretValues

	// Namespace is the Namespace in which the Secret will be created.
	Namespace string
}

func NewRegistrySecretK8sObject() *RegistrySecretK8sObject {
	return &RegistrySecretK8sObject{
		Namespace: commonState.DataPlaneNamespaceName,
	}
}

func (obj *RegistrySecretK8sObject) UpdateRegistrySecretValues(registrySecretValues *models.RegistrySecretValues) {
	obj.registrySecretValues = registrySecretValues
}

func (obj *RegistrySecretK8sObject) EmptyK8sObject() client.Object { return &coreV1.Secret{} }

func (obj *RegistrySecretK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: commonState.RegistrySecretName, Namespace: obj.Namespace}
}

func (obj *RegistrySecretK8sObject) MutateK8sObject(k8sObject client.Object, spec *cbcontainersv1.CBContainersAgentSpec) error {
	secret, ok := k8sObject.(*coreV1.Secret)
	if !ok {
		return fmt.Errorf("expected Secret K8s object")
	}

	if obj.registrySecretValues == nil {
		return fmt.Errorf("wasn't given with the desired registry secret values")
	}

	secret.Type = obj.registrySecretValues.Type
	secret.Data = obj.registrySecretValues.Data
	secret.Namespace = spec.Namespace

	return nil
}
