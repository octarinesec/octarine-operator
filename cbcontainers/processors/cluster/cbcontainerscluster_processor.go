package cluster

import (
	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"reflect"
)

type Gateway interface {
	RegisterCluster() error
	GetRegistrySecret() (*models.RegistrySecretValues, error)
}

type GatewayCreator interface {
	CreateGateway(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) (Gateway, error)
}

type CBContainerClusterProcessor struct {
	gatewayCreator GatewayCreator

	lastRegistrySecretValues *models.RegistrySecretValues

	lastProcessedObject *cbcontainersv1.CBContainersCluster

	log logr.Logger
}

func NewCBContainerClusterProcessor(log logr.Logger, clusterRegistrarCreator GatewayCreator) *CBContainerClusterProcessor {
	return &CBContainerClusterProcessor{
		gatewayCreator:      clusterRegistrarCreator,
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
	return processor.lastRegistrySecretValues != nil &&
		processor.lastProcessedObject != nil &&
		reflect.DeepEqual(processor.lastProcessedObject, cbContainersCluster)
}

func (processor *CBContainerClusterProcessor) initializeIfNeeded(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) error {
	if processor.isInitialized(cbContainersCluster) {
		return nil
	}

	processor.log.Info("Initializing CBContainerClusterProcessor components")
	gateway, err := processor.gatewayCreator.CreateGateway(cbContainersCluster, accessToken)
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
	return nil
}
