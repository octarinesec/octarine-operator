package cluster

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

type ClusterRegistrar interface {
	RegisterCluster() error
	GetRegistrySecret() (*models.RegistrySecretValues, error)
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

func (processor *CBContainerClusterProcessor) Process(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) (*models.RegistrySecretValues, error) {
	clusterRegistrar := processor.clusterRegistrarCreator.CreateClusterRegistrar(cbContainersCluster, accessToken)

	registrySecret, err := clusterRegistrar.GetRegistrySecret()
	if err != nil {
		return nil, err
	}

	if err := clusterRegistrar.RegisterCluster(); err != nil {
		return nil, err
	}

	return registrySecret, nil
}
