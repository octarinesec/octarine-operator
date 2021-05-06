package cluster

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"reflect"
)

type Gateway interface {
	RegisterCluster() error
	GetRegistrySecret() (*models.RegistrySecretValues, error)
	GetCertificates(name string, privateKey *rsa.PrivateKey) (*x509.CertPool, *tls.Certificate, error)
}

type GatewayCreator interface {
	CreateGateway(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) Gateway
}

type MonitorCreator interface {
	CreateMonitor(cbContainersCluster *cbcontainersv1.CBContainersCluster, gateway Gateway) (Monitor, error)
}

type Monitor interface {
	Start()
	Stop()
}

type CBContainerClusterProcessor struct {
	gatewayCreator GatewayCreator
	monitorCreator MonitorCreator

	monitor Monitor

	lastRegistrySecretValues *models.RegistrySecretValues

	lastProcessedObject *cbcontainersv1.CBContainersCluster

	log logr.Logger
}

func NewCBContainerClusterProcessor(log logr.Logger, clusterRegistrarCreator GatewayCreator, monitorCreator MonitorCreator) *CBContainerClusterProcessor {
	return &CBContainerClusterProcessor{
		gatewayCreator:      clusterRegistrarCreator,
		monitorCreator:      monitorCreator,
		monitor:             nil,
		lastProcessedObject: nil,
		log:                 log,
	}
}

func (processor *CBContainerClusterProcessor) Process(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) (*models.RegistrySecretValues, error) {
	if err := processor.initializeIfNeeded(cbContainersCluster, accessToken); err != nil {
		return nil, err
	}

	return processor.lastRegistrySecretValues, nil
}

func (processor *CBContainerClusterProcessor) isInitialized(cbContainersCluster *cbcontainersv1.CBContainersCluster) bool {
	return processor.monitor != nil &&
		processor.lastRegistrySecretValues != nil &&
		processor.lastProcessedObject != nil &&
		reflect.DeepEqual(processor.lastProcessedObject, cbContainersCluster)
}

func (processor *CBContainerClusterProcessor) initializeIfNeeded(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) error {
	if processor.isInitialized(cbContainersCluster) {
		return nil
	}

	processor.log.Info("Initializing CBContainerClusterProcessor components")
	gateway := processor.gatewayCreator.CreateGateway(cbContainersCluster, accessToken)
	monitor, err := processor.monitorCreator.CreateMonitor(cbContainersCluster, gateway)
	if err != nil {
		return err
	}

	processor.log.Info("Calling get registry secret")
	registrySecretValues, err := gateway.GetRegistrySecret()
	if err != nil {
		return err
	}

	processor.log.Info("Calling register cluster")
	if err := gateway.RegisterCluster(); err != nil {
		return err
	}

	processor.lastRegistrySecretValues = registrySecretValues
	processor.lastProcessedObject = cbContainersCluster

	if processor.monitor != nil {
		processor.log.Info("Stopping old monitor")
		processor.monitor.Stop()
	}
	processor.monitor = monitor
	processor.log.Info("Starting new monitor")
	processor.monitor.Start()

	return nil
}
