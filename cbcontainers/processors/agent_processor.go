package processors

import (
	"reflect"

	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

type APIGateway interface {
	RegisterCluster() error
	GetRegistrySecret() (*models.RegistrySecretValues, error)
	GetCompatibilityMatrixEntryFor(operatorVersion string) (*models.OperatorCompatibility, error)
}

type APIGatewayCreator interface {
	CreateGateway(cbContainersCluster *cbcontainersv1.CBContainersAgent, accessToken string) (APIGateway, error)
}

type AgentProcessor struct {
	gatewayCreator APIGatewayCreator

	lastRegistrySecretValues *models.RegistrySecretValues

	lastProcessedObject *cbcontainersv1.CBContainersAgent

	log logr.Logger
}

func NewAgentProcessor(log logr.Logger, clusterRegistrarCreator APIGatewayCreator) *AgentProcessor {
	return &AgentProcessor{
		gatewayCreator:      clusterRegistrarCreator,
		lastProcessedObject: nil,
		log:                 log,
	}
}

func (processor *AgentProcessor) Process(cbContainersAgent *cbcontainersv1.CBContainersAgent, accessToken string) (*models.RegistrySecretValues, error) {
	if err := processor.initializeIfNeeded(cbContainersAgent, accessToken); err != nil {
		return nil, err
	}

	return processor.lastRegistrySecretValues, nil
}

func (processor *AgentProcessor) isInitialized(cbContainersCluster *cbcontainersv1.CBContainersAgent) bool {
	return processor.lastRegistrySecretValues != nil &&
		processor.lastProcessedObject != nil &&
		reflect.DeepEqual(processor.lastProcessedObject, cbContainersCluster)
}

func (processor *AgentProcessor) initializeIfNeeded(cbContainersCluster *cbcontainersv1.CBContainersAgent, accessToken string) error {
	if processor.isInitialized(cbContainersCluster) {
		return nil
	}

	if cbContainersCluster.Spec.Gateways.GatewayTLS.InsecureSkipVerify {
		processor.log.Info("'tls insecure skip verify' set to true. In this mode, TLS is susceptible to machine-in-the-middle attacks")
		if len(cbContainersCluster.Spec.Gateways.GatewayTLS.RootCAsBundle) > 0 {
			processor.log.Info("root CAs are redundant due to 'tls insecure skip verify' set to true")
		}
	}

	processor.log.Info("Initializing AgentProcessor components")
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
