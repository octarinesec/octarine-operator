package cluster

import (
	"context"
	"fmt"
	operatorcontainerscarbonblackiov1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CBContainersClusterStateController struct {
	desiredConfigMap *ConfigurationK8sObject
}

func NewStateController() *CBContainersClusterStateController {
	return &CBContainersClusterStateController{
		desiredConfigMap: NewConfigurationK8sObject(),
	}
}

func (c *CBContainersClusterStateController) ApplyState(ctx context.Context, namespacedName types.NamespacedName, client client.Client) (bool, error) {
	cbContainersCluster := &operatorcontainerscarbonblackiov1.CBContainersCluster{}
	if err := client.Get(ctx, namespacedName, cbContainersCluster); err != nil {
		return false, fmt.Errorf("couldn't find CBContainersCluster object: %v", err)
	}

	c.desiredConfigMap.UpdateCbContainersCluster(cbContainersCluster)
	ok, err := applyment.ApplyDesiredK8sObject(ctx, client, c.desiredConfigMap)
	if err != nil {
		return false, err
	}

	return ok, nil
}
