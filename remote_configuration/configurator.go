package remote_configuration

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"time"
)

// TODO: Check which type sshould be exposed

const (
	timeoutSingleIteration = time.Second * 60
)

type ApiGateway interface {
	GetSensorMetadata() ([]models.SensorMetadata, error)
	GetCompatibilityMatrixEntryFor(operatorVersion string) (*models.OperatorCompatibility, error)

	GetConfigurationChanges(ctx context.Context, clusterIdentifier string) ([]models.ConfigurationChange, error)
	UpdateConfigurationChangeStatus(context.Context, models.ConfigurationChangeStatusUpdate) error
}

type AccessTokenProvider interface {
	GetCBAccessToken(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersAgent, deployedNamespace string) (string, error)
}

type ApiCreator func(cbContainersCluster *cbcontainersv1.CBContainersAgent, accessToken string) (ApiGateway, error)

type Configurator struct {
	k8sClient           client.Client
	logger              logr.Logger
	accessTokenProvider AccessTokenProvider
	apiCreator          ApiCreator
	operatorVersion     string
	deployedNamespace   string
	clusterIdentifier   string
}

func NewConfigurator(
	k8sClient client.Client,
	gatewayCreator ApiCreator,
	logger logr.Logger,
	accessTokenProvider AccessTokenProvider,
	operatorVersion string,
	deployedNamespace string,
	clusterIdentifier string,
) *Configurator {
	return &Configurator{
		k8sClient:           k8sClient,
		logger:              logger,
		apiCreator:          gatewayCreator,
		accessTokenProvider: accessTokenProvider,
		operatorVersion:     operatorVersion,
		deployedNamespace:   deployedNamespace,
		clusterIdentifier:   clusterIdentifier,
	}
}

func (configurator *Configurator) RunIteration(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutSingleIteration)
	defer cancel()

	configurator.logger.Info("Checking for installed agent...")
	cr, err := configurator.getCR(ctx)
	if err != nil {
		configurator.logger.Error(err, "Failed to get CBContainerAgent resource, cannot continue")
		return err
	}
	if cr == nil {
		configurator.logger.Info("No CBContainerAgent installed, there is nothing to configure")
		return nil
	}

	if remoteConfigSettings := cr.Spec.Components.Settings.RemoteConfiguration; remoteConfigSettings != nil &&
		remoteConfigSettings.EnabledForAgent != nil &&
		*remoteConfigSettings.EnabledForAgent == false {
		configurator.logger.Info("Remote configuration feature is disabled, no changes will be made")
		return nil

	}
	apiGateway, err := configurator.createAPIGateway(ctx, cr)
	if err != nil {
		configurator.logger.Error(err, "Failed to create a valid CB API Gateway, cannot continue")
		return err
	}

	configurator.logger.Info("Checking for pending remote configuration changes...")
	change, errGettingChanges := configurator.getPendingChange(ctx, apiGateway)
	if errGettingChanges != nil {
		configurator.logger.Error(errGettingChanges, "Failed to get pending configuration changes")
		return errGettingChanges
	}

	if change == nil {
		configurator.logger.Info("No pending remote configuration changes found")
		return nil
	}

	configurator.logger.Info("Applying remote configuration change to CBContainerAgent resource", "change", change)
	errApplyingCR := configurator.applyChangeToCR(ctx, apiGateway, *change, cr)
	if errApplyingCR != nil {
		configurator.logger.Error(errApplyingCR, "Failed to apply configuration changes to CBContainerAGent resource")
		// Intentional fallthrough as we want to report the change application as failed to the backend
	} else {
		configurator.logger.Info("Successfully applied configuration changes to CBContainerAgent resource")
	}

	if err := configurator.updateChangeStatus(ctx, apiGateway, *change, cr, errApplyingCR); err != nil {
		configurator.logger.Error(err, "Failed to update the status of a configuration change; it might be re-applied again in the future")
		return err
	}

	// If we failed to apply the CR, we report this to the backend but want to return the apply error here to indicate a failure
	return errApplyingCR
}

// getCR loads exactly 0 or 1 CBContainersAgent definitions
// if no resource is defined, nil is returned
// in case more than 1 resource is defined (which is not generally supported), only the first one is returned
func (configurator *Configurator) getCR(ctx context.Context) (*cbcontainersv1.CBContainersAgent, error) {
	// keep implementation in-sync with CBContainersAgentController.getContainersAgentObject() to ensure both operate on the same agent instance

	cbContainersAgentsList := &cbcontainersv1.CBContainersAgentList{}
	if err := configurator.k8sClient.List(ctx, cbContainersAgentsList); err != nil {
		return nil, fmt.Errorf("couldn't list CBContainersAgent k8s objects: %w", err)
	}

	if len(cbContainersAgentsList.Items) == 0 {
		return nil, nil
	}

	// We don't log a warning if len >=2 as the controller already warns users about that
	return &cbContainersAgentsList.Items[0], nil
}

func (configurator *Configurator) getPendingChange(ctx context.Context, apiGateway ApiGateway) (*models.ConfigurationChange, error) {
	changes, err := apiGateway.GetConfigurationChanges(ctx, configurator.clusterIdentifier)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(changes, func(i, j int) bool {
		return changes[i].Timestamp < changes[j].Timestamp
	})

	for _, change := range changes {
		if change.Status == models.ChangeStatusPending {
			return &change, nil
		}
	}
	return nil, nil
}

func (configurator *Configurator) applyChangeToCR(ctx context.Context, apiGateway ApiGateway, change models.ConfigurationChange, cr *cbcontainersv1.CBContainersAgent) error {
	validator, err := NewConfigurationChangeValidator(configurator.operatorVersion, apiGateway)
	if err != nil {
		return fmt.Errorf("failed to create configuration change validator")
	}
	if err := validator.ValidateChange(change, cr); err != nil {
		return err
	}
	c := ChangeApplier{}
	c.ApplyConfigChangeToCR(change, cr)
	return configurator.k8sClient.Update(ctx, cr)
}

func (configurator *Configurator) updateChangeStatus(
	ctx context.Context,
	apiGateway ApiGateway,
	change models.ConfigurationChange,
	cr *cbcontainersv1.CBContainersAgent,
	encounteredError error,
) error {
	statusUpdate := models.ConfigurationChangeStatusUpdate{
		ID:                change.ID,
		ClusterIdentifier: configurator.clusterIdentifier,
	}

	if encounteredError == nil {
		statusUpdate.Status = models.ChangeStatusAcked
		statusUpdate.AppliedGeneration = cr.Generation
		statusUpdate.AppliedTimestamp = time.Now().UTC().Format(time.RFC3339)
	} else {
		statusUpdate.Status = models.ChangeStatusFailed
		statusUpdate.Error = encounteredError.Error()
		// Validation change is the only thing we can safely give information to the user about
		if errors.As(encounteredError, &invalidChangeError{}) {
			statusUpdate.ErrorReason = encounteredError.Error()
		}
	}

	return apiGateway.UpdateConfigurationChangeStatus(ctx, statusUpdate)
}

func (configurator *Configurator) createAPIGateway(ctx context.Context, cr *cbcontainersv1.CBContainersAgent) (ApiGateway, error) {
	accessToken, err := configurator.accessTokenProvider.GetCBAccessToken(ctx, cr, configurator.deployedNamespace)
	if err != nil {
		return nil, err
	}
	return configurator.apiCreator(cr, accessToken)
}
