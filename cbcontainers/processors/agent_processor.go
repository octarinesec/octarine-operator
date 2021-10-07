package processors

import (
	"errors"
	"reflect"

	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/operator"
)

type APIGateway interface {
	RegisterCluster() error
	GetRegistrySecret() (*models.RegistrySecretValues, error)
	GetCompatibilityMatrixEntryFor(operatorVersion string) (*models.OperatorCompatibility, error)
}

type APIGatewayCreator interface {
	CreateGateway(cbContainersCluster *cbcontainersv1.CBContainersAgent, accessToken string) (APIGateway, error)
}

type OperatorVersionProvider interface {
	GetOperatorVersion() (string, error)
}

type AgentProcessor struct {
	gatewayCreator APIGatewayCreator
	// operatorVersionProvider provides the version of the running operator
	operatorVersionProvider OperatorVersionProvider

	lastRegistrySecretValues *models.RegistrySecretValues

	lastProcessedObject *cbcontainersv1.CBContainersAgent

	log logr.Logger
}

func NewAgentProcessor(log logr.Logger, clusterRegistrarCreator APIGatewayCreator, operatorVersionProvider OperatorVersionProvider) *AgentProcessor {
	return &AgentProcessor{
		gatewayCreator:          clusterRegistrarCreator,
		lastProcessedObject:     nil,
		operatorVersionProvider: operatorVersionProvider,
		log:                     log,
	}
}

func (processor *AgentProcessor) Process(cbContainersAgent *cbcontainersv1.CBContainersAgent, accessToken string) (*models.RegistrySecretValues, error) {
	if err := processor.initializeIfNeeded(cbContainersAgent, accessToken); err != nil {
		return nil, err
	}

	if err := processor.checkCompatibility(cbContainersAgent, accessToken); err != nil {
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

// checkCompatibility makes a backend call to check whether the given
// operatorVersion is compatible with the desired agent version.
//
// If we fail to build the API Gateway or the API call fails
// we will skip the compatibility check and not block the instalation
// (even in that case, if the operator and agent are not compatibility that will be seen later).
//
// This method will only return an error if we succesfully fetch the compatibility matrix and
// see that the operator is not compatible with the agent.
func (processor *AgentProcessor) checkCompatibility(cbContainersAgent *cbcontainersv1.CBContainersAgent, accessToken string) error {
	operatorVersion, err := processor.operatorVersionProvider.GetOperatorVersion()
	if err != nil {
		if errors.Is(err, operator.ErrNotSemVer) {
			processor.log.Error(err, "skipping compatibility check, operator version is not a semantic version")
			return nil
		}
		return err
	}
	gateway, err := processor.gatewayCreator.CreateGateway(cbContainersAgent, accessToken)
	if err != nil {
		processor.log.Error(err, "skipping compatibility check, error while building API gateway")
		// if there is an error while building the gateway log it and skip the check
		return nil
	}
	m, err := gateway.GetCompatibilityMatrixEntryFor(operatorVersion)
	if err != nil {
		// if there is an error while getting the compatibility matrix log it and skip the check
		processor.log.Error(err, "skipping compatibility check, error while getting compatibility matrix from backend")
		return nil
	}

	// if there is no error check the compatibility and return the result
	return m.CheckCompatibility(cbContainersAgent.Spec.Version)
}
