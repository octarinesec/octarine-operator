package cluster

import (
	"fmt"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RegistrySecretName = "cbcontainers-registry-secret"
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
	return types.NamespacedName{Name: RegistrySecretName, Namespace: obj.cbContainersCluster.Namespace}
}

func (obj *RegistrySecretK8sObject) EmptyK8sObject() client.Object { return &coreV1.Secret{} }

func (obj *RegistrySecretK8sObject) MutateK8sObject(k8sObject client.Object) (bool, error) {
	secret, ok := k8sObject.(*coreV1.Secret)
	if !ok {
		return false, fmt.Errorf("expected Secret K8s object")
	}

	if obj.registrySecretValues == nil {
		return false, fmt.Errorf("wasn't given with the desired registry secret values")
	}

	mutated := false
	actualSecretType := string(secret.Type)
	mutated = applyment.MutateString(string(obj.registrySecretValues.Type), func() *string { return &actualSecretType }, func(value string) { secret.Type = coreV1.SecretType(value) }) || mutated

	desiredData := make(map[string]string)
	for key, value := range obj.registrySecretValues.Data {
		desiredData[key] = string(value)
	}

	if !reflect.DeepEqual(secret.StringData, desiredData) {
		secret.StringData = desiredData
		mutated = true
	}

	return mutated, nil
}
