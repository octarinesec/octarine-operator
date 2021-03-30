package cluster

import (
	"context"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CBContainersClusterStateApplier struct {
	desiredConfigMap      *ConfigurationK8sObject
	desiredRegistrySecret *RegistrySecretK8sObject
}

func NewClusterStateApplier() *CBContainersClusterStateApplier {
	return &CBContainersClusterStateApplier{
		desiredConfigMap:      NewConfigurationK8sObject(),
		desiredRegistrySecret: NewRegistrySecretK8sObject(),
	}
}

func (c *CBContainersClusterStateApplier) ApplyDesiredState(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster, registrySecret *models.RegistrySecretValues, client client.Client, setOwner applyment.OwnerSetter) (bool, error) {
	mutated := false

	c.desiredConfigMap.UpdateCbContainersCluster(cbContainersCluster)
	mutatedConfigmap, err := applyment.ApplyDesiredK8sObject(ctx, client, c.desiredConfigMap, setOwner)
	if err != nil {
		return false, err
	}
	mutated = mutatedConfigmap || mutated

	c.desiredRegistrySecret.UpdateCbContainersCluster(cbContainersCluster)
	c.desiredRegistrySecret.UpdateRegistrySecretValues(registrySecret)
	mutatedRegistrySecret, err := applyment.ApplyDesiredK8sObject(ctx, client, c.desiredRegistrySecret, setOwner)
	if err != nil {
		return false, err
	}
	mutated = mutatedRegistrySecret || mutated

	return mutated, nil
}
