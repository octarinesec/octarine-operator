package cluster

import (
	"context"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	stateTypes "github.com/vmware/cbcontainers-operator/cbcontainers/state/types"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clusterChildK8sObject interface {
	MutateClusterChildK8sObject(k8sObject client.Object, cbContainersCluster *cbcontainersv1.CBContainersCluster) error
	ClusterChildNamespacedName(cbContainersCluster *cbcontainersv1.CBContainersCluster) types.NamespacedName
	stateTypes.DesiredK8sObjectInitializer
}

func ApplyClusterChildK8sObject(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster, client client.Client, clusterChildK8sObject clusterChildK8sObject, applyOptionsList ...*applymentOptions.ApplyOptions) (bool, client.Object, error) {
	clusterChildWrapper := NewCBContainersClusterChildK8sObject(cbContainersCluster, clusterChildK8sObject)
	return applyment.ApplyDesiredK8sObject(ctx, client, clusterChildWrapper, applyOptionsList...)
}

type CBContainersClusterChildK8sObject struct {
	cbContainersCluster *cbcontainersv1.CBContainersCluster
	clusterChildK8sObject
}

func NewCBContainersClusterChildK8sObject(cbContainersCluster *cbcontainersv1.CBContainersCluster, clusterChildK8sObject clusterChildK8sObject) *CBContainersClusterChildK8sObject {
	return &CBContainersClusterChildK8sObject{
		cbContainersCluster:   cbContainersCluster,
		clusterChildK8sObject: clusterChildK8sObject,
	}
}

func (hardeningChildWrapper *CBContainersClusterChildK8sObject) NamespacedName() types.NamespacedName {
	return hardeningChildWrapper.ClusterChildNamespacedName(hardeningChildWrapper.cbContainersCluster)
}

func (hardeningChildWrapper *CBContainersClusterChildK8sObject) MutateK8sObject(k8sObject client.Object) error {
	return hardeningChildWrapper.MutateClusterChildK8sObject(k8sObject, hardeningChildWrapper.cbContainersCluster)
}
