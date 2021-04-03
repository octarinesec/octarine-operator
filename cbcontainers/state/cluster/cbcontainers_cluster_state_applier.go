package cluster

import (
	"context"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
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

func (c *CBContainersClusterStateApplier) ApplyDesiredState(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster, registrySecret *models.RegistrySecretValues, client client.Client, setOwner applymentOptions.OwnerSetter) (bool, error) {
	applyOptions := applymentOptions.NewApplyOptions().SetOwnerSetter(setOwner)

	c.desiredConfigMap.UpdateCbContainersCluster(cbContainersCluster)
	mutatedConfigmap, err := applyment.ApplyDesiredK8sObject(ctx, client, c.desiredConfigMap, applyOptions)
	if err != nil {
		return false, err
	}

	c.desiredRegistrySecret.UpdateCbContainersCluster(cbContainersCluster)
	c.desiredRegistrySecret.UpdateRegistrySecretValues(registrySecret)
	mutatedRegistrySecret, err := applyment.ApplyDesiredK8sObject(ctx, client, c.desiredRegistrySecret, applyOptions)
	if err != nil {
		return false, err
	}

	return mutatedConfigmap || mutatedRegistrySecret, nil
}
