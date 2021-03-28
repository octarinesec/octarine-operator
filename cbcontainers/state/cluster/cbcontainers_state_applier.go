package cluster

import (
	"context"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CBContainersClusterStateApplier struct {
	desiredConfigMap *ConfigurationK8sObject
}

func NewStateApplier() *CBContainersClusterStateApplier {
	return &CBContainersClusterStateApplier{
		desiredConfigMap: NewConfigurationK8sObject(),
	}
}

func (c *CBContainersClusterStateApplier) ApplyDesiredState(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster, client client.Client, setOwner applyment.OwnerSetter) (bool, error) {
	c.desiredConfigMap.UpdateCbContainersCluster(cbContainersCluster)
	ok, err := applyment.ApplyDesiredK8sObject(ctx, client, c.desiredConfigMap, setOwner)
	if err != nil {
		return false, err
	}

	return ok, nil
}
