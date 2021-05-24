package cluster

import (
	"context"
	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	clusterObjects "github.com/vmware/cbcontainers-operator/cbcontainers/state/cluster/objects"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterStateApplier struct {
	desiredConfigMap      *clusterObjects.ConfigurationK8sObject
	desiredRegistrySecret *clusterObjects.RegistrySecretK8sObject
	desiredPriorityClass  *clusterObjects.PriorityClassK8sObject
	childApplier          ClusterChildK8sObjectApplier
	log                   logr.Logger
}

type ClusterChildK8sObjectApplier interface {
	ApplyClusterChildK8sObject(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster, client client.Client, clusterChildK8sObject ClusterChildK8sObject, applyOptionsList ...*applymentOptions.ApplyOptions) (bool, client.Object, error)
}

func NewClusterStateApplier(log logr.Logger, k8sVersion string, childApplier ClusterChildK8sObjectApplier) *ClusterStateApplier {
	return &ClusterStateApplier{
		desiredConfigMap:      clusterObjects.NewConfigurationK8sObject(),
		desiredRegistrySecret: clusterObjects.NewRegistrySecretK8sObject(),
		desiredPriorityClass:  clusterObjects.NewPriorityClassK8sObject(k8sVersion),
		childApplier:          childApplier,
		log:                   log,
	}
}

func (c *ClusterStateApplier) GetPriorityClassEmptyK8sObject() client.Object {
	return c.desiredPriorityClass.EmptyK8sObject()
}

func (c *ClusterStateApplier) ApplyDesiredState(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster, registrySecret *models.RegistrySecretValues, client client.Client, setOwner applymentOptions.OwnerSetter) (bool, error) {
	applyOptions := applymentOptions.NewApplyOptions().SetOwnerSetter(setOwner)

	mutatedConfigmap, _, err := c.childApplier.ApplyClusterChildK8sObject(ctx, cbContainersCluster, client, c.desiredConfigMap, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied config map", "Mutated", mutatedConfigmap)

	c.desiredRegistrySecret.UpdateRegistrySecretValues(registrySecret)
	mutatedRegistrySecret, _, err := c.childApplier.ApplyClusterChildK8sObject(ctx, cbContainersCluster, client, c.desiredRegistrySecret, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied registry secret", "Mutated", mutatedRegistrySecret)

	mutatedPriorityClass, _, err := c.childApplier.ApplyClusterChildK8sObject(ctx, cbContainersCluster, client, c.desiredPriorityClass, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied priority class", "Mutated", mutatedPriorityClass)

	return mutatedConfigmap || mutatedRegistrySecret || mutatedPriorityClass, nil
}
