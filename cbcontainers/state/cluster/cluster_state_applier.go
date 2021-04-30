package cluster

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	clusterObjects "github.com/vmware/cbcontainers-operator/cbcontainers/state/cluster/objects"
	coreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterStateApplier struct {
	desiredConfigMap      *clusterObjects.ConfigurationK8sObject
	desiredRegistrySecret *clusterObjects.RegistrySecretK8sObject
	desiredPriorityClass  *clusterObjects.PriorityClassK8sObject
	log                   logr.Logger
}

func NewClusterStateApplier(log logr.Logger) *ClusterStateApplier {
	return &ClusterStateApplier{
		desiredConfigMap:      clusterObjects.NewConfigurationK8sObject(),
		desiredRegistrySecret: clusterObjects.NewRegistrySecretK8sObject(),
		desiredPriorityClass:  clusterObjects.NewPriorityClassK8sObject(),
		log:                   log,
	}
}

func (c *ClusterStateApplier) ApplyDesiredState(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster, registrySecret *models.RegistrySecretValues, client client.Client, setOwner applymentOptions.OwnerSetter) (bool, error) {
	applyOptions := applymentOptions.NewApplyOptions().SetOwnerSetter(setOwner)

	c.log.Info("Getting Nodes list")
	nodesList := &coreV1.NodeList{}
	if err := client.List(ctx, nodesList); err != nil || nodesList.Items == nil || len(nodesList.Items) < 1 {
		return false, fmt.Errorf("couldn't get nodes list")
	}

	mutatedConfigmap, _, err := ApplyClusterChildK8sObject(ctx, cbContainersCluster, client, c.desiredConfigMap, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied config map", "Mutated", mutatedConfigmap)

	c.desiredRegistrySecret.UpdateRegistrySecretValues(registrySecret)
	mutatedRegistrySecret, _, err := ApplyClusterChildK8sObject(ctx, cbContainersCluster, client, c.desiredRegistrySecret, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied registry secret", "Mutated", mutatedRegistrySecret)

	c.desiredPriorityClass.UpdateKubeletVersion(nodesList.Items[0].Status.NodeInfo.KubeletVersion)
	mutatedPriorityClass, _, err := ApplyClusterChildK8sObject(ctx, cbContainersCluster, client, c.desiredPriorityClass, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied priority class", "Mutated", mutatedPriorityClass)

	return mutatedConfigmap || mutatedRegistrySecret || mutatedPriorityClass, nil
}
