package cluster

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

type ClusterRegistrar interface {
	RegisterCluster() error
	GetRegistrySecret() (*models.RegistrySecretValues, error)
	GetCertificates(name string) (*x509.CertPool, *tls.Certificate, error)
}

type ClusterRegistrarCreator interface {
	CreateClusterRegistrar(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) ClusterRegistrar
}

type monitor interface {
	Start()
	Stop()
}

type monitorCreator func(host string, port int, certPool *x509.CertPool, cert *tls.Certificate) monitor

type CBContainerClusterProcessor struct {
	clusterRegistrarCreator ClusterRegistrarCreator
	createMonitor           monitorCreator
	monitor                 monitor
}

func NewCBContainerClusterProcessor(clusterRegistrarCreator ClusterRegistrarCreator, createMonitor monitorCreator) *CBContainerClusterProcessor {
	return &CBContainerClusterProcessor{
		clusterRegistrarCreator: clusterRegistrarCreator,
		createMonitor:           createMonitor,
		monitor:                 nil,
	}
}

func (processor *CBContainerClusterProcessor) GetRegistrySecretValues(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) (*models.RegistrySecretValues, error) {
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

func (processor *CBContainerClusterProcessor) UpdateMonitor(ctx context.Context, cluster *cbcontainersv1.CBContainersCluster) {
	if processor.monitor != nil {
		processor.monitor.Stop()
	}

	processor.monitor = processor.createMonitor(cluster.Spec.EventsGatewaySpec.Host, cluster.Spec.EventsGatewaySpec.Port)
	processor.monitor.Start()
}
