package cluster

import (
	"context"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	clusterObjects "github.com/vmware/cbcontainers-operator/cbcontainers/state/cluster/objects"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterStateApplier struct {
	desiredConfigMap      *clusterObjects.ConfigurationK8sObject
	desiredRegistrySecret *clusterObjects.RegistrySecretK8sObject
}

func NewClusterStateApplier() *ClusterStateApplier {
	return &ClusterStateApplier{
		desiredConfigMap:      clusterObjects.NewConfigurationK8sObject(),
		desiredRegistrySecret: clusterObjects.NewRegistrySecretK8sObject(),
	}
}

func (c *ClusterStateApplier) ApplyDesiredState(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster, registrySecret *models.RegistrySecretValues, client client.Client, setOwner applymentOptions.OwnerSetter) (bool, error) {
	applyOptions := applymentOptions.NewApplyOptions().SetOwnerSetter(setOwner)

	mutatedConfigmap, _, err := ApplyClusterChildK8sObject(ctx, cbContainersCluster, client, c.desiredConfigMap, applyOptions)
	if err != nil {
		return false, err
	}

	c.desiredRegistrySecret.UpdateRegistrySecretValues(registrySecret)
	mutatedRegistrySecret, _, err := ApplyClusterChildK8sObject(ctx, cbContainersCluster, client, c.desiredRegistrySecret, applyOptions)
	if err != nil {
		return false, err
	}

	return mutatedConfigmap || mutatedRegistrySecret, nil
}
