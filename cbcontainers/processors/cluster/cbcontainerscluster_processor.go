package cluster

import (
	"crypto/tls"
	"crypto/x509"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"reflect"
)

type Gateway interface {
	RegisterCluster() error
	GetRegistrySecret() (*models.RegistrySecretValues, error)
	GetCertificates(name string) (*x509.CertPool, *tls.Certificate, error)
}

type gatewayCreator interface {
	CreateGateway(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) Gateway
}

type monitorCreator interface {
	CreateMonitor(cbContainersCluster *cbcontainersv1.CBContainersCluster, gateway Gateway) (Monitor, error)
}

type Monitor interface {
	Start()
	Stop()
}

type CBContainerClusterProcessor struct {
	gatewayCreator gatewayCreator
	monitorCreator monitorCreator

	gateway Gateway
	monitor Monitor

	lastRegistrySecretValues *models.RegistrySecretValues

	lastProcessedObject *cbcontainersv1.CBContainersCluster
}

func NewCBContainerClusterProcessor(clusterRegistrarCreator gatewayCreator, monitorCreator monitorCreator) *CBContainerClusterProcessor {
	return &CBContainerClusterProcessor{
		gatewayCreator:      clusterRegistrarCreator,
		monitorCreator:      monitorCreator,
		gateway:             nil,
		monitor:             nil,
		lastProcessedObject: nil,
	}
}

func (processor *CBContainerClusterProcessor) Process(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) (*models.RegistrySecretValues, error) {
	if err := processor.initializeIfNeeded(cbContainersCluster, accessToken); err != nil {
		return nil, err
	}

	return processor.lastRegistrySecretValues, nil
}

func (processor *CBContainerClusterProcessor) isInitialized(cbContainersCluster *cbcontainersv1.CBContainersCluster) bool {
	return processor.gateway != nil &&
		processor.monitor != nil &&
		processor.lastRegistrySecretValues != nil &&
		processor.lastProcessedObject != nil &&
		reflect.DeepEqual(processor.lastProcessedObject, cbContainersCluster)
}

func (processor *CBContainerClusterProcessor) initializeIfNeeded(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) error {
	if processor.isInitialized(cbContainersCluster) {
		return nil
	}

	gateway := processor.gatewayCreator.CreateGateway(cbContainersCluster, accessToken)
	monitor, err := processor.monitorCreator.CreateMonitor(cbContainersCluster, gateway)
	if err != nil {
		return err
	}

	registrySecretValues, err := processor.gateway.GetRegistrySecret()
	if err != nil {
		return err
	}

	if err := processor.gateway.RegisterCluster(); err != nil {
		return err
	}

	processor.gateway = gateway
	processor.lastRegistrySecretValues = registrySecretValues
	processor.lastProcessedObject = cbContainersCluster

	if processor.monitor != nil {
		processor.monitor.Stop()
	}
	processor.monitor = monitor
	processor.monitor.Start()

	return nil
}
