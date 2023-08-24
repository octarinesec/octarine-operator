package config_applier

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"time"
)

// TODO: Env_var to enable
// TODO: Respect proxy config

const (
	timeoutSingleIteration = time.Second * 60
)

type ConfigurationChangesAPI interface {
	// Get Compatibility matrix

	GetConfigurationChanges(context.Context) ([]ConfigurationChange, error)
	UpdateConfigurationChangeStatus(context.Context, ConfigurationChangeStatusUpdate) error
}

type Applier struct {
	k8sClient  client.Client
	logger     logr.Logger
	changesAPI ConfigurationChangesAPI
}

func NewApplier(k8sClient client.Client, api ConfigurationChangesAPI, logger logr.Logger) *Applier {
	return &Applier{k8sClient: k8sClient, logger: logger, changesAPI: api}
}

func (applier *Applier) RunIteration(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutSingleIteration)
	defer cancel()

	applier.logger.Info("Checking for pending remote configuration changes...")

	change, errGettingChanges := applier.getPendingChange(ctx)
	if errGettingChanges != nil {
		applier.logger.Error(errGettingChanges, "Failed to get pending configuration changes")
		return errGettingChanges
	}

	if change == nil {
		applier.logger.Info("No pending remote configuration changes found")
		return nil
	}

	applier.logger.Info("Applying remote configuration change", "change", change)
	cr, errApplyingCR := applier.applyChange(ctx, *change)
	if errApplyingCR != nil {
		applier.logger.Error(errApplyingCR, "Failed to apply configuration change", "changeID", change.ID)
		// Intentional fallthrough as we always update the status of the change on the backend, including failed status
	}

	if errStatusUpdate := applier.updateChangeStatus(ctx, *change, cr, errApplyingCR); errStatusUpdate != nil {
		applier.logger.Error(errStatusUpdate, "Failed to update the status of a configuration change; it might be re-applied again in the future")
		return errStatusUpdate
	}

	return errApplyingCR
}

func (applier *Applier) getPendingChange(ctx context.Context) (*ConfigurationChange, error) {
	changes, err := applier.changesAPI.GetConfigurationChanges(ctx)
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

func (applier *Applier) updateChangeStatus(ctx context.Context, change ConfigurationChange, cr *cbcontainersv1.CBContainersAgent, encounteredError error) error {
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

	return applier.changesAPI.UpdateConfigurationChangeStatus(ctx, statusUpdate)
}

func (applier *Applier) applyChange(ctx context.Context, change ConfigurationChange) (*cbcontainersv1.CBContainersAgent, error) {
	cr, err := applier.getContainerAgentCR(ctx)
	if err != nil {
		return nil, err
	}
	if cr == nil {
		return nil, fmt.Errorf("no CBContainerAgent instance found, cannot apply change")
	}

	applyChangesToCR(change, cr)

	err = applier.k8sClient.Update(ctx, cr)
	return cr, err
}

// getContainerAgentCR loads exactly 0 or 1 CBContainersAgent definitions
// if no resource is defined, nil is returned
// in case more than 1 resource is defined (which is not supported), only the first one is returned
func (applier *Applier) getContainerAgentCR(ctx context.Context) (*cbcontainersv1.CBContainersAgent, error) {
	// keep implementation in-sync with CBContainersAgentController.getContainersAgentObject() to ensure both operate on the same agent instance

	cbContainersAgentsList := &cbcontainersv1.CBContainersAgentList{}
	if err := applier.k8sClient.List(ctx, cbContainersAgentsList); err != nil {
		return nil, fmt.Errorf("couldn't list CBContainersAgent k8s objects: %w", err)
	}

	if len(cbContainersAgentsList.Items) == 0 {
		return nil, nil
	}

	// We don't log a warning if len >=2 as the controller already warns users about that
	return &cbContainersAgentsList.Items[0], nil
}