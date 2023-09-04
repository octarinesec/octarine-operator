package remote_configuration

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"time"
)

// TODO: Respect proxy config

const (
	timeoutSingleIteration = time.Second * 60
)

type ConfigurationChangesAPI interface {
	// TODO: Get Compatibility matrix
	// TODO: Get sensor data

	GetConfigurationChanges(context.Context) ([]ConfigurationChange, error)
	UpdateConfigurationChangeStatus(context.Context, ConfigurationChangeStatusUpdate) error
}

type ChangeValidator interface {
	ValidateChange(change ConfigurationChange, cr *cbcontainersv1.CBContainersAgent) (bool, string)
}

type Configurator struct {
	k8sClient       client.Client
	logger          logr.Logger
	changesAPI      ConfigurationChangesAPI
	changeValidator ChangeValidator
}

func NewConfigurator(k8sClient client.Client, configChangesAPI ConfigurationChangesAPI, changeValidator ChangeValidator, logger logr.Logger) *Configurator {
	return &Configurator{
		k8sClient:       k8sClient,
		logger:          logger,
		changesAPI:      configChangesAPI,
		changeValidator: changeValidator,
	}
}

func (configurator *Configurator) RunIteration(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutSingleIteration)
	defer cancel()

	configurator.logger.Info("Checking for installed agent...")
	cr, errGettingCR := configurator.getContainerAgentCR(ctx)
	if errGettingCR != nil {
		configurator.logger.Error(errGettingCR, "Failed to get CBContainerAgent resource, cannot continue")
		return errGettingCR
	}
	if cr == nil {
		configurator.logger.Info("No CBContainerAgent installed, there is nothing to configure")
		return nil
	}

	configurator.logger.Info("Checking for pending remote configuration changes...")
	change, errGettingChanges := configurator.getPendingChange(ctx)
	if errGettingChanges != nil {
		configurator.logger.Error(errGettingChanges, "Failed to get pending configuration changes")
		return errGettingChanges
	}

	if change == nil {
		configurator.logger.Info("No pending remote configuration changes found")
		return nil
	}

	configurator.logger.Info("Applying remote configuration change to CBContainerAgent resource", "change", change)
	errApplyingCR := configurator.applyChange(ctx, *change, cr)
	if errApplyingCR != nil {
		configurator.logger.Error(errApplyingCR, "Failed to apply configuration change", "changeID", change.ID)
		// Intentional fallthrough as we always update the status of the change on the backend, including failed status
	}

	if errStatusUpdate := configurator.updateChangeStatus(ctx, *change, cr, errApplyingCR); errStatusUpdate != nil {
		configurator.logger.Error(errStatusUpdate, "Failed to update the status of a configuration change; it might be re-applied again in the future")
		return errStatusUpdate
	}

	// If we failed to apply the CR, we still report this to the backend but want to return the apply error here to propagate properly
	return errApplyingCR
}

func (configurator *Configurator) getPendingChange(ctx context.Context) (*ConfigurationChange, error) {
	changes, err := configurator.changesAPI.GetConfigurationChanges(ctx)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(changes, func(i, j int) bool {
		return changes[i].Timestamp < changes[j].Timestamp
	})

	for _, change := range changes {
		if change.Status == string(statusPending) {
			return &change, nil
		}
	}
	return nil, nil
}

// applyChange will sync the required changes and push them to the k8s api-server
// the input agent will be modified after this function and will no longer match the original
func (configurator *Configurator) applyChange(ctx context.Context, change ConfigurationChange, agent *cbcontainersv1.CBContainersAgent) error {
	if validChange, reason := configurator.changeValidator.ValidateChange(change, agent); !validChange {
		return fmt.Errorf("provided change with ID (%s) is not applicable due to (%s)", change.ID, reason)
	}

	ApplyChangeToCR(change, agent)

	return configurator.k8sClient.Update(ctx, agent)
}

func (configurator *Configurator) updateChangeStatus(ctx context.Context, change ConfigurationChange, cr *cbcontainersv1.CBContainersAgent, encounteredError error) error {
	var statusUpdate ConfigurationChangeStatusUpdate
	if encounteredError == nil {
		statusUpdate = ConfigurationChangeStatusUpdate{
			ID:                change.ID,
			Status:            string(statusAcknowledged),
			Reason:            "", // TODO
			AppliedGeneration: cr.Generation,
			AppliedTimestamp:  time.Now().UTC().Format(time.RFC3339),
		}
	} else {
		statusUpdate = ConfigurationChangeStatusUpdate{
			ID:     change.ID,
			Status: string(statusFailed),
			Reason: encounteredError.Error(), // TODO
		}
	}

	return configurator.changesAPI.UpdateConfigurationChangeStatus(ctx, statusUpdate)
}

// getContainerAgentCR loads exactly 0 or 1 CBContainersAgent definitions
// if no resource is defined, nil is returned
// in case more than 1 resource is defined (which is not supported), only the first one is returned
func (configurator *Configurator) getContainerAgentCR(ctx context.Context) (*cbcontainersv1.CBContainersAgent, error) {
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
