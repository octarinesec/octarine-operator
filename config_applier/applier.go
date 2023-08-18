package config_applier

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"time"
)

// TODO: Use interfaces for dependencies?
// TODO: Env_var to enable
// TODO: Configurable polling interval

const (
	timeoutSingleIteration = time.Second * 30
)

var versions = []string{"2.12.1", "2.10.0", "2.12.0", "2.11.0"}

var (
	tr  = true
	fal = false
)

type ConfigurationAPI interface {
	// Get Compatibility matrix
	// Update status of change (ack/error)
	// Get pending changes
	// Set status for change (acknowledge/error)

	GetConfigurationChanges(context.Context) ([]ConfigurationChange, error)
	UpdateConfigurationChangeStatus(context.Context, ConfigurationChangeStatusUpdate) error
}

type Applier struct {
	K8sClient client.Client
	Logger    logr.Logger
	Api       ConfigurationAPI
}

type pendingChangesResponse struct {
	ConfigurationChanges []ConfigurationChange `json:"configuration_changes"`
}

type ConfigurationChange struct {
	ID                    string  `json:"id"`
	Status                string  `json:"status"`
	AgentVersion          *string `json:"agent_version"`
	EnableClusterScanning *bool   `json:"enable_cluster_scanning"`
	EnableRuntime         *bool   `json:"enable_runtime"`
}

type ConfigurationChangeStatusUpdate struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Reason string `json:"reason"`
	// AppliedGeneration tracks the generation of the Custom resource where the change was applied
	AppliedGeneration int64 `json:"applied_generation"`
	// AppliedTimestamp records when the change was applied in RFC3339 format
	AppliedTimestamp string `json:"applied_timestamp"`

	// TODO: CLuster and group. Cluster identifier?
}

type changeStatus string

var (
	statusPending      changeStatus = "PENDING"
	statusAcknowledged changeStatus = "ACKNOWLEDGED" // TODO: Acknowledged or applied?
	statusFailed       changeStatus = "FAILED"
)

func (applier *Applier) RunLoop(signalsContext context.Context) {
	pollingSleepDuration := 20 * time.Second
	pollingTimer := time.NewTicker(pollingSleepDuration)
	defer pollingTimer.Stop()

	for {
		select {
		case <-signalsContext.Done():
			applier.Logger.Info("Received cancel signal, turning off configuration applier")
			return
		case <-pollingTimer.C:
			// Nothing to do; this is the polling sleep case
		}
		// TODO: Pass context down?
		applier.Logger.Info("RUNNING ITERATION")
		applier.RunIteration(signalsContext) // TODO!!!
	}
}

func (applier *Applier) RunIteration(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutSingleIteration)
	defer cancel()

	applier.Logger.Info("Checking for pending remote configuration changes...")

	change, err := applier.getPendingChange(ctx)
	if err != nil {
		applier.Logger.Error(err, "Failed to get pending configuration changes")
		return err // TODO
	}

	if change == nil {
		applier.Logger.Info("No pending remote configuration changes found")
		return nil
	}

	applier.Logger.Info("Applying remote configuration change", "change", change)
	cr, err := applier.applyChange(ctx, change)
	if err != nil {
		applier.Logger.Error(err, "Failed to apply configuration change", "changeID", change.ID)
		// Intentional fallthrough so we always update the status of the change on the backend, including failed status
	}

	if errStatusUpdate := applier.updateChangeStatus(ctx, change, cr, err); errStatusUpdate != nil {
		applier.Logger.Error(err, "Failed to update the status of a configuration change; it might be re-applied again in the future")
		return errStatusUpdate // TODO
	}

	return nil
}

func (applier *Applier) getPendingChange(ctx context.Context) (*ConfigurationChange, error) {
	changes, err := applier.Api.GetConfigurationChanges(ctx)
	if err != nil {
		return nil, err
	}

	for _, change := range changes {
		if change.Status == string(statusPending) {
			return &change, nil
		}
	}
	return nil, nil
}

func (applier *Applier) updateChangeStatus(ctx context.Context, change *ConfigurationChange, cr *cbcontainersv1.CBContainersAgent, err error) error {
	statusUpdate := ConfigurationChangeStatusUpdate{
		ID:                change.ID,
		Status:            string(statusAcknowledged),
		Reason:            "", // TODO
		AppliedGeneration: cr.Generation,
		AppliedTimestamp:  time.Now().UTC().Format(time.RFC3339),
	}
	return applier.Api.UpdateConfigurationChangeStatus(ctx, statusUpdate)
}

func (applier *Applier) applyChange(ctx context.Context, change *ConfigurationChange) (*cbcontainersv1.CBContainersAgent, error) {
	cr, err := applier.getContainerAgentCR(ctx)
	if err != nil {
		return nil, err
	}
	if cr == nil {
		applier.Logger.Info("No CBContainersAgent instance found")
		return nil, nil
	}

	// TODO: Validation!
	if change.AgentVersion != nil {
		cr.Spec.Version = *change.AgentVersion
	}
	if change.EnableClusterScanning != nil {
		cr.Spec.Components.ClusterScanning.Enabled = change.EnableClusterScanning
	}
	if change.EnableRuntime != nil {
		cr.Spec.Components.RuntimeProtection.Enabled = change.EnableRuntime
	}

	generationBefore := cr.ObjectMeta.Generation
	// TODO:  Handle Conflict response and retry
	err = applier.K8sClient.Update(ctx, cr)
	generationAfter := cr.ObjectMeta.Generation
	applier.Logger.Info("Updated object", "oldGeneration", generationBefore, "newGeneration", generationAfter, "err", err)
	return cr, nil
}

// getContainerAgentCR loads exactly 0 or 1 CBContainersAgent definitions
// if no resource is defined, nil is returned
// in case more than 1 resource is defined (which is not supported), only the first one is returned
func (applier *Applier) getContainerAgentCR(ctx context.Context) (*cbcontainersv1.CBContainersAgent, error) {
	// keep implementation in-sync with CBContainersAgentController.getContainersAgentObject() to ensure both operate on the same agent instance

	cbContainersAgentsList := &cbcontainersv1.CBContainersAgentList{}
	if err := applier.K8sClient.List(ctx, cbContainersAgentsList); err != nil {
		return nil, fmt.Errorf("couldn't list CBContainersAgent k8s objects: %v", err)
	}

	if cbContainersAgentsList.Items == nil || len(cbContainersAgentsList.Items) == 0 {
		return nil, nil
	}

	// We don't log a warning if len >=2 as the controller already warns users about that
	return &cbContainersAgentsList.Items[0], nil
}

func RandomChange() *ConfigurationChange {
	csRand, runtimeRand, versionRand := rand.Int(), rand.Int(), rand.Intn(len(versions)+1)

	csRand, runtimeRand, versionRand = 1, 2, 3
	if versionRand == len(versions) {
		return nil
	}

	changeVersion := &versions[versionRand]

	var changeClusterScanning *bool
	var changeRuntime *bool

	switch csRand % 5 {
	case 1, 3:
		changeClusterScanning = &tr
	case 2, 4:
		changeClusterScanning = &fal
	default:
		changeClusterScanning = nil
	}

	switch runtimeRand % 5 {
	case 1, 3:
		changeRuntime = &tr
	case 2, 4:
		changeRuntime = &fal
	default:
		changeRuntime = nil
	}

	return &ConfigurationChange{
		ID:                    strconv.Itoa(rand.Int()),
		AgentVersion:          changeVersion,
		EnableClusterScanning: changeClusterScanning,
		EnableRuntime:         changeRuntime,
		Status:                string(statusPending),
	}
}
