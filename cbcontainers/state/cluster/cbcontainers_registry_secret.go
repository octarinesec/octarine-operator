package cluster

import (
	"fmt"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RegistrySecretK8sObject struct {
	CBContainersClusterChildK8sObject
	registrySecretValues *models.RegistrySecretValues
}

func NewRegistrySecretK8sObject() *RegistrySecretK8sObject { return &RegistrySecretK8sObject{} }

func (obj *RegistrySecretK8sObject) UpdateRegistrySecretValues(registrySecretValues *models.RegistrySecretValues) {
	obj.registrySecretValues = registrySecretValues
}

func (obj *RegistrySecretK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: commonState.RegistrySecretName, Namespace: obj.cbContainersCluster.Namespace}
}

func (obj *RegistrySecretK8sObject) EmptyK8sObject() client.Object { return &coreV1.Secret{} }

func (obj *RegistrySecretK8sObject) MutateK8sObject(k8sObject client.Object) error {
	secret, ok := k8sObject.(*coreV1.Secret)
	if !ok {
		return fmt.Errorf("expected Secret K8s object")
	}

	if obj.registrySecretValues == nil {
		return fmt.Errorf("wasn't given with the desired registry secret values")
	}

	desiredData := make(map[string]string)
	for key, value := range obj.registrySecretValues.Data {
		desiredData[key] = string(value)
	}

	secret.Type = obj.registrySecretValues.Type
	secret.StringData = desiredData

	return nil
}
