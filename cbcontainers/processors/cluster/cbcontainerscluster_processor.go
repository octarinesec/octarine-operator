package cluster

import cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"

type ClusterRegistrar interface {
	RegisterCluster() error
}

type ClusterRegistrarCreator interface {
	CreateClusterRegistrar(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) ClusterRegistrar
}

type CBContainerClusterProcessor struct {
	clusterRegistrarCreator ClusterRegistrarCreator
}

func NewCBContainerClusterProcessor(clusterRegistrarCreator ClusterRegistrarCreator) *CBContainerClusterProcessor {
	return &CBContainerClusterProcessor{
		clusterRegistrarCreator: clusterRegistrarCreator,
	}
}

func (processor *CBContainerClusterProcessor) Process(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) error {
	clusterRegistrar := processor.clusterRegistrarCreator.CreateClusterRegistrar(cbContainersCluster, accessToken)
	return clusterRegistrar.RegisterCluster()
}
